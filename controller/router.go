package controller

import (
	"github.com/filecoin-project/go-jsonrpc"
	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
)

func SetupRouter(
	router server.Router,
	rpcServer *jsonrpc.RPCServer,
	logger log.Logger,
) {

	router.POST(
		"/v1/rpc/:key", echo.WrapHandler(rpcServer),
		server.SetExecutionContextMw,
		server.SetCallerContextMw,
		server.SetResponseHeaderMw,
	)
}
