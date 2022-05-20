package logging

import (
	libp2p "github.com/ipfs/go-log"
	waku "github.com/status-im/go-waku/waku/v2/utils"
	"go.uber.org/zap"
)

// IsDebugLevel returns true if the Waku loggers are set to DEBUG level.
func IsDebugLevel() bool {
	return waku.Logger().Desugar().Core().Enabled(zap.DebugLevel)
}

// IfDebug aides conditional addition of logging fields.
//
// log.Info("query",
//	zap.String("from", peer.ID),
// 	logging.IfDebug(zap.Object("query", query)),
// )
func IfDebug(field zap.Field) zap.Field {
	if IsDebugLevel() {
		return field
	}
	return zap.Skip()
}

// ToggleDebugLevel toggles the log level between DEBUG and INFO level.
func ToggleDebugLevel() {
	levelWaku := "DEBUG"
	levelLibP2P := libp2p.LevelDebug
	if IsDebugLevel() {
		levelWaku = "INFO"
		levelLibP2P = libp2p.LevelInfo
	}
	_ = waku.SetLogLevel(levelWaku)
	libp2p.SetAllLoggers(levelLibP2P)
}
