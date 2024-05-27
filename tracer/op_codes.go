package tracer

import "github.com/ethereum/go-ethereum/core/vm"

var opCodeToString = [256]string{

	vm.ADD:        "ADD",
	vm.MUL:        "MUL",
	vm.SUB:        "SUB",
	vm.STOP:       "STOP",
	vm.DIV:        "DIV",
	vm.SDIV:       "SDIV",
	vm.MOD:        "MOD",
	vm.SMOD:       "SMOD",
	vm.EXP:        "EXP",
	vm.NOT:        "NOT",
	vm.LT:         "LT",
	vm.GT:         "GT",
	vm.SLT:        "SLT",
	vm.SGT:        "SGT",
	vm.EQ:         "EQ",
	vm.ISZERO:     "ISZERO",
	vm.SIGNEXTEND: "SIGNEXTEND",

	vm.AND:    "AND",
	vm.OR:     "OR",
	vm.XOR:    "XOR",
	vm.BYTE:   "BYTE",
	vm.SHL:    "SHL",
	vm.SHR:    "SHR",
	vm.SAR:    "SAR",
	vm.ADDMOD: "ADDMOD",
	vm.MULMOD: "MULMOD",

	vm.KECCAK256: "KECCAK256",

	vm.ADDRESS:        "ADDRESS",
	vm.BALANCE:        "BALANCE",
	vm.ORIGIN:         "ORIGIN",
	vm.CALLER:         "CALLER",
	vm.CALLVALUE:      "CALLVALUE",
	vm.CALLDATALOAD:   "CALLDATALOAD",
	vm.CALLDATASIZE:   "CALLDATASIZE",
	vm.CALLDATACOPY:   "CALLDATACOPY",
	vm.CODESIZE:       "CODESIZE",
	vm.CODECOPY:       "CODECOPY",
	vm.GASPRICE:       "GASPRICE",
	vm.EXTCODESIZE:    "EXTCODESIZE",
	vm.EXTCODECOPY:    "EXTCODECOPY",
	vm.RETURNDATASIZE: "RETURNDATASIZE",
	vm.RETURNDATACOPY: "RETURNDATACOPY",
	vm.EXTCODEHASH:    "EXTCODEHASH",

	vm.BLOCKHASH:   "BLOCKHASH",
	vm.COINBASE:    "COINBASE",
	vm.TIMESTAMP:   "TIMESTAMP",
	vm.NUMBER:      "NUMBER",
	vm.DIFFICULTY:  "DIFFICULTY", // TODO (MariusVanDerWijden) rename to PREVRANDAO post merge
	vm.GASLIMIT:    "GASLIMIT",
	vm.CHAINID:     "CHAINID",
	vm.SELFBALANCE: "SELFBALANCE",
	vm.BASEFEE:     "BASEFEE",
	vm.BLOBHASH:    "BLOBHASH",
	vm.BLOBBASEFEE: "BLOBBASEFEE",

	vm.POP:      "POP",
	vm.MLOAD:    "MLOAD",
	vm.MSTORE:   "MSTORE",
	vm.MSTORE8:  "MSTORE8",
	vm.SLOAD:    "SLOAD",
	vm.SSTORE:   "SSTORE",
	vm.JUMP:     "JUMP",
	vm.JUMPI:    "JUMPI",
	vm.PC:       "PC",
	vm.MSIZE:    "MSIZE",
	vm.GAS:      "GAS",
	vm.JUMPDEST: "JUMPDEST",
	vm.TLOAD:    "TLOAD",
	vm.TSTORE:   "TSTORE",
	vm.MCOPY:    "MCOPY",
	vm.PUSH0:    "PUSH0",

	vm.PUSH1:  "PUSH1",
	vm.PUSH2:  "PUSH2",
	vm.PUSH3:  "PUSH3",
	vm.PUSH4:  "PUSH4",
	vm.PUSH5:  "PUSH5",
	vm.PUSH6:  "PUSH6",
	vm.PUSH7:  "PUSH7",
	vm.PUSH8:  "PUSH8",
	vm.PUSH9:  "PUSH9",
	vm.PUSH10: "PUSH10",
	vm.PUSH11: "PUSH11",
	vm.PUSH12: "PUSH12",
	vm.PUSH13: "PUSH13",
	vm.PUSH14: "PUSH14",
	vm.PUSH15: "PUSH15",
	vm.PUSH16: "PUSH16",
	vm.PUSH17: "PUSH17",
	vm.PUSH18: "PUSH18",
	vm.PUSH19: "PUSH19",
	vm.PUSH20: "PUSH20",
	vm.PUSH21: "PUSH21",
	vm.PUSH22: "PUSH22",
	vm.PUSH23: "PUSH23",
	vm.PUSH24: "PUSH24",
	vm.PUSH25: "PUSH25",
	vm.PUSH26: "PUSH26",
	vm.PUSH27: "PUSH27",
	vm.PUSH28: "PUSH28",
	vm.PUSH29: "PUSH29",
	vm.PUSH30: "PUSH30",
	vm.PUSH31: "PUSH31",
	vm.PUSH32: "PUSH32",

	vm.DUP1:  "DUP1",
	vm.DUP2:  "DUP2",
	vm.DUP3:  "DUP3",
	vm.DUP4:  "DUP4",
	vm.DUP5:  "DUP5",
	vm.DUP6:  "DUP6",
	vm.DUP7:  "DUP7",
	vm.DUP8:  "DUP8",
	vm.DUP9:  "DUP9",
	vm.DUP10: "DUP10",
	vm.DUP11: "DUP11",
	vm.DUP12: "DUP12",
	vm.DUP13: "DUP13",
	vm.DUP14: "DUP14",
	vm.DUP15: "DUP15",
	vm.DUP16: "DUP16",

	vm.SWAP1:  "SWAP1",
	vm.SWAP2:  "SWAP2",
	vm.SWAP3:  "SWAP3",
	vm.SWAP4:  "SWAP4",
	vm.SWAP5:  "SWAP5",
	vm.SWAP6:  "SWAP6",
	vm.SWAP7:  "SWAP7",
	vm.SWAP8:  "SWAP8",
	vm.SWAP9:  "SWAP9",
	vm.SWAP10: "SWAP10",
	vm.SWAP11: "SWAP11",
	vm.SWAP12: "SWAP12",
	vm.SWAP13: "SWAP13",
	vm.SWAP14: "SWAP14",
	vm.SWAP15: "SWAP15",
	vm.SWAP16: "SWAP16",

	vm.LOG0: "LOG0",
	vm.LOG1: "LOG1",
	vm.LOG2: "LOG2",
	vm.LOG3: "LOG3",
	vm.LOG4: "LOG4",

	vm.CREATE:       "CREATE",
	vm.CALL:         "CALL",
	vm.RETURN:       "RETURN",
	vm.CALLCODE:     "CALLCODE",
	vm.DELEGATECALL: "DELEGATECALL",
	vm.CREATE2:      "CREATE2",
	vm.STATICCALL:   "STATICCALL",
	vm.REVERT:       "REVERT",
	vm.INVALID:      "INVALID",
	vm.SELFDESTRUCT: "SELFDESTRUCT",
}
