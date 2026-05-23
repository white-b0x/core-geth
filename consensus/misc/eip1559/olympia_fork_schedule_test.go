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

	"github.com/ethereum/go-ethereum/params/types/goethereum"
)

// TestForkScheduleOverridesOperatorGasCeil_PreOlympia verifies that ForkGasTarget returns
// 8 M for all Spiral-era blocks, regardless of any operator --miner.gaslimit setting.
// Pre-Olympia ETC nodes must be held to exactly 8 M by the fork schedule.
func TestForkScheduleOverridesOperatorGasCeil_PreOlympia(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock) // spiralBlock = 50

	for _, bn := range []uint64{50, 75, 99} {
		got := ForkGasTarget(cfg, new(big.Int).SetUint64(bn))
		if got == nil {
			t.Fatalf("block %d: ForkGasTarget = nil, want 8_000_000", bn)
		}
		if *got != 8_000_000 {
			t.Fatalf("block %d: ForkGasTarget = %d, want 8_000_000 (Spiral schedule)", bn, *got)
		}
	}
}

// TestForkScheduleUsed_PostOlympia verifies that ForkGasTarget returns 60 M at and after
// Olympia activation, driving gas limit convergence to 60 M.
func TestForkScheduleUsed_PostOlympia(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)

	for _, bn := range []uint64{100, 101, 500, 2055} {
		got := ForkGasTarget(cfg, new(big.Int).SetUint64(bn))
		if got == nil {
			t.Fatalf("block %d: ForkGasTarget = nil, want 60_000_000", bn)
		}
		if *got != 60_000_000 {
			t.Fatalf("block %d: ForkGasTarget = %d, want 60_000_000 (Olympia schedule)", bn, *got)
		}
	}
}

// TestForkScheduleFallback_WhenNil verifies that ForkGasTarget returns nil for non-ETC
// chains (goethereum.ChainConfig has no gas schedule fields), signalling that the miner
// should fall back to the operator --miner.gaslimit.
func TestForkScheduleFallback_WhenNil(t *testing.T) {
	// goethereum.ChainConfig does not implement SpiralGasTarget/OlympiaGasTarget.
	cfg := &goethereum.ChainConfig{
		ChainID:             big.NewInt(1),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
	}
	for _, bn := range []uint64{0, 100, 1_000_000} {
		got := ForkGasTarget(cfg, new(big.Int).SetUint64(bn))
		if got != nil {
			t.Fatalf("ETH chain block %d: ForkGasTarget = %d, want nil", bn, *got)
		}
	}
}

// TestForkScheduleProducesSameTrajectory verifies that ForkGasTarget returns identical
// values at each Olympia-era block regardless of any prior operator configuration.
// The schedule is network-authoritative: two nodes that differ only in --miner.gaslimit
// must produce the exact same gas limit trajectory post-Olympia.
func TestForkScheduleProducesSameTrajectory(t *testing.T) {
	const olympiaBlock = 100
	cfg := newETCTestConfig(olympiaBlock)

	// All post-Olympia blocks return 60M, never anything else.
	for i := uint64(0); i < 200; i++ {
		bn := olympiaBlock + i
		got := ForkGasTarget(cfg, new(big.Int).SetUint64(bn))
		if got == nil || *got != 60_000_000 {
			t.Fatalf("block %d: ForkGasTarget = %v, want 60_000_000", bn, got)
		}
	}
}
