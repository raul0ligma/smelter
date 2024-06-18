package tracer

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/rahul0tripathi/smelter/entity"
	internal "github.com/rahul0tripathi/smelter/types"
	"github.com/rahul0tripathi/smelter/utils"
)

type TraceLog struct {
	Type  string `json:"type"`
	Depth uint64 `json:"depth"`
	From  string `json:"from"`
	To    string `json:"to"`
	Text  string `json:"text"`
	Value string `json:"value"`
}

type LogTracer struct {
	Logs         []TraceLog
	logOP        bool
	currentDepth uint64
}

func (l *LogTracer) Fmt() string {
	op := ""
	var calc string
	for _, l := range l.Logs {
		if l.From != "" && l.To != "" {
			calc = fmt.Sprintf("%s => %s", l.From, l.To)
		}

		op += fmt.Sprintf("\n %s[%s] %s [%s] (%s)", strings.Repeat("	", int(l.Depth)), l.Type, calc, l.Value, l.Text)
	}

	return op
}

func (l *LogTracer) OtterTrace() entity.TransactionTraces {
	traces := make(entity.TransactionTraces, 0)

	for _, log := range l.Logs {
		if log.Type == "RETURN" {
			traces = append(traces, entity.TransactionTrace{
				Type:   log.Type,
				Depth:  uint(log.Depth),
				From:   common.HexToAddress("").Hex(),
				To:     common.HexToAddress("").Hex(),
				Value:  "0x00",
				Input:  "0x",
				Output: "0x",
			})
			continue
		}

		traces = append(traces, entity.TransactionTrace{
			Type:   log.Type,
			Depth:  uint(log.Depth),
			From:   log.From,
			To:     log.To,
			Value:  log.Value,
			Input:  log.Text,
			Output: "0x",
		})
	}

	return traces
}

func (l *LogTracer) Hooks() *tracing.Hooks {
	return &tracing.Hooks{
		OnTxStart: func(vm *tracing.VMContext, tx *types.Transaction, from common.Address) {
			l.Logs = append(l.Logs, TraceLog{
				Type:  "TX_START",
				Depth: l.currentDepth,
				From:  from.Hex(),
				To:    tx.To().Hex(),
				Text:  hexutil.Encode(tx.Data()),
				Value: fmt.Sprintf("%+v", tx.Value()),
			})

			l.currentDepth++
		},
		OnTxEnd: func(receipt *types.Receipt, err error) {
			l.currentDepth--
			l.Logs = append(l.Logs, TraceLog{
				Type:  "TX_END",
				Depth: l.currentDepth,
				From:  receipt.ContractAddress.Hex(),
				To:    "RECEIPT",
				Text:  receipt.TxHash.Hex(),
				Value: "",
			})
		},
		OnEnter: func(depth int, op byte, from, to common.Address, input []byte, gas uint64, value *big.Int) {
			l.Logs = append(l.Logs, TraceLog{
				Type:  opCodeToString[op],
				Depth: l.currentDepth,
				From:  from.Hex(),
				To:    to.Hex(),
				Text:  hexutil.Encode(input),
				Value: utils.Big2Hex(value),
			})
			l.currentDepth++
		},
		OnExit: func(depth int, output []byte, gasUsed uint64, err error, reverted bool) {
			l.currentDepth--
			l.Logs = append(l.Logs, TraceLog{
				Type:  "RETURN",
				Depth: l.currentDepth,
				From:  "",
				To:    "",
				Text:  fmt.Sprintf("%s (%d) ERR: (%v) REVERTED: %t", hexutil.Encode(output), gasUsed, err, reverted),
				Value: "",
			})

		},
		OnOpcode: func(
			pc uint64,
			op byte,
			gas, cost uint64,
			scope tracing.OpContext,
			rData []byte,
			depth int,
			err error,
		) {
			if l.logOP {
				l.Logs = append(l.Logs, TraceLog{
					Type:  "OP",
					Depth: l.currentDepth,
					From:  "",
					To:    "",
					Text:  fmt.Sprintf("PC: %d OP: %s GAS: %d COST %d DEPTH: %d rDATA: %s err: %+v", pc, opCodeToString[op], gas, cost, depth, hexutil.Encode(rData), err),
					Value: "",
				})
			}
		},
		OnFault: func(pc uint64, op byte, gas, cost uint64, scope tracing.OpContext, depth int, err error) {
			l.Logs = append(l.Logs, TraceLog{
				Type:  "FAULT:" + opCodeToString[op],
				Depth: l.currentDepth,
				From:  scope.Caller().Hex(),
				To:    scope.Address().Hex(),
				Text:  fmt.Sprintf("INPUT: %s (%d) ERR: (%v) ", hexutil.Encode(scope.CallInput()), scope.CallValue().Uint64(), err),
				Value: "",
			})
		},
		OnGasChange: func(prevGas, newGas uint64, reason tracing.GasChangeReason) {

		},
		OnBlockchainInit: func(chainConfig *params.ChainConfig) {

		},
		OnClose: func() {

		},
		OnBlockStart: func(event tracing.BlockEvent) {

		},
		OnBlockEnd: func(err error) {

		},
		OnSkippedBlock: func(event tracing.BlockEvent) {

		},
		OnGenesisBlock: func(genesis *types.Block, alloc types.GenesisAlloc) {

		},
		OnSystemCallStart: func() {

		},
		OnSystemCallEnd: func() {

		},
		OnBalanceChange: func(addr common.Address, prev, new *big.Int, reason tracing.BalanceChangeReason) {

		},
		OnNonceChange: func(address common.Address, prevNonce, newNonce uint64) {

		},
		OnCodeChange: func(
			addr common.Address,
			prevCodeHash common.Hash,
			prevCode []byte,
			codeHash common.Hash,
			code []byte,
		) {

		},
		OnStorageChange: func(addr common.Address, slot common.Hash, prev, new common.Hash) {

		},
		OnLog: func(log *types.Log) {
			l.Logs = append(l.Logs, TraceLog{
				Type:  "EMIT",
				Depth: l.currentDepth,
				From:  log.Address.Hex(),
				To:    internal.Address0x.Hex(),
				Text:  fmt.Sprintf("%+v", log),
				Value: "",
			})
		},
	}
}

func NewTracer(logOp bool) *LogTracer {
	return &LogTracer{
		Logs:         make([]TraceLog, 0),
		logOP:        logOp,
		currentDepth: 0,
	}
}
