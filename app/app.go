package app

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/raul0ligma/smelter/controller"
	"github.com/raul0ligma/smelter/entity"
	"github.com/raul0ligma/smelter/pkg/log"
	"github.com/raul0ligma/smelter/pkg/server"
	"github.com/raul0ligma/smelter/provider"
	"github.com/raul0ligma/smelter/services"
	"go.uber.org/zap"
)

func Run(
	ctx context.Context,
	rpcURL string,
	forkBlock uint64,
	chainID *big.Int,
	stateTTL time.Duration,
	cleanupInterval time.Duration,
	startHook chan<- struct{},
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger, err := log.NewZapLogger(false)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	httpserver := server.New(":6969", logger)

	forkConfig := entity.ForkConfig{
		ChainID:   chainID.Uint64(),
		ForkBlock: new(big.Int).SetUint64(forkBlock),
	}

	stateReader, err := provider.NewJsonRPCProvider(rpcURL)
	if err != nil {
		return fmt.Errorf("state reader error: %w", err)
	}

	storage := services.NewExecutionStorage(forkConfig, stateReader, stateTTL)
	go storage.Watcher(ctx, cleanupInterval)
	ethRpcService := services.NewRpcService(storage, forkConfig, stateReader)
	smelterRpcService := services.NewSmelterRpc(storage)
	otterscanRpcService := services.NewOtterscanRpc(ethRpcService, storage)
	erigonRpcService := services.NewErigonRpc(ethRpcService)

	rpcServer := jsonrpc.NewServer(
		jsonrpc.WithServerMethodNameFormatter(
			func(namespace, method string) string {
				r := []rune(method)
				r[0] = unicode.ToLower(r[0])
				return namespace + "_" + string(r)
			},
		),

	)

	rpcServer.Register("eth", ethRpcService)
	rpcServer.Register("smelter", smelterRpcService)
	rpcServer.Register("ots", otterscanRpcService)
	rpcServer.Register("erigon", erigonRpcService)

	controller.SetupRouter(httpserver.Router(), rpcServer, logger)

	httpserver.Start()

	if startHook != nil {
		select {
		case startHook <- struct{}{}:
		default:
		}
	}

	// Waiting for signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("context canceled")
	case s := <-interrupt:
		logger.Info("signal -> " + s.String())
	case err = <-httpserver.Notify():
		return fmt.Errorf("notify -> %w", err)
	}

	if err := httpserver.Shutdown(); err != nil {
		logger.Error("app::shutdown", zap.Error(err))
	}

	return nil
}
