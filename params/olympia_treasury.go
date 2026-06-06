package params

import "github.com/ethereum/go-ethereum/common"

// OlympiaTreasuryAddr is the ECIP-1112 treasury vault address used by both
// Mordor testnet and ETC mainnet chain configs.
//
// v0.3 demo deployment — deterministic address per ECIP-1112 §Treasury Address.
// Matches Besu, Fukuii, and Nethermind reference implementations.
var OlympiaTreasuryAddr = common.HexToAddress("0x60d0A7394f9Cd5C469f9F5Ec4F9C803F5294d79b")
