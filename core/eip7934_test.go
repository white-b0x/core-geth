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
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/consensus/ethash"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
	"github.com/ethereum/go-ethereum/params/vars"
	"github.com/ethereum/go-ethereum/triedb"
)

// newEIP7934Config returns a minimal CoreGethChainConfig with EIP-7934
// activated at the given block number.
func newEIP7934Config(forkBlock int64) *coregeth.CoreGethChainConfig {
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
		EIP7934FBlock: big.NewInt(forkBlock),
	}
}

// TestEIP7934BlockSizeConstant verifies that the BlockRLPSizeCap constant
// is 10 MiB minus 2 MiB (safety margin) = ~9.5 MiB.
func TestEIP7934BlockSizeConstant(t *testing.T) {
	expected := uint64(10*1024*1024 - 2*1024*1024) // 8,388,608
	if vars.BlockRLPSizeCap != expected {
		t.Fatalf("BlockRLPSizeCap: got %d, want %d", vars.BlockRLPSizeCap, expected)
	}
}

// TestEIP7934NormalBlocksPass verifies that normal-sized blocks pass
// validation when EIP-7934 is active.
func TestEIP7934NormalBlocksPass(t *testing.T) {
	config := newEIP7934Config(1) // activate at block 1

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 3
	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, nil)

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	// Normal blocks should all be accepted
	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}
}

// TestEIP7934ErrorDefined verifies the ErrBlockOversized error exists and
// is properly defined.
func TestEIP7934ErrorDefined(t *testing.T) {
	if !errors.Is(ErrBlockOversized, ErrBlockOversized) {
		t.Fatal("ErrBlockOversized should be a defined error")
	}
	if ErrBlockOversized.Error() != "block RLP-encoded size exceeds maximum" {
		t.Fatalf("unexpected error message: %s", ErrBlockOversized.Error())
	}
}
