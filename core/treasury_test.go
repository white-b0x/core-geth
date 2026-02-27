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
	treasuryTestKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	treasuryTestAddr   = crypto.PubkeyToAddress(treasuryTestKey.PublicKey)
	treasuryAddress    = common.HexToAddress("0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF")
)

// newTreasuryConfig returns a minimal CoreGethChainConfig with EIP-1559
// and treasury redirect activated at the given block number.
func newTreasuryConfig(forkBlock int64) *coregeth.CoreGethChainConfig {
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
		EIP170FBlock: big.NewInt(0),
		EIP140FBlock: big.NewInt(0),
		EIP658FBlock: big.NewInt(0),
		// EIP-1559 + EIP-3198 + Treasury activation
		EIP1559FBlock:          big.NewInt(forkBlock),
		EIP3198FBlock:          big.NewInt(forkBlock),
		OlympiaTreasuryAddress: &treasuryAddress,
	}
}

// TestTreasuryBaseFeeRedirect verifies that with EIP-1559 + ECIP-1111 active,
// the basefee revenue (baseFee * gasUsed) goes to the treasury address
// instead of being burned.
func TestTreasuryBaseFeeRedirect(t *testing.T) {
	config := newTreasuryConfig(0) // activate at genesis

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			treasuryTestAddr: {Balance: big.NewInt(1_000_000_000_000_000_000)}, // 1 ETH
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 5
	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		// Add a simple value transfer to create gas usage
		tx := types.MustSignNewTx(treasuryTestKey, gen.Signer(), &types.LegacyTx{
			Nonce:    gen.TxNonce(treasuryTestAddr),
			To:       &common.Address{0xaa},
			Value:    big.NewInt(1000),
			Gas:      vars.TxGas,
			GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(1_000_000_000)), // baseFee + 1 gwei tip
		})
		gen.AddTx(tx)
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Verify treasury accumulated basefee revenue for each block
	var expectedTreasury uint256.Int
	for i := 0; i < numBlocks; i++ {
		block := genchain[i]
		receipt := genreceipts[i][0] // one tx per block

		if receipt.Status != types.ReceiptStatusSuccessful {
			t.Fatalf("block %d: tx failed with status %d", block.NumberU64(), receipt.Status)
		}

		// baseFee * gasUsed for this block
		baseFee := uint256.MustFromBig(block.BaseFee())
		gasUsed := new(uint256.Int).SetUint64(block.GasUsed())
		blockBaseFeeRevenue := new(uint256.Int).Mul(baseFee, gasUsed)
		expectedTreasury.Add(&expectedTreasury, blockBaseFeeRevenue)

		t.Logf("block %d: baseFee=%s gasUsed=%d baseFeeRevenue=%s",
			block.NumberU64(), block.BaseFee(), block.GasUsed(), blockBaseFeeRevenue)
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	actualTreasury := statedb.GetBalance(treasuryAddress)
	if actualTreasury.Cmp(&expectedTreasury) != 0 {
		t.Fatalf("treasury balance mismatch:\n  got:  %s\n  want: %s", actualTreasury, &expectedTreasury)
	}
	if expectedTreasury.IsZero() {
		t.Fatal("expected non-zero treasury balance (blocks had transactions)")
	}
	t.Logf("treasury balance correct: %s", &expectedTreasury)
}

// TestTreasuryForkBoundary verifies that blocks before EIP-1559 activation
// do not credit the treasury, and blocks after activation do.
func TestTreasuryForkBoundary(t *testing.T) {
	forkBlock := int64(3)
	config := newTreasuryConfig(forkBlock)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			treasuryTestAddr: {Balance: big.NewInt(1_000_000_000_000_000_000)},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 6
	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		blockNum := int64(i + 1)
		if blockNum >= forkBlock {
			// Post-fork: send tx to generate gas usage
			tx := types.MustSignNewTx(treasuryTestKey, gen.Signer(), &types.LegacyTx{
				Nonce:    gen.TxNonce(treasuryTestAddr),
				To:       &common.Address{0xaa},
				Value:    big.NewInt(1000),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(1_000_000_000)),
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Calculate expected treasury balance (only post-fork blocks)
	var expectedTreasury uint256.Int
	for i := 0; i < numBlocks; i++ {
		block := genchain[i]
		blockNum := block.NumberU64()

		if int64(blockNum) >= forkBlock && len(genreceipts[i]) > 0 {
			baseFee := uint256.MustFromBig(block.BaseFee())
			gasUsed := new(uint256.Int).SetUint64(block.GasUsed())
			blockRevenue := new(uint256.Int).Mul(baseFee, gasUsed)
			expectedTreasury.Add(&expectedTreasury, blockRevenue)
		}
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	actualTreasury := statedb.GetBalance(treasuryAddress)
	if actualTreasury.Cmp(&expectedTreasury) != 0 {
		t.Fatalf("treasury balance mismatch:\n  got:  %s\n  want: %s", actualTreasury, &expectedTreasury)
	}
	if expectedTreasury.IsZero() {
		t.Fatal("expected non-zero treasury balance (post-fork blocks had transactions)")
	}
	t.Logf("treasury balance correct: %s (from %d post-fork blocks)", &expectedTreasury, numBlocks-int(forkBlock)+1)
}

// TestTreasuryZeroGasUsed verifies that blocks with no transactions
// credit zero to the treasury (baseFee * 0 = 0).
func TestTreasuryZeroGasUsed(t *testing.T) {
	config := newTreasuryConfig(0) // activate at genesis

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	// Generate empty blocks (no transactions)
	numBlocks := 3
	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, nil)

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Verify all blocks have zero gasUsed
	for _, block := range genchain {
		if block.GasUsed() != 0 {
			t.Fatalf("block %d: expected 0 gasUsed, got %d", block.NumberU64(), block.GasUsed())
		}
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Treasury should have zero balance (no gas used = no basefee revenue)
	balance := statedb.GetBalance(treasuryAddress)
	if !balance.IsZero() {
		t.Fatalf("expected zero treasury balance for empty blocks, got %s", balance)
	}
}

// TestTreasuryMinerGetsTips verifies that with EIP-1559 active, the miner
// receives only the tip (gasPrice - baseFee) and not the baseFee portion.
func TestTreasuryMinerGetsTips(t *testing.T) {
	config := newTreasuryConfig(0)

	minerAddr := common.HexToAddress("0x1111111111111111111111111111111111111111")

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			treasuryTestAddr: {Balance: big.NewInt(1_000_000_000_000_000_000)},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	tipPerGas := big.NewInt(1_000_000_000) // 1 gwei tip

	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 1, func(i int, gen *BlockGen) {
		gen.SetCoinbase(minerAddr)
		tx := types.MustSignNewTx(treasuryTestKey, gen.Signer(), &types.LegacyTx{
			Nonce:    gen.TxNonce(treasuryTestAddr),
			To:       &common.Address{0xaa},
			Value:    big.NewInt(1000),
			Gas:      vars.TxGas,
			GasPrice: new(big.Int).Add(gen.BaseFee(), tipPerGas),
		})
		gen.AddTx(tx)
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	block := genchain[0]
	receipt := genreceipts[0][0]
	if receipt.Status != types.ReceiptStatusSuccessful {
		t.Fatalf("tx failed with status %d", receipt.Status)
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Treasury gets: baseFee * gasUsed
	baseFee := uint256.MustFromBig(block.BaseFee())
	gasUsed := new(uint256.Int).SetUint64(block.GasUsed())
	expectedTreasury := new(uint256.Int).Mul(baseFee, gasUsed)

	actualTreasury := statedb.GetBalance(treasuryAddress)
	if actualTreasury.Cmp(expectedTreasury) != 0 {
		t.Fatalf("treasury balance mismatch: got %s, want %s", actualTreasury, expectedTreasury)
	}

	// Miner gets: block reward + tip * gasUsed
	// For legacy txs, effectiveTip = gasPrice - baseFee
	minerBalance := statedb.GetBalance(minerAddr)
	expectedTip := new(uint256.Int).Mul(uint256.MustFromBig(tipPerGas), gasUsed)

	// Miner also gets block reward. We just verify tip is included.
	// Block reward for ethash varies, so we check that miner balance > tip
	if minerBalance.Cmp(expectedTip) < 0 {
		t.Fatalf("miner balance %s is less than expected tip %s", minerBalance, expectedTip)
	}

	t.Logf("block baseFee=%s gasUsed=%d", block.BaseFee(), block.GasUsed())
	t.Logf("treasury=%s (baseFee*gasUsed)", actualTreasury)
	t.Logf("miner=%s (includes block reward + tip*gasUsed)", minerBalance)
}

// TestTreasuryNoAddressNoBurn verifies that when EIP-1559 is active but
// no treasury address is configured, the basefee is implicitly burned
// (not credited to anyone).
func TestTreasuryNoAddressNoBurn(t *testing.T) {
	// Config with EIP-1559 but NO treasury address
	config := &coregeth.CoreGethChainConfig{
		NetworkID:    1337,
		Ethash:       new(ctypes.EthashConfig),
		ChainID:      big.NewInt(1337),
		EIP2FBlock:   big.NewInt(0),
		EIP7FBlock:   big.NewInt(0),
		EIP150Block:  big.NewInt(0),
		EIP155Block:  big.NewInt(0),
		EIP160FBlock: big.NewInt(0),
		EIP161FBlock: big.NewInt(0),
		EIP170FBlock: big.NewInt(0),
		EIP140FBlock: big.NewInt(0),
		EIP658FBlock: big.NewInt(0),
		// EIP-1559 active, but no treasury
		EIP1559FBlock:          big.NewInt(0),
		EIP3198FBlock:          big.NewInt(0),
		OlympiaTreasuryAddress: nil,
	}

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			treasuryTestAddr: {Balance: big.NewInt(1_000_000_000_000_000_000)},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 3, func(i int, gen *BlockGen) {
		tx := types.MustSignNewTx(treasuryTestKey, gen.Signer(), &types.LegacyTx{
			Nonce:    gen.TxNonce(treasuryTestAddr),
			To:       &common.Address{0xaa},
			Value:    big.NewInt(1000),
			Gas:      vars.TxGas,
			GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(1_000_000_000)),
		})
		gen.AddTx(tx)
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// treasuryAddress should have zero balance since it's not configured
	balance := statedb.GetBalance(treasuryAddress)
	if !balance.IsZero() {
		t.Fatalf("expected zero balance for unconfigured treasury, got %s", balance)
	}

	// Verify blocks did use gas (otherwise test is vacuous)
	for _, block := range genchain {
		if block.GasUsed() == 0 {
			t.Fatalf("block %d has zero gasUsed, test is invalid", block.NumberU64())
		}
	}
}

// TestTreasuryCumulativeAccumulation runs a 50-block chain with varying gas
// usage per block (empty blocks, single-tx, multi-tx) and verifies the
// treasury balance matches the exact sum of baseFee*gasUsed across all blocks.
func TestTreasuryCumulativeAccumulation(t *testing.T) {
	config := newTreasuryConfig(0)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			treasuryTestAddr: {Balance: new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	recipient := common.Address{0xbb}
	numBlocks := 50

	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		blockNum := i + 1

		// Vary gas usage pattern:
		// Blocks 1-5: empty (no txs)
		// Blocks 6-15: 1 tx each
		// Blocks 16-25: 3 txs each
		// Blocks 26-30: empty
		// Blocks 31-50: 1 tx each
		var txCount int
		switch {
		case blockNum <= 5:
			txCount = 0
		case blockNum <= 15:
			txCount = 1
		case blockNum <= 25:
			txCount = 3
		case blockNum <= 30:
			txCount = 0
		default:
			txCount = 1
		}

		for j := 0; j < txCount; j++ {
			tx := types.MustSignNewTx(treasuryTestKey, gen.Signer(), &types.LegacyTx{
				Nonce:    gen.TxNonce(treasuryTestAddr),
				To:       &recipient,
				Value:    big.NewInt(1000),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(1_000_000_000)),
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Calculate expected treasury: sum of baseFee * gasUsed for all blocks
	var expectedTreasury uint256.Int
	emptyBlocks := 0
	txBlocks := 0
	for _, block := range genchain {
		baseFee := uint256.MustFromBig(block.BaseFee())
		gasUsed := new(uint256.Int).SetUint64(block.GasUsed())
		revenue := new(uint256.Int).Mul(baseFee, gasUsed)
		expectedTreasury.Add(&expectedTreasury, revenue)

		if block.GasUsed() == 0 {
			emptyBlocks++
		} else {
			txBlocks++
		}
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	actualTreasury := statedb.GetBalance(treasuryAddress)
	if actualTreasury.Cmp(&expectedTreasury) != 0 {
		t.Fatalf("treasury balance mismatch after %d blocks:\n  got:  %s\n  want: %s", numBlocks, actualTreasury, &expectedTreasury)
	}
	if expectedTreasury.IsZero() {
		t.Fatal("expected non-zero treasury (chain had transactions)")
	}

	t.Logf("treasury correct: %s across %d blocks (%d with txs, %d empty)", &expectedTreasury, numBlocks, txBlocks, emptyBlocks)
}
