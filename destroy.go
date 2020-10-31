package main

import (
	"context"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

// Implement the Destroyer interface
func (p *Platform) DestroyFunc() interface{} {
	return p.destroy
}

func (p *Platform) destroy(ctx context.Context, ui terminal.UI, deployment *Deployment) error {
	return nil
}
