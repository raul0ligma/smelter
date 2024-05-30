package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/go-jsonrpc"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"github.com/rahul0tripathi/smelter/pkg/server"
)

func SetupRouter(
	router server.Router,
	rpcServer *jsonrpc.RPCServer,
	logger log.Logger,
) {
	router.POST("/v1/rpc", echo.WrapHandler(rpcServer), func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Content-Type", "application/json")
			return next(c)
		}
	})
}
