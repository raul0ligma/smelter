package provider

import (
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

const (
	MethodCodeAt      = "eth_getCode"
	MethodNonceAt     = "eth_getTransactionCount"
	MethodBalanceAt   = "eth_getBalance"
	MethodBlockNumber = "eth_blockNumber"
)

func Func2RpcMethod(in any) string {
	fnName := runtime.FuncForPC(reflect.ValueOf(in).Pointer()).Name()
	parts := strings.Split(fnName, ".")
	methodName := parts[len(parts)-1]

	if idx := strings.Index(methodName, "-"); idx > 0 {
		methodName = methodName[:idx]
	}

	r := []rune(methodName)
	r[0] = unicode.ToLower(r[0])
	return "eth_" + string(r)
}
