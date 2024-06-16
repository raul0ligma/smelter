package main

import (
	"errors"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rahul0tripathi/smelter/app"
	"github.com/rahul0tripathi/smelter/utils"
	clitool "github.com/urfave/cli/v2"
)

func main() {
	var (
		rpcURL          string
		forkBlock       uint64
		stateTTL        time.Duration
		cleanupInterval time.Duration
		chainID         *big.Int
	)

	cli := &clitool.App{
		Name:  "smelter",
		Usage: "run a local node by passing in --rpcURL and --forkBlock",
		Flags: []clitool.Flag{
			&clitool.StringFlag{
				Required:    true,
				Name:        "rpcURL",
				Value:       "http://localhost:8485",
				Usage:       "rpc url of the chain to fork",
				Destination: &rpcURL,
			},
			&clitool.Uint64Flag{
				Name:        "forkBlock",
				Value:       0,
				Usage:       "block number of the chain to create a fork from",
				Destination: &forkBlock,
			},
			&clitool.DurationFlag{
				Name:        "stateTTL",
				Value:       time.Minute * 5,
				Usage:       "TTL for a fork state before it's cleaned",
				Destination: &stateTTL,
			},
			&clitool.DurationFlag{
				Name:        "cleanupInterval",
				Value:       time.Minute * 10,
				Usage:       "periodic interval to check and clean unused fork states",
				Destination: &cleanupInterval,
			},
		},
		Action: func(cCtx *clitool.Context) error {
			if rpcURL == "" {
				return errors.New("invalid rpc url")
			}

			client, err := ethclient.Dial(rpcURL)
			if err != nil {
				return err
			}

			if forkBlock == 0 {
				forkBlock, err = client.BlockNumber(cCtx.Context)
				if err != nil {
					return err
				}
			}

			chainID, err = client.ChainID(cCtx.Context)
			if err != nil {
				return err
			}

			utils.PrintSmelter()
			utils.PrintConfig(rpcURL, chainID, forkBlock)
			return app.Run(cCtx.Context, rpcURL, forkBlock, chainID, stateTTL, cleanupInterval, nil)
		},
	}

	if err := cli.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
