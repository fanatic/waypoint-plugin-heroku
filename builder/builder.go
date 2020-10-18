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

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	heroku "github.com/heroku/heroku-go/v5"
)

type BuildConfig struct {
	From     string `hcl:"from"`
	Source   string `hcl:"source,optional"`
	Pipeline string `hcl:"pipeline"`
	App      string `hcl:"app,optional"`
}

type Builder struct {
	H      *heroku.Service
	config BuildConfig
}

// Implement Configurable
func (b *Builder) Config() interface{} {
	return &b.config
}

// Implement Builder
func (b *Builder) BuildFunc() interface{} {
	// return a function which will be called by Waypoint
	return b.build
}

func (b *Builder) build(ctx context.Context, ui terminal.UI, src *component.Source, log hclog.Logger) (*Binary, error) {
	log.Info(
		"Start build",
		"src", src,
		"config", b.config,
	)
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return nil, err
	}

	if b.config.From == "source" {
		sg := ui.StepGroup()
		step := sg.Add("Sending source to Heroku to build...")
		defer step.Abort()

		if b.config.Source == "" {
			b.config.Source = src.Path
		}

		source, err := b.H.SourceCreate(ctx)
		if err != nil {
			return nil, err
		}

		pr, pw := io.Pipe()
		req, err := http.NewRequest("PUT", source.SourceBlob.PutURL, pr)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "")

		if err := Tar(b.config.Source, gzip.NewWriter(pw)); err != nil {
			return nil, err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		fmt.Printf("%d: %s\n", resp.StatusCode, string(body))

		buildOpts := heroku.BuildCreateOpts{}
		buildOpts.SourceBlob.URL = &source.SourceBlob.GetURL
		buildOpts.SourceBlob.Version = String("sha1")
		build, err := b.H.BuildCreate(ctx, "appIdentity", buildOpts)
		if err != nil {
			return nil, err
		}
		step.Done()

		step = sg.Add("Building image...")
		resp, err = http.Get(build.OutputStreamURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		io.Copy(stdout, resp.Body)
		step.Done()
	} else {
		return nil, fmt.Errorf("Must supply valid 'from' parameter: source")
	}

	return &Binary{
		Location: "abc",
	}, nil
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
