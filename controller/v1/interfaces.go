package v1

import (
	"context"

	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/pkg/log"
)

type Rpc interface {
	HandleRPCRequest(
		ctx context.Context,
		logger log.Logger,
		message *entity.JsonrpcMessage,
	) *entity.JsonrpcMessage
}
