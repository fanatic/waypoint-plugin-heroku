package main

import (
	"github.com/fanatic/waypoint-plugin-heroku/builder"
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
)

func main() {
	// Main sets up all the go-plugin requirements
	sdk.Main(sdk.WithComponents(
		&builder.Builder{},
		//&registry.Registry{},
		//&platform.Platform{},
		//&release.ReleaseManager{},
	))
}
