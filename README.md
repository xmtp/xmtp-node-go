# xmtp-node-go

This repo provides software for the nodes that currently form the XMTP network. **No new development is planned for this node software.**

At this time, all nodes in the XMTP network are run by XMTP Labs, whose mission is to promote and support the development and global adoption of XMTP.

All new development is focused on [xmtpd](https://github.com/xmtp/xmtpd), an **experimental** version of XMTP node software.

After `xmtpd` meets specific functional requirements, the plan is for it to become the node software that powers the XMTP network. In the future, anyone will be able to run an `xmtpd` node that participates in the XMTP network.

## Instructions

### Install prerequisites

- [Go](https://go.dev/doc/install)
- [Docker](https://www.docker.com/get-started/)

You must have the _exact_ go version listed in `go.mod` - you can verify this by running `go version`.

### Install dependencies and start the DB

1. `dev/up`

### Run the tests

1. `dev/test`

### Start a local node

1. `dev/start`

### Lint your files

1. `dev/lint`

### Fix formatting issues

1. `golangci-lint --config .golangci.yaml run --fix ./...`

### Create a migration for the message database

1. `dev/migrate-message $MIGRATION_NAME`

### Create a migration for the MLS database

1. `dev/migrate-mls $MIGRATION_NAME`

### Create a migration for the authz database

1. `dev/migrate-authz $MIGRATION_NAME`

### Updating the SQLC Queries for the MLSStore

If you modify `pkg/mls/store/queries.sql` you need to run `./dev/sqlc` from the root of this package to regenerate any generated code.

### Debugging metrics

1. `dev/run --metrics`
2. Browse to http://localhost:9090 to see prometheus interface

## Deployments

Merging a PR to the `main` branch will trigger a new deployment via GitHub Actions and Terraform.

The default behavior is to deploy `main` to both the `dev` and `production` environments. If you'd like to deploy a different branch to `dev`, open a PR with an update to [.github/workflows/deploy.yml](https://github.com/xmtp/xmtp-node-go/blob/main/.github/workflows/deploy.yml#L29) switching from `main` to your branch. Remember to PR it back to `main` when you're done.

Changes will be automatically applied and no user action is required after merging.
