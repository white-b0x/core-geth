// Copyright 2021 The go-ethereum Authors
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
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
)

// ForkGasTarget returns the fork-parameterised gas limit target for the given
// block number, or nil if no schedule is configured (ETH and other non-ETC chains
// return nil, meaning the operator --miner.gaslimit flag controls the target).
//
// For ETC: Spiral-era blocks return SpiralGasTarget (8M); Olympia+ blocks return
// OlympiaGasTarget (60M). The values come from chain config so they are
// network-authoritative and cannot be overridden by operator flags.
func ForkGasTarget(config ctypes.ChainConfigurator, blockNum *big.Int) *uint64 {
	if !blockNum.IsUint64() {
		return nil
	}
	n := blockNum.Uint64()
	olympia := config.GetEIP1559Transition()
	if olympia != nil && n >= *olympia {
		return config.GetOlympiaGasTarget()
	}
	spiral := config.GetEIP3855Transition()
	if spiral != nil && n >= *spiral {
		return config.GetSpiralGasTarget()
	}
	return nil
}

// VerifyEIP1559Header verifies some header attributes which were changed in EIP-1559,
// - gas limit check
// - basefee check
func VerifyEIP1559Header(config ctypes.ChainConfigurator, parent, header *types.Header) error {
	// Verify that the gas limit remains within allowed bounds.
	// ETH London: at the EIP-1559 activation block, the parent gas limit is doubled
	// so that gasTarget = new gasLimit / 2 = old gasLimit (effective throughput preserved).
	// ETC Olympia: this doubling must NOT apply — ETC uses a fork-parameterised gas
	// schedule (ForkGasTarget) for a gradual 8M→60M ramp. Skip the 2× for ETC chains.
	parentGasLimit := parent.GasLimit
	if !config.IsEnabled(config.GetEIP1559Transition, parent.Number) {
		if ForkGasTarget(config, header.Number) == nil {
			parentGasLimit = parent.GasLimit * config.GetElasticityMultiplier()
		}
	}
	if err := misc.VerifyGaslimit(parentGasLimit, header.GasLimit); err != nil {
		return err
	}
	// Verify the header is not malformed
	if header.BaseFee == nil {
		return errors.New("header is missing baseFee")
	}
	// Verify the baseFee is correct based on the parent header.
	expectedBaseFee := CalcBaseFee(config, parent)
	if header.BaseFee.Cmp(expectedBaseFee) != 0 {
		return fmt.Errorf("invalid baseFee: have %s, want %s, parentBaseFee %s, parentGasUsed %d",
			header.BaseFee, expectedBaseFee, parent.BaseFee, parent.GasUsed)
	}
	return nil
}

// CalcBaseFee calculates the basefee of the header.
func CalcBaseFee(config ctypes.ChainConfigurator, parent *types.Header) *big.Int {
	// If the current block is the first EIP-1559 block, return the InitialBaseFee.
	if !config.IsEnabled(config.GetEIP1559Transition, parent.Number) {
		return new(big.Int).SetUint64(vars.InitialBaseFee)
	}

	parentGasTarget := parent.GasLimit / config.GetElasticityMultiplier()
	// If the parent gasUsed is the same as the target, the baseFee remains unchanged.
	if parent.GasUsed == parentGasTarget {
		return new(big.Int).Set(parent.BaseFee)
	}

	var (
		num   = new(big.Int)
		denom = new(big.Int)
	)

	if parent.GasUsed > parentGasTarget {
		// If the parent block used more gas than its target, the baseFee should increase.
		// max(1, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		num.SetUint64(parent.GasUsed - parentGasTarget)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(config.GetBaseFeeChangeDenominator()))
		baseFeeDelta := math.BigMax(num, common.Big1)

		return num.Add(parent.BaseFee, baseFeeDelta)
	} else {
		// Otherwise if the parent block used less gas than its target, the baseFee should decrease.
		// max(0, parentBaseFee * gasUsedDelta / parentGasTarget / baseFeeChangeDenominator)
		//
		// The delta is floored at 1 before subtraction (matches go-ethereum and Fukuii canonical
		// behaviour). Without this floor, when baseFee = 1 wei and gasUsed = 0, integer division
		// truncates the delta to 0, leaving baseFee stuck at 1 wei forever instead of decaying
		// to 0 on the next empty block.
		num.SetUint64(parentGasTarget - parent.GasUsed)
		num.Mul(num, parent.BaseFee)
		num.Div(num, denom.SetUint64(parentGasTarget))
		num.Div(num, denom.SetUint64(config.GetBaseFeeChangeDenominator()))
		if num.Sign() == 0 {
			num.SetUint64(1) // floor delta at 1
		}
		baseFee := num.Sub(parent.BaseFee, num)

		return math.BigMax(baseFee, common.Big0)
	}
}
