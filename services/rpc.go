package services

import (
	"context"

	"github.com/rahul0tripathi/smelter/entity"
	"github.com/rahul0tripathi/smelter/pkg/log"
	"go.uber.org/zap"
)

type Rpc struct {
}

func (r *Rpc) HandleRPCRequest(
	ctx context.Context,
	logger log.Logger,
	message *entity.JsonrpcMessage,
) *entity.JsonrpcMessage {
	logger.Info("message received", zap.Any("message", message))
	return entity.EmptyJsonMessage(message)
}
