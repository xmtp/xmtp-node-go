package authn

import (
	"github.com/xmtp/xmtp-node-go/pkg/authz"
	"github.com/xmtp/xmtp-node-go/pkg/ratelimiter"
)

// Options bundle command line options associated with the authn package.
type Options struct {
	Enable     bool `long:"enable" description:"require client authentication via wallet tokens"`
	Ratelimits bool `long:"ratelimits" description:"apply rate limits per wallet"`
	AllowLists bool `long:"allowlists" description:"apply higher limits for allow listed wallets (requires authz and ratelimits)"`
}

// Config bundles Options and other parameters needed to set up an authorizer.
type Config struct {
	Options
	Limiter     ratelimiter.RateLimiter
	AllowLister authz.WalletAllowLister
}
