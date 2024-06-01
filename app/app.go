package app

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"unicode"

	"github.com/rahul0tripathi/go-jsonrpc"
	"github.com/rahul0tripathi/smelter/config"
	"github.com/rahul0tripathi/smelter/controller"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/executor"
	"github.com/rahul0tripathi/smelter/fork"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
	"github.com/rahul0tripathi/smelter/provider"
	"github.com/rahul0tripathi/smelter/services"
	"go.uber.org/zap"
)

func Run(ctx context.Context, rpcURL string, forkBlock uint64, chainID *big.Int) error {
	var err error
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger, err := log.NewZapLogger(false)
	if err != nil {
		return fmt.Errorf("failed to create logger, %w", err)
	}

	httpserver := server.New(":6969", logger)

	forkConfig := entity.ForkConfig{
		ChainID:   chainID.Uint64(),
		ForkBlock: new(big.Int).SetUint64(forkBlock),
	}

	stateReader, err := provider.NewJsonRPCProvider(rpcURL)
	if err != nil {
		return fmt.Errorf("state reader err %w", err)
	}

	forkDB := fork.NewDB(stateReader, forkConfig, entity.NewAccountsStorage(), entity.NewAccountsState())
	cfg := config.NewConfigWithDefaults()
	cfg.ForkConfig = &forkConfig

	exec, err := executor.NewExecutor(ctx, cfg, forkDB, stateReader)
	if err != nil {
		return fmt.Errorf("new executor err %w", err)
	}

	rpcService := services.NewRpcService(exec, forkDB, &forkConfig, stateReader)

	rpcServer := jsonrpc.NewServer(jsonrpc.WithNamespaceSeparator("_"), jsonrpc.WithMethodTransformer(func(s string) string {
		r := []rune(s)
		r[0] = unicode.ToLower(r[0])
		return string(r)
	}))

	rpcServer.Register("eth", rpcService)

	controller.SetupRouter(httpserver.Router(), rpcServer, logger)

	httpserver.Start()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		logger.Info("M::context canceled")
	case s := <-interrupt:
		logger.Info("M::signal -> " + s.String())
	case err = <-httpserver.Notify():
		return fmt.Errorf("M::notify ->, %w", err)
	}

	err = httpserver.Shutdown()
	if err != nil {
		logger.Error("APP::shutdown, %s", zap.Error(err))
	}

	return nil

}
