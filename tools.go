//go:build tools
// +build tools

package tools

import (
	_ "github.com/yoheimuta/protolint/cmd/protolint"
	_ "go.uber.org/mock/mockgen"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
