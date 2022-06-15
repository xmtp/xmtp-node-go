/*
This is a tool for triggering TF Cloud runs/deploys from master commits.
Usage:
	go run ./scripts/deploy \
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
)

const NODE_IMAGE_PREFIX = "xmtp/node-go@sha256:"

var options struct {
	TFToken      string `long:"tf-token" description:"Terraform token"`
	Workspace    string `long:"workspace" description:"TF cloud workspace" choice:"dev" choice:"testnet"`
	Organization string `long:"organization" default:"xmtp" choice:"xmtp"`
	NodeImage    string `long:"xmtp-node-image"`
	Apply        bool   `long:"apply"`
	Commit       string `long:"git-commit"`
}

func main() {
	_, err := flags.NewParser(&options, flags.Default).Parse()
	failIfError(err, "parsing options")

	if !strings.HasPrefix(options.NodeImage, NODE_IMAGE_PREFIX) {
		log.Fatalf("Invalid node image %s", options.NodeImage)
	}

	c := newClient()
	c.updateVar("xmtp_node_image", options.NodeImage)
	c.startRun(options.Commit, options.Apply)

}

func failIfError(err error, msg string, params ...interface{}) {
	if err == nil {
		return
	}
	m := fmt.Sprintf(msg, params...)
	log.Fatalf("%s:%s", m, err.Error())
}
