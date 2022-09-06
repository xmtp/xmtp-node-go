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
	"fmt"
	"log"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/xmtp/xmtp-node-go/internal/terraform"
)

const (
	nodeImagePrefix = "xmtp/node-go@sha256:"
)

var options struct {
	TFToken      string `long:"tf-token" description:"Terraform token"`
	Workspace    string `long:"workspace" description:"TF cloud workspace" choice:"dev" choice:"production"`
	Organization string `long:"organization" default:"xmtp" choice:"xmtp"`
	NodeImage    string `long:"xmtp-node-image"`
	Apply        bool   `long:"apply"`
	Commit       string `long:"git-commit"`
}

func main() {
	_, err := flags.NewParser(&options, flags.Default).Parse()
	failIfError(err, "parsing options")

	c, err := terraform.NewClient(&terraform.Config{
		Token:        options.TFToken,
		Workspace:    options.Workspace,
		Organization: options.Organization,
	})
	failIfError(err, "creating client")

	if options.NodeImage == "" {
		log.Fatal("Must specify xmtp-node-image")
	}

	if !strings.HasPrefix(options.NodeImage, nodeImagePrefix) {
		log.Fatalf("Invalid node image %s", options.NodeImage)
	}
	c.UpdateVar("xmtp_node_image", options.NodeImage)

	c.StartRun(options.Commit, options.Apply)
}

func failIfError(err error, msg string, params ...interface{}) {
	if err == nil {
		return
	}
	m := fmt.Sprintf(msg, params...)
	log.Fatalf("%s:%s", m, err.Error())
}
