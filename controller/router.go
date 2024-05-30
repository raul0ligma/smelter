package controller

import (
	v1 "github.com/rahul0tripathi/smelter/controller/v1"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
)

func SetupRouter(
	router server.Router,
	rpcServer v1.Rpc,
	logger log.Logger,
) {
	handler := v1.NewHandler()

	router.POST("/v1/rpc", handler.MakeRpcHandler(logger, rpcServer))
}
