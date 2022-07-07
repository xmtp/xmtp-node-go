# xmtp-node-go

XMTP Node software

## Instructions

### Start the server

1. `make build`
2. `./build/xmtp`

### Build the Docker image

1. `make docker-image`

### Run the tests

1. `make test`

### Create a migration for the message database

1. `docker-compose up -d`
2. `make build`
3. `./build/xmtp --message-db-connection-string "postgres://postgres:xmtp@localhost:5432/postgres?sslmode=disable" --create-message-migration $MIGRATION_NAME`

### Create a migration for the authz database

1. `docker-compose up -d`
2. `make build`
3. `./build/xmtp --authz-db-connection-string "postgres://postgres:xmtp@localhost:6543/postgres?sslmode=disable" --create-authz-migration $MIGRATION_NAME`

### Debugging metrics

1. `docker-compose up -d`
2. `make build`
3. `./build/xmtp --message-db-connection-string "postgres://postgres:xmtp@localhost:6543/postgres?sslmode=disable" --metrics`
4. browse to http://localhost:9090 to see prometheus interface

## Deployments

Merging a PR to either the `main` branch will trigger a new deployment via Github Actions and Terraform.

The default behavior is to deploy `main` to both the `dev` and `production` environments. If you'd like to deploy a different branch to `dev`, open a PR with an update to [.github/workflows/deploy.yml](https://github.com/xmtp/xmtp-node-go/blob/main/.github/workflows/deploy.yml#L29) switching from `main` to your branch. Remember to PR it back to `main` when you're done.

Changes will be automatically applied and no user action is required after merging.
