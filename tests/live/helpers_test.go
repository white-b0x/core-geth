//go:build live

package live

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

// Mordor Olympia fork parameters
const (
	MordorChainID       = 63
	ETCMainnetChainID   = 61
	MordorForkBlock     = 15800850
	TreasuryAddressHex  = "0xCfE1e0ECbff745e6c800fF980178a8dDEf94bEe2"
	InitialBaseFeeGwei  = 1_000_000_000 // 1 gwei
	BlockRLPSizeCap     = 8_388_608
	TxGasLimitCap       = 16_777_216 // 2^24, EIP-7825 final spec
	EIP2935ContractAddr = "0x0000000000000000000000000000000000000F0A"
	P256VerifyAddr      = "0x0000000000000000000000000000000000000100"
)

// BLS12-381 precompile addresses (EIP-2537 final: 7 precompiles at 0x0b-0x11)
var BLSPrecompileAddrs = []string{
	"0x000000000000000000000000000000000000000b", // G1Add
	"0x000000000000000000000000000000000000000c", // G1MultiExp (covers G1Mul at k=1)
	"0x000000000000000000000000000000000000000d", // G2Add
	"0x000000000000000000000000000000000000000e", // G2MultiExp (covers G2Mul at k=1)
	"0x000000000000000000000000000000000000000f", // Pairing
	"0x0000000000000000000000000000000000000010", // MapG1
	"0x0000000000000000000000000000000000000011", // MapG2
}

// Known Mordor genesis block hash
var MordorGenesisHash = common.HexToHash("0xa68ebde7932f0bf2579b075499416f0a693de84c26b05cd01de86e60aad05ec0")

// Known ETC mainnet genesis block hash
var ETCGenesisHash = common.HexToHash("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3")

// ECIP-1017 era length (blocks per era)
const ECIP1017EraLength = 5_000_000

// ECIP-1099 epoch calibration fork block on Mordor
const MordorECIP1099Block = 2_520_000

// rpcBlock is a minimal block structure for JSON-RPC responses.
type rpcBlock struct {
	Number       *hexutil.Big      `json:"number"`
	Hash         common.Hash       `json:"hash"`
	ParentHash   common.Hash       `json:"parentHash"`
	GasUsed      *hexutil.Big      `json:"gasUsed"`
	GasLimit     *hexutil.Big      `json:"gasLimit"`
	BaseFee      *hexutil.Big      `json:"baseFeePerGas"`
	StateRoot    common.Hash       `json:"stateRoot"`
	ReceiptsRoot common.Hash       `json:"receiptsRoot"`
	Size         *hexutil.Big      `json:"size"`
	Miner        common.Address    `json:"miner"`
	Difficulty   *hexutil.Big      `json:"difficulty"`
	Nonce        string            `json:"nonce"`
	MixHash      common.Hash       `json:"mixHash"`
	Timestamp    *hexutil.Big      `json:"timestamp"`
	Uncles       []common.Hash     `json:"uncles"`
	Transactions []json.RawMessage `json:"transactions"`
}

// rpcTx is a minimal transaction structure for JSON-RPC responses.
type rpcTx struct {
	Hash     common.Hash     `json:"hash"`
	Type     hexutil.Uint64  `json:"type"`
	GasPrice *hexutil.Big    `json:"gasPrice"`
	Gas      *hexutil.Big    `json:"gas"`
	To       *common.Address `json:"to"`
	Value    *hexutil.Big    `json:"value"`
	// EIP-1559 fields
	MaxFeePerGas         *hexutil.Big `json:"maxFeePerGas"`
	MaxPriorityFeePerGas *hexutil.Big `json:"maxPriorityFeePerGas"`
	// EIP-7702 fields
	AuthorizationList json.RawMessage `json:"authorizationList"`
}

// getMordorRPC returns the Mordor RPC URL from env or default.
func getMordorRPC() string {
	if url := os.Getenv("MORDOR_RPC"); url != "" {
		return url
	}
	return "http://localhost:8545"
}

// getETCRPC returns the ETC mainnet RPC URL from env or default.
func getETCRPC() string {
	if url := os.Getenv("ETC_RPC"); url != "" {
		return url
	}
	return "https://etc.rivet.link"
}

// dialRPC creates an RPC client with timeout.
func dialRPC(t *testing.T, url string) *rpc.Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := rpc.DialContext(ctx, url)
	if err != nil {
		t.Skipf("cannot connect to RPC at %s: %v", url, err)
	}
	return client
}

// getChainID fetches the chain ID via eth_chainId.
func getChainID(t *testing.T, client *rpc.Client) uint64 {
	t.Helper()
	var result hexutil.Uint64
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &result, "eth_chainId"); err != nil {
		t.Fatalf("eth_chainId failed: %v", err)
	}
	return uint64(result)
}

// getBlockNumber fetches the latest block number.
func getBlockNumber(t *testing.T, client *rpc.Client) uint64 {
	t.Helper()
	var result hexutil.Uint64
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &result, "eth_blockNumber"); err != nil {
		t.Fatalf("eth_blockNumber failed: %v", err)
	}
	return uint64(result)
}

// getBlockByNumber fetches a block by number. fullTxs controls whether transactions are full objects.
func getBlockByNumber(t *testing.T, client *rpc.Client, number uint64, fullTxs bool) *rpcBlock {
	t.Helper()
	var block rpcBlock
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &block, "eth_getBlockByNumber", hexutil.EncodeUint64(number), fullTxs); err != nil {
		t.Fatalf("eth_getBlockByNumber(%d) failed: %v", number, err)
	}
	return &block
}

// getLatestBlock fetches the latest block.
func getLatestBlock(t *testing.T, client *rpc.Client) *rpcBlock {
	t.Helper()
	var block rpcBlock
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &block, "eth_getBlockByNumber", "latest", false); err != nil {
		t.Fatalf("eth_getBlockByNumber(latest) failed: %v", err)
	}
	return &block
}

// getBalance fetches the balance of an address at the given block.
func getBalance(t *testing.T, client *rpc.Client, addr common.Address, blockTag string) *big.Int {
	t.Helper()
	var result hexutil.Big
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &result, "eth_getBalance", addr, blockTag); err != nil {
		t.Fatalf("eth_getBalance(%s, %s) failed: %v", addr.Hex(), blockTag, err)
	}
	return result.ToInt()
}

// getCode fetches the code at an address.
func getCode(t *testing.T, client *rpc.Client, addr common.Address, blockTag string) []byte {
	t.Helper()
	var result hexutil.Bytes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &result, "eth_getCode", addr, blockTag); err != nil {
		t.Fatalf("eth_getCode(%s, %s) failed: %v", addr.Hex(), blockTag, err)
	}
	return result
}

// getStorageAt fetches storage at a given position.
func getStorageAt(t *testing.T, client *rpc.Client, addr common.Address, pos common.Hash, blockTag string) common.Hash {
	t.Helper()
	var result common.Hash
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &result, "eth_getStorageAt", addr, pos, blockTag); err != nil {
		t.Fatalf("eth_getStorageAt(%s, %s, %s) failed: %v", addr.Hex(), pos.Hex(), blockTag, err)
	}
	return result
}

// ethCall performs an eth_call to the given address with the given data.
func ethCall(t *testing.T, client *rpc.Client, to common.Address, data []byte, blockTag string) ([]byte, error) {
	t.Helper()
	type callMsg struct {
		To   common.Address `json:"to"`
		Data hexutil.Bytes  `json:"data"`
		Gas  hexutil.Uint64 `json:"gas"`
	}
	msg := callMsg{To: to, Data: data, Gas: 1_000_000}
	var result hexutil.Bytes
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := client.CallContext(ctx, &result, "eth_call", msg, blockTag)
	return result, err
}

// getTransaction fetches a transaction by hash.
func getTransaction(t *testing.T, client *rpc.Client, hash common.Hash) *rpcTx {
	t.Helper()
	var tx rpcTx
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.CallContext(ctx, &tx, "eth_getTransactionByHash", hash); err != nil {
		t.Fatalf("eth_getTransactionByHash(%s) failed: %v", hash.Hex(), err)
	}
	return &tx
}

// requireForkReached skips the test if the fork block hasn't been reached yet.
func requireForkReached(t *testing.T, client *rpc.Client, forkBlock uint64) {
	t.Helper()
	currentBlock := getBlockNumber(t, client)
	if currentBlock < forkBlock {
		t.Skipf("fork block %d not yet reached (current: %d)", forkBlock, currentBlock)
	}
}

// blockTag returns the hex-encoded block number for use as a block tag.
func blockTag(number uint64) string {
	return fmt.Sprintf("0x%x", number)
}
