# xmtp-node-go

XMTP Node software

## Instructions

### Start the server

1. `make build`
2. `./build/xmtp`

### Build the Docker image

1. `make docker-image`

### Create a migration for the message table

1. `docker-compose up -d`
2. `make build`
3. `./build/xmtp --message-db-connection-string "postgres://postgres:xmtp@localhost:5432/postgres?sslmode=disable" --create-message-migration $MIGRATION_NAME`

### Create a migration for the authz table

1. `docker-compose up -d`
2. `make build`
3. `./build/xmtp --authz-db-connection-string "postgres://postgres:xmtp@localhost:6543/postgres?sslmode=disable" --create-authz-migration $MIGRATION_NAME`
