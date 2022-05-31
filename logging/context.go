package logging

import (
	"github.com/status-im/go-waku/logging"
)

var (
	// Re-export the go-waku helpers.
	With = logging.With
	From = logging.From
)
