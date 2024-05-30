package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/pkg/log"
)

func handleRpcCall(
	c echo.Context,
	logger log.Logger,
	server Rpc,
	raw []byte,
) error {
	request := &entity.JsonrpcMessage{}
	if err := json.Unmarshal(raw, request); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, server.HandleRPCRequest(c.Request().Context(), logger, request))
}

func (h *Handler) MakeRpcHandler(logger log.Logger, server Rpc) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		return handleRpcCall(c, logger, server, raw)
	}

}
