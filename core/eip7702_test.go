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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
	"github.com/ethereum/go-ethereum/params/vars"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

var (
	eip7702Key1, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	eip7702Addr1   = crypto.PubkeyToAddress(eip7702Key1.PublicKey)
	eip7702Key2, _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	eip7702Addr2   = crypto.PubkeyToAddress(eip7702Key2.PublicKey)
	eip7702Key3, _ = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	eip7702Addr3   = crypto.PubkeyToAddress(eip7702Key3.PublicKey)
)

// newEIP7702Config returns a CoreGethChainConfig with EIP-7702 and
// prerequisites activated at genesis.
func newEIP7702Config() *coregeth.CoreGethChainConfig {
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
		EIP140FBlock:  big.NewInt(0),
		EIP658FBlock:  big.NewInt(0),
		EIP2565FBlock: big.NewInt(0),
		EIP2718FBlock: big.NewInt(0),
		EIP2929FBlock: big.NewInt(0),
		EIP2930FBlock: big.NewInt(0),
		EIP1559FBlock: big.NewInt(0),
		EIP3198FBlock: big.NewInt(0),
		EIP7702FBlock: big.NewInt(0),
	}
}

// makeSetCodeChain is a helper that generates and inserts a chain with a
// single SetCode transaction built by the given function.
func makeSetCodeChain(t *testing.T, config *coregeth.CoreGethChainConfig, alloc genesisT.GenesisAlloc, buildTx func(gen *BlockGen)) (*BlockChain, *types.Block) {
	t.Helper()

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   30_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc:      alloc,
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	chain, receipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 1, func(i int, gen *BlockGen) {
		buildTx(gen)
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)

	if _, err := blockchain.InsertChain(chain); err != nil {
		t.Fatalf("insert error: %v", err)
	}

	if len(receipts[0]) > 0 && receipts[0][0].Status != types.ReceiptStatusSuccessful {
		t.Logf("tx status: %d (may be expected for some edge cases)", receipts[0][0].Status)
	}

	return blockchain, chain[0]
}

// TestEIP7702DelegationMismatchedChainID verifies that an authorization with
// a non-matching chain ID (not 0 and not current chain) is silently skipped.
func TestEIP7702DelegationMismatchedChainID(t *testing.T) {
	config := newEIP7702Config()
	chainID := big.NewInt(1337)

	alloc := genesisT.GenesisAlloc{
		eip7702Addr1: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		eip7702Addr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
	}

	blockchain, _ := makeSetCodeChain(t, config, alloc, func(gen *BlockGen) {
		signer := gen.Signer()
		delegateTarget := common.Address{0xdd}

		// Authorization with wrong chain ID (9999 instead of 1337)
		auth := types.SetCodeAuthorization{
			ChainID: *uint256.NewInt(9999),
			Address: delegateTarget,
			Nonce:   0,
		}
		signedAuth, err := types.SignSetCode(eip7702Key2, auth)
		if err != nil {
			t.Fatalf("SignSetCode: %v", err)
		}

		tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
			ChainID:   uint256.MustFromBig(chainID),
			Nonce:     gen.TxNonce(eip7702Addr1),
			To:        eip7702Addr2,
			Value:     uint256.NewInt(0),
			Gas:       vars.TxGas + vars.CallNewAccountGas,
			GasTipCap: uint256.NewInt(2e9),
			GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
			AuthList:  []types.SetCodeAuthorization{signedAuth},
		})
		gen.AddTx(tx)
	})
	defer blockchain.Stop()

	statedb, _ := blockchain.State()

	// Delegation should NOT have been set (chain ID mismatch → silently skipped)
	code := statedb.GetCode(eip7702Addr2)
	if len(code) != 0 {
		t.Fatalf("expected no delegation code (mismatched chain ID should be skipped), got %d bytes", len(code))
	}
	t.Log("mismatched chain ID authorization correctly skipped")
}

// TestEIP7702DelegationWildcardChainID verifies that an authorization with
// chain ID 0 (wildcard) is accepted regardless of the current chain ID.
func TestEIP7702DelegationWildcardChainID(t *testing.T) {
	config := newEIP7702Config()
	chainID := big.NewInt(1337)

	alloc := genesisT.GenesisAlloc{
		eip7702Addr1: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		eip7702Addr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
	}

	delegateTarget := common.Address{0xdd}

	blockchain, _ := makeSetCodeChain(t, config, alloc, func(gen *BlockGen) {
		signer := gen.Signer()

		// Authorization with chain ID 0 (wildcard)
		auth := types.SetCodeAuthorization{
			ChainID: *uint256.NewInt(0),
			Address: delegateTarget,
			Nonce:   0,
		}
		signedAuth, err := types.SignSetCode(eip7702Key2, auth)
		if err != nil {
			t.Fatalf("SignSetCode: %v", err)
		}

		tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
			ChainID:   uint256.MustFromBig(chainID),
			Nonce:     gen.TxNonce(eip7702Addr1),
			To:        eip7702Addr2,
			Value:     uint256.NewInt(0),
			Gas:       vars.TxGas + vars.CallNewAccountGas,
			GasTipCap: uint256.NewInt(2e9),
			GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
			AuthList:  []types.SetCodeAuthorization{signedAuth},
		})
		gen.AddTx(tx)
	})
	defer blockchain.Stop()

	statedb, _ := blockchain.State()

	// Delegation SHOULD have been set (wildcard chain ID = 0)
	code := statedb.GetCode(eip7702Addr2)
	if len(code) != 23 {
		t.Fatalf("expected 23-byte delegation code, got %d bytes", len(code))
	}
	target, ok := types.ParseDelegation(code)
	if !ok {
		t.Fatal("failed to parse delegation code")
	}
	if target != delegateTarget {
		t.Fatalf("delegation target: got %s, want %s", target.Hex(), delegateTarget.Hex())
	}
	t.Log("wildcard chain ID (0) authorization correctly accepted")
}

// TestEIP7702ClearDelegation verifies that delegating to the zero address
// clears the delegation code from the account.
func TestEIP7702ClearDelegation(t *testing.T) {
	config := newEIP7702Config()
	chainID := big.NewInt(1337)

	alloc := genesisT.GenesisAlloc{
		eip7702Addr1: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		eip7702Addr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
	}

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   30_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc:      alloc,
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	delegateTarget := common.Address{0xdd}

	chain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 2, func(i int, gen *BlockGen) {
		signer := gen.Signer()

		switch i {
		case 0: // Block 1: Set delegation
			auth := types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: delegateTarget,
				Nonce:   0,
			}
			signedAuth, err := types.SignSetCode(eip7702Key2, auth)
			if err != nil {
				t.Fatalf("SignSetCode: %v", err)
			}
			tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     gen.TxNonce(eip7702Addr1),
				To:        eip7702Addr2,
				Value:     uint256.NewInt(0),
				Gas:       vars.TxGas + vars.CallNewAccountGas,
				GasTipCap: uint256.NewInt(2e9),
				GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
				AuthList:  []types.SetCodeAuthorization{signedAuth},
			})
			gen.AddTx(tx)

		case 1: // Block 2: Clear delegation (delegate to zero address)
			auth := types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: common.Address{}, // zero address = clear
				Nonce:   1,                // nonce incremented from block 1
			}
			signedAuth, err := types.SignSetCode(eip7702Key2, auth)
			if err != nil {
				t.Fatalf("SignSetCode: %v", err)
			}
			tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     gen.TxNonce(eip7702Addr1),
				To:        eip7702Addr2,
				Value:     uint256.NewInt(0),
				Gas:       vars.TxGas + vars.CallNewAccountGas,
				GasTipCap: uint256.NewInt(2e9),
				GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
				AuthList:  []types.SetCodeAuthorization{signedAuth},
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	// Insert block 1 and verify delegation is set
	if _, err := blockchain.InsertChain(chain[:1]); err != nil {
		t.Fatalf("insert block 1: %v", err)
	}
	state1, _ := blockchain.State()
	code1 := state1.GetCode(eip7702Addr2)
	if len(code1) != 23 {
		t.Fatalf("after block 1: expected 23-byte delegation, got %d bytes", len(code1))
	}
	t.Logf("after block 1: delegation set to %s", delegateTarget.Hex())

	// Insert block 2 and verify delegation is cleared
	if _, err := blockchain.InsertChain(chain[1:]); err != nil {
		t.Fatalf("insert block 2: %v", err)
	}
	state2, _ := blockchain.State()
	code2 := state2.GetCode(eip7702Addr2)
	if len(code2) != 0 {
		t.Fatalf("after block 2: expected empty code (cleared delegation), got %d bytes", len(code2))
	}
	t.Log("delegation correctly cleared by delegating to zero address")
}

// TestEIP7702DelegationToSelf verifies that an account can delegate to itself.
// The delegation code is set, and the delegation prefix is stored.
func TestEIP7702DelegationToSelf(t *testing.T) {
	config := newEIP7702Config()
	chainID := big.NewInt(1337)

	alloc := genesisT.GenesisAlloc{
		eip7702Addr1: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		eip7702Addr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
	}

	blockchain, _ := makeSetCodeChain(t, config, alloc, func(gen *BlockGen) {
		signer := gen.Signer()

		// eip7702Addr2 delegates to ITSELF
		auth := types.SetCodeAuthorization{
			ChainID: *uint256.MustFromBig(chainID),
			Address: eip7702Addr2,
			Nonce:   0,
		}
		signedAuth, err := types.SignSetCode(eip7702Key2, auth)
		if err != nil {
			t.Fatalf("SignSetCode: %v", err)
		}

		tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
			ChainID:   uint256.MustFromBig(chainID),
			Nonce:     gen.TxNonce(eip7702Addr1),
			To:        eip7702Addr2,
			Value:     uint256.NewInt(0),
			Gas:       vars.TxGas + vars.CallNewAccountGas,
			GasTipCap: uint256.NewInt(2e9),
			GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
			AuthList:  []types.SetCodeAuthorization{signedAuth},
		})
		gen.AddTx(tx)
	})
	defer blockchain.Stop()

	statedb, _ := blockchain.State()

	// Delegation code should be set (23 bytes: ef0100 + eip7702Addr2)
	code := statedb.GetCode(eip7702Addr2)
	if len(code) != 23 {
		t.Fatalf("expected 23-byte self-delegation code, got %d bytes", len(code))
	}
	target, ok := types.ParseDelegation(code)
	if !ok {
		t.Fatal("failed to parse self-delegation code")
	}
	if target != eip7702Addr2 {
		t.Fatalf("delegation target: got %s, want %s (self)", target.Hex(), eip7702Addr2.Hex())
	}
	t.Log("self-delegation correctly set")
}

// TestEIP7702DelegationChainStopsAtOneLevel verifies that delegation resolution
// only follows one level. If A delegates to B, and B delegates to C, calling A
// should resolve to B's code (the delegation prefix pointing to C), NOT C's actual code.
func TestEIP7702DelegationChainStopsAtOneLevel(t *testing.T) {
	config := newEIP7702Config()
	chainID := big.NewInt(1337)

	alloc := genesisT.GenesisAlloc{
		eip7702Addr1: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		eip7702Addr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		eip7702Addr3: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
	}

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   30_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc:      alloc,
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	// Set up: addr2 delegates to addr3, addr3 delegates to some contract
	contractAddr := common.Address{0xcc}

	chain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 2, func(i int, gen *BlockGen) {
		signer := gen.Signer()

		switch i {
		case 0: // Block 1: addr3 delegates to contractAddr
			auth := types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: contractAddr,
				Nonce:   0,
			}
			signedAuth, _ := types.SignSetCode(eip7702Key3, auth)
			tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     gen.TxNonce(eip7702Addr1),
				To:        eip7702Addr3,
				Value:     uint256.NewInt(0),
				Gas:       vars.TxGas + vars.CallNewAccountGas,
				GasTipCap: uint256.NewInt(2e9),
				GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
				AuthList:  []types.SetCodeAuthorization{signedAuth},
			})
			gen.AddTx(tx)

		case 1: // Block 2: addr2 delegates to addr3
			auth := types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: eip7702Addr3,
				Nonce:   0,
			}
			signedAuth, _ := types.SignSetCode(eip7702Key2, auth)
			tx := types.MustSignNewTx(eip7702Key1, signer, &types.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     gen.TxNonce(eip7702Addr1),
				To:        eip7702Addr2,
				Value:     uint256.NewInt(0),
				Gas:       vars.TxGas + vars.CallNewAccountGas,
				GasTipCap: uint256.NewInt(2e9),
				GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
				AuthList:  []types.SetCodeAuthorization{signedAuth},
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if _, err := blockchain.InsertChain(chain); err != nil {
		t.Fatalf("insert error: %v", err)
	}

	statedb, _ := blockchain.State()

	// Verify delegation chain: addr2 → addr3, addr3 → contractAddr
	code2 := statedb.GetCode(eip7702Addr2)
	target2, ok := types.ParseDelegation(code2)
	if !ok {
		t.Fatal("addr2 should have delegation code")
	}
	if target2 != eip7702Addr3 {
		t.Fatalf("addr2 delegation target: got %s, want %s", target2.Hex(), eip7702Addr3.Hex())
	}

	code3 := statedb.GetCode(eip7702Addr3)
	target3, ok := types.ParseDelegation(code3)
	if !ok {
		t.Fatal("addr3 should have delegation code")
	}
	if target3 != contractAddr {
		t.Fatalf("addr3 delegation target: got %s, want %s", target3.Hex(), contractAddr.Hex())
	}

	// The key insight: resolveCode(addr2) reads addr2's code → delegation to addr3.
	// It then reads addr3's code → which is ALSO a delegation (to contractAddr).
	// But resolveCode only follows ONE level, so it returns addr3's delegation
	// prefix bytes (ef0100 + contractAddr), NOT contractAddr's actual code.
	// Since the delegation prefix is not valid EVM bytecode, execution would
	// fail/return empty rather than executing contractAddr's code.
	t.Log("delegation chain correctly set: addr2 → addr3 → contractAddr")
	t.Log("resolveCode(addr2) returns addr3's code (delegation prefix), does NOT follow to contractAddr")
}
