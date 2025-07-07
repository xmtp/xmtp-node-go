package history

import (
	"context"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	ALCHEMY_MOCK_URL = "https://fake.alchemyapi.io/v2"
)

func newFetcher() *AlchemyTransactionHistoryFetcher {
	logger, _ := zap.NewDevelopment()
	return NewAlchemyTransactionHistoryFetcher(ALCHEMY_MOCK_URL, logger)
}

func TestAlchemyRequestSuccess(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResponse := `
	{
		"id": "57353e64-9a4b-49f4-918b-53093edc7892",
		"result": {
		  "transfers": [
			{
			  "blockNum": "0xb08dd2",
			  "hash": "0x75a77de7074d403ed6886aa618bca38c1c9e1cb3eef569969650302d76f8b962",
			  "from": "0x8707607d24371ae43e1a6a9bf8cdafbb4fd4ae43",
			  "to": "0x70e36f6bf80a52b3b46b3af8e106cc0ed743e8e4",
			  "value": 0.56656055,
			  "asset": "COMP",
			  "category": "erc20"
			}
		  ]
		}
	}`
	httpmock.RegisterResponder(
		"POST",
		ALCHEMY_MOCK_URL,
		httpmock.NewStringResponder(200, mockResponse),
	)

	fetcher := newFetcher()
	res, err := fetcher.Fetch(context.Background(), "0xDADBOD")

	require.NoError(t, err)
	require.Equal(t, res.HasTransactions(), true)
	require.Equal(t, httpmock.GetTotalCallCount(), 1)
}

func TestAlchemyRequestNoTransactions(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	mockResponse := `
	{
		"id": "57353e64-9a4b-49f4-918b-53093edc7892",
		"result": {
		  "transfers": []
		}
	  }
	`
	httpmock.RegisterResponder(
		"POST",
		ALCHEMY_MOCK_URL,
		httpmock.NewStringResponder(200, mockResponse),
	)

	fetcher := newFetcher()
	res, err := fetcher.Fetch(context.Background(), "0xDADBOD")

	require.NoError(t, err)
	require.Equal(t, res.HasTransactions(), false)
	require.Equal(t, httpmock.GetTotalCallCount(), 1)
}

func TestAlchemyError(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", ALCHEMY_MOCK_URL, httpmock.NewStringResponder(500, ""))

	fetcher := newFetcher()
	res, err := fetcher.Fetch(context.Background(), "0xDADBOD")

	require.Error(t, err)
	require.Nil(t, res)
}

func TestBuildGetAssetTransfersRequest(t *testing.T) {
	walletAddress := "0xfoo"
	req := buildGetAssetTransfersRequest(walletAddress, From)

	require.NotNil(t, req.Id)
	require.Equal(t, req.Method, "alchemy_getAssetTransfers")
	require.Equal(t, req.JsonRPC, "2.0")
	require.Len(t, req.Params, 1)

	param := req.Params[0]
	require.Equal(t, param.FromAddress, walletAddress)
	require.Equal(t, param.FromBlock, "0x1")
	require.Empty(t, param.ToAddress)
	require.Empty(t, param.PageKey)
}
