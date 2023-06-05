package blockcypher_client

import (
	"fmt"
	go_http "github.com/pefish/go-http"
	go_logger "github.com/pefish/go-logger"
	"time"
)

type BlockcypherClient struct {
	timeout time.Duration
	logger  go_logger.InterfaceLogger
	baseUrl string
	key     string
}

func NewBlockcypherClient(
	logger go_logger.InterfaceLogger,
	httpTimeout time.Duration,
	baseUrl string,
	key string,
) *BlockcypherClient {
	return &BlockcypherClient{
		timeout: httpTimeout,
		logger:  logger,
		baseUrl: baseUrl,
		key:     key,
	}
}

type ListTransactionsResult struct {
	TxId        string `json:"hash"`
	BlockNumber int64  `json:"block_height"`
	DoubleSpend bool   `json:"double_spend"` // false
	Inputs      []struct {
		PrevHash    string   `json:"prev_hash"`
		OutputIndex uint64   `json:"output_index"`
		Addresses   []string `json:"addresses"`
	} `json:"inputs"`
	Outputs []struct {
		Value     uint64   `json:"value"`
		Addresses []string `json:"addresses"`
	} `json:"outputs"`
	Confirmations uint64 `json:"confirmations"`
}

func (bc *BlockcypherClient) ListTransactions(index uint64, address string) ([]ListTransactionsResult, error) {
	results := make([]ListTransactionsResult, 0)

	var httpResult struct {
		Txs     []ListTransactionsResult `json:"txs"`
		HasMore bool                     `json:"hasMore"`
		Error   string                   `json:"error"`
	}
	_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
		Url: fmt.Sprintf("%s/addrs/%s/full", bc.baseUrl, address),
		Params: map[string]interface{}{
			"after": index,
			"token": bc.key,
		},
	}, &httpResult)
	if err != nil {
		return nil, err
	}
	if httpResult.Error != "" {
		return nil, fmt.Errorf(httpResult.Error)
	}

	results = append(results, httpResult.Txs...)
	if !httpResult.HasMore {
		return results, nil
	}

	before := httpResult.Txs[len(httpResult.Txs)-1].BlockNumber
	for {
		var httpResult struct {
			Txs     []ListTransactionsResult `json:"txs"`
			HasMore bool                     `json:"hasMore"`
			Error   string                   `json:"error"`
		}
		_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
			Url: fmt.Sprintf("%s/addrs/%s/full", bc.baseUrl, address),
			Params: map[string]interface{}{
				"after":  index,
				"before": before,
				"token":  bc.key,
			},
		}, &httpResult)
		if err != nil {
			return nil, err
		}
		if httpResult.Error != "" {
			return nil, fmt.Errorf(httpResult.Error)
		}
		results = append(results, httpResult.Txs...)
		if !httpResult.HasMore {
			break
		}
		before = httpResult.Txs[len(httpResult.Txs)-1].BlockNumber
	}
	return results, nil
}

type GetTransactionResult struct {
	Confirmations uint64 `json:"confirmations"`
	Error         string `json:"error"`
}

func (bc *BlockcypherClient) GetTransaction(hash string) (*GetTransactionResult, error) {
	var result GetTransactionResult
	_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
		Url: fmt.Sprintf("%s/txs/%s", bc.baseUrl, hash),
		Params: map[string]interface{}{
			"token": bc.key,
		},
	}, &result)
	if err != nil {
		return nil, err
	}
	if result.Error != "" {
		return nil, fmt.Errorf(result.Error)
	}
	return &result, nil
}
