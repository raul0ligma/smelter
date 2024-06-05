package tests

import (
	"context"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-resty/resty/v2"
)

type chainListResponse struct {
	PageProps struct {
		Chain struct {
			Name  string `json:"name"`
			Chain string `json:"chain"`
			Icon  string `json:"icon"`
			Rpc   []struct {
				Url             string `json:"url"`
				Tracking        string `json:"tracking,omitempty"`
				TrackingDetails string `json:"trackingDetails,omitempty"`
				IsOpenSource    bool   `json:"isOpenSource,omitempty"`
			} `json:"rpc"`
			Features []struct {
				Name string `json:"name"`
			} `json:"features"`
			Faucets        []interface{} `json:"faucets"`
			NativeCurrency struct {
				Name     string `json:"name"`
				Symbol   string `json:"symbol"`
				Decimals int    `json:"decimals"`
			} `json:"nativeCurrency"`
			InfoURL   string `json:"infoURL"`
			ShortName string `json:"shortName"`
			ChainId   int    `json:"chainId"`
			NetworkId int    `json:"networkId"`
			Slip44    int    `json:"slip44"`
			Ens       struct {
				Registry string `json:"registry"`
			} `json:"ens"`
			Explorers []interface{} `json:"explorers"`
			Tvl       float64       `json:"tvl"`
			ChainSlug string        `json:"chainSlug"`
		} `json:"chain"`
	} `json:"pageProps"`
	NSSG bool `json:"__N_SSG"`
}

func GetRPClient(ctx context.Context, chainID uint64) (string, error) {
	resp := &chainListResponse{}
	_, err := resty.New().R().SetContext(ctx).SetResult(resp).Get(fmt.Sprintf("https://chainlist.org/_next/data/2a6bi8IKX51_EWouG4vWk/chain/%d.json?chain=%d", chainID, chainID))
	if err != nil {
		return "", err
	}

	fmt.Println(resp)

	for _, rpc := range resp.PageProps.Chain.Rpc {
		_, err := ethclient.DialContext(ctx, rpc.Url)
		if err != nil {
			continue
		}

		return rpc.Url, nil
	}

	return "", errors.New("no provider found")
}
