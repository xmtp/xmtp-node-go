package context

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

const (
	ClientVersionMetadataKey = "x-client-version"
	AppVersionMetadataKey    = "x-app-version"
)

type requesterInfo struct {
	ClientName    string
	ClientVersion string

	AppName    string
	AppVersion string
}

func NewRequesterInfo(ctx context.Context) *requesterInfo {
	ri := &requesterInfo{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ri
	}
	ri.ClientName, _, ri.ClientVersion = parseVersionHeaderValue(md.Get(ClientVersionMetadataKey))
	ri.AppName, _, ri.AppVersion = parseVersionHeaderValue(md.Get(AppVersionMetadataKey))
	md.Append("X-User-Id", "real_user_id")
	return ri
}

func (ri *requesterInfo) ZapFields() []zap.Field {
	return []zap.Field{
		zap.String("client", ri.ClientName),
		zap.String("client_version", ri.ClientVersion),
		zap.String("app", ri.AppName),
		zap.String("app_version", ri.AppVersion),
	}
}

func parseVersionHeaderValue(vals []string) (name string, version string, full string) {
	if len(vals) == 0 {
		return
	}
	full = vals[0]
	parts := strings.Split(full, "/")
	if len(parts) > 0 {
		name = parts[0]
		if len(parts) > 1 {
			version = parts[1]
		}
	}
	return
}
