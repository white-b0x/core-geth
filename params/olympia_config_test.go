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

package params

// Olympia chain-config correctness tests.
// These tests verify that the ETC mainnet and Mordor testnet configs have the correct
// EIP activation blocks and fork-parameterised gas schedule as required by ECIP-1121.

import (
	"math/big"
	"testing"
)

// TestEIP1559NotAtMystique_ClassicMainnet verifies that EIP-1559 is NOT active at
// Mystique (block 14,525,000) or any earlier block on ETC mainnet.
// This is the critical invariant: EIP-1559 activates only at Olympia, not Mystique.
func TestEIP1559NotAtMystique_ClassicMainnet(t *testing.T) {
	cfg := ClassicChainConfig
	for _, bn := range []int64{0, 14_524_999, 14_525_000, 19_250_000} {
		if cfg.IsEnabled(cfg.GetEIP1559Transition, big.NewInt(bn)) {
			t.Fatalf("EIP-1559 must not be active at block %d on ETC mainnet (Mystique/Spiral)", bn)
		}
	}
}

// TestEIP1559AtOlympia_ClassicMainnet verifies that EIP-1559 IS active at the Olympia
// sentinel block on ETC mainnet.
func TestEIP1559AtOlympia_ClassicMainnet(t *testing.T) {
	cfg := ClassicChainConfig
	// olympiaMainnetBlock = 1_000_000_000_000_000_000
	olympia := big.NewInt(int64(olympiaMainnetBlock))
	if !cfg.IsEnabled(cfg.GetEIP1559Transition, olympia) {
		t.Fatalf("EIP-1559 must be active at Olympia sentinel block %s on ETC mainnet", olympia)
	}
}

// TestEIP7825AtOlympia_ClassicMainnet verifies that EIP-7825 (per-tx gas limit cap)
// activates at Olympia on ETC mainnet, not before.
func TestEIP7825AtOlympia_ClassicMainnet(t *testing.T) {
	cfg := ClassicChainConfig
	// Must NOT be active at Mystique/Spiral
	if cfg.IsEnabled(cfg.GetEIP7825Transition, big.NewInt(19_250_000)) {
		t.Fatal("EIP-7825 must not be active at Spiral on ETC mainnet")
	}
	// Must be active at Olympia
	olympia := big.NewInt(int64(olympiaMainnetBlock))
	if !cfg.IsEnabled(cfg.GetEIP7825Transition, olympia) {
		t.Fatal("EIP-7825 must be active at Olympia on ETC mainnet")
	}
}

// TestEIP7623AtOlympia_ClassicMainnet verifies that EIP-7623 (calldata floor gas cost)
// activates at Olympia on ETC mainnet, not before.
func TestEIP7623AtOlympia_ClassicMainnet(t *testing.T) {
	cfg := ClassicChainConfig
	if cfg.IsEnabled(cfg.GetEIP7623Transition, big.NewInt(19_250_000)) {
		t.Fatal("EIP-7623 must not be active at Spiral on ETC mainnet")
	}
	olympia := big.NewInt(int64(olympiaMainnetBlock))
	if !cfg.IsEnabled(cfg.GetEIP7623Transition, olympia) {
		t.Fatal("EIP-7623 must be active at Olympia on ETC mainnet")
	}
}

// TestForkGasSchedule_ClassicMainnet verifies that ClassicChainConfig has both
// SpiralGasTarget (8 M) and OlympiaGasTarget (60 M) set as required by ECIP-1121.
func TestForkGasSchedule_ClassicMainnet(t *testing.T) {
	cfg := ClassicChainConfig
	spiral := cfg.GetSpiralGasTarget()
	if spiral == nil {
		t.Fatal("SpiralGasTarget is nil on ClassicChainConfig — ECIP-1121 requires 8_000_000")
	}
	if *spiral != 8_000_000 {
		t.Fatalf("SpiralGasTarget = %d, want 8_000_000", *spiral)
	}
	olympia := cfg.GetOlympiaGasTarget()
	if olympia == nil {
		t.Fatal("OlympiaGasTarget is nil on ClassicChainConfig — ECIP-1121 requires 60_000_000")
	}
	if *olympia != 60_000_000 {
		t.Fatalf("OlympiaGasTarget = %d, want 60_000_000", *olympia)
	}
}

// TestForkGasSchedule_Mordor verifies that MordorChainConfig has the same gas schedule
// targets as mainnet (both use ECIP-1121 values).
func TestForkGasSchedule_Mordor(t *testing.T) {
	cfg := MordorChainConfig
	spiral := cfg.GetSpiralGasTarget()
	if spiral == nil || *spiral != 8_000_000 {
		t.Fatalf("Mordor SpiralGasTarget = %v, want 8_000_000", spiral)
	}
	olympia := cfg.GetOlympiaGasTarget()
	if olympia == nil || *olympia != 60_000_000 {
		t.Fatalf("Mordor OlympiaGasTarget = %v, want 60_000_000", olympia)
	}
}
