package history

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	HTTP_TIMEOUT_SECONDS = 5
	PAGE_SIZE            = 100
)

// AssetTransferDirection determines whether we are looking for transfers to or from the given wallet address
type AssetTransferDirection int64

const (
	To   AssetTransferDirection = 0
	From AssetTransferDirection = 1
)

// An implementation of the TransactionHistoryFetcher interface for Alchemy
type AlchemyTransactionHistoryFetcher struct {
	client *http.Client
	log    *zap.Logger
	apiUrl string
}

// GetAssetTransfersRequest is the format expected by the Alchemy API
// https://docs.alchemy.com/alchemy/enhanced-apis/transfers-api
type GetAssetTransfersRequest struct {
	JsonRPC string                  `json:"jsonrpc"`
	Id      string                  `json:"id"`
	Method  string                  `json:"method"`
	Params  []GetAssetTransferParam `json:"params"`
}

// GetAssetTransferParam is the structure of the params expected in GetAssetTransfersRequest
type GetAssetTransferParam struct {
	FromAddress string   `json:"fromAddress,omitempty"`
	FromBlock   string   `json:"fromBlock,omitempty"`
	ToAddress   string   `json:"toAddress,omitempty"`
	MaxCount    string   `json:"maxCount,omitempty"`
	Category    []string `json:"category,omitempty"`
	PageKey     *string  `json:"pageKey,omitempty"`
}

// AlchemyResult is a simplified version of the response struct
type AlchemyResult struct {
	Result struct {
		PageKey   string `json:"pageKey"`
		Transfers []struct {
			Category string   `json:"category"`
			From     string   `json:"from"`
			To       *string  `json:"to"`
			Value    *float64 `json:"value"`
			Asset    *string  `json:"asset"`
		} `json:"transfers"`
	} `json:"result"`
}

// An implementation of the TokenDataResult interface
type AlchemyTokenDataResult struct {
	hasTransactions bool
}

// Return a new instance of the AlchemyTransactionHistoryFetcher
func NewAlchemyTransactionHistoryFetcher(apiUrl string, log *zap.Logger) *AlchemyTransactionHistoryFetcher {
	fetcher := new(AlchemyTransactionHistoryFetcher)
	fetcher.log = log
	fetcher.apiUrl = apiUrl
	fetcher.client = &http.Client{
		Timeout: time.Second * HTTP_TIMEOUT_SECONDS,
	}

	return fetcher
}

// Fetch in parallel transactions to and from the wallet address
// Question: is it really necessary to have to AND from. Seems impossible to have a transaction from a wallet without a transaction to the wallet
func (fetcher AlchemyTransactionHistoryFetcher) Fetch(ctx context.Context, walletAddress string) (res TransactionHistoryResult, err error) {
	fetcher.log.Info("Fetching Alchemy data for wallet", zap.String("wallet_address", walletAddress))
	var response AlchemyResult
	response, err = fetcher.alchemyRequest(ctx, walletAddress, From)
	if err != nil {
		fetcher.log.Error("Error fetching from Alchemy", zap.Error(err))
		return res, err
	}
	hasTransactions := verifyTransactions(response)
	res = AlchemyTokenDataResult{hasTransactions: hasTransactions}
	fetcher.log.Info("Got Alchemy data for wallet", zap.String("wallet_address", walletAddress), zap.Bool("has_transactions", hasTransactions))
	return res, err
}

func (fetcher AlchemyTransactionHistoryFetcher) alchemyRequest(ctx context.Context, walletAddress string, direction AssetTransferDirection) (res AlchemyResult, err error) {
	transferRequest := buildGetAssetTransfersRequest(walletAddress, direction)
	requestPayload, err := json.Marshal(transferRequest)
	if err != nil {
		return res, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fetcher.apiUrl, bytes.NewBuffer(requestPayload))
	if err != nil {
		return res, err
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := fetcher.client.Do(request)
	if err != nil {
		return res, err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return res, err
	}

	err = json.Unmarshal(responseBody, &res)
	return res, err
}

func verifyTransactions(result AlchemyResult) (hasTransactions bool) {
	for _, transfer := range result.Result.Transfers {
		// TODO fancier logic for validating transfers
		// This check is just for safety. Alchemy's API should be omitting 0 value transfers
		if transfer.Value != nil && *transfer.Value > 0 {
			hasTransactions = true
			break
		}
	}
	return
}

func (r AlchemyTokenDataResult) HasTransactions() bool {
	return r.hasTransactions
}

func buildGetAssetTransfersRequest(walletAddress string, direction AssetTransferDirection) GetAssetTransfersRequest {
	param := GetAssetTransferParam{
		// Need to set an early FromBlock as the default behaviour is to search latest -> latest apparently
		FromBlock: "0x1",
		MaxCount:  fmt.Sprintf("0x%x", PAGE_SIZE),
		Category:  []string{"token", "erc20", "erc721", "erc1155"},
	}
	if direction == To {
		param.ToAddress = walletAddress
	} else {
		param.FromAddress = walletAddress
	}

	return GetAssetTransfersRequest{
		JsonRPC: "2.0",
		Id:      uuid.NewString(),
		Method:  "alchemy_getAssetTransfers",
		Params:  []GetAssetTransferParam{param},
	}
}
