package main

import (
	"github.com/fanatic/waypoint-plugin-heroku/builder"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
	heroku "github.com/heroku/heroku-go/v5"
)

func main() {
	h := heroku.NewService(heroku.DefaultClient)

	// Main sets up all the go-plugin requirements
	sdk.Main(sdk.WithComponents(
		&builder.Builder{H: h},
		//&registry.Registry{},
		//&platform.Platform{},
		//&release.ReleaseManager{},
	))
}
