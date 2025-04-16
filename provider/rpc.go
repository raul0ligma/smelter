package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/smelter/entity"
)

type RpcProvider struct {
	*ethclient.Client
}

func NewJsonRPCProvider(rpcURL string) (*RpcProvider, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &RpcProvider{Client: client}, nil
}

func (r *RpcProvider) SupportsBatching() bool {
	return false
}

func (p *RpcProvider) Batch(ctx context.Context, requests []entity.BatchReq) ([]json.RawMessage, error) {
	return nil, errors.New("batching not supported")
}

func (p *RpcProvider) BatchWithUnmarshal(ctx context.Context, requests []entity.BatchReq, outputs []any) error {
	return errors.New("batching not supported")
}

var methodNameMap = map[string]string{
	"CodeAt":      "eth_getCode",
	"BalanceAt":   "eth_getBalance",
	"StorageAt":   "eth_getStorageAt",
	"NonceAt":     "eth_getTransactionCount",
	"BlockNumber": "eth_blockNumber",
}

type BatchRpcProvider struct {
	*ethclient.Client
	rpcURL string
	client *http.Client
}

func NewBatchJSONRPcProvider(rpcURL string) (*BatchRpcProvider, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &BatchRpcProvider{
		Client: client,
		rpcURL: rpcURL,
		client: http.DefaultClient,
	}, nil
}

func (p *BatchRpcProvider) SupportsBatching() bool {
	return true
}

func (p *BatchRpcProvider) Batch(ctx context.Context, requests []entity.BatchReq) ([]json.RawMessage, error) {
	if len(requests) == 0 {
		return nil, nil
	}

	rpcRequests := make([]map[string]any, len(requests))

	for i, req := range requests {

		fnName := runtime.FuncForPC(reflect.ValueOf(req.Method).Pointer()).Name()
		parts := strings.Split(fnName, ".")
		methodName := parts[len(parts)-1]

		if idx := strings.Index(methodName, "-"); idx > 0 {
			methodName = methodName[:idx]
		}

		rpcMethod, ok := methodNameMap[methodName]
		if !ok {
			r := []rune(methodName)
			r[0] = unicode.ToLower(r[0])
			rpcMethod = "eth_" + string(r)
		}

		rpcParams := make([]any, len(req.Params))
		for j, param := range req.Params {
			// check if this is a block number parameter, mostly for all eth methods
			if j == len(req.Params)-1 && methodName != "BlockNumber" {
				if blockNum, ok := param.(*big.Int); ok {
					if blockNum == nil {
						rpcParams[j] = "latest"
					} else {
						rpcParams[j] = hexutil.Encode(blockNum.Bytes())
					}
					continue
				}
			}
			rpcParams[j] = param
		}

		rpcRequests[i] = map[string]any{
			"jsonrpc": "2.0",
			"method":  rpcMethod,
			"params":  rpcParams,
			"id":      i + 1,
		}
	}

	reqBody, err := json.Marshal(rpcRequests)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.rpcURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	var responses []map[string]json.RawMessage
	if err := json.NewDecoder(httpResp.Body).Decode(&responses); err != nil {
		return nil, err
	}

	results := make([]json.RawMessage, len(responses))
	for _, resp := range responses {
		var id int
		if err := json.Unmarshal(resp["id"], &id); err != nil {
			return nil, err
		}

		if errResp, ok := resp["error"]; ok && len(errResp) > 0 {
			var jsonErr struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(errResp, &jsonErr); err == nil {
				return nil, fmt.Errorf("rpc error (request %d): %s", id, jsonErr.Message)
			}
			return nil, fmt.Errorf("rpc error in request %d", id)
		}

		results[id-1] = resp["result"]
	}

	return results, nil
}

func (p *BatchRpcProvider) BatchWithUnmarshal(ctx context.Context, requests []entity.BatchReq, outputs []any) error {
	if len(requests) != len(outputs) {
		return fmt.Errorf("mismatch between requests and outputs count")
	}

	results, err := p.Batch(ctx, requests)
	if err != nil {
		return err
	}

	for i, result := range results {
		if err := json.Unmarshal(result, outputs[i]); err != nil {
			return fmt.Errorf("unmarshal result %d: %w", i, err)
		}
	}

	return nil
}
