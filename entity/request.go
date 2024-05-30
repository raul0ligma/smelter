package entity

import (
	"encoding/json"
	"fmt"
)

const (
	_genericRpcErrorCode = -3201
	_genericRpcError     = "failed to handle request, internal server error"
)

type JsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type JsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *JsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func JsonMessageWithError(request *JsonrpcMessage, err error, data interface{}) *JsonrpcMessage {
	return &JsonrpcMessage{
		Version: request.Version,
		ID:      request.ID,
		Error: &JsonError{
			Code:    _genericRpcErrorCode,
			Message: fmt.Errorf(_genericRpcError, err).Error(),
			Data:    data,
		},
	}
}

func EmptyJsonMessage(request *JsonrpcMessage) *JsonrpcMessage {
	return &JsonrpcMessage{
		Version: request.Version,
		ID:      request.ID,
	}
}

type DebugData struct {
	request *JsonrpcMessage
	Method  string
}

func NewDeubgData(request *JsonrpcMessage, method string) *DebugData {
	return &DebugData{
		request: request,
		Method:  method,
	}
}
