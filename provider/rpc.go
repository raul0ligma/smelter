package provider

import "github.com/ethereum/go-ethereum/ethclient"

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
