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

Merging a PR to either the `dev` or `main` branch will trigger a new deployment via Github Actions and Terraform.

- The `dev` branch will deploy to the `dev` environment and push to the Docker tag `latest`
- The `main` branch will deploy to the `production` environment and push to the Docker tag `production`

Changes will be automatically applied and no user action is required after merging.
