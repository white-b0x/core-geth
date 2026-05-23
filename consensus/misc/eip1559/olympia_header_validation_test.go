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

package eip1559

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params/vars"
)

// TestFirstOlympiaBlock_CorrectBaseFee_Valid verifies that the first Olympia block
// (parent is pre-Olympia) is accepted when baseFee = InitialBaseFee and
// gasLimit = parent + 1/1024 step.
func TestFirstOlympiaBlock_CorrectBaseFee_Valid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	child := olympiaHeader(olympiaBlock, 8_007_811, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	if err := VerifyEIP1559Header(cfg, parent, child); err != nil {
		t.Fatalf("first Olympia block should be valid: %v", err)
	}
}

// TestFirstOlympiaBlock_WrongBaseFee_Invalid verifies that the first Olympia block is
// rejected when baseFee deviates from InitialBaseFee by even 1 wei.
func TestFirstOlympiaBlock_WrongBaseFee_Invalid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	badBaseFee := new(big.Int).SetUint64(vars.InitialBaseFee + 1)
	child := olympiaHeader(olympiaBlock, 8_007_811, 0, badBaseFee)
	if err := VerifyEIP1559Header(cfg, parent, child); err == nil {
		t.Fatal("first Olympia block with wrong baseFee should be rejected")
	}
}

// TestFirstOlympiaBlock_16MGasLimit_Invalid is the regression test for Bug A.
// With the fix applied, VerifyEIP1559Header must reject a 16 M gasLimit at the first
// Olympia block — the ETH 2× doubling must NOT apply to ETC.
func TestFirstOlympiaBlock_16MGasLimit_Invalid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	child := olympiaHeader(olympiaBlock, 16_000_000, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	if err := VerifyEIP1559Header(cfg, parent, child); err == nil {
		t.Fatal("16 M gasLimit at first Olympia block should be rejected — Bug A not fixed")
	}
}

// TestFirstOlympiaBlock_MissingBaseFee_Invalid verifies that the first Olympia block
// is rejected when the baseFee field is nil.
func TestFirstOlympiaBlock_MissingBaseFee_Invalid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	// olympiaHeader always sets BaseFee; build without it manually.
	child := preOlympiaHeader(olympiaBlock, 8_007_811, 0) // no BaseFee
	if err := VerifyEIP1559Header(cfg, parent, child); err == nil {
		t.Fatal("first Olympia block without baseFee should be rejected")
	}
}

// TestFirstOlympiaBlock_TooBigGasLimitStep_Invalid verifies the off-by-one boundary:
// a diff of exactly parentGasLimit/1024 (= 7812) is rejected (must be strictly less).
func TestFirstOlympiaBlock_TooBigGasLimitStep_Invalid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parent := preOlympiaHeader(olympiaBlock-1, 8_000_000, 0)
	// 8_000_000 / 1024 = 7812; diff must be < 7812, so 8_007_812 is invalid
	child := olympiaHeader(olympiaBlock, 8_007_812, 0, new(big.Int).SetUint64(vars.InitialBaseFee))
	if err := VerifyEIP1559Header(cfg, parent, child); err == nil {
		t.Fatal("gasLimit diff = 7812 (= limit) should be rejected (need diff < limit)")
	}
}

// TestPostOlympiaBlock_CorrectBaseFee_Valid verifies that a post-activation block with
// a correctly derived baseFee and valid gasLimit step is accepted.
func TestPostOlympiaBlock_CorrectBaseFee_Valid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	// Block 100 as parent: gasLimit=30M, empty, baseFee=1Gwei
	parentBaseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	parent := olympiaHeader(olympiaBlock, 30_000_000, 0, parentBaseFee)
	childBaseFee := CalcBaseFee(cfg, parent)
	// One step: 30M + (30M/1024 - 1) = 30M + 29295 = 30,029,295
	child := olympiaHeader(olympiaBlock+1, 30_029_295, 0, childBaseFee)
	if err := VerifyEIP1559Header(cfg, parent, child); err != nil {
		t.Fatalf("valid post-Olympia block rejected: %v", err)
	}
}

// TestPostOlympiaBlock_MissingBaseFee_Invalid verifies that any post-activation block
// without a baseFee field is rejected.
func TestPostOlympiaBlock_MissingBaseFee_Invalid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	parentBaseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	parent := olympiaHeader(olympiaBlock, 30_000_000, 0, parentBaseFee)
	// Block 101 with no BaseFee
	child := preOlympiaHeader(olympiaBlock+1, 30_029_295, 0)
	if err := VerifyEIP1559Header(cfg, parent, child); err == nil {
		t.Fatal("post-Olympia block without baseFee should be rejected")
	}
}

// TestPostOlympiaBlock_StableGasLimit_Valid verifies a block at the target gas limit
// (gasUsed = gasTarget) maintains the same baseFee and same gasLimit.
func TestPostOlympiaBlock_StableGasLimit_Valid(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)
	gasLimit := uint64(30_000_000)
	gasTarget := gasLimit / 2 // elasticity multiplier = 2
	baseFee := new(big.Int).SetUint64(vars.InitialBaseFee)
	parent := olympiaHeader(olympiaBlock, gasLimit, gasTarget, baseFee)
	// At target: baseFee unchanged, gasLimit unchanged
	child := olympiaHeader(olympiaBlock+1, gasLimit, 0, new(big.Int).Set(baseFee))
	if err := VerifyEIP1559Header(cfg, parent, child); err != nil {
		t.Fatalf("stable post-Olympia block rejected: %v", err)
	}
}
