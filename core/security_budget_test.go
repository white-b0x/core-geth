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
	"fmt"
	"math/big"
	"strings"
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

// Test keys — one per tx type to keep nonce tracking clean
var (
	budgetLegacyKey, _     = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
	budgetLegacyAddr       = crypto.PubkeyToAddress(budgetLegacyKey.PublicKey)
	budgetAccessListKey, _ = crypto.HexToECDSA("8a1f9a8f95be41cd7ccb6168179afb4504aefe388d1e14474d32c45c72ce7b7a")
	budgetAccessListAddr   = crypto.PubkeyToAddress(budgetAccessListKey.PublicKey)
	budgetDynFeeKey, _     = crypto.HexToECDSA("49a7b37aa6f6645917e7b807e9d1c00d4fa71f18343b0d4122a4d2df64dd6fee")
	budgetDynFeeAddr       = crypto.PubkeyToAddress(budgetDynFeeKey.PublicKey)

	budgetTreasuryAddr = common.HexToAddress("0xDeaDbeefdEAdbeefdEadbEEFdeadbeEFdEaDbeeF")
	budgetMinerAddr    = common.HexToAddress("0x1111111111111111111111111111111111111111")
	budgetRecipient    = common.Address{0xcc}
)

// blockMetrics tracks per-block fee distribution.
type blockMetrics struct {
	number      uint64
	era         int
	blockReward *big.Int // era-adjusted block reward (wei)
	gasUsed     uint64
	gasLimit    uint64
	baseFee     *big.Int
	legacyTips  *big.Int
	alTips      *big.Int // access list
	dynFeeTips  *big.Int // dynamic fee
	treasuryRev *big.Int
	txCount     int
	legacyCount int
	alCount     int
	dynFeeCount int
}

func newBudgetConfig() *coregeth.CoreGethChainConfig {
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
		// Berlin EIPs (needed for AccessList and DynamicFee tx types)
		EIP2565FBlock: big.NewInt(0),
		EIP2718FBlock: big.NewInt(0),
		EIP2929FBlock: big.NewInt(0),
		EIP2930FBlock: big.NewInt(0),
		EIP2028FBlock: big.NewInt(0),
		// London EIPs
		EIP1559FBlock: big.NewInt(0),
		EIP3198FBlock: big.NewInt(0),
		// ECIP-1017 monetary policy (10 blocks/era for test — maps to 5M on mainnet)
		ECIP1017FBlock:    big.NewInt(0),
		ECIP1017EraRounds: big.NewInt(10),
		// Treasury
		OlympiaTreasuryAddress: &budgetTreasuryAddr,
	}
}

// weiToETC formats a *big.Int wei value as ETC with 6 decimal places.
func weiToETC(wei *big.Int) string {
	f := new(big.Float).SetInt(wei)
	f.Quo(f, new(big.Float).SetFloat64(1e18))
	return f.Text('f', 6)
}

// TestSecurityBudgetAnalysis runs a 60-block simulation at Olympia's 60M gas
// limit (EIP-7935) with ECIP-1017 era-based rewards reflecting current ETC
// mainnet conditions (era 4 at ~24M blocks, era 5 approaching at 25M).
//
// Baseline: ETC mainnet averages ~30,000 txs/day ÷ ~6,200 blocks/day ≈ 5 txs/block.
// This floor is used for normal phases to show realistic miner earnings before congestion.
//
// Three scenarios:
//   - Era 4 "Activation": Olympia goes live, 2.048 ETC/block, baseline 5 txs/block
//   - Era 5 "Normal": after 25M, 1.6384 ETC/block, baseline 5 txs/block
//   - Era 5 "Congestion": same era, 2,000 txs/block flooding the network
//
// Blocks 1-40 fast-forward through eras 0-3 (empty). The meaningful simulation
// runs blocks 41-60 with eraLength=10.
func TestSecurityBudgetAnalysis(t *testing.T) {
	config := newBudgetConfig()
	eraRounds := big.NewInt(10)

	// Olympia gas limit per EIP-7935: 60M
	gasLimit := uint64(60_000_000)

	// 100,000 ETC per account — enough for thousands of txs at any gas price
	accountBalance := new(big.Int).Mul(big.NewInt(100_000), big.NewInt(1e18))

	gspec := &genesisT.Genesis{
		Config:     config,
		GasLimit:   gasLimit,
		Difficulty: vars.MinimumDifficulty,
		BaseFee:    big.NewInt(vars.InitialBaseFee),
		Alloc: genesisT.GenesisAlloc{
			budgetLegacyAddr:     {Balance: new(big.Int).Set(accountBalance)},
			budgetAccessListAddr: {Balance: new(big.Int).Set(accountBalance)},
			budgetDynFeeAddr:     {Balance: new(big.Int).Set(accountBalance)},
		},
	}

	gendb := rawdb.NewMemoryDatabase()
	db := rawdb.NewMemoryDatabase()
	genesis := MustCommitGenesis(gendb, triedb.NewDatabase(gendb, triedb.HashDefaults), gspec)

	numBlocks := 60
	tipAmount := big.NewInt(2_000_000_000) // 2 Gwei tip

	// ETC mainnet floor: ~30,000 txs/day ÷ ~6,200 blocks/day ≈ 5 txs/block
	baselineTxsPerBlock := 5

	t.Log("Generating 60 blocks (40 fast-forward + 20 meaningful) at 60M gas limit...")

	genchain, _ := GenerateChain(gspec.Config, genesis, ethash.NewFaker(), gendb, numBlocks, func(i int, gen *BlockGen) {
		gen.SetCoinbase(budgetMinerAddr)
		blockNum := i + 1

		switch {
		// Blocks 1-40: Fast-forward through eras 0-3 (empty)
		case blockNum <= 40:
			// No transactions — advancing era counter only

		// Era 4 "Activation" (blocks 41-50, 2.048 ETC/block)
		// Baseline traffic: 5 txs/block (30K txs/day ÷ 6,200 blocks/day)
		// Mixed types: 3 Legacy + 1 AccessList + 1 DynamicFee
		case blockNum <= 50:
			for j := 0; j < 3; j++ {
				tx := types.MustSignNewTx(budgetLegacyKey, gen.Signer(), &types.LegacyTx{
					Nonce:    gen.TxNonce(budgetLegacyAddr),
					To:       &budgetRecipient,
					Value:    big.NewInt(1000),
					Gas:      vars.TxGas,
					GasPrice: new(big.Int).Add(gen.BaseFee(), tipAmount),
				})
				gen.AddTx(tx)
			}
			tx := types.MustSignNewTx(budgetAccessListKey, gen.Signer(), &types.AccessListTx{
				ChainID:  config.ChainID,
				Nonce:    gen.TxNonce(budgetAccessListAddr),
				To:       &budgetRecipient,
				Value:    big.NewInt(1000),
				Gas:      vars.TxGas,
				GasPrice: new(big.Int).Add(gen.BaseFee(), tipAmount),
			})
			gen.AddTx(tx)
			tx = types.MustSignNewTx(budgetDynFeeKey, gen.Signer(), &types.DynamicFeeTx{
				ChainID:   config.ChainID,
				Nonce:     gen.TxNonce(budgetDynFeeAddr),
				To:        &budgetRecipient,
				Value:     big.NewInt(1000),
				Gas:       vars.TxGas,
				GasFeeCap: new(big.Int).Add(gen.BaseFee(), new(big.Int).Mul(tipAmount, big.NewInt(3))),
				GasTipCap: tipAmount,
			})
			gen.AddTx(tx)

		// Era 5 "Normal" (blocks 51-55, 1.6384 ETC/block, baseline traffic)
		// 5 Legacy txs/block = 105K gas = 0.175% utilization (30K txs/day floor)
		case blockNum <= 55:
			for j := 0; j < baselineTxsPerBlock; j++ {
				tx := types.MustSignNewTx(budgetLegacyKey, gen.Signer(), &types.LegacyTx{
					Nonce:    gen.TxNonce(budgetLegacyAddr),
					To:       &budgetRecipient,
					Value:    big.NewInt(1000),
					Gas:      vars.TxGas,
					GasPrice: new(big.Int).Add(gen.BaseFee(), tipAmount),
				})
				gen.AddTx(tx)
			}

		// Era 5 "Congestion" (blocks 56-60, 1.6384 ETC/block, heavy traffic)
		// 1,800 DynFee + 200 Legacy = 2,000 txs/block = 42M gas = 70% util
		default:
			for j := 0; j < 1800; j++ {
				tx := types.MustSignNewTx(budgetDynFeeKey, gen.Signer(), &types.DynamicFeeTx{
					ChainID:   config.ChainID,
					Nonce:     gen.TxNonce(budgetDynFeeAddr),
					To:        &budgetRecipient,
					Value:     big.NewInt(1000),
					Gas:       vars.TxGas,
					GasFeeCap: new(big.Int).Add(gen.BaseFee(), new(big.Int).Mul(tipAmount, big.NewInt(5))),
					GasTipCap: tipAmount,
				})
				gen.AddTx(tx)
			}
			for j := 0; j < 200; j++ {
				tx := types.MustSignNewTx(budgetLegacyKey, gen.Signer(), &types.LegacyTx{
					Nonce:    gen.TxNonce(budgetLegacyAddr),
					To:       &budgetRecipient,
					Value:    big.NewInt(1000),
					Gas:      vars.TxGas,
					GasPrice: new(big.Int).Add(gen.BaseFee(), tipAmount),
				})
				gen.AddTx(tx)
			}
		}
	})

	blockchain, _ := NewBlockChain(db, nil, gspec, nil, ethash.NewFaker(), vm.Config{}, nil, nil)
	defer blockchain.Stop()

	if i, err := blockchain.InsertChain(genchain); err != nil {
		t.Fatalf("insert error (block %d): %v", genchain[i].NumberU64(), err)
	}

	// ── Compute per-block metrics ──
	metrics := make([]blockMetrics, numBlocks)
	for i, block := range genchain {
		m := &metrics[i]
		m.number = block.NumberU64()
		m.gasUsed = block.GasUsed()
		m.gasLimit = block.GasLimit()
		m.baseFee = new(big.Int).Set(block.BaseFee())
		m.legacyTips = new(big.Int)
		m.alTips = new(big.Int)
		m.dynFeeTips = new(big.Int)
		m.treasuryRev = new(big.Int).Mul(block.BaseFee(), new(big.Int).SetUint64(block.GasUsed()))

		// Compute ECIP-1017 era and per-block reward
		era := mutations.GetBlockEra(block.Number(), eraRounds)
		m.era = int(era.Int64())
		eraReward := mutations.GetBlockWinnerRewardByEra(era, vars.FrontierBlockReward)
		m.blockReward = eraReward.ToBig()

		for _, tx := range block.Transactions() {
			m.txCount++
			effectiveTip := new(big.Int)
			gasUsed := new(big.Int).SetUint64(vars.TxGas)
			switch tx.Type() {
			case types.LegacyTxType:
				m.legacyCount++
				effectiveTip.Sub(tx.GasPrice(), block.BaseFee())
				m.legacyTips.Add(m.legacyTips, new(big.Int).Mul(effectiveTip, gasUsed))
			case types.AccessListTxType:
				m.alCount++
				effectiveTip.Sub(tx.GasPrice(), block.BaseFee())
				m.alTips.Add(m.alTips, new(big.Int).Mul(effectiveTip, gasUsed))
			case types.DynamicFeeTxType:
				m.dynFeeCount++
				maxTip := new(big.Int).Sub(tx.GasFeeCap(), block.BaseFee())
				if tx.GasTipCap().Cmp(maxTip) < 0 {
					effectiveTip.Set(tx.GasTipCap())
				} else {
					effectiveTip.Set(maxTip)
				}
				m.dynFeeTips.Add(m.dynFeeTips, new(big.Int).Mul(effectiveTip, gasUsed))
			}
		}
	}

	// ── Verify state ──
	statedb, err := blockchain.State()
	if err != nil {
		t.Fatal(err)
	}

	// Treasury verification: on-chain balance must match sum(baseFee * gasUsed)
	var expectedTreasury uint256.Int
	for _, m := range metrics {
		rev := uint256.MustFromBig(m.treasuryRev)
		expectedTreasury.Add(&expectedTreasury, rev)
	}
	actualTreasury := statedb.GetBalance(budgetTreasuryAddr)
	if actualTreasury.Cmp(&expectedTreasury) != 0 {
		t.Fatalf("treasury mismatch: got %s, want %s", actualTreasury, &expectedTreasury)
	}

	// BaseFee assertion: must rise during congestion (blocks 56-60) vs pre-congestion
	preCongestionBaseFee := metrics[54].baseFee // block 55 (last normal block)
	congestionPeak := new(big.Int)
	for i := 55; i <= 59; i++ { // blocks 56-60 (0-indexed)
		if metrics[i].baseFee.Cmp(congestionPeak) > 0 {
			congestionPeak = new(big.Int).Set(metrics[i].baseFee)
		}
	}
	if congestionPeak.Cmp(preCongestionBaseFee) <= 0 {
		t.Errorf("baseFee did not rise during congestion: pre=%s, peak=%s", preCongestionBaseFee, congestionPeak)
	}

	// Era verification: block 41 should be era 4, block 51 should be era 5
	if metrics[40].era != 4 {
		t.Errorf("block 41 expected era 4, got era %d", metrics[40].era)
	}
	if metrics[50].era != 5 {
		t.Errorf("block 51 expected era 5, got era %d", metrics[50].era)
	}

	// ── Generate Report ──
	var sb strings.Builder
	eip1559Target := gasLimit / 2

	// Era reward values for display
	era4Reward := mutations.GetBlockWinnerRewardByEra(big.NewInt(4), vars.FrontierBlockReward)
	era5Reward := mutations.GetBlockWinnerRewardByEra(big.NewInt(5), vars.FrontierBlockReward)

	sb.WriteString("\n")
	sb.WriteString("╔═══════════════════════════════════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║                         ETC OLYMPIA SECURITY BUDGET REPORT                                   ║\n")
	sb.WriteString("║                         ECIP-1017 Era 4/5 — Current Mainnet Conditions                       ║\n")
	sb.WriteString("╠═══════════════════════════════════════════════════════════════════════════════════════════════╣\n")
	sb.WriteString("\n")
	sb.WriteString("  CHAIN PARAMETERS:\n")
	sb.WriteString(fmt.Sprintf("    Gas Limit: %d (EIP-7935) | EIP-1559 Target: %d (limit/2)\n", gasLimit, eip1559Target))
	sb.WriteString(fmt.Sprintf("    Txs to exceed target: %d simple transfers\n", eip1559Target/vars.TxGas))
	sb.WriteString(fmt.Sprintf("    Treasury: %s (ECIP-1111/1112)\n", budgetTreasuryAddr.Hex()))
	sb.WriteString(fmt.Sprintf("    Tip: 2 Gwei | Miner: %s\n", budgetMinerAddr.Hex()))
	sb.WriteString("\n")
	sb.WriteString("  ECIP-1017 ERA SCHEDULE (5M blocks/era on mainnet, 10 blocks/era in test):\n")
	sb.WriteString(fmt.Sprintf("    Era 4 (mainnet blocks 20M-25M): %s ETC/block  ← CURRENT (Olympia activates here)\n", weiToETC(era4Reward.ToBig())))
	sb.WriteString(fmt.Sprintf("    Era 5 (mainnet blocks 25M-30M): %s ETC/block  ← ~3 months after activation\n", weiToETC(era5Reward.ToBig())))
	sb.WriteString("\n")

	// Phase summary
	sb.WriteString("  BASELINE: 30,000 txs/day ÷ ~6,200 blocks/day ≈ 5 txs/block (ETC mainnet floor)\n")
	sb.WriteString("\n")
	sb.WriteString("  SIMULATION PHASES (blocks 1-40 fast-forward through eras 0-3):\n")
	sb.WriteString("    Phase 1 (blocks 41-50): Era 4 Activation      5 txs/block    105K gas     baseline mixed\n")
	sb.WriteString("    Phase 2 (blocks 51-55): Era 5 Normal           5 txs/block    105K gas     baseline\n")
	sb.WriteString("    Phase 3 (blocks 56-60): Era 5 Congestion   2,000 txs/block   42.0M gas     baseFee RISES\n")
	sb.WriteString("\n")

	// Per-block detail (only blocks 41-60)
	sb.WriteString("  PER-BLOCK DETAIL (blocks 41-60):\n")
	sb.WriteString("  Block  Era  Reward(ETC)  Txs    GasUsed      Util%%    BaseFee(Gwei)  Legacy  AL   DynFee  Treasury(ETC)\n")
	sb.WriteString("  ─────  ───  ──────────   ────   ──────────   ──────   ─────────────  ──────  ───  ──────  ─────────────\n")

	// Phase accumulators
	type phaseStats struct {
		name        string
		blocks      int
		txs         int
		rewards     *big.Int
		tips        *big.Int
		treasury    *big.Int
		legacyTips  *big.Int
		alTips      *big.Int
		dynFeeTips  *big.Int
		legacyCount int
		alCount     int
		dynFeeCount int
	}
	phases := []*phaseStats{
		{name: "Era 4 Activation", rewards: new(big.Int), tips: new(big.Int), treasury: new(big.Int), legacyTips: new(big.Int), alTips: new(big.Int), dynFeeTips: new(big.Int)},
		{name: "Era 5 Normal", rewards: new(big.Int), tips: new(big.Int), treasury: new(big.Int), legacyTips: new(big.Int), alTips: new(big.Int), dynFeeTips: new(big.Int)},
		{name: "Era 5 Congestion", rewards: new(big.Int), tips: new(big.Int), treasury: new(big.Int), legacyTips: new(big.Int), alTips: new(big.Int), dynFeeTips: new(big.Int)},
	}

	for i := 40; i < numBlocks; i++ { // blocks 41-60 (0-indexed 40-59)
		m := &metrics[i]
		util := float64(0)
		if m.gasLimit > 0 {
			util = float64(m.gasUsed) / float64(m.gasLimit) * 100
		}
		baseFeeGwei := new(big.Float).Quo(
			new(big.Float).SetInt(m.baseFee),
			new(big.Float).SetFloat64(1e9),
		)

		sb.WriteString(fmt.Sprintf("  %-5d  %-3d  %-11s  %-5d  %-11d  %5.1f%%   %-13s  %-6d  %-3d  %-6d  %s\n",
			m.number, m.era, weiToETC(m.blockReward),
			m.txCount, m.gasUsed, util,
			baseFeeGwei.Text('f', 4),
			m.legacyCount, m.alCount, m.dynFeeCount,
			weiToETC(m.treasuryRev)))

		// Accumulate into phase
		var p *phaseStats
		switch {
		case m.number <= 50:
			p = phases[0]
		case m.number <= 55:
			p = phases[1]
		default:
			p = phases[2]
		}
		p.blocks++
		p.txs += m.txCount
		p.rewards.Add(p.rewards, m.blockReward)
		blockTips := new(big.Int).Add(new(big.Int).Add(m.legacyTips, m.alTips), m.dynFeeTips)
		p.tips.Add(p.tips, blockTips)
		p.treasury.Add(p.treasury, m.treasuryRev)
		p.legacyTips.Add(p.legacyTips, m.legacyTips)
		p.alTips.Add(p.alTips, m.alTips)
		p.dynFeeTips.Add(p.dynFeeTips, m.dynFeeTips)
		p.legacyCount += m.legacyCount
		p.alCount += m.alCount
		p.dynFeeCount += m.dynFeeCount
	}

	// Per-phase summaries
	sb.WriteString("\n")
	sb.WriteString("  PHASE SUMMARIES:\n")
	for _, p := range phases {
		minerIncome := new(big.Int).Add(p.rewards, p.tips)
		totalEconomy := new(big.Int).Add(minerIncome, p.treasury)
		sb.WriteString(fmt.Sprintf("\n    ── %s (%d blocks, %d txs) ──\n", p.name, p.blocks, p.txs))
		sb.WriteString(fmt.Sprintf("      Block rewards:     %s ETC\n", weiToETC(p.rewards)))
		sb.WriteString(fmt.Sprintf("      Tips (all types):  %s ETC  (Legacy: %s, AL: %s, DynFee: %s)\n",
			weiToETC(p.tips), weiToETC(p.legacyTips), weiToETC(p.alTips), weiToETC(p.dynFeeTips)))
		sb.WriteString(fmt.Sprintf("      Total miner:       %s ETC\n", weiToETC(minerIncome)))
		sb.WriteString(fmt.Sprintf("      Treasury:          %s ETC\n", weiToETC(p.treasury)))

		if minerIncome.Sign() > 0 {
			rewardPct, _ := new(big.Float).Quo(
				new(big.Float).Mul(new(big.Float).SetInt(p.rewards), new(big.Float).SetFloat64(100)),
				new(big.Float).SetInt(minerIncome),
			).Float64()
			tipPct, _ := new(big.Float).Quo(
				new(big.Float).Mul(new(big.Float).SetInt(p.tips), new(big.Float).SetFloat64(100)),
				new(big.Float).SetInt(minerIncome),
			).Float64()
			treasuryPct, _ := new(big.Float).Quo(
				new(big.Float).Mul(new(big.Float).SetInt(p.treasury), new(big.Float).SetFloat64(100)),
				new(big.Float).SetInt(totalEconomy),
			).Float64()
			sb.WriteString(fmt.Sprintf("      Rewards %%:         %.2f%% of miner income\n", rewardPct))
			sb.WriteString(fmt.Sprintf("      Tips %%:            %.2f%% of miner income\n", tipPct))
			sb.WriteString(fmt.Sprintf("      Treasury %%:        %.4f%% of total economy\n", treasuryPct))
		}
	}

	// Overall totals (blocks 41-60 only)
	totalRewards := new(big.Int)
	totalTips := new(big.Int)
	totalTreasury := new(big.Int)
	for _, p := range phases {
		totalRewards.Add(totalRewards, p.rewards)
		totalTips.Add(totalTips, p.tips)
		totalTreasury.Add(totalTreasury, p.treasury)
	}
	totalMiner := new(big.Int).Add(totalRewards, totalTips)
	totalEconomy := new(big.Int).Add(totalMiner, totalTreasury)

	sb.WriteString("\n")
	sb.WriteString("  OVERALL (blocks 41-60):\n")
	sb.WriteString(fmt.Sprintf("    Total block rewards:  %s ETC\n", weiToETC(totalRewards)))
	sb.WriteString(fmt.Sprintf("    Total tips:           %s ETC\n", weiToETC(totalTips)))
	sb.WriteString(fmt.Sprintf("    Total miner income:   %s ETC\n", weiToETC(totalMiner)))
	sb.WriteString(fmt.Sprintf("    Total treasury:       %s ETC\n", weiToETC(totalTreasury)))

	if totalMiner.Sign() > 0 {
		rewardDom, _ := new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(totalRewards), new(big.Float).SetFloat64(100)),
			new(big.Float).SetInt(totalMiner),
		).Float64()
		treasuryOfEconomy, _ := new(big.Float).Quo(
			new(big.Float).Mul(new(big.Float).SetInt(totalTreasury), new(big.Float).SetFloat64(100)),
			new(big.Float).SetInt(totalEconomy),
		).Float64()
		sb.WriteString(fmt.Sprintf("    Block reward dominance: %.2f%%\n", rewardDom))
		sb.WriteString(fmt.Sprintf("    Treasury %% of economy:  %.4f%%\n", treasuryOfEconomy))
	}

	// BaseFee trajectory
	startBaseFee := new(big.Float).Quo(new(big.Float).SetInt(metrics[40].baseFee), new(big.Float).SetFloat64(1e9))
	peakBaseFeeF := new(big.Float).Quo(new(big.Float).SetInt(congestionPeak), new(big.Float).SetFloat64(1e9))
	endBaseFee := new(big.Float).Quo(new(big.Float).SetInt(metrics[numBlocks-1].baseFee), new(big.Float).SetFloat64(1e9))
	sb.WriteString("\n")
	sb.WriteString("    BaseFee Trajectory:\n")
	sb.WriteString(fmt.Sprintf("      Block 41 (era 4 start): %s Gwei\n", startBaseFee.Text('f', 4)))
	sb.WriteString(fmt.Sprintf("      Congestion peak:        %s Gwei\n", peakBaseFeeF.Text('f', 4)))
	sb.WriteString(fmt.Sprintf("      Block 60 (end):         %s Gwei\n", endBaseFee.Text('f', 4)))

	// Security budget insight
	sb.WriteString("\n")
	sb.WriteString("    Security Budget Insight:\n")
	sb.WriteString("      At ETC's baseline of ~30K txs/day (5 txs/block), fee revenue is negligible\n")
	sb.WriteString("      — block rewards dominate miner income at both era 4 and era 5 rates.\n")
	sb.WriteString("      At Era 4 (2.048 ETC/block), rewards are 59% lower than Era 1 (5 ETC).\n")
	sb.WriteString("      At Era 5 (1.6384 ETC/block), they drop another 20%. The treasury captures\n")
	sb.WriteString("      baseFee revenue (redirected from burn as on Ethereum), but at baseline\n")
	sb.WriteString("      traffic it is minimal. Under congestion (>1,428 txs/block), baseFee rises,\n")
	sb.WriteString("      significantly increasing both treasury funding and tip-based miner income.\n")

	sb.WriteString("\n")
	sb.WriteString("═══════════════════════════════════════════════════════════════════════════════════════════════════\n")

	t.Log(sb.String())

	// Final verification log
	t.Logf("treasury verified: %s ETC (on-chain matches computed)", weiToETC(expectedTreasury.ToBig()))
	t.Logf("era 4 reward: %s ETC, era 5 reward: %s ETC", weiToETC(era4Reward.ToBig()), weiToETC(era5Reward.ToBig()))
	t.Logf("baseFee: pre-congestion=%s, peak=%s (confirmed rise during era 5 congestion)",
		preCongestionBaseFee, congestionPeak)
}
