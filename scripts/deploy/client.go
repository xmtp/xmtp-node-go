package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/go-tfe"
)

type client struct {
	*tfe.Client
	ctx context.Context
	wsp *tfe.Workspace
}

func newClient() *client {
	config := &tfe.Config{Token: options.TFToken}
	tfc, err := tfe.NewClient(config)
	failIfError(err, "creating client")

	ctx := context.Background()
	wsp, err := tfc.Workspaces.Read(ctx, options.Organization, options.Workspace)
	failIfError(err, "getting workspace %s/%s", options.Organization, options.Workspace)

	return &client{tfc, ctx, wsp}
}

func (c *client) updateVar(name string, value string) {
	vars, err := c.Variables.List(c.ctx, c.wsp.ID, nil)
	failIfError(err, "listing workspace vars")

	for _, v := range vars.Items {
		if v.Key == name {
			_, err := c.Variables.Update(c.ctx, c.wsp.ID, v.ID, tfe.VariableUpdateOptions{Value: &value})
			failIfError(err, "updating variable %s", name)
			return
		}
	}
	log.Fatalf("variable %s not found", name)
}

func (c client) startRun(commit string) {
	msg := fmt.Sprintf("triggered from xtmp-node-go commit %s", commit)
	_, err := c.Runs.Create(c.ctx, tfe.RunCreateOptions{
		Message:   &msg,
		Workspace: c.wsp,
	})
	failIfError(err, "creating run")
}
