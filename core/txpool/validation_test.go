// Copyright 2026 The core-geth Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package txpool

import (
	"bytes"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
)

var (
	valTestKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	valTestAddr   = crypto.PubkeyToAddress(valTestKey.PublicKey)
)

// newValidationConfig returns a minimal CoreGethChainConfig with Olympia EIPs
// activated at the given block (or nil to leave inactive).
func newValidationConfig(forkBlock *big.Int) *coregeth.CoreGethChainConfig {
	return &coregeth.CoreGethChainConfig{
		NetworkID:     1337,
		Ethash:        new(ctypes.EthashConfig),
		ChainID:       big.NewInt(1337),
		EIP2FBlock:    big.NewInt(0),
		EIP7FBlock:    big.NewInt(0),
		EIP150Block:   big.NewInt(0),
		EIP155Block:   big.NewInt(0),
		EIP160FBlock:  big.NewInt(0),
		EIP161FBlock:  big.NewInt(0),
		EIP170FBlock:  big.NewInt(0),
		EIP140FBlock:  big.NewInt(0),
		EIP658FBlock:  big.NewInt(0),
		EIP2565FBlock: big.NewInt(0), // Berlin (needed for non-legacy tx types)
		EIP2718FBlock: big.NewInt(0),
		EIP2929FBlock: big.NewInt(0),
		EIP2930FBlock: big.NewInt(0),
		EIP1559FBlock: forkBlock,
		EIP3198FBlock: forkBlock,
		EIP7702FBlock: forkBlock,
		EIP7825FBlock: forkBlock,
		EIP7623FBlock: forkBlock,
		EIP2028FBlock: big.NewInt(0),
	}
}

func validationOpts(config *coregeth.CoreGethChainConfig) *ValidationOptions {
	return &ValidationOptions{
		Config:  config,
		Accept:  1<<types.LegacyTxType | 1<<types.AccessListTxType | 1<<types.DynamicFeeTxType,
		MaxSize: 128 * 1024,
		MinTip:  big.NewInt(0),
	}
}

func testHeader(blockNum int64, gasLimit uint64) *types.Header {
	return &types.Header{
		Number:   big.NewInt(blockNum),
		GasLimit: gasLimit,
		BaseFee:  big.NewInt(vars.InitialBaseFee),
	}
}

// TestEIP7825GasCapRejectsOverLimit verifies that ValidateTransaction rejects
// transactions with gas > 30M when EIP-7825 is active.
func TestEIP7825GasCapRejectsOverLimit(t *testing.T) {
	config := newValidationConfig(big.NewInt(0))
	signer := types.MakeSigner(config, big.NewInt(0), 0)
	head := testHeader(0, 60_000_000)
	opts := validationOpts(config)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Value:    big.NewInt(0),
		Gas:      vars.MaxTxGas + 1, // 30_000_001
		GasPrice: big.NewInt(1e10),
	})

	err := ValidateTransaction(tx, head, signer, opts)
	if err == nil {
		t.Fatal("expected error for gas > MaxTxGas, got nil")
	}
	if !errors.Is(err, core.ErrGasLimitTooHigh) {
		t.Fatalf("expected ErrGasLimitTooHigh, got: %v", err)
	}
}

// TestEIP7825GasCapAllowsAtLimit verifies that transactions with gas exactly
// at or below 2^24 (16,777,216) are accepted when EIP-7825 is active.
func TestEIP7825GasCapAllowsAtLimit(t *testing.T) {
	config := newValidationConfig(big.NewInt(0))
	signer := types.MakeSigner(config, big.NewInt(0), 0)
	head := testHeader(0, 60_000_000)
	opts := validationOpts(config)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Value:    big.NewInt(0),
		Gas:      vars.MaxTxGas, // exactly 2^24 = 16,777,216
		GasPrice: big.NewInt(1e10),
	})

	err := ValidateTransaction(tx, head, signer, opts)
	if err != nil {
		t.Fatalf("expected tx at MaxTxGas to be accepted, got: %v", err)
	}
}

// TestEIP7825GasCapInactivePreFork verifies that transactions with gas > 2^24
// are accepted when EIP-7825 is NOT active.
func TestEIP7825GasCapInactivePreFork(t *testing.T) {
	config := newValidationConfig(nil) // no fork
	signer := types.MakeSigner(config, big.NewInt(0), 0)
	head := testHeader(0, 60_000_000)
	opts := validationOpts(config)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Value:    big.NewInt(0),
		Gas:      vars.MaxTxGas + 1_000_000, // 31M, above cap
		GasPrice: big.NewInt(1e10),
	})

	err := ValidateTransaction(tx, head, signer, opts)
	if err != nil {
		t.Fatalf("expected tx above MaxTxGas to be accepted pre-fork, got: %v", err)
	}
}

// TestEIP7623FloorDataGasRejectsLowGas verifies that transactions with
// insufficient gas for the calldata floor cost are rejected when EIP-7623 is active.
func TestEIP7623FloorDataGasRejectsLowGas(t *testing.T) {
	config := newValidationConfig(big.NewInt(0))
	signer := types.MakeSigner(config, big.NewInt(0), 0)
	head := testHeader(0, 60_000_000)
	opts := validationOpts(config)

	// 1000 non-zero bytes of calldata
	// Floor = TxGas + (nz * TxTokenPerNonZeroByte + z) * TxCostFloorPerToken
	//       = 21000 + (1000 * 4 + 0) * 10 = 21000 + 40000 = 61000
	data := bytes.Repeat([]byte{0xff}, 1000)
	floorGas, _ := core.FloorDataGas(data)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Value:    big.NewInt(0),
		Gas:      floorGas - 1, // just below floor
		Data:     data,
		GasPrice: big.NewInt(1e10),
	})

	err := ValidateTransaction(tx, head, signer, opts)
	if err == nil {
		t.Fatal("expected error for gas below floor data gas, got nil")
	}
	if !errors.Is(err, core.ErrFloorDataGas) {
		t.Fatalf("expected ErrFloorDataGas, got: %v", err)
	}
}

// TestEIP7623FloorDataGasAllowsSufficientGas verifies that transactions with
// enough gas for the calldata floor cost are accepted.
func TestEIP7623FloorDataGasAllowsSufficientGas(t *testing.T) {
	config := newValidationConfig(big.NewInt(0))
	signer := types.MakeSigner(config, big.NewInt(0), 0)
	head := testHeader(0, 60_000_000)
	opts := validationOpts(config)

	data := bytes.Repeat([]byte{0xff}, 1000)
	floorGas, _ := core.FloorDataGas(data)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Value:    big.NewInt(0),
		Gas:      floorGas, // exactly at floor
		Data:     data,
		GasPrice: big.NewInt(1e10),
	})

	err := ValidateTransaction(tx, head, signer, opts)
	if err != nil {
		t.Fatalf("expected tx at floor gas to be accepted, got: %v", err)
	}
}

// TestEIP7623FloorDataGasInactivePreFork verifies that calldata-heavy transactions
// with gas below the floor are accepted when EIP-7623 is NOT active.
func TestEIP7623FloorDataGasInactivePreFork(t *testing.T) {
	config := newValidationConfig(nil) // no fork
	signer := types.MakeSigner(config, big.NewInt(0), 0)
	head := testHeader(0, 60_000_000)
	opts := validationOpts(config)

	// 1000 non-zero bytes, floor would be 61000 if active
	data := bytes.Repeat([]byte{0xff}, 1000)

	// Intrinsic gas = 21000 + 1000*16 = 37000 (below floor of 61000)
	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Value:    big.NewInt(0),
		Gas:      37000, // covers intrinsic but not floor
		Data:     data,
		GasPrice: big.NewInt(1e10),
	})

	err := ValidateTransaction(tx, head, signer, opts)
	if err != nil {
		t.Fatalf("expected tx below floor gas to be accepted pre-fork, got: %v", err)
	}
}
