package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/docker/cli/cli/config"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/registry"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	wpdocker "github.com/hashicorp/waypoint/builtin/docker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegistryConfig struct {
	Pipeline string `hcl:"pipeline"`
	App      string `hcl:"app,optional"`
}

type Registry struct {
	config RegistryConfig
}

// Implement Configurable
func (r *Registry) Config() (interface{}, error) {
	return &r.config, nil
}

// Implement Registry
func (r *Registry) PushFunc() interface{} {
	// return a function which will be called by Waypoint
	return r.push
}

func (r *Registry) push(
	ctx context.Context,
	img *wpdocker.Image,
	ui terminal.UI,
	log hclog.Logger,
) (*Artifact, error) {
	stdout, _, err := ui.OutputWriters()
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to create output for logs:%s", err)
	}

	sg := ui.StepGroup()
	step := sg.Add("Initializing Docker client...")
	defer func() { step.Abort() }()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "unable to create Docker client:%s", err)
	}
	cli.NegotiateAPIVersion(ctx)

	target := "registry.heroku.com/" + r.config.App
	step.Update("Tagging Docker image: %s => %s", img.Name(), target)

	imgInspect, _, err := cli.ImageInspectWithRaw(ctx, img.Name())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to inspect image:%s", err)
	}

	err = cli.ImageTag(ctx, img.Name(), target)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to tag image:%s", err)
	}

	step.Done()

	ref, err := reference.ParseNormalizedNamed(target)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to parse image name: %s", err)
	}

	// Resolve the Repository name from fqn to RepositoryInfo
	repoInfo, err := registry.ParseRepositoryInfo(ref)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to parse repository info from image name: %s", err)
	}

	var server string

	if repoInfo.Index.Official {
		info, err := cli.Info(ctx)
		if err != nil || info.IndexServerAddress == "" {
			server = registry.IndexServer
		} else {
			server = info.IndexServerAddress
		}
	} else {
		server = repoInfo.Index.Name
	}

	var errBuf bytes.Buffer
	cf := config.LoadDefaultConfigFile(&errBuf)
	if errBuf.Len() > 0 {
		// NOTE(mitchellh): I don't know why we ignore this, but we always have.
		log.Warn("error loading Docker config file", "err", err)
	}

	log.Info("server", server)
	authConfig, _ := cf.GetAuthConfig(server)
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to generate authentication info for registry: %s", err)
	}
	encodedAuth := base64.URLEncoding.EncodeToString(buf)

	step = sg.Add("Pushing Docker image...")

	options := types.ImagePushOptions{
		RegistryAuth: encodedAuth,
	}

	responseBody, err := cli.ImagePush(ctx, reference.FamiliarString(ref), options)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to push image to registry: %s", err)
	}

	defer responseBody.Close()

	var termFd uintptr
	if f, ok := stdout.(*os.File); ok {
		termFd = f.Fd()
	}

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, step.TermOutput(), termFd, true, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to stream Docker logs to terminal: %s", err)
	}

	step.Done()

	step = sg.Add("Docker image pushed: %s", target)
	step.Done()

	return &Artifact{ContainerImageDigest: imgInspect.ID}, nil
}

var (
	_ component.Registry     = (*Registry)(nil)
	_ component.Configurable = (*Registry)(nil)
)
