package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rahul0tripathi/go-jsonrpc"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
)

func SetupRouter(
	router server.Router,
	rpcServer *jsonrpc.RPCServer,
	logger log.Logger,
) {
	router.POST("/v1/rpc/:key", echo.WrapHandler(rpcServer),
		middleware.CORSWithConfig(middleware.DefaultCORSConfig),
		server.SetExecutionContextMw,
		server.SetCallerContextMw,
		server.SetResponseHeaderMw,
	)
}
