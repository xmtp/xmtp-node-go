package context

import (
	"context"
	"os"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/mod/semver"
	"google.golang.org/grpc/metadata"
)

const (
	AppVersionMetadataKey     = "x-app-version"
	ClientVersionMetadataKey  = "x-client-version"
	LibxmtpVersionMetadataKey = "x-libxmtp-version"
)

type requesterInfo struct {
	AppName    string
	AppVersion string

	ClientName    string
	ClientVersion string

	IsSupportedClient bool

	LibxmtpVersion string
}

func NewRequesterInfo(ctx context.Context) *requesterInfo {
	ri := &requesterInfo{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ri
	}
	ri.AppName, _, ri.AppVersion = parseVersionHeaderValue(md.Get(AppVersionMetadataKey))
	ri.ClientName, _, ri.ClientVersion = parseVersionHeaderValue(md.Get(ClientVersionMetadataKey))
	_, _, ri.LibxmtpVersion = parseVersionHeaderValue(md.Get(LibxmtpVersionMetadataKey))
	ri.IsSupportedClient = ri.isSupportedClient()
	md.Append("X-User-Id", "real_user_id")
	return ri
}

func (ri *requesterInfo) ZapFields() []zap.Field {
	return []zap.Field{
		zap.String("app", ri.AppName),
		zap.String("app_version", ri.AppVersion),
		zap.String("client", ri.ClientName),
		zap.String("client_version", ri.ClientVersion),
		zap.Bool("is_supported_client", ri.IsSupportedClient),
		zap.String("libxmtp_version", ri.LibxmtpVersion),
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

func (ri *requesterInfo) isSupportedClient() bool {
	// Only version-gate on production
	if os.Getenv("ENV") != "production" {
		return true
	}
	// Err on the side of caution for unknown versions
	if ri.LibxmtpVersion == "" || !semver.IsValid(ri.LibxmtpVersion) {
		return true
	}
	if ri.LibxmtpVersion == "1.2.0-dev" {
		return true
	}
	if semver.Compare(ri.LibxmtpVersion, "1.1.5") >= 0 {
		return true
	}

	return false
}
