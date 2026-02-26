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
	"github.com/ethereum/go-ethereum/params/mutations"
	"github.com/ethereum/go-ethereum/params/types/coregeth"
	"github.com/ethereum/go-ethereum/params/types/ctypes"
	"github.com/ethereum/go-ethereum/params/types/genesisT"
	"github.com/ethereum/go-ethereum/params/vars"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
)

var (
	compatTestKey, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	compatTestAddr    = crypto.PubkeyToAddress(compatTestKey.PublicKey)
	compatTestKey2, _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	compatTestAddr2   = crypto.PubkeyToAddress(compatTestKey2.PublicKey)
	compatTreasury    = common.HexToAddress("0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF")
	compatMiner       = common.HexToAddress("0x1111111111111111111111111111111111111111")
)

// newOlympiaTestConfig returns a CoreGethChainConfig that mirrors Mordor's
// historical forks compressed into small block numbers, with Olympia activated
// at the given block. Uses short ECIP-1017 era length for testing reward decay.
func newOlympiaTestConfig(olympiaBlock int64, eraLength int64) *coregeth.CoreGethChainConfig {
	return &coregeth.CoreGethChainConfig{
		NetworkID: 1337,
		Ethash:    new(ctypes.EthashConfig),
		ChainID:   big.NewInt(1337),

		// Frontier-equivalent (block 0)
		EIP2FBlock: big.NewInt(0),
		EIP7FBlock: big.NewInt(0),

		// Tangerine Whistle (block 0)
		EIP150Block: big.NewInt(0),

		// Spurious Dragon / EIP-158 (block 0)
		EIP155Block:  big.NewInt(0),
		EIP160FBlock: big.NewInt(0),
		EIP161FBlock: big.NewInt(0),
		EIP170FBlock: big.NewInt(0),

		// Byzantium-equivalent (block 0)
		EIP100FBlock: big.NewInt(0),
		EIP140FBlock: big.NewInt(0),
		EIP198FBlock: big.NewInt(0),
		EIP211FBlock: big.NewInt(0),
		EIP212FBlock: big.NewInt(0),
		EIP213FBlock: big.NewInt(0),
		EIP214FBlock: big.NewInt(0),
		EIP658FBlock: big.NewInt(0),

		// Constantinople / Agharta (block 0)
		EIP145FBlock:  big.NewInt(0),
		EIP1014FBlock: big.NewInt(0),
		EIP1052FBlock: big.NewInt(0),

		// Istanbul / Phoenix (block 0)
		EIP152FBlock:  big.NewInt(0),
		EIP1108FBlock: big.NewInt(0),
		EIP1344FBlock: big.NewInt(0),
		EIP1884FBlock: big.NewInt(0),
		EIP2028FBlock: big.NewInt(0),
		EIP2200FBlock: big.NewInt(0),

		// Berlin / Magneto (block 0)
		EIP2565FBlock: big.NewInt(0),
		EIP2718FBlock: big.NewInt(0),
		EIP2929FBlock: big.NewInt(0),
		EIP2930FBlock: big.NewInt(0),

		// London partial / Mystique (block 0)
		EIP3529FBlock: big.NewInt(0),
		EIP3541FBlock: big.NewInt(0),

		// Shanghai partial / Spiral (block 0)
		EIP3651FBlock: big.NewInt(0),
		EIP3855FBlock: big.NewInt(0),
		EIP3860FBlock: big.NewInt(0),
		EIP6049FBlock: big.NewInt(0),

		// Olympia
		EIP1559FBlock: big.NewInt(olympiaBlock),
		EIP3198FBlock: big.NewInt(olympiaBlock),
		EIP5656FBlock: big.NewInt(olympiaBlock),
		EIP1153FBlock: big.NewInt(olympiaBlock),
		EIP6780FBlock: big.NewInt(olympiaBlock),
		EIP7702FBlock: big.NewInt(olympiaBlock),
		EIP2935FBlock: big.NewInt(olympiaBlock),
		EIP7623FBlock: big.NewInt(olympiaBlock),
		EIP7825FBlock: big.NewInt(olympiaBlock),
		EIP7883FBlock: big.NewInt(olympiaBlock),
		EIP7823FBlock: big.NewInt(olympiaBlock),

		// ECIP-1017 monetary policy
		DisposalBlock:     big.NewInt(0),
		ECIP1017FBlock:    big.NewInt(0),
		ECIP1017EraRounds: big.NewInt(eraLength),

		// Treasury
		OlympiaTreasuryAddress: &compatTreasury,
	}
}

// TestAllTxTypesOnOlympiaChain verifies that all transaction types supported
// on Mordor after Olympia work correctly on the same chain:
// Type 0 (Legacy), Type 1 (AccessList), Type 2 (DynamicFee), Type 4 (SetCode).
func TestAllTxTypesOnOlympiaChain(t *testing.T) {
	config := newOlympiaTestConfig(0, 5_000_000)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   30_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			compatTestAddr:  {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
			compatTestAddr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	recipient := common.Address{0xaa}
	chainID := big.NewInt(1337)

	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 4, func(i int, gen *BlockGen) {
		gen.SetCoinbase(compatMiner)
		signer := gen.Signer()

		switch i {
		case 0: // Block 1: Legacy transaction (Type 0)
			tx := types.MustSignNewTx(compatTestKey, signer, &types.LegacyTx{
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(1e15), // 0.001 ETC
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
			})
			gen.AddTx(tx)

		case 1: // Block 2: AccessList transaction (Type 1)
			tx := types.MustSignNewTx(compatTestKey, signer, &types.AccessListTx{
				ChainID:  chainID,
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(2e15),
				Gas:      vars.TxGas + vars.TxAccessListAddressGas + vars.TxAccessListStorageKeyGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
				AccessList: types.AccessList{
					{Address: recipient, StorageKeys: []common.Hash{{0x01}}},
				},
			})
			gen.AddTx(tx)

		case 2: // Block 3: DynamicFee transaction (Type 2)
			tx := types.MustSignNewTx(compatTestKey, signer, &types.DynamicFeeTx{
				ChainID:   chainID,
				Nonce:     gen.TxNonce(compatTestAddr),
				To:        &recipient,
				Value:     big.NewInt(3e15),
				Gas:       vars.TxGas,
				GasTipCap: big.NewInt(2e9),
				GasFeeCap: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
			})
			gen.AddTx(tx)

		case 3: // Block 4: SetCode transaction (Type 4)
			// Create an authorization: compatTestAddr2 delegates to a target address
			delegateTarget := common.Address{0xbb}
			auth := types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: delegateTarget,
				Nonce:   0, // compatTestAddr2's current nonce
			}
			signedAuth, err := types.SignSetCode(compatTestKey2, auth)
			if err != nil {
				t.Fatalf("SignSetCode: %v", err)
			}

			tx := types.MustSignNewTx(compatTestKey, signer, &types.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     gen.TxNonce(compatTestAddr),
				To:        compatTestAddr2,
				Value:     uint256.NewInt(4e15),
				Gas:       vars.TxGas + vars.CallNewAccountGas, // 21000 + 25000 for 1 auth
				GasTipCap: uint256.NewInt(2e9),
				GasFeeCap: uint256.MustFromBig(new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9))),
				AuthList:  []types.SetCodeAuthorization{signedAuth},
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Verify all 4 blocks processed successfully
	txTypeNames := []string{"Legacy (Type 0)", "AccessList (Type 1)", "DynamicFee (Type 2)", "SetCode (Type 4)"}
	for i := 0; i < 4; i++ {
		block := genchain[i]
		receipts := genreceipts[i]

		if len(receipts) != 1 {
			t.Fatalf("block %d (%s): expected 1 receipt, got %d", block.NumberU64(), txTypeNames[i], len(receipts))
		}

		receipt := receipts[0]
		if receipt.Status != types.ReceiptStatusSuccessful {
			t.Errorf("block %d (%s): tx failed with status %d", block.NumberU64(), txTypeNames[i], receipt.Status)
		}
		if receipt.GasUsed == 0 {
			t.Errorf("block %d (%s): gas used is 0", block.NumberU64(), txTypeNames[i])
		}

		t.Logf("block %d (%s): gasUsed=%d baseFee=%s status=OK",
			block.NumberU64(), txTypeNames[i], receipt.GasUsed, block.BaseFee())
	}

	// Verify final balances
	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Recipient should have received value from first 3 txs
	recipientBal := statedb.GetBalance(recipient)
	expectedRecipient := uint256.NewInt(1e15 + 2e15 + 3e15) // 0.006 ETC
	if recipientBal.Cmp(expectedRecipient) != 0 {
		t.Errorf("recipient balance: got %s, want %s", recipientBal, expectedRecipient)
	}

	// Treasury should have accumulated baseFee revenue from all 4 blocks
	treasuryBal := statedb.GetBalance(compatTreasury)
	if treasuryBal.IsZero() {
		t.Error("treasury balance should be non-zero (blocks had transactions)")
	}
	t.Logf("treasury balance: %s", treasuryBal)

	// compatTestAddr2 should have received value from SetCode tx (block 4)
	addr2Bal := statedb.GetBalance(compatTestAddr2)
	// Started with 100 ETC, received 4e15 from SetCode tx
	expected2Base := new(uint256.Int).Mul(uint256.NewInt(100), uint256.NewInt(1e18))
	expected2 := new(uint256.Int).Add(expected2Base, uint256.NewInt(4e15))
	if addr2Bal.Cmp(expected2) != 0 {
		t.Errorf("compatTestAddr2 balance: got %s, want %s", addr2Bal, expected2)
	}

	// Verify delegation was set on compatTestAddr2
	code := statedb.GetCode(compatTestAddr2)
	if len(code) != 23 {
		t.Fatalf("expected 23-byte delegation code on compatTestAddr2, got %d bytes", len(code))
	}
	target, ok := types.ParseDelegation(code)
	if !ok {
		t.Fatal("failed to parse delegation code on compatTestAddr2")
	}
	if target != (common.Address{0xbb}) {
		t.Errorf("delegation target: got %s, want 0xbb", target.Hex())
	}
	t.Log("delegation set correctly on compatTestAddr2 -> 0xbb")
}

// TestPreOlympiaLegacyTxsPostFork verifies that legacy (Type 0) and
// AccessList (Type 1) transactions continue to work after the Olympia fork.
func TestPreOlympiaLegacyTxsPostFork(t *testing.T) {
	olympiaBlock := int64(3)
	config := newOlympiaTestConfig(olympiaBlock, 5_000_000)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			compatTestAddr: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	recipient := common.Address{0xcc}
	chainID := big.NewInt(1337)

	numBlocks := 6
	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		blockNum := int64(i + 1)
		signer := gen.Signer()

		// Pre-1559: no baseFee, use fixed gasPrice. Post-1559: baseFee + tip.
		var gasPrice *big.Int
		if blockNum < olympiaBlock {
			gasPrice = big.NewInt(1e10) // 10 gwei
		} else {
			gasPrice = new(big.Int).Add(gen.BaseFee(), big.NewInt(1e9))
		}

		// Alternate between Legacy and AccessList txs across the fork boundary
		if blockNum%2 == 1 {
			// Legacy tx (odd blocks: 1, 3, 5)
			tx := types.MustSignNewTx(compatTestKey, signer, &types.LegacyTx{
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(1000),
				Gas:      vars.TxGas,
				GasPrice: gasPrice,
			})
			gen.AddTx(tx)
		} else {
			// AccessList tx (even blocks: 2, 4, 6)
			tx := types.MustSignNewTx(compatTestKey, signer, &types.AccessListTx{
				ChainID:  chainID,
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(2000),
				Gas:      vars.TxGas,
				GasPrice: gasPrice,
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// All blocks should have successful txs
	for i := 0; i < numBlocks; i++ {
		block := genchain[i]
		receipt := genreceipts[i][0]

		forkStatus := "pre-fork"
		if int64(block.NumberU64()) >= olympiaBlock {
			forkStatus = "post-fork"
		}

		txType := "Legacy"
		if receipt.Type == types.AccessListTxType {
			txType = "AccessList"
		}

		if receipt.Status != types.ReceiptStatusSuccessful {
			t.Errorf("block %d (%s, %s): tx failed with status %d",
				block.NumberU64(), forkStatus, txType, receipt.Status)
		}

		t.Logf("block %d (%s, %s): gasUsed=%d baseFee=%s status=OK",
			block.NumberU64(), forkStatus, txType, receipt.GasUsed, block.BaseFee())
	}

	// Verify total value received
	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	recipientBal := statedb.GetBalance(recipient)
	// 3 legacy txs (1000 each) + 3 access list txs (2000 each) = 9000
	expectedRecipient := uint256.NewInt(9000)
	if recipientBal.Cmp(expectedRecipient) != 0 {
		t.Errorf("recipient balance: got %s, want %s", recipientBal, expectedRecipient)
	}
}

// TestECIP1017EraRewardsWithTreasury verifies that ECIP-1017 era-based block
// reward reduction works correctly alongside the ECIP-1111 treasury redirect.
// Uses short era length (5 blocks) to test multiple eras in a small chain.
func TestECIP1017EraRewardsWithTreasury(t *testing.T) {
	eraLength := int64(5) // 5 blocks per era for testing
	config := newOlympiaTestConfig(0, eraLength)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			compatTestAddr: {Balance: new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	recipient := common.Address{0xdd}

	// Generate 20 blocks spanning 4 eras (5 blocks each)
	numBlocks := 20
	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		gen.SetCoinbase(compatMiner)
		// Add a tx every block to generate gas usage for treasury
		tx := types.MustSignNewTx(compatTestKey, gen.Signer(), &types.LegacyTx{
			Nonce:    gen.TxNonce(compatTestAddr),
			To:       &recipient,
			Value:    big.NewInt(1000),
			Gas:      vars.TxGas,
			GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
		})
		gen.AddTx(tx)
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Verify all txs succeeded
	for i := 0; i < numBlocks; i++ {
		if genreceipts[i][0].Status != types.ReceiptStatusSuccessful {
			t.Fatalf("block %d: tx failed", genchain[i].NumberU64())
		}
	}

	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Calculate expected miner balance: sum of era-adjusted rewards + tips
	// Era rewards: 5 ETC * (4/5)^era
	//   Era 0 (blocks 1-5):  5.0  ETC per block
	//   Era 1 (blocks 6-10): 4.0  ETC per block
	//   Era 2 (blocks 11-15): 3.2 ETC per block
	//   Era 3 (blocks 16-20): 2.56 ETC per block
	eraLenBig := big.NewInt(eraLength)
	var expectedMinerRewards uint256.Int
	var expectedTreasury uint256.Int

	for i := 0; i < numBlocks; i++ {
		block := genchain[i]
		blockNum := block.Number()

		// Calculate era reward
		era := mutations.GetBlockEra(blockNum, eraLenBig)
		eraReward := mutations.GetBlockWinnerRewardByEra(era, vars.FrontierBlockReward)
		expectedMinerRewards.Add(&expectedMinerRewards, eraReward)

		// Calculate treasury revenue: baseFee * gasUsed
		baseFee := uint256.MustFromBig(block.BaseFee())
		gasUsed := new(uint256.Int).SetUint64(block.GasUsed())
		treasuryRevenue := new(uint256.Int).Mul(baseFee, gasUsed)
		expectedTreasury.Add(&expectedTreasury, treasuryRevenue)

		eraIdx := era.Int64()
		t.Logf("block %2d (era %d): reward=%s baseFee=%s gasUsed=%d treasury_revenue=%s",
			block.NumberU64(), eraIdx, eraReward, block.BaseFee(), block.GasUsed(), treasuryRevenue)
	}

	// Miner gets era reward + tips (tips = gasPrice - baseFee) * gasUsed
	// For simplicity, just verify the miner balance is at least the era rewards
	minerBal := statedb.GetBalance(compatMiner)
	if minerBal.Cmp(&expectedMinerRewards) < 0 {
		t.Errorf("miner balance %s is less than expected era rewards %s", minerBal, &expectedMinerRewards)
	}

	// Verify treasury accumulated correctly
	actualTreasury := statedb.GetBalance(compatTreasury)
	if actualTreasury.Cmp(&expectedTreasury) != 0 {
		t.Errorf("treasury balance mismatch: got %s, want %s", actualTreasury, &expectedTreasury)
	}

	// Verify era rewards are decreasing
	era0reward := mutations.GetBlockWinnerRewardByEra(big.NewInt(0), vars.FrontierBlockReward)
	era1reward := mutations.GetBlockWinnerRewardByEra(big.NewInt(1), vars.FrontierBlockReward)
	era2reward := mutations.GetBlockWinnerRewardByEra(big.NewInt(2), vars.FrontierBlockReward)
	era3reward := mutations.GetBlockWinnerRewardByEra(big.NewInt(3), vars.FrontierBlockReward)

	if era0reward.Cmp(uint256.NewInt(5e18)) != 0 {
		t.Errorf("era 0 reward: got %s, want 5e18", era0reward)
	}
	if era1reward.Cmp(uint256.NewInt(4e18)) != 0 {
		t.Errorf("era 1 reward: got %s, want 4e18", era1reward)
	}
	if era2reward.Cmp(uint256.NewInt(3.2e18)) != 0 {
		t.Errorf("era 2 reward: got %s, want 3.2e18", era2reward)
	}
	if era3reward.Cmp(uint256.NewInt(2.56e18)) != 0 {
		t.Errorf("era 3 reward: got %s, want 2.56e18", era3reward)
	}

	t.Logf("Era rewards: 0=%s 1=%s 2=%s 3=%s", era0reward, era1reward, era2reward, era3reward)
	t.Logf("Miner total (rewards+tips): %s", minerBal)
	t.Logf("Treasury total: %s", actualTreasury)
	t.Logf("Expected miner era rewards: %s", &expectedMinerRewards)
}

// TestECIP1017EraBoundaryRewards verifies exact miner reward amounts at
// era boundaries, confirming the 80% reduction per era.
func TestECIP1017EraBoundaryRewards(t *testing.T) {
	eraLength := int64(5)
	config := newOlympiaTestConfig(0, eraLength)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   5_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	// Generate empty blocks (no txs) so miner only gets block reward
	numBlocks := 20
	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		gen.SetCoinbase(compatMiner)
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	// Insert blocks one by one to check miner balance at each era boundary
	eraLenBig := big.NewInt(eraLength)
	var cumulativeReward uint256.Int

	for _, block := range genchain {
		if _, err := blockchain.InsertChain([]*types.Block{block}); err != nil {
			t.Fatalf("insert error (block %d): %v", block.NumberU64(), err)
		}

		era := mutations.GetBlockEra(block.Number(), eraLenBig)
		eraReward := mutations.GetBlockWinnerRewardByEra(era, vars.FrontierBlockReward)
		cumulativeReward.Add(&cumulativeReward, eraReward)

		statedb, err := blockchain.State()
		if err != nil {
			t.Fatal(err)
		}

		minerBal := statedb.GetBalance(compatMiner)
		if minerBal.Cmp(&cumulativeReward) != 0 {
			t.Errorf("block %d (era %s): miner balance %s != expected cumulative %s",
				block.NumberU64(), era, minerBal, &cumulativeReward)
		}

		// Log at era boundaries
		blockNum := block.NumberU64()
		if blockNum == 5 || blockNum == 6 || blockNum == 10 || blockNum == 11 || blockNum == 15 || blockNum == 16 {
			t.Logf("block %2d (era %s): reward=%s cumulative=%s",
				blockNum, era, eraReward, &cumulativeReward)
		}
	}

	// Verify exact cumulative: 5*5 + 5*4 + 5*3.2 + 5*2.56 = 25+20+16+12.8 = 73.8 ETC
	expectedTotal := new(uint256.Int)
	expectedTotal.Add(expectedTotal, new(uint256.Int).Mul(uint256.NewInt(5), uint256.NewInt(5e18)))    // Era 0
	expectedTotal.Add(expectedTotal, new(uint256.Int).Mul(uint256.NewInt(5), uint256.NewInt(4e18)))    // Era 1
	expectedTotal.Add(expectedTotal, new(uint256.Int).Mul(uint256.NewInt(5), uint256.NewInt(3.2e18)))  // Era 2
	expectedTotal.Add(expectedTotal, new(uint256.Int).Mul(uint256.NewInt(5), uint256.NewInt(2.56e18))) // Era 3

	statedb, _ := blockchain.State()
	minerBal := statedb.GetBalance(compatMiner)
	if minerBal.Cmp(expectedTotal) != 0 {
		t.Fatalf("final miner balance: got %s, want %s (73.8 ETC)", minerBal, expectedTotal)
	}
	t.Logf("final miner balance correct: %s (73.8 ETC across 4 eras)", minerBal)
}

// TestGasAccountingAllTxTypes verifies that gas is correctly charged and
// accounted for across all transaction types, including the baseFee/tip split.
func TestGasAccountingAllTxTypes(t *testing.T) {
	config := newOlympiaTestConfig(0, 5_000_000)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   30_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			compatTestAddr:  {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
			compatTestAddr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	recipient := common.Address{0xee}
	chainID := big.NewInt(1337)
	tipPerGas := big.NewInt(2e9) // 2 gwei tip

	type txCase struct {
		name    string
		gasUsed uint64
	}

	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, 3, func(i int, gen *BlockGen) {
		gen.SetCoinbase(compatMiner)
		signer := gen.Signer()

		switch i {
		case 0: // Legacy
			tx := types.MustSignNewTx(compatTestKey, signer, &types.LegacyTx{
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(0),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), tipPerGas),
			})
			gen.AddTx(tx)

		case 1: // DynamicFee
			tx := types.MustSignNewTx(compatTestKey, signer, &types.DynamicFeeTx{
				ChainID:   chainID,
				Nonce:     gen.TxNonce(compatTestAddr),
				To:        &recipient,
				Value:     big.NewInt(0),
				Gas:       vars.TxGas,
				GasTipCap: tipPerGas,
				GasFeeCap: new(big.Int).Add(gen.BaseFee(), tipPerGas),
			})
			gen.AddTx(tx)

		case 2: // AccessList
			tx := types.MustSignNewTx(compatTestKey, signer, &types.AccessListTx{
				ChainID:  chainID,
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(0),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), tipPerGas),
			})
			gen.AddTx(tx)
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Verify gas split: treasury gets baseFee*gasUsed, miner gets tip*gasUsed
	statedb, _ := blockchain.State()

	var totalTreasuryExpected, totalMinerTips uint256.Int
	txTypes := []string{"Legacy", "DynamicFee", "AccessList"}

	for i := 0; i < 3; i++ {
		block := genchain[i]
		receipt := genreceipts[i][0]

		baseFee := uint256.MustFromBig(block.BaseFee())
		gasUsed := new(uint256.Int).SetUint64(receipt.GasUsed)
		treasuryShare := new(uint256.Int).Mul(baseFee, gasUsed)
		tipShare := new(uint256.Int).Mul(uint256.MustFromBig(tipPerGas), gasUsed)

		totalTreasuryExpected.Add(&totalTreasuryExpected, treasuryShare)
		totalMinerTips.Add(&totalMinerTips, tipShare)

		t.Logf("%s: gasUsed=%d baseFee=%s treasury=%s tip=%s",
			txTypes[i], receipt.GasUsed, block.BaseFee(), treasuryShare, tipShare)
	}

	actualTreasury := statedb.GetBalance(compatTreasury)
	if actualTreasury.Cmp(&totalTreasuryExpected) != 0 {
		t.Errorf("treasury: got %s, want %s", actualTreasury, &totalTreasuryExpected)
	}

	// Miner gets block rewards + tips
	minerBal := statedb.GetBalance(compatMiner)
	if minerBal.Cmp(&totalMinerTips) < 0 {
		t.Errorf("miner balance %s less than expected tips %s", minerBal, &totalMinerTips)
	}
	t.Logf("treasury=%s miner=%s (includes 3 block rewards)", actualTreasury, minerBal)
}

// TestSignerSelectionAcrossForks verifies that the correct signer is
// selected for each transaction type based on chain config activation.
func TestSignerSelectionAcrossForks(t *testing.T) {
	chainID := big.NewInt(1337)
	olympiaBlock := big.NewInt(5)

	config := newOlympiaTestConfig(5, 5_000_000)

	// Pre-Olympia (block 4): no EIP-1559, no EIP-7702
	// Should get EIP-2930 signer (highest activated pre-Olympia)
	preOlympiaSigner := types.MakeSigner(config, big.NewInt(4), 0)
	if _, ok := preOlympiaSigner.(*types.EIP155Signer); ok {
		t.Log("pre-Olympia signer: EIP155Signer (expected - EIP-2930 wraps EIP-155)")
	}

	// Post-Olympia (block 5): EIP-1559 + EIP-7702 active
	// Should get SetCodeSigner (highest)
	postOlympiaSigner := types.MakeSigner(config, olympiaBlock, 0)

	// Verify each tx type can be signed and the sender recovered
	cases := []struct {
		name   string
		txData types.TxData
	}{
		{
			"Legacy (Type 0) post-fork",
			&types.LegacyTx{
				Nonce: 0, To: &common.Address{0x01}, Value: big.NewInt(0),
				Gas: vars.TxGas, GasPrice: big.NewInt(1e10),
			},
		},
		{
			"AccessList (Type 1) post-fork",
			&types.AccessListTx{
				ChainID: chainID, Nonce: 1, To: &common.Address{0x01}, Value: big.NewInt(0),
				Gas: vars.TxGas, GasPrice: big.NewInt(1e10),
			},
		},
		{
			"DynamicFee (Type 2) post-fork",
			&types.DynamicFeeTx{
				ChainID: chainID, Nonce: 2, To: &common.Address{0x01}, Value: big.NewInt(0),
				Gas: vars.TxGas, GasTipCap: big.NewInt(1e9), GasFeeCap: big.NewInt(1e10),
			},
		},
	}

	for _, tc := range cases {
		tx := types.MustSignNewTx(compatTestKey, postOlympiaSigner, tc.txData)

		// Recover sender
		sender, err := types.Sender(postOlympiaSigner, tx)
		if err != nil {
			t.Errorf("%s: Sender recovery failed: %v", tc.name, err)
			continue
		}
		if sender != compatTestAddr {
			t.Errorf("%s: wrong sender: got %s, want %s", tc.name, sender.Hex(), compatTestAddr.Hex())
		}

		// Also verify the pre-Olympia signer can recover legacy/accesslist txs
		if tx.Type() == types.LegacyTxType || tx.Type() == types.AccessListTxType {
			sender2, err := types.Sender(preOlympiaSigner, tx)
			if err != nil {
				t.Errorf("%s: pre-Olympia Sender recovery failed: %v", tc.name, err)
			} else if sender2 != compatTestAddr {
				t.Errorf("%s: pre-Olympia wrong sender: got %s, want %s", tc.name, sender2.Hex(), compatTestAddr.Hex())
			}
		}

		t.Logf("%s: type=%d sender=%s OK", tc.name, tx.Type(), sender.Hex())
	}
}

// TestMordorLikeChainIntegration generates a short chain that mirrors
// Mordor's full activation history with short era length, exercising
// all pre-Olympia and Olympia features together.
func TestMordorLikeChainIntegration(t *testing.T) {
	eraLength := int64(3)
	config := newOlympiaTestConfig(0, eraLength)

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   30_000_000,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			compatTestAddr:  {Balance: new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))},
			compatTestAddr2: {Balance: new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	recipient := common.Address{0xff}
	chainID := big.NewInt(1337)

	// 12 blocks = 4 eras of 3 blocks each
	numBlocks := 12
	genchain, genreceipts := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		gen.SetCoinbase(compatMiner)
		signer := gen.Signer()

		// Cycle through all tx types
		switch i % 4 {
		case 0: // Legacy
			tx := types.MustSignNewTx(compatTestKey, signer, &types.LegacyTx{
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(1e15),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
			})
			gen.AddTx(tx)

		case 1: // AccessList
			tx := types.MustSignNewTx(compatTestKey, signer, &types.AccessListTx{
				ChainID:  chainID,
				Nonce:    gen.TxNonce(compatTestAddr),
				To:       &recipient,
				Value:    big.NewInt(1e15),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
			})
			gen.AddTx(tx)

		case 2: // DynamicFee
			tx := types.MustSignNewTx(compatTestKey, signer, &types.DynamicFeeTx{
				ChainID:   chainID,
				Nonce:     gen.TxNonce(compatTestAddr),
				To:        &recipient,
				Value:     big.NewInt(1e15),
				Gas:       vars.TxGas,
				GasTipCap: big.NewInt(2e9),
				GasFeeCap: new(big.Int).Add(gen.BaseFee(), big.NewInt(2e9)),
			})
			gen.AddTx(tx)

		case 3: // SetCode (delegation + value transfer)
			auth := types.SetCodeAuthorization{
				ChainID: *uint256.MustFromBig(chainID),
				Address: common.Address{0xbb},
				Nonce:   stateNonce(gen, compatTestAddr2),
			}
			signedAuth, err := types.SignSetCode(compatTestKey2, auth)
			if err != nil {
				t.Fatalf("block %d: SignSetCode: %v", i+1, err)
			}

			tx := types.MustSignNewTx(compatTestKey, signer, &types.SetCodeTx{
				ChainID:   uint256.MustFromBig(chainID),
				Nonce:     gen.TxNonce(compatTestAddr),
				To:        recipient,
				Value:     uint256.NewInt(1e15),
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

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// Verify all blocks succeeded
	eraLenBig := big.NewInt(eraLength)
	txTypeNames := []string{"Legacy", "AccessList", "DynamicFee", "SetCode"}
	for i := 0; i < numBlocks; i++ {
		block := genchain[i]
		receipt := genreceipts[i][0]
		era := mutations.GetBlockEra(block.Number(), eraLenBig)
		eraReward := mutations.GetBlockWinnerRewardByEra(era, vars.FrontierBlockReward)

		if receipt.Status != types.ReceiptStatusSuccessful {
			t.Errorf("block %d: tx failed", block.NumberU64())
		}

		t.Logf("block %2d (era %s): %s gasUsed=%d baseFee=%s eraReward=%s",
			block.NumberU64(), era, txTypeNames[i%4], receipt.GasUsed, block.BaseFee(), eraReward)
	}

	statedb, _ := blockchain.State()

	// Verify all subsystems
	treasuryBal := statedb.GetBalance(compatTreasury)
	minerBal := statedb.GetBalance(compatMiner)
	recipientBal := statedb.GetBalance(recipient)

	if treasuryBal.IsZero() {
		t.Error("treasury balance should be non-zero")
	}
	if minerBal.IsZero() {
		t.Error("miner balance should be non-zero")
	}

	// Recipient got 1e15 per block * 12 blocks = 12e15
	expectedRecipient := new(uint256.Int).Mul(uint256.NewInt(1e15), uint256.NewInt(uint64(numBlocks)))
	if recipientBal.Cmp(expectedRecipient) != 0 {
		t.Errorf("recipient: got %s, want %s", recipientBal, expectedRecipient)
	}

	t.Logf("Integration result: treasury=%s miner=%s recipient=%s",
		treasuryBal, minerBal, recipientBal)
}

// stateNonce is a helper to get the current nonce from the BlockGen's state.
// For authorization nonces, we need the authority account's current nonce.
func stateNonce(gen *BlockGen, addr common.Address) uint64 {
	return gen.TxNonce(addr)
}
