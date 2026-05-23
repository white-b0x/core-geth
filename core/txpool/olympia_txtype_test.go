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

// Olympia transaction-type admission tests.
// These tests verify that Type-2 (EIP-1559 DynamicFee) and Type-4 (EIP-7702 SetCode)
// transactions are correctly gated by the Olympia fork block.
//
// Helpers (newValidationConfig, validationOpts, testHeader, valTestKey) are defined in
// validation_test.go in the same package.

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// TestType2RejectedPreOlympia verifies EIP-1559 DynamicFee (Type-2) transactions are
// rejected before the Olympia activation block.
func TestType2RejectedPreOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(99, 8_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(1337),
		Nonce:     0,
		To:        &common.Address{0x01},
		Gas:       21000,
		GasFeeCap: big.NewInt(1e10),
		GasTipCap: big.NewInt(1e9),
	})
	err := ValidateTransaction(tx, head, signer, opts)
	if !errors.Is(err, core.ErrTxTypeNotSupported) {
		t.Fatalf("Type-2 pre-Olympia: want ErrTxTypeNotSupported, got %v", err)
	}
}

// TestType2AcceptedAtOlympia verifies EIP-1559 DynamicFee (Type-2) transactions are
// accepted at and after the Olympia activation block.
func TestType2AcceptedAtOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(100, 60_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(1337),
		Nonce:     0,
		To:        &common.Address{0x01},
		Gas:       21000,
		GasFeeCap: big.NewInt(1e10),
		GasTipCap: big.NewInt(1e9),
	})
	if err := ValidateTransaction(tx, head, signer, opts); errors.Is(err, core.ErrTxTypeNotSupported) {
		t.Fatalf("Type-2 at Olympia: should not be rejected as unsupported type, got %v", err)
	}
}

// TestType4RejectedPreOlympia verifies EIP-7702 SetCode (Type-4) transactions are
// rejected before the Olympia activation block.
func TestType4RejectedPreOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(99, 8_000_000)
	opts := &ValidationOptions{
		Config:  cfg,
		Accept:  1<<types.LegacyTxType | 1<<types.AccessListTxType | 1<<types.DynamicFeeTxType | 1<<types.SetCodeTxType,
		MaxSize: 128 * 1024,
		MinTip:  big.NewInt(0),
	}

	// The SetCode type gate fires before signature checks, so an unsigned tx is sufficient.
	tx := types.NewTx(&types.SetCodeTx{
		ChainID:   uint256.NewInt(1337),
		Nonce:     0,
		To:        common.Address{0x01},
		Gas:       21000,
		GasFeeCap: uint256.NewInt(1e10),
		GasTipCap: uint256.NewInt(1e9),
		AuthList:  []types.SetCodeAuthorization{{}},
		V:         uint256.NewInt(0),
		R:         uint256.NewInt(0),
		S:         uint256.NewInt(0),
	})
	err := ValidateTransaction(tx, head, signer, opts)
	if !errors.Is(err, core.ErrTxTypeNotSupported) {
		t.Fatalf("Type-4 pre-Olympia: want ErrTxTypeNotSupported, got %v", err)
	}
}

// TestType4AcceptedAtOlympia verifies EIP-7702 SetCode (Type-4) transactions are not
// rejected as an unsupported type at and after the Olympia activation block.
func TestType4AcceptedAtOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(100, 60_000_000)
	opts := &ValidationOptions{
		Config:  cfg,
		Accept:  1<<types.LegacyTxType | 1<<types.AccessListTxType | 1<<types.DynamicFeeTxType | 1<<types.SetCodeTxType,
		MaxSize: 128 * 1024,
		MinTip:  big.NewInt(0),
	}

	tx := types.NewTx(&types.SetCodeTx{
		ChainID:   uint256.NewInt(1337),
		Nonce:     0,
		To:        common.Address{0x01},
		Gas:       21000,
		GasFeeCap: uint256.NewInt(1e10),
		GasTipCap: uint256.NewInt(1e9),
		AuthList:  []types.SetCodeAuthorization{{}},
		V:         uint256.NewInt(0),
		R:         uint256.NewInt(0),
		S:         uint256.NewInt(0),
	})
	err := ValidateTransaction(tx, head, signer, opts)
	if errors.Is(err, core.ErrTxTypeNotSupported) {
		t.Fatalf("Type-4 at Olympia must not be rejected as unsupported type, got %v", err)
	}
}

// TestType0AcceptedPreOlympia verifies legacy (Type-0) transactions are accepted before
// Olympia activation.
func TestType0AcceptedPreOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(99, 8_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Gas:      21000,
		GasPrice: big.NewInt(1e10),
	})
	if err := ValidateTransaction(tx, head, signer, opts); err != nil {
		t.Fatalf("Type-0 pre-Olympia: should be accepted, got %v", err)
	}
}

// TestType0AcceptedPostOlympia verifies legacy (Type-0) transactions remain accepted
// after Olympia activation (backwards compatibility).
func TestType0AcceptedPostOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(100, 60_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{0x01},
		Gas:      21000,
		GasPrice: big.NewInt(1e10),
	})
	if err := ValidateTransaction(tx, head, signer, opts); err != nil {
		t.Fatalf("Type-0 post-Olympia: should be accepted, got %v", err)
	}
}

// TestType1AcceptedPreOlympia verifies EIP-2930 access-list (Type-1) transactions are
// accepted before Olympia.
func TestType1AcceptedPreOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(99, 8_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.AccessListTx{
		ChainID:  big.NewInt(1337),
		Nonce:    0,
		To:       &common.Address{0x01},
		Gas:      21000,
		GasPrice: big.NewInt(1e10),
	})
	if err := ValidateTransaction(tx, head, signer, opts); err != nil {
		t.Fatalf("Type-1 pre-Olympia: should be accepted, got %v", err)
	}
}

// TestType1AcceptedPostOlympia verifies EIP-2930 access-list (Type-1) transactions
// remain accepted after Olympia activation.
func TestType1AcceptedPostOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(100, 60_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.AccessListTx{
		ChainID:  big.NewInt(1337),
		Nonce:    0,
		To:       &common.Address{0x01},
		Gas:      21000,
		GasPrice: big.NewInt(1e10),
	})
	if err := ValidateTransaction(tx, head, signer, opts); err != nil {
		t.Fatalf("Type-1 post-Olympia: should be accepted, got %v", err)
	}
}

// TestType0ContractDeployPreOlympia verifies that contract-creation legacy transactions
// (To = nil) are accepted before Olympia.
func TestType0ContractDeployPreOlympia(t *testing.T) {
	cfg := newValidationConfig(big.NewInt(100))
	signer := types.MakeSigner(cfg, big.NewInt(100), 0)
	head := testHeader(99, 8_000_000)
	opts := validationOpts(cfg)

	tx := types.MustSignNewTx(valTestKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       nil, // contract creation
		Gas:      60000,
		GasPrice: big.NewInt(1e10),
		Data:     []byte{0x60, 0x00, 0x60, 0x00, 0xf3}, // minimal init code
	})
	if err := ValidateTransaction(tx, head, signer, opts); err != nil {
		t.Fatalf("Type-0 contract deploy pre-Olympia: should be accepted, got %v", err)
	}
}
