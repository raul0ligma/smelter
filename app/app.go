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

	"github.com/rahul0tripathi/go-jsonrpc"
	"github.com/rahul0tripathi/smelter/controller"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
	"github.com/rahul0tripathi/smelter/provider"
	"github.com/rahul0tripathi/smelter/services"
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

	rpcServer := jsonrpc.NewServer(
		jsonrpc.WithNamespaceSeparator("_"),
		jsonrpc.WithMethodTransformer(func(s string) string {
			r := []rune(s)
			r[0] = unicode.ToLower(r[0])
			return string(r)
		}),
	)

	rpcServer.Register("eth", ethRpcService)
	rpcServer.Register("smelter", smelterRpcService)

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
