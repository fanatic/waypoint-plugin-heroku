package main

import (
	sdk "github.com/hashicorp/waypoint-plugin-sdk"
)

func main() {
	// Main sets up all the go-plugin requirements
	sdk.Main(sdk.WithComponents(
		&Builder{},
		&Registry{},
		&Platform{},
		//&Releaser{},
	))
}
