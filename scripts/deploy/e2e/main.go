/*
This is a tool for triggering TF Cloud runs/deploys from master commits.
Usage:
	go run ./scripts/deploy/e2e \
		--tf-token XXX \
		--workspace dev \
		--xmtp-e2e-image xmtp/xmtpd-e2e@sha256:XXX \
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
	e2eRunnerImagePrefix = "xmtp/xmtpd-e2e@sha256:"
)

var options struct {
	TFToken        string `long:"tf-token" description:"Terraform token"`
	Workspace      string `long:"workspace" description:"TF cloud workspace" choice:"dev" choice:"production"`
	Organization   string `long:"organization" default:"xmtp" choice:"xmtp"`
	E2ERunnerImage string `long:"xmtpd-e2e-image"`
	Apply          bool   `long:"apply"`
	Commit         string `long:"git-commit"`
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

	if options.E2ERunnerImage == "" {
		log.Fatal("Must specify xmtpd-e2e-image")
	}

	if !strings.HasPrefix(options.E2ERunnerImage, e2eRunnerImagePrefix) {
		log.Fatalf("Invalid e2e image %s", options.E2ERunnerImage)
	}
	c.UpdateVar("xmtpd_e2e_image", options.E2ERunnerImage)

	c.StartRun(options.Commit, options.Apply)
}

func failIfError(err error, msg string, params ...interface{}) {
	if err == nil {
		return
	}
	m := fmt.Sprintf(msg, params...)
	log.Fatalf("%s:%s", m, err.Error())
}
