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

const (
	nodeImagePrefix      = "xmtp/node-go@sha256:"
	e2eRunnerImagePrefix = "xmtp/xmtpd-e2e@sha256:"
)

var options struct {
	TFToken        string `long:"tf-token" description:"Terraform token"`
	Workspace      string `long:"workspace" description:"TF cloud workspace" choice:"dev" choice:"production"`
	Organization   string `long:"organization" default:"xmtp" choice:"xmtp"`
	NodeImage      string `long:"xmtp-node-image"`
	E2ERunnerImage string `long:"xmtpd-e2e-image"`
	Apply          bool   `long:"apply"`
	Commit         string `long:"git-commit"`
}

func main() {
	_, err := flags.NewParser(&options, flags.Default).Parse()
	failIfError(err, "parsing options")

	c := newClient()

	if options.NodeImage == "" && options.E2ERunnerImage == "" {
		log.Fatal("Must specify either xmtp-node-image or xmtpd-e2e-image")
	}

	if options.NodeImage != "" {
		if !strings.HasPrefix(options.NodeImage, nodeImagePrefix) {
			log.Fatalf("Invalid node image %s", options.NodeImage)
		}
		c.updateVar("xmtp_node_image", options.NodeImage)
	}

	if options.E2ERunnerImage != "" {
		if !strings.HasPrefix(options.E2ERunnerImage, e2eRunnerImagePrefix) {
			log.Fatalf("Invalid e2e image %s", options.E2ERunnerImage)
		}
		c.updateVar("xmtpd_e2e_image", options.E2ERunnerImage)
	}

	c.startRun(options.Commit, options.Apply)

}

func failIfError(err error, msg string, params ...interface{}) {
	if err == nil {
		return
	}
	m := fmt.Sprintf(msg, params...)
	log.Fatalf("%s:%s", m, err.Error())
}
