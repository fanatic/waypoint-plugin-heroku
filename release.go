package main

import (
	"context"
	"runtime/debug"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

// Releaser is the ReleaseManager implementation
type Releaser struct {
}

// ReleaseFunc implements component.ReleaseManager
func (r *Releaser) ReleaseFunc() interface{} {
	return r.release
}

func (r *Release) URL() string { return r.Url }

func (r *Releaser) release(
	ctx context.Context,
	ui terminal.UI,
	src *component.Source,
	job *component.JobInfo,
	log hclog.Logger,
	deployment *Deployment,
) (*Release, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("Panic: %s %s", r, string(debug.Stack()))
		}
	}()

	return &Release{
		Url: deployment.Url,
	}, nil
}

var (
	_ component.ReleaseManager = (*Releaser)(nil)
	_ component.Release        = (*Release)(nil)
)
