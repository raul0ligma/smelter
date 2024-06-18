
# SMELTER 🛠
A local Ethereum node written in go powered by geth, supports multiple forks and simulation, with support for otterscan block explorer

### Installation
```bash
git clone github.com/rahul0tripathi/smelter
cd smelter
go run cmd/main.go --rpcURL https://eth.llamarpc.com --stateTTL 5m --cleanupInterval 3m
```

### Request
```bash
curl --request POST \
  --url http://localhost:6969/v1/rpc/:key \
  --header 'Content-Type: application/json' \
  --data '{
  "method": "eth_getBlockByNumber",
  "params": [
    "0x132fba9",
    false
  ],
  "id": 6,
  "jsonrpc": "2.0"
}'
```

> The key param is used to assign and manage the fork state, each key identifies a state which is cleared after --stateTTL value (default 10m)

```bash


	███████╗███╗   ███╗███████╗██╗  ████████╗███████╗██████╗ 
	██╔════╝████╗ ████║██╔════╝██║  ╚══██╔══╝██╔════╝██╔══██╗
	███████╗██╔████╔██║█████╗  ██║     ██║   █████╗  ██████╔╝
	╚════██║██║╚██╔╝██║██╔══╝  ██║     ██║   ██╔══╝  ██╔══██╗
	███████║██║ ╚═╝ ██║███████╗███████╗██║   ███████╗██║  ██║
	╚══════╝╚═╝     ╚═╝╚══════╝╚══════╝╚═╝   ╚══════╝╚═╝  ╚═╝

	============================================================
	RPC_URL		https://eth.llamarpc.com
	CHAIN_ID	1
	FORK_BLOCK	20120497
	============================================================

⇨ http server started on [::]:6969
2024-06-19T00:11:58.048+0530	DEBUG	services/eth_rpc.go:220	Called SendRawTransaction	{"encoded": "0xe8018082753094c02aaa39b223fe8d0a0e5c4f27ead9083c756cc285e8d4a5100084d0e30db0808080"}
2024-06-19T00:12:02.206+0530	DEBUG	services/eth_rpc.go:257	trace

 [CALL] 0x0000000000000000000000000000000000000069 => 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2 [0xe8d4a51000] (0xd0e30db0)
 [RETURN] 0x0000000000000000000000000000000000000069 => 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2 [] (0x (26074) ERR: (<nil>) REVERTED: false)

2024-06-19T00:12:02.207+0530	DEBUG	services/eth_rpc.go:283	Called GetTransactionByHash	{"txHash": "0x77846b3841fded7bcd150910a14fa6402105b0166c8323f6e29f105a6682e322"}
2024-06-19T00:12:02.207+0530	DEBUG	services/eth_rpc.go:267	Called GetTransactionReceipt	{"txHash": "0x77846b3841fded7bcd150910a14fa6402105b0166c8323f6e29f105a6682e322"}
2024-06-19T00:12:02.208+0530	DEBUG	services/eth_rpc.go:267	Called GetTransactionReceipt	{"txHash": "0x77846b3841fded7bcd150910a14fa6402105b0166c8323f6e29f105a6682e322"}
2024-06-19T00:12:02.208+0530	DEBUG	services/eth_rpc.go:174	Called Call	{"msg": {"From":"0x0000000000000000000000000000000000000000","To":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","Gas":0,"GasPrice":"","GasFeeCap":"","GasTipCap":"","Value":"","Input":"0x70a082310000000000000000000000000000000000000000000000000000000000000069","Data":""}, "blockNumber": "0x1330386"}
2024-06-19T00:12:02.564+0530	DEBUG	services/eth_rpc.go:337	Called GetBalance	{"account": "0x0000000000000000000000000000000000000069", "blockNumber": "0x1330386"}
2024-06-19T00:12:02.884+0530	DEBUG	services/eth_rpc.go:174	Called Call	{"msg": {"From":"0x0000000000000000000000000000000000000000","To":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","Gas":0,"GasPrice":"","GasFeeCap":"","GasTipCap":"","Value":"","Input":"0x70a082310000000000000000000000000000000000000000000000000000000000000069","Data":""}, "blockNumber": "latest"}
2024-06-19T00:12:03.932+0530	DEBUG	services/eth_rpc.go:337	Called GetBalance	{"account": "0x0000000000000000000000000000000000000069", "blockNumber": "latest"}
2024-06-19T00:12:03.935+0530	DEBUG	services/eth_rpc.go:391	Called SetBalance	{"account": "0x0000000000000000000000000000000000000069", "balance": "5000000"}
2024-06-19T00:12:03.936+0530	DEBUG	services/eth_rpc.go:220	Called SendRawTransaction	{"encoded": "0xf8650280834c4b4094c02aaa39b223fe8d0a0e5c4f27ead9083c756cc280b844a9059cbb00000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000064808080"}
2024-06-19T00:12:04.617+0530	DEBUG	services/eth_rpc.go:257	trace

 [CALL] 0x0000000000000000000000000000000000000069 => 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2 [0x0] (0xa9059cbb00000000000000000000000000000000000000000000000000000000000000070000000000000000000000000000000000000000000000000000000000000064)
 [RETURN] 0x0000000000000000000000000000000000000069 => 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2 [] (0x0000000000000000000000000000000000000000000000000000000000000001 (33362) ERR: (<nil>) REVERTED: false)

2024-06-19T00:12:04.617+0530	DEBUG	services/eth_rpc.go:283	Called GetTransactionByHash	{"txHash": "0xd84e3a421f8792a48bc168f9375fb994ab4a94be4e6d21220c1e59febf192909"}
2024-06-19T00:12:04.617+0530	DEBUG	services/eth_rpc.go:267	Called GetTransactionReceipt	{"txHash": "0xd84e3a421f8792a48bc168f9375fb994ab4a94be4e6d21220c1e59febf192909"}
2024-06-19T00:12:04.618+0530	DEBUG	services/eth_rpc.go:267	Called GetTransactionReceipt	{"txHash": "0xd84e3a421f8792a48bc168f9375fb994ab4a94be4e6d21220c1e59febf192909"}
2024-06-19T00:12:04.618+0530	DEBUG	services/eth_rpc.go:174	Called Call	{"msg": {"From":"0x0000000000000000000000000000000000000000","To":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","Gas":0,"GasPrice":"","GasFeeCap":"","GasTipCap":"","Value":"","Input":"0x70a082310000000000000000000000000000000000000000000000000000000000000007","Data":""}, "blockNumber": "0x1330387"}
2024-06-19T00:12:05.016+0530	DEBUG	services/eth_rpc.go:174	Called Call	{"msg": {"From":"0x0000000000000000000000000000000000000000","To":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","Gas":0,"GasPrice":"","GasFeeCap":"","GasTipCap":"","Value":"","Input":"0x70a082310000000000000000000000000000000000000000000000000000000000000007","Data":""}, "blockNumber": "latest"}
2024-06-19T00:12:05.018+0530	DEBUG	services/eth_rpc.go:174	Called Call	{"msg": {"From":"0x0000000000000000000000000000000000000000","To":"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2","Gas":0,"GasPrice":"","GasFeeCap":"","GasTipCap":"","Value":"","Input":"0x70a082310000000000000000000000000000000000000000000000000000000000000069","Data":""}, "blockNumber": "latest"}




```

## Supported Methods

### ETH JSON RPC Namespace
| Method Name |   
|---------------------------------|
| eth_chainId  | 
| eth_blockNumber | 
| eth_getBlockByHash |
| eth_getStorageAt | 
| eth_getHeaderByHash |
| eth_getHeaderByNumber |
| eth_call |  
| eth_sendRawTransaction |
| eth_getTransactionReceipt | 
| eth_getTransactionByHash | 
| eth_estimateGas | 
| eth_gasPrice | 
| eth_getBlockByNumber | 
| eth_getBalance | 
| eth_getCode | 
| eth_setBalance | 
| eth_getTransactionCount | 

- [ETH JSON RPC Spec](https://ethereum.github.io/execution-apis/api-documentation/)

### OTTERSCAN Namespace
| Method Name |   
|---------------------------------|
| ots_getApiLevel |
| ots_hasCode |
| ots_getContractCreator |
| ots_searchTransactionsBefore |
| ots_getBlockDetails |
| ots_getTransactionError |
| ots_getBlockTransactions |
| ots_traceTransaction | 

- [OTTERSCAN RPC Spec](https://github.com/otterscan/otterscan/blob/develop/docs/custom-jsonrpc.md)

### SMELTER Namespace
| Method Name | Description |
|--------------------------------------|--------------------------------------------------|
| `smelter_impersonateAccount` | Impersonates an account with the given address.all further executions are executed with this as sender |
| `smelter_stopImpersonatingAccount` | Stops impersonating the current account. |
| `smelter_getState` | Retrieves the current state as a JSON message. |
| `smelter_setStateOverrides` | Sets state overrides with the provided values. all further executions are executed with these values |

## Demo

https://github.com/rahul0tripathi/smelter/assets/48456755/43de1c0d-7ac2-43fd-92f7-4a829e2b8b57

<img width="1796" alt="smelter" src="https://github.com/rahul0tripathi/smelter/assets/48456755/ea8852b5-5e3d-4f61-9a34-955929815c08">
<img width="1796" alt="otterscan" src="https://github.com/rahul0tripathi/smelter/assets/48456755/daeafac8-8e13-4662-9528-4bb0ebbf8dbe">

### Refrences
- [https://github.com/ethereum/go-ethereum](https://github.com/ethereum/go-ethereum)
- [https://github.com/foundry-rs/foundry/blob/master/crates/anvil](https://github.com/foundry-rs/foundry/blob/master/crates/anvil/)
