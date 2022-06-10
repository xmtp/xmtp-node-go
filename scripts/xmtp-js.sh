#!/bin/sh
#
# Run the node on the host set up so that xmtp-js tests can run against it if LOCAL_NODE is set.
# ```
#   xmtp-js % LOCAL_NODE=true npm test
# ```
# Expects the binary to have been built via `make build`
#
build/xmtp \
--nodekey 8a30dcb604b0b53627a5adc054dbf434b446628d4bd1eccc681d223f0550ce67 \
--ws \
--ws-port=9002 \
--store \
--message-db-connection-string=postgres://postgres:xmtp@localhost:5432/postgres?sslmode=disable \
--authz-db-connection-string=postgres://postgres:xmtp@localhost:6543/postgres?sslmode=disable \
--lightpush \
--filter \
--metrics \
--metrics-period 5s
