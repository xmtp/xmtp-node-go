# xmtp-node-go

XMTP node software

## Instructions

### Install dependencies and start the DB

1. `dev/up`

### Run the tests

1. `dev/test`

### Start a local node

1. `dev/start`

### Debugging metrics

1. `dev/run --metrics`
2. Browse to http://localhost:9090 to see prometheus interface

## Deployments

Merging a PR to the `main` branch will trigger a new deployment via Github Actions and Terraform.

The default behavior is to deploy `main` to both the `dev` and `production` environments. If you'd like to deploy a different branch to `dev`, open a PR with an update to [.github/workflows/deploy.yml](https://github.com/xmtp/xmtp-node-go/blob/main/.github/workflows/deploy.yml#L29) switching from `main` to your branch. Remember to PR it back to `main` when you're done.

Changes will be automatically applied and no user action is required after merging.
