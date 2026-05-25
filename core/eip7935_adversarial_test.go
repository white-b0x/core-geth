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
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/params/vars"
)

// simulateMixedMining simulates a population of miners with different gas limit
// targets over a number of blocks. Each block is "won" by a miner according to
// the hashrate split — honestPct% of blocks target honestTarget, the rest
// target adversaryTarget. Returns the final gas limit.
func simulateMixedMining(startLimit, honestTarget, adversaryTarget uint64, honestPct int, blocks int, seed int64) uint64 {
	rng := rand.New(rand.NewSource(seed))
	limit := startLimit
	for i := 0; i < blocks; i++ {
		target := adversaryTarget
		if rng.Intn(100) < honestPct {
			target = honestTarget
		}
		limit = CalcGasLimit(limit, target)
	}
	return limit
}

// TestEIP7935AdversarialMinority verifies that a 30% adversarial hashrate
// targeting 30M cannot significantly affect the equilibrium when 70% of
// miners correctly target 60M.
func TestEIP7935AdversarialMinority(t *testing.T) {
	startLimit := uint64(8_000_000)
	honestTarget := uint64(60_000_000)
	adversaryTarget := uint64(30_000_000)

	// Run multiple seeds to reduce randomness sensitivity
	for _, seed := range []int64{42, 123, 999, 7777} {
		final := simulateMixedMining(startLimit, honestTarget, adversaryTarget, 70, 10_000, seed)

		// With 70% honest, equilibrium should be near 60M (within 5%)
		low := honestTarget * 95 / 100   // 57M
		high := honestTarget * 105 / 100 // 63M

		if final < low || final > high {
			t.Errorf("seed=%d: 70/30 split equilibrium %d outside expected range [%d, %d]",
				seed, final, low, high)
		}
		t.Logf("seed=%d: 70/30 split → final gas limit %d (%.1f%% of 60M target)",
			seed, final, float64(final)/float64(honestTarget)*100)
	}
}

// TestEIP7935AdversarialMajority verifies that when adversarial miners
// control >50% hashrate, the equilibrium shifts toward the adversary's
// target. This documents the expected behavior — it is not a vulnerability
// per se (>50% hashrate is already a majority attack scenario).
func TestEIP7935AdversarialMajority(t *testing.T) {
	startLimit := uint64(8_000_000)
	honestTarget := uint64(60_000_000)
	adversaryTarget := uint64(30_000_000)

	// 50/50 split: equilibrium should settle near the lower target
	for _, seed := range []int64{42, 123, 999} {
		final := simulateMixedMining(startLimit, honestTarget, adversaryTarget, 50, 10_000, seed)

		// With 50/50, the lower target wins because both sides push when
		// limit is below their target, but above the lower target, the
		// adversary pushes down while honest pushes up — net zero.
		// Below the lower target, both push up. So equilibrium ≈ lower target.
		low := adversaryTarget * 90 / 100   // 27M
		high := adversaryTarget * 120 / 100 // 36M

		if final < low || final > high {
			t.Errorf("seed=%d: 50/50 split equilibrium %d outside expected range [%d, %d]",
				seed, final, low, high)
		}
		t.Logf("seed=%d: 50/50 split → final gas limit %d (%.1f%% of 30M adversary target)",
			seed, final, float64(final)/float64(adversaryTarget)*100)
	}

	// 40/60 split (adversary has 60%): equilibrium should be near adversary's target
	adversaryTarget52 := uint64(52_000_000)
	for _, seed := range []int64{42, 123} {
		final := simulateMixedMining(startLimit, honestTarget, adversaryTarget52, 40, 10_000, seed)

		low := adversaryTarget52 * 95 / 100
		high := adversaryTarget52 * 105 / 100

		if final < low || final > high {
			t.Errorf("seed=%d: 40/60 split equilibrium %d outside expected range [%d, %d]",
				seed, final, low, high)
		}
		t.Logf("seed=%d: 40/60 split (adversary=52M) → final gas limit %d",
			seed, final)
	}
}

// TestEIP7935ConsensusRejectsInvalidDelta verifies that VerifyGaslimit
// enforces the ±1/1024 bound and rejects blocks with gas limit changes
// exceeding the allowed delta.
func TestEIP7935ConsensusRejectsInvalidDelta(t *testing.T) {
	parentLimit := uint64(60_000_000)
	maxDelta := parentLimit/vars.GasLimitBoundDivisor - 1

	tests := []struct {
		name      string
		newLimit  uint64
		wantError bool
	}{
		{"no change", parentLimit, false},
		{"increase by 1", parentLimit + 1, false},
		{"decrease by 1", parentLimit - 1, false},
		{"max valid increase", parentLimit + maxDelta, false},
		{"max valid decrease", parentLimit - maxDelta, false},
		{"one over max increase", parentLimit + maxDelta + 1, true},
		{"one over max decrease", parentLimit - maxDelta - 1, true},
		{"double jump increase", parentLimit + 2*maxDelta, true},
		{"dramatic jump to 120M", 120_000_000, true},
		{"dramatic drop to 8M", 8_000_000, true},
		{"below MinGasLimit", vars.MinGasLimit - 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := misc.VerifyGaslimit(parentLimit, tt.newLimit)
			if tt.wantError && err == nil {
				t.Errorf("VerifyGaslimit(%d, %d) should have failed but didn't",
					parentLimit, tt.newLimit)
			}
			if !tt.wantError && err != nil {
				t.Errorf("VerifyGaslimit(%d, %d) unexpected error: %v",
					parentLimit, tt.newLimit, err)
			}
		})
	}
}

// TestEIP7935ConvergenceFromVariousStarts verifies convergence behavior
// from different starting points, documenting the expected block counts
// to reach 99% of the 60M target.
func TestEIP7935ConvergenceFromVariousStarts(t *testing.T) {
	target := uint64(60_000_000)
	threshold := target * 99 / 100

	starts := []struct {
		name  string
		start uint64
	}{
		{"ETC pre-Olympia 8M", 8_000_000},
		{"ETH-like 30M", 30_000_000},
		{"already at 60M", 60_000_000},
		{"above target 80M", 80_000_000},
		{"minimum gas limit", vars.MinGasLimit},
	}

	for _, s := range starts {
		t.Run(s.name, func(t *testing.T) {
			limit := s.start
			blocks := 0

			if s.start >= target {
				// For at-or-above-target, verify stability/decrease
				next := CalcGasLimit(limit, target)
				if s.start == target && next != target {
					t.Fatalf("at target: CalcGasLimit(%d, %d) = %d, want stable", s.start, target, next)
				}
				if s.start > target && next >= s.start {
					t.Fatalf("above target: CalcGasLimit(%d, %d) = %d, should decrease", s.start, target, next)
				}
				t.Logf("start=%d: next=%d (stable or decreasing as expected)", s.start, next)
				return
			}

			for limit < threshold {
				limit = CalcGasLimit(limit, target)
				blocks++
				if blocks > 200_000 {
					t.Fatalf("did not converge within 200K blocks (stuck at %d)", limit)
				}
			}
			hours := float64(blocks) * 13 / 3600
			t.Logf("start=%d: convergence to 99%% of 60M in %d blocks (%.1f hours at 13s/block)",
				s.start, blocks, hours)
		})
	}
}
