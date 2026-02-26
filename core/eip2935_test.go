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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
	"github.com/ethereum/go-ethereum/params/vars"
	"github.com/ethereum/go-ethereum/triedb"
)

// newEIP2935Config returns a minimal CoreGethChainConfig with EIP-2935
// activated at the given block number.
func newEIP2935Config(forkBlock int64) *coregeth.CoreGethChainConfig {
	return &coregeth.CoreGethChainConfig{
		NetworkID:    1337,
		Ethash:       new(ctypes.EthashConfig),
		ChainID:      big.NewInt(1337),
		EIP2FBlock:   big.NewInt(0),
		EIP7FBlock:   big.NewInt(0),
		EIP150Block:  big.NewInt(0),
		EIP155Block:  big.NewInt(0),
		EIP160FBlock: big.NewInt(0),
		EIP161FBlock: big.NewInt(0),
		EIP170FBlock:  big.NewInt(0),
		EIP140FBlock:  big.NewInt(0), // REVERT opcode (used by history contract)
		EIP658FBlock:  big.NewInt(0),
		EIP3855FBlock: big.NewInt(0), // PUSH0 opcode (used by history contract bytecode)
		// EIP-2935 activation
		EIP2935FBlock: big.NewInt(forkBlock),
	}
}

// TestEIP2935HistoryStorage verifies that EIP-2935 deploys the history
// storage contract at the fork block and stores parent hashes correctly.
func TestEIP2935HistoryStorage(t *testing.T) {
	config := newEIP2935Config(1) // activate at block 1

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 5
	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, nil)

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Verify contract was deployed
	code := statedb.GetCode(vars.HistoryStorageAddress)
	if len(code) == 0 {
		t.Fatal("EIP-2935 history storage contract not deployed")
	}
	if nonce := statedb.GetNonce(vars.HistoryStorageAddress); nonce != 1 {
		t.Fatalf("expected nonce 1, got %d", nonce)
	}

	// Verify parent hashes in storage for each block
	for i := 0; i < numBlocks; i++ {
		block := genchain[i]
		blockNum := block.NumberU64()
		slot := (blockNum - 1) % vars.HistoryServeWindow
		got := statedb.GetState(vars.HistoryStorageAddress, common.BigToHash(new(big.Int).SetUint64(slot)))
		want := block.ParentHash()
		if got != want {
			t.Errorf("block %d: wrong parent hash at slot %d\ngot:  %s\nwant: %s", blockNum, slot, got.Hex(), want.Hex())
		}
	}
}

// TestEIP2935ForkBoundary verifies that blocks before the EIP-2935
// activation do not store parent hashes in the history contract.
func TestEIP2935ForkBoundary(t *testing.T) {
	forkBlock := int64(3)
	config := newEIP2935Config(forkBlock)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 5
	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, nil)

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Blocks before fork should have no storage entries
	for i := 0; i < int(forkBlock)-1; i++ {
		block := genchain[i]
		blockNum := block.NumberU64()
		slot := (blockNum - 1) % vars.HistoryServeWindow
		got := statedb.GetState(vars.HistoryStorageAddress, common.BigToHash(new(big.Int).SetUint64(slot)))
		if got != (common.Hash{}) {
			t.Errorf("block %d (pre-fork): expected empty storage at slot %d, got %s", blockNum, slot, got.Hex())
		}
	}

	// Fork block and after should have parent hashes stored
	for i := int(forkBlock) - 1; i < numBlocks; i++ {
		block := genchain[i]
		blockNum := block.NumberU64()
		slot := (blockNum - 1) % vars.HistoryServeWindow
		got := statedb.GetState(vars.HistoryStorageAddress, common.BigToHash(new(big.Int).SetUint64(slot)))
		want := block.ParentHash()
		if got != want {
			t.Errorf("block %d (post-fork): wrong parent hash at slot %d\ngot:  %s\nwant: %s", blockNum, slot, got.Hex(), want.Hex())
		}
	}
}

// TestEIP2935ContractDeployment verifies that the history storage contract
// is deployed exactly at the fork activation block with nonce=1.
func TestEIP2935ContractDeployment(t *testing.T) {
	forkBlock := int64(3)
	config := newEIP2935Config(forkBlock)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 5
	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, nil)

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	// Insert blocks one by one to check state at each step
	for i, block := range genchain {
		if _, err := blockchain.InsertChain([]*types.Block{block}); err != nil {
			t.Fatalf("insert error (block %d): %v", block.NumberU64(), err)
		}

		statedb, err := blockchain.State()
		if err != nil {
			t.Fatal(err)
		}

		code := statedb.GetCode(vars.HistoryStorageAddress)
		blockNum := block.NumberU64()

		if blockNum < uint64(forkBlock) {
			// Before fork: contract should NOT exist
			if len(code) > 0 {
				t.Errorf("block %d (pre-fork): contract should not exist yet", blockNum)
			}
		} else {
			// At and after fork: contract MUST exist
			if len(code) == 0 {
				t.Errorf("block %d (post-fork): contract not deployed", blockNum)
			}
			if nonce := statedb.GetNonce(vars.HistoryStorageAddress); nonce != 1 {
				t.Errorf("block %d: expected nonce 1, got %d", blockNum, nonce)
			}
		}
		_ = i
	}
}
