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
