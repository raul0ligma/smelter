package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unicode"

	"github.com/rahul0tripathi/go-jsonrpc"
	"github.com/rahul0tripathi/smelter/controller"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
	"github.com/rahul0tripathi/smelter/services"
	"go.uber.org/zap"
)

func Run() error {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, err := log.NewZapLogger(false)
	if err != nil {
		return fmt.Errorf("failed to create logger, %w", err)
	}

	httpserver := server.New(":6969", logger)

	rpcService := &services.Rpc{}
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
