// logging implements custom logging field types for commonly
// logged values like host ID or wallet address.
//
// implementation purposely does as little as possible at field creation time,
// and postpones any transformation to output time by relying on the generic
// zap types like zap.Stringer, zap.Array, zap.Object
package logging

import (
	"fmt"

	"github.com/status-im/go-waku/logging"
	proto "github.com/xmtp/proto/v3/go/message_api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Re-export relevant go-waku helpers
	MultiAddrs = logging.MultiAddrs
	HostID     = logging.HostID
	Time       = logging.Time
	Filters    = logging.Filters
	PagingInfo = logging.PagingInfo
	HexBytes   = logging.HexBytes
	ENode      = logging.ENode
	TCPAddr    = logging.TCPAddr
	UDPAddr    = logging.UDPAddr
	String     = zap.String
	Bool       = zap.Bool
	Int        = zap.Int
)

// WalletAddress creates a field for a wallet address.
func WalletAddress(address string) zapcore.Field {
	return zap.String("wallet_address", address)
}

func WalletAddressLabelled(label string, address types.WalletAddr) zapcore.Field {
	return zap.String(label, string(address))
}

type valueType struct{ val interface{} }

func ValueType(key string, val interface{}) zap.Field {
	return zap.Stringer(key, valueType{val})
}

func (vt valueType) String() string {
	return fmt.Sprintf("%T", vt.val)
}

type queryParameters struct{ *proto.QueryRequest }

func (qp queryParameters) String() string {
	return QueryShape(qp.QueryRequest)
}

func QueryParameters(req *proto.QueryRequest) zap.Field {
	return zap.Stringer("parameters", queryParameters{req})
}

func QueryShape(req *proto.QueryRequest) string {
	params := []byte("-------")
	if req.StartTimeNs > 0 {
		params[0] = 'S'
	}
	if req.EndTimeNs > 0 {
		params[1] = 'E'
	}
	if pi := req.PagingInfo; pi != nil {
		if pi.Direction == proto.SortDirection_SORT_DIRECTION_ASCENDING {
			params[2] = 'A'
		} else {
			params[2] = 'D'
		}
		if pi.Limit > 0 {
			params[3] = 'L'
		}
		if ci := pi.Cursor.GetIndex(); ci != nil {
			params[4] = 'C'
			if ci.Digest != nil {
				params[5] = 'D'
			}
			if ci.SenderTimeNs > 0 {
				params[6] = 'T'
			}
		}
	}
	return string(params)
}
