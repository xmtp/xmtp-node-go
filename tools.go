//go:build tools
// +build tools

package tools

import (
	_ "github.com/yoheimuta/protolint/cmd/protolint"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
