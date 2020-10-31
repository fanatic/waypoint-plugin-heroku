package main

import (
	"context"
	"fmt"

	"github.com/fanatic/waypoint-plugin-heroku/heroku"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	herokuSDK "github.com/heroku/heroku-go/v5"
)

type DeployConfig struct {
	Pipeline string `hcl:"pipeline,optional"`
	App      string `hcl:"app,optional"`
}

func (d *Deployment) URL() string { return d.Url }

type Platform struct {
	config DeployConfig
}

// Implement Configurable
func (p *Platform) Config() (interface{}, error) {
	return &p.config, nil
}

// Implement Builder
func (p *Platform) DeployFunc() interface{} {
	// return a function which will be called by Waypoint
	return p.deploy
}

// DefaultReleaserFunc implements component.PlatformReleaser
func (p *Platform) DefaultReleaserFunc() interface{} {
	return func() *Releaser { return &Releaser{} }
}

func (p *Platform) deploy(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	job *component.JobInfo,
	log hclog.Logger,
	artifact *Artifact,
	//slug *builder.Slug,
) (*Deployment, error) {
	log.Info(
		"Start deploy",
		"src", src,
		"config", p.config,
		"artifact", artifact,
	)

	h, err := heroku.New()
	if err != nil {
		return nil, err
	}

	// TODO: support dynamically creating an app in the pipeline per deploy
	// if p.config.App == "" {
	// 	 p.createHerokuApp()
	// }

	if artifact.ContainerImageDigest != "" {
		if err := p.releaseHerokuContainer(ctx, log, h, p.config.App, artifact.ContainerImageDigest); err != nil {
			return nil, err
		}
	} else if artifact.SlugID != "" {
		if err := p.releaseHerokuSlug(ctx, log, h, job, p.config.App, artifact.SlugID); err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("missing either container or slug artifact")
	}

	app, err := h.AppInfo(ctx, p.config.App)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		Url: app.WebURL,
	}, nil
}

func (p *Platform) releaseHerokuContainer(ctx context.Context, log hclog.Logger, h *herokuSDK.Service, app, dockerImage string) error {
	type Update struct {
		DockerImage string `json:"docker_image" url:"docker_image,key"`
		Process     string `json:"process" url:"process,key"`
	}

	opts := struct {
		Updates []Update `json:"updates" url:"updates,key"`
	}{}
	opts.Updates = append(opts.Updates, Update{Process: "web", DockerImage: dockerImage})
	log.Info(
		"About to update formation",
		"app", app,
		"dockerImage", dockerImage,
		"opts", opts,
	)

	var formation herokuSDK.FormationBatchUpdateResult
	if err := h.Patch(ctx, &formation, fmt.Sprintf("/apps/%v/formation", app), opts); err != nil {
		return err
	}

	log.Info(
		"Formation updated",
		"formation", formation,
	)
	return nil
}

func (p *Platform) releaseHerokuSlug(ctx context.Context, log hclog.Logger, h *herokuSDK.Service, job *component.JobInfo, app, slugID string) error {
	desc := "Deployed " + job.Id
	release, err := h.ReleaseCreate(ctx, app, herokuSDK.ReleaseCreateOpts{
		Description: &desc,
		Slug:        slugID,
	})
	if err != nil {
		return err
	}

	log.Info(
		"Release created",
		"release", release,
	)
	return nil
}

var (
	_ component.Platform         = (*Platform)(nil)
	_ component.Configurable     = (*Platform)(nil)
	_ component.PlatformReleaser = (*Platform)(nil)
	_ component.Deployment       = (*Deployment)(nil)
)
