package blockcypher_client

import (
	"fmt"
	go_decimal "github.com/pefish/go-decimal"
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
	key string,
) *BlockcypherClient {
	return &BlockcypherClient{
		timeout: httpTimeout,
		logger:  logger,
		baseUrl: "https://api.blockcypher.com/v1/btc/main",
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

func (bc *BlockcypherClient) ListTransactions(index uint64, address string, isIncludePending bool) ([]ListTransactionsResult, error) {
	confirmations := 0
	if !isIncludePending {
		confirmations = 1
	}

	results := make([]ListTransactionsResult, 0)

	var httpResult struct {
		Txs     []ListTransactionsResult `json:"txs"`
		HasMore bool                     `json:"hasMore"`
		Error   string                   `json:"error"`
	}
	_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
		Url: fmt.Sprintf("%s/addrs/%s/full", bc.baseUrl, address),
		Params: map[string]interface{}{
			"after":         index,
			"token":         bc.key,
			"confirmations": confirmations,
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
				"after":         index,
				"before":        before,
				"token":         bc.key,
				"confirmations": confirmations,
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

func (bc *BlockcypherClient) GetBtcBalance(address string) (string, error) {
	var result struct {
		Balance uint64 `json:"balance"`
		Error   string `json:"error"`
	}
	_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
		Url: fmt.Sprintf("%s/addrs/%s/balance", bc.baseUrl, address),
		Params: map[string]interface{}{
			"token": bc.key,
		},
	}, &result)
	if err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf(result.Error)
	}
	return go_decimal.Decimal.Start(result.Balance).MustUnShiftedBy(8).EndForString(), nil
}

type ListUnspentResult struct {
	TxHash        string `json:"tx_hash"`
	TxOutputN     uint64 `json:"tx_output_n"`
	Value         uint64 `json:"value"`
	Confirmations uint64 `json:"confirmations"`
	BlockHeight   uint64 `json:"block_height"`
}

func (bc *BlockcypherClient) ListUnspent(address string) ([]ListUnspentResult, error) {
	results := make([]ListUnspentResult, 0)

	var httpResult struct {
		TxRefs  []ListUnspentResult `json:"txrefs"`
		HasMore bool                `json:"hasMore"`
		Error   string              `json:"error"`
	}
	_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
		Url: fmt.Sprintf("%s/addrs/%s", bc.baseUrl, address),
		Params: map[string]interface{}{
			"after":       0,
			"unspentOnly": true,
			"token":       bc.key,
		},
	}, &httpResult)
	if err != nil {
		return nil, err
	}
	if httpResult.Error != "" {
		return nil, fmt.Errorf(httpResult.Error)
	}

	results = append(results, httpResult.TxRefs...)
	if !httpResult.HasMore {
		return results, nil
	}

	before := httpResult.TxRefs[len(httpResult.TxRefs)-1].BlockHeight
	for {
		var httpResult struct {
			TxRefs  []ListUnspentResult `json:"txrefs"`
			HasMore bool                `json:"hasMore"`
			Error   string              `json:"error"`
		}
		_, err := go_http.NewHttpRequester(go_http.WithTimeout(bc.timeout), go_http.WithLogger(bc.logger)).GetForStruct(go_http.RequestParam{
			Url: fmt.Sprintf("%s/addrs/%s", bc.baseUrl, address),
			Params: map[string]interface{}{
				"after":       0,
				"unspentOnly": true,
				"before":      before,
				"token":       bc.key,
			},
		}, &httpResult)
		if err != nil {
			return nil, err
		}
		if httpResult.Error != "" {
			return nil, fmt.Errorf(httpResult.Error)
		}
		results = append(results, httpResult.TxRefs...)
		if !httpResult.HasMore {
			break
		}
		before = httpResult.TxRefs[len(httpResult.TxRefs)-1].BlockHeight
	}
	return results, nil
}
