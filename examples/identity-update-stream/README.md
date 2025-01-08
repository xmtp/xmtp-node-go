# Identity Update Stream Example

## Background

XMTP nodes expose a streaming API that provides callers with all identity association changes on the network in real-time.

An identity association change event is created any time a blockchain account (wallet address) is associated with an Inbox ID or revoked from an Inbox ID.

This API method is not supported in our web or mobile SDKs, since the intended use is to call it from a backend service.

## Use-cases

- An application may want to keep an up-to-date registry of all account address -> inbox ID mappings to power a search API.
- An application may want to notify users when a new wallet association is added to their InboxID

## Usage

This API can be called from any environment that supports GRPC and Protobuf. You can generate a client in your language of choice following the instructions [here](https://protobuf.dev/reference/). The .proto can be found [here](https://github.com/xmtp/proto/blob/main/proto/identity/api/v1/identity.proto).

This [example](./main.go) uses a Golang client generated via Buf and connects to a local instance of our node software. To connect to the `dev` or `production` network you would need to replace `localhost:50051` with `grpc.dev.xmtp.network:443` or `grpc.production.xmtp.network:443` and configure the client to use TLS.
