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

package core

import (
	"testing"

	"github.com/ethereum/go-ethereum/params/vars"
)

// TestEIP7935CalcGasLimitConverges verifies that CalcGasLimit gradually
// adjusts from the old ETC 8M gas limit toward the 60M target.
func TestEIP7935CalcGasLimitConverges(t *testing.T) {
	parentLimit := uint64(8_000_000) // pre-Olympia ETC gas limit
	target := uint64(60_000_000)     // EIP-7935 target

	// After one adjustment, gas limit should increase by delta
	next := CalcGasLimit(parentLimit, target)
	delta := parentLimit/vars.GasLimitBoundDivisor - 1
	expectedNext := parentLimit + delta

	if next != expectedNext {
		t.Fatalf("CalcGasLimit(%d, %d) = %d, want %d", parentLimit, target, next, expectedNext)
	}
	if next <= parentLimit {
		t.Fatal("gas limit should increase toward 60M target")
	}

	// Run 100 iterations — should converge toward 60M
	limit := parentLimit
	for i := 0; i < 100; i++ {
		limit = CalcGasLimit(limit, target)
	}
	if limit <= parentLimit {
		t.Fatalf("after 100 blocks, gas limit should be well above 8M: got %d", limit)
	}
	t.Logf("gas limit after 100 blocks: %d (%.1f%% of 60M target)", limit, float64(limit)/float64(target)*100)
}

// TestEIP7935CalcGasLimitAtTarget verifies that CalcGasLimit is stable
// when the gas limit already equals the target.
func TestEIP7935CalcGasLimitAtTarget(t *testing.T) {
	target := uint64(60_000_000)
	next := CalcGasLimit(target, target)
	if next != target {
		t.Fatalf("CalcGasLimit(%d, %d) = %d, want %d (should be stable at target)", target, target, next, target)
	}
}

// TestEIP7935CalcGasLimitAboveTarget verifies that CalcGasLimit decreases
// if the parent gas limit is above the desired target.
func TestEIP7935CalcGasLimitAboveTarget(t *testing.T) {
	parentLimit := uint64(80_000_000) // above 60M target
	target := uint64(60_000_000)

	next := CalcGasLimit(parentLimit, target)
	if next >= parentLimit {
		t.Fatalf("CalcGasLimit(%d, %d) = %d, should decrease toward target", parentLimit, target, next)
	}
	if next < target {
		t.Fatalf("CalcGasLimit should not overshoot below target: got %d, target %d", next, target)
	}
}

// TestEIP7935MinGasLimit verifies that CalcGasLimit enforces the minimum
// gas limit even if the desired limit is below it.
func TestEIP7935MinGasLimit(t *testing.T) {
	parentLimit := uint64(8_000_000)
	tooLow := uint64(1000) // below MinGasLimit (5000)

	next := CalcGasLimit(parentLimit, tooLow)
	if next < vars.MinGasLimit {
		t.Fatalf("CalcGasLimit should not go below MinGasLimit: got %d, min %d", next, vars.MinGasLimit)
	}
}

// TestEIP7935ConvergenceTime estimates how many blocks it takes to converge
// from the pre-Olympia 8M gas limit to the 60M EIP-7935 target.
func TestEIP7935ConvergenceTime(t *testing.T) {
	limit := uint64(8_000_000)
	target := uint64(60_000_000)
	threshold := target * 99 / 100 // 99% of target

	blocks := 0
	for limit < threshold {
		limit = CalcGasLimit(limit, target)
		blocks++
		if blocks > 100_000 {
			t.Fatalf("did not converge within 100K blocks (stuck at %d)", limit)
		}
	}
	t.Logf("convergence from 8M to 99%% of 60M: %d blocks (%.1f hours at 13s/block)",
		blocks, float64(blocks)*13/3600)
}
