package builder

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fanatic/waypoint-plugin-heroku/heroku"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	herokuSDK "github.com/heroku/heroku-go/v5"
	"github.com/paketo-buildpacks/procfile/procfile"
)

type BuildConfig struct {
	From     string `hcl:"from"`
	Source   string `hcl:"source,optional"`
	Pipeline string `hcl:"pipeline"`
	App      string `hcl:"app,optional"`
}

type Builder struct {
	config BuildConfig
}

// Implement Configurable
func (b *Builder) Config() (interface{}, error) {
	return &b.config, nil
}

// Implement Builder
func (b *Builder) BuildFunc() interface{} {
	// return a function which will be called by Waypoint
	return b.build
}

func (b *Builder) build(ctx context.Context, ui terminal.UI, job *component.JobInfo, src *component.Source, log hclog.Logger) (*Slug, error) {
	h, err := heroku.New()
	if err != nil {
		return nil, err
	}

	log.Info(
		"Start build",
		"src", src,
		"config", b.config,
	)

	if b.config.From == "source" {
		sg := ui.StepGroup()

		if b.config.Source == "" {
			b.config.Source = src.Path
		}

		step := sg.Add("Archiving source...")
		tf, err := b.createLocalArchive(log, b.config.Source)
		if err != nil {
			step.Abort()
			return nil, err
		}
		defer os.Remove(tf.Name())
		defer tf.Close()
		step.Done()

		step = sg.Add("Sending source to Heroku...")
		sourceURL, err := b.createHerokuSource(ctx, h, log, tf)
		if err != nil {
			step.Abort()
			return nil, err
		}
		step.Done()

		step = sg.Add("Building image...")
		slugID, err := b.createHerokuBuild(ctx, h, sourceURL, job.Id, step.TermOutput())
		if err != nil {
			step.Abort()
			return nil, err
		}
		step.Done()

		return &Slug{
			Id: slugID,
		}, nil
	} else if b.config.From == "archive" {
		sg := ui.StepGroup()

		if b.config.Source == "" {
			b.config.Source = src.Path
		}

		step := sg.Add("Archiving slug...")
		tf, err := b.createLocalArchive(log, b.config.Source)
		if err != nil {
			return nil, err
		}
		defer os.Remove(tf.Name())
		defer tf.Close()
		step.Done()

		step = sg.Add("Sending slug to Heroku...")
		slugID, err := b.createHerokuSlug(ctx, h, log, tf)
		if err != nil {
			return nil, err
		}
		step.Done()

		return &Slug{
			Id: slugID,
		}, nil
	}

	return nil, fmt.Errorf("Must supply valid 'from' parameter: source")
}

func (b *Builder) createLocalArchive(log hclog.Logger, source string) (*os.File, error) {
	log.Info("Tar started", "source", source)
	tf, err := ioutil.TempFile("", "source-tar.")
	if err != nil {
		return nil, err
	}

	if err := Tar(source, tf); err != nil {
		return nil, err
	}
	log.Info("Tar finished", "source", source)
	return tf, nil
}

func (b *Builder) createHerokuSource(ctx context.Context, h *herokuSDK.Service, log hclog.Logger, tf *os.File) (string, error) {
	source, err := h.SourceCreate(ctx)
	if err != nil {
		return "", err
	}
	log.Info(
		"Source created",
		"source", source,
	)

	if _, err := tf.Seek(0, 0); err != nil {
		return "", err
	}
	req, err := http.NewRequest("PUT", source.SourceBlob.PutURL, tf)
	if err != nil {
		return "", err
	}
	stat, err := tf.Stat()
	if err != nil {
		return "", err
	}
	req.ContentLength = stat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Info("Source upload complete", "code", resp.StatusCode, "body", string(body))

	return source.SourceBlob.GetURL, nil
}

func (b *Builder) createHerokuBuild(ctx context.Context, h *herokuSDK.Service, sourceURL, sourceVersion string, w io.Writer) (string, error) {
	buildOpts := herokuSDK.BuildCreateOpts{}
	buildOpts.SourceBlob.URL = &sourceURL
	buildOpts.SourceBlob.Version = &sourceVersion
	build, err := h.BuildCreate(ctx, b.config.App, buildOpts)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(build.OutputStreamURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	io.Copy(w, resp.Body)

	return build.Slug.ID, nil
}

func (b *Builder) createHerokuSlug(ctx context.Context, h *herokuSDK.Service, log hclog.Logger, tf *os.File) (string, error) {
	processTypesRaw, err := procfile.NewProcfileFromPath(b.config.Source)
	if err != nil {
		return "", err
	}

	processTypes := map[string]string{}
	for k, v := range processTypesRaw {
		processTypes[k], _ = v.(string)
	}

	o := herokuSDK.SlugCreateOpts{
		BuildpackProvidedDescription: String("waypoint-plugin-heroku"),
		ProcessTypes:                 processTypes,
	}
	slug, err := h.SlugCreate(ctx, b.config.App, o)
	if err != nil {
		return "", err
	}
	log.Info(
		"Slug created",
		"source", slug,
	)

	if _, err := tf.Seek(0, 0); err != nil {
		return "", err
	}
	req, err := http.NewRequest("PUT", slug.Blob.URL, tf)
	if err != nil {
		return "", err
	}
	stat, err := tf.Stat()
	if err != nil {
		return "", err
	}
	req.ContentLength = stat.Size()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	log.Info("Slug upload complete", "code", resp.StatusCode, "body", string(body))

	return slug.ID, nil
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer; the purpose for accepting multiple writers is to allow
// for multiple outputs (for example a file, or md5 hash)
func Tar(src string, writers ...io.Writer) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("Unable to tar files - %v", err.Error())
	}

	mw := io.MultiWriter(writers...)

	gzw := gzip.NewWriter(mw)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// walk path
	return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})
}

func String(s string) *string {
	return &s
}
