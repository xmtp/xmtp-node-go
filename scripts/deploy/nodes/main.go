/*
This is a tool for triggering TF Cloud runs/deploys from master commits.
Usage:
	go run ./scripts/deploy/nodes \
		--tf-token XXX \
		--workspace dev \
		--xmtp-node-image xmtp/node-go@sha256:XXX \
		--git-commit=$(git rev-parse HEAD)
*/
package main

import (
	"context"
	"strings"

	"github.com/hashicorp/go-tfe"
	"github.com/jessevdk/go-flags"
	"github.com/xmtp/xmtp-node-go/internal/terraform"
	"go.uber.org/zap"
)

const (
	nodeImagePrefix = "xmtp/node-go@sha256:"
)

var options struct {
	TFToken        string `long:"tf-token" description:"Terraform token"`
	Workspace      string `long:"workspace" description:"TF cloud workspace" choice:"dev" choice:"production"`
	Organization   string `long:"organization" default:"xmtp" choice:"xmtp"`
	ContainerImage string `long:"container-image"`
	GitCommit      string `long:"git-commit"`
}

func main() {
	ctx := context.Background()

	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	_, err = flags.NewParser(&options, flags.Default).Parse()
	if err != nil {
		log.Fatal("parsing options", zap.Error(err))
	}

	tfc, err := tfe.NewClient(&tfe.Config{
		Token: options.TFToken,
	})
	if err != nil {
		log.Fatal("creating terraform client", zap.Error(err))
	}

	deployer, err := terraform.NewDeployer(ctx, log, tfc, options.Organization, options.TFToken)
	if err != nil {
		log.Fatal("creating deployer", zap.Error(err))
	}

	if options.ContainerImage == "" {
		log.Fatal("Must specify container-image")
	}

	if !strings.HasPrefix(options.ContainerImage, nodeImagePrefix) {
		log.Fatal("Invalid node image %s", zap.String("image", options.ContainerImage))
	}

	deployer.Deploy("xmtp_node_image", options.ContainerImage, options.GitCommit)
}
