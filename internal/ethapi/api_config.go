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

package ethapi

import (
	"fmt"
	"math/big"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params/confp"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/vars"
)

// ForkConfig represents the configuration of a single fork point.
type ForkConfig struct {
	ActivationBlock string            `json:"activationBlock"`
	ChainID         string            `json:"chainId"`
	Precompiles     map[string]string `json:"precompiles"`
	SystemContracts map[string]string `json:"systemContracts"`
}

// ChainConfigResult is the result of the eth_config RPC call.
type ChainConfigResult struct {
	Current *ForkConfig `json:"current"`
	Next    *ForkConfig `json:"next,omitempty"`
	Last    *ForkConfig `json:"last,omitempty"`
}

// Config returns the current, next, and last fork configurations (EIP-7910).
func (s *EthereumAPI) Config() (*ChainConfigResult, error) {
	config := s.b.ChainConfig()
	head := s.b.CurrentHeader()
	if head == nil {
		return nil, fmt.Errorf("current header not available")
	}

	headNum := head.Number.Uint64()

	// Gather all block-based fork numbers, sorted and deduplicated.
	forkBlocks := confp.BlockForks(config)
	sort.Slice(forkBlocks, func(i, j int) bool { return forkBlocks[i] < forkBlocks[j] })

	// Determine current, last, and next fork blocks.
	var currentFork, lastFork, nextFork uint64
	var hasNext, hasLast bool

	for _, fb := range forkBlocks {
		if fb == 0 {
			continue // Skip genesis forks
		}
		if fb <= headNum {
			lastFork = currentFork
			if currentFork > 0 {
				hasLast = true
			}
			currentFork = fb
		} else {
			nextFork = fb
			hasNext = true
			break
		}
	}

	result := &ChainConfigResult{}

	// Current fork config (at head block)
	result.Current = buildForkConfig(config, head.Number, headNum)

	// Next fork (if scheduled)
	if hasNext {
		result.Next = buildForkConfig(config, new(big.Int).SetUint64(nextFork), nextFork)
	}

	// Last fork (most recently passed)
	if hasLast {
		result.Last = buildForkConfig(config, new(big.Int).SetUint64(lastFork), lastFork)
	}

	return result, nil
}

// buildForkConfig creates a ForkConfig for a given block number.
func buildForkConfig(config ctypes.ChainConfigurator, bn *big.Int, blockNum uint64) *ForkConfig {
	fc := &ForkConfig{
		ActivationBlock: fmt.Sprintf("%d", blockNum),
		Precompiles:     gatherPrecompiles(config, bn),
		SystemContracts: gatherSystemContracts(config, bn),
	}

	// Chain ID
	if cid := config.GetChainID(); cid != nil {
		fc.ChainID = fmt.Sprintf("0x%x", cid)
	}

	return fc
}

// precompileNames maps addresses to canonical precompile names.
var precompileNames = map[common.Address]string{
	common.BytesToAddress([]byte{1}):    "ecrecover",
	common.BytesToAddress([]byte{2}):    "sha256",
	common.BytesToAddress([]byte{3}):    "ripemd160",
	common.BytesToAddress([]byte{4}):    "identity",
	common.BytesToAddress([]byte{5}):    "modexp",
	common.BytesToAddress([]byte{6}):    "bn256Add",
	common.BytesToAddress([]byte{7}):    "bn256ScalarMul",
	common.BytesToAddress([]byte{8}):    "bn256Pairing",
	common.BytesToAddress([]byte{9}):    "blake2f",
	common.BytesToAddress([]byte{0x0b}): "bls12381G1Add",
	common.BytesToAddress([]byte{0x0c}): "bls12381G1MultiExp",
	common.BytesToAddress([]byte{0x0d}): "bls12381G2Add",
	common.BytesToAddress([]byte{0x0e}): "bls12381G2MultiExp",
	common.BytesToAddress([]byte{0x0f}): "bls12381Pairing",
	common.BytesToAddress([]byte{0x10}): "bls12381MapG1",
	common.BytesToAddress([]byte{0x11}): "bls12381MapG2",
	common.BytesToAddress([]byte{1, 0}): "p256Verify",
}

// gatherPrecompiles returns the active precompiles at a given block number.
func gatherPrecompiles(config ctypes.ChainConfigurator, bn *big.Int) map[string]string {
	precompiles := vm.PrecompiledContractsForConfig(config, bn, nil)
	result := make(map[string]string, len(precompiles))
	for addr := range precompiles {
		name, ok := precompileNames[addr]
		if !ok {
			name = addr.Hex()
		}
		result[name] = addr.Hex()
	}
	return result
}

// gatherSystemContracts returns the active system contracts at a given block number.
func gatherSystemContracts(config ctypes.ChainConfigurator, bn *big.Int) map[string]string {
	contracts := make(map[string]string)

	// EIP-2935: History storage contract
	if config.IsEnabled(config.GetEIP2935Transition, bn) {
		contracts["historyStorage"] = vars.HistoryStorageAddress.Hex()
	}

	return contracts
}
