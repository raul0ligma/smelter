package server

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
)

const (
	_headerCaller = "X-Caller"
	_paramKey     = "key"
)

type Caller struct{}
type Key struct{}

func SetResponseHeaderMw(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/json")
		return next(c)
	}
}

func SetCallerContextMw(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		caller := c.Request().Header.Get(_headerCaller)
		if common.IsHexAddress(caller) {
			ctx := context.WithValue(c.Request().Context(), Caller{}, common.HexToAddress(caller))
			c.SetRequest(c.Request().WithContext(ctx))
		}

		return next(c)
	}
}

func SetExecutionContextMw(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		caller := c.Param(_paramKey)
		if caller != "" {
			ctx := context.WithValue(c.Request().Context(), Key{}, caller)
			c.SetRequest(c.Request().WithContext(ctx))
		}

		return next(c)
	}
}
