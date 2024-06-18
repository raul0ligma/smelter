package utils

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

func Big2Hex(v *big.Int) string {
	if v == nil || v.String() == "0" {
		return "0x0"
	}

	return hexutil.Encode(v.Bytes())
}

func ToEvenLength(hex string) string {
	if len(hex) < 2 || hex[:2] != "0x" {
		return hex
	}

	if len(hex[2:])%2 != 0 {
		return "0x0" + hex[2:]
	}

	return hex
}
