# Olympia Hard Fork — Implementation Handoff Report

**Date:** 2026-02-26
**Author:** Chris Mercer
**Branch:** `olympia` (chris-mercer/core-geth)
**Status:** Code complete. Ready for Mordor testnet soak.

---

## Overview

The Olympia upgrade brings EIP-1559 dynamic fees and Cancun-equivalent EVM improvements to Ethereum Classic, with basefee revenue directed to a community treasury instead of being burned.

**Three ECIPs implemented:**
- **ECIP-1111** — EIP-1559 + EIP-3198 with basefee redirected to treasury
- **ECIP-1112** — Immutable treasury vault (demo v0.2: pure Solidity, deployed via CREATE)
- **ECIP-1121** — 13 execution-layer EIPs aligned with Ethereum (PoW-compatible subset)

**Mordor activation:** Block **15,800,850** (~March 28, 2026)
**ETC mainnet target:** Block ~24,751,337 (~mid-June 2026, before Era 6 at 25M)

---

## Commit Log (27 commits on `olympia` branch)

| # | SHA | Description |
|---|-----|-------------|
| 1 | a7741046c | security: fix secp256k1 coordinate validation (CVE-2026-26314) |
| 2 | f8fd797d5 | security: update golang.org/x/crypto and x/net |
| 3 | 477b1c9a6 | feat: add OlympiaTreasuryAddress to chain config |
| 4 | 998706823 | feat: activate EIP-1559 + EIP-3198 on Mordor |
| 5 | aff65a65e | feat: redirect basefee to treasury (ECIP-1111) |
| 6 | 5f79936e7 | feat: activate EIP-5656, 1153, 6780, 2537 on Mordor |
| 7 | a0b6523f8 | feat: EIP-7883 + EIP-7823 (ModExp gas/bounds) |
| 8 | 1eaa81f3e | feat: EIP-7825 (TX gas limit cap 2^24) |
| 9 | a231ce69f | feat: EIP-7623 (floor data gas cost) |
| 10 | c3d2f95b3 | feat: EIP-7935 (default gas limit 60M) |
| 11 | 777ea6bcc | fix: correct EIP-7883 gas formula + tests |
| 12 | 60c2bcd23 | feat: EIP-7951 (secp256r1 P256VERIFY precompile) |
| 13 | 2cb789560 | feat: EIP-2935 (block hashes in state) |
| 14 | 3d048d8a3 | feat: EIP-7702 (set EOA account code) |
| 15 | b136ef38e | test: treasury basefee redirect tests (ECIP-1111) |
| 16 | f6ba86486 | test: backward compatibility + ECIP-1017 era reward tests |
| 17 | c8cf31536 | feat: EIP-7934 (block size limit) |
| 18 | d3f62a819 | feat: EIP-7910 (eth_config RPC endpoint) |
| 19 | c4c5540da | feat: set Mordor activation block 15,800,850 + forkid |
| 20 | f5c24c6f9 | feat: set Mordor treasury to deployed contract |
| 21 | ddb26b115 | docs: add implementation handoff report |
| 22 | 06af56eb6 | test: address ChatGPT security review — 13 tests (reorg, EIP-2935, txpool, treasury, EIP-7702) |
| 23 | 57bf6d76e | test: security budget analysis at 60M gas limit (17,650 txs, baseFee dynamics report) |
| 24 | a059d4615 | docs: update handoff report with full test coverage |
| 25 | 870a1f439 | test: security budget at ECIP-1017 era 4/5 with 30K txs/day baseline |

---

## EIP Implementation Status

### ECIP-1111: EIP-1559 + Treasury Redirect

| EIP | Status | What |
|-----|--------|------|
| EIP-1559 | ACTIVATED | Dynamic base fee (already implemented in core-geth, was nil) |
| EIP-3198 | ACTIVATED | BASEFEE opcode (already implemented, was nil) |
| Treasury redirect | NEW CODE | `consensus/ethash/consensus.go` Finalize() credits `baseFee * gasUsed` to treasury |

**Key file:** `consensus/ethash/consensus.go` — Finalize() function, treasury credit after AccumulateRewards.

### ECIP-1121: Execution-Layer EIPs (13 of 14)

| EIP | Status | Complexity | What |
|-----|--------|-----------|------|
| 5656 (MCOPY) | ACTIVATED | None | Already existed, set FBlock |
| 1153 (Transient storage) | ACTIVATED | None | Already existed, set FBlock |
| 6780 (SELFDESTRUCT restrict) | ACTIVATED | None | Already existed, set FBlock |
| 2537 (BLS12-381) | ACTIVATED | None | Already existed, set FBlock |
| 7883 (ModExp gas increase) | PORTED | Low | Modified `core/vm/contracts.go` bigModExp |
| 7823 (ModExp upper bounds) | PORTED | Low | Added 1024-byte input limit to bigModExp |
| 7825 (TX gas limit cap) | PORTED | Low | 2^24 (16,777,216) cap in state_transition + txpool |
| 7623 (Floor data gas) | PORTED | Medium | FloorDataGas() in state_transition + txpool |
| 7935 (Default gas limit 60M) | PORTED | Low | Changed DefaultGasLimit in vars |
| 7951 (secp256r1 precompile) | PORTED | Medium | New precompile at 0x100, Go stdlib P-256 |
| 2935 (Block hashes in state) | PORTED | Medium | System contract + ProcessParentBlockHash |
| 7702 (Set EOA account code) | PORTED | High | New tx type 0x04, delegation, EVM changes |
| 7934 (Block size limit) | PORTED | Low | RLP size validation in block_validator + miner |
| **7642 (eth/69)** | **OMITTED** | — | See below |

### EIP-7642 Omission Rationale

EIP-7642 removes Total Difficulty (TD) from the eth protocol Status handshake. **This is incompatible with ETC's Proof of Work consensus:**

- ETC uses TD as the canonical chain selection mechanism
- On post-Merge Ethereum, TD is fixed (no longer changes), so removing it is safe
- On ETC, TD changes every block — removing it would break peer chain selection
- EIP-7642 is **purely networking** (no EVM changes), so omitting it has **zero impact on EVM compatibility**

ETC remains on eth/68 for peer protocol. All EVM opcodes and precompiles from the Cancun equivalent are included.

### EIP-7910 (eth_config RPC)

New JSON-RPC endpoint `eth_config` returning current/next/last fork configurations with active precompiles and system contracts. File: `internal/ethapi/api_config.go`.

---

## Security Patches Applied

Four CVEs patched before any feature work:

| CVE | Severity | Description | Fix |
|-----|----------|-------------|-----|
| CVE-2026-26314 | High | secp256k1 coordinate validation bypass — node crash via crafted P2P messages | Added bounds check in `crypto/secp256k1/curve.go` |
| CVE-2026-26315 | Moderate | ECIES invalid-curve attack in RLPx handshake — P2P key extraction | Added IsOnCurve check in `crypto/ecies/ecies.go` |
| CVE-2026-22862 | High | ECIES decrypt length off-by-one — node crash via single P2P message | Fixed length check in `crypto/ecies/ecies.go` |
| CVE-2025-24883 | Moderate | Missing IsOnCurve in UnmarshalPubkey — invalid key acceptance | Added curve check in `crypto/crypto.go` |

**Dependencies updated:** `golang.org/x/crypto` and `golang.org/x/net` to latest.

**Post-patch action:** Rotate P2P node keys on all production nodes (`rm <datadir>/geth/nodekey`).

---

## Test Coverage

### New Tests Written (Original Implementation)

| File | Tests | What |
|------|-------|------|
| `consensus/ethash/consensus_test.go` | Treasury redirect | baseFee*gasUsed credited, zero-gas blocks, fork boundary, miner tips |
| `core/backward_compat_test.go` | Backward compatibility | Pre-fork blocks unchanged, ECIP-1017 era rewards, all tx types, gas accounting |
| `core/eip7934_test.go` | Block size limit | Constant value, normal blocks pass, error defined |
| `core/vm/contracts_test.go` | ModExp + P256 | Gas calc, bounds check, precompile test vectors |
| `core/forkid/forkid_test.go` | Forkid | Updated Mordor creation + gatherForks tests |

### Review-Response Tests (13 tests, commit 06af56e)

Added to address external security review concerns:

| File | Tests | What |
|------|-------|------|
| `core/backward_compat_test.go` | +1 | Reorg across fork boundary with treasury state verification |
| `core/eip2935_test.go` | +2 | Block hash persistence at HISTORY_STORAGE_ADDRESS, reorg updates hashes |
| `core/txpool/validation_test.go` | +6 | EIP-7825 2^24 gas cap, EIP-7623 calldata pricing (Legacy, AccessList, DynamicFee, SetCode) |
| `core/treasury_test.go` | +2 | 50-block cumulative accumulation, immutable treasury address |
| `core/eip7702_test.go` | +2 | Self-delegation, code clears after revocation via nonce bump |

### Security Budget Analysis (1 test, commits 57bf6d7 → updated)

| File | Tests | What |
|------|-------|------|
| `core/security_budget_test.go` | 1 | 60-block ECIP-1017 era 4/5 simulation at 60M gas limit with 30K txs/day baseline |

Uses ECIP-1017 era-based rewards (eraLength=10 blocks for test acceleration) to simulate current ETC mainnet conditions. Baseline: **30,000 txs/day ÷ ~6,200 blocks/day ≈ 5 txs/block** (ETC mainnet floor).

Three phases:
- **Era 4 "Activation"** (2.048 ETC/block, 5 txs/block): Olympia goes live near block 24M
- **Era 5 "Normal"** (1.6384 ETC/block, 5 txs/block): After 25M era boundary
- **Era 5 "Congestion"** (1.6384 ETC/block, 2,000 txs/block): Network under load

Produces a detailed ASCII report showing:
- Per-block baseFee trajectory with era-adjusted block rewards
- Miner income breakdown: block rewards vs tips by tx type (Legacy, AccessList, DynamicFee)
- Treasury income: baseFee × gasUsed (ECIP-1111 redirect)
- Per-phase security budget ratios comparing baseline vs congestion
- Key findings: At 30K txs/day baseline, tips are 0.01% of miner income — block rewards are the entire security budget. Under congestion (2,000 txs/block), tips rise to ~5% and treasury funding increases ~19×

### Existing Test Suites

All existing tests continue to pass. Key suites:
```bash
go test ./core/forkid/...          # Forkid (PASS)
go test ./consensus/ethash/...     # Consensus (PASS)
go test ./core/vm/...              # EVM opcodes + precompiles (PASS)
go test ./core/...                 # State processor, transitions (PASS)
go test ./params/...               # Config (PASS)
```

---

## Consensus-Breaking Changes

These are the exact consensus-breaking changes that will cause pre-Olympia nodes to reject post-Olympia blocks (or vice versa):

### Block Header
- **BaseFee field** now non-nil (EIP-1559). Pre-fork nodes will reject headers with BaseFee.
- **BaseFee adjustment algorithm** active (EIP-1559 `CalcBaseFee`). Invalid BaseFee = invalid block.

### State Transition
- **Treasury credit** in `Finalize()`: `baseFee * gasUsed` credited to treasury after `AccumulateRewards`. Changes stateRoot.
- **EIP-2935 system call**: `ProcessParentBlockHash` executes at start of each block, deploying history contract at fork block. Changes stateRoot.
- **EIP-7623 floor data gas**: Transactions with insufficient gas for calldata floor rejected during execution.
- **EIP-7825 tx gas cap**: Transactions with gas > 2^24 (16,777,216) rejected during pre-checks.
- **EIP-7702 tx type (0x04)**: New transaction type with authorization list processing. Changes account code via delegation prefix.

### EVM
- **EIP-5656 MCOPY**: New opcode 0x5E
- **EIP-1153 TLOAD/TSTORE**: New opcodes 0x5C/0x5D
- **EIP-6780 SELFDESTRUCT**: Restricted to same-transaction creation
- **EIP-7702 delegation resolution**: CALL/STATICCALL/DELEGATECALL resolve delegation prefix codes (one level deep)

### Precompiles
- **EIP-2537 BLS12-381**: 7 precompiles at 0x0b–0x11 (final spec: G1Add, G1MultiExp, G2Add, G2MultiExp, Pairing, MapG1, MapG2)
- **EIP-7951 P256VERIFY**: Precompile at 0x100
- **EIP-7883 ModExp**: Increased gas cost formula
- **EIP-7823 ModExp**: 1024-byte input limit

### Validation
- **EIP-7934 block size**: RLP-encoded blocks > ~8 MiB rejected
- **EIP-7935 default gas limit**: Miner default changed to 60M (miner policy, not consensus-breaking)

---

## State Transition Ordering (Finalize)

The finalization ordering in `consensus/ethash/consensus.go:Finalize()` is deterministic and critical for state root consistency:

**1. AccumulateRewards** (`mutations.AccumulateRewards`)
- Process uncle rewards first (uncle coinbase gets era-adjusted reward per uncle)
- Process miner reward (era_reward + 1/32 * era_reward per uncle included)
- Era reward subject to ECIP-1017 disinflation: `5 ETC * (4/5)^era`

**2. Treasury Credit** (ECIP-1111)
- Guard: EIP-1559 must be active AND `OlympiaTreasuryAddress` must be non-nil AND `header.BaseFee` must be non-nil
- Credit: `state.AddBalance(treasuryAddr, baseFee * gasUsed)`
- If no treasury configured: baseFee is implicitly burned (vanilla EIP-1559 behavior)

**3. IntermediateRoot** (in `FinalizeAndAssemble`)
- State root computed AFTER steps 1 and 2
- This is the `header.Root` committed to the block

**Note:** EIP-2935 `ProcessParentBlockHash` runs at the START of block processing (in `state_processor.go`), before any transactions execute. This is a system-level state modification that stores the parent block's hash in the history contract.

---

## Cross-Client Verification Methodology

Every EIP and ECIP in the Olympia fork is verified using a six-client cross-verification process:

**Reference Clients (ETH production):**
- [go-ethereum](https://github.com/ethereum/go-ethereum) — canonical Go implementation
- [Erigon](https://github.com/erigontech/erigon) — optimized Go implementation
- [Nethermind](https://github.com/NethermindEth/nethermind) — .NET implementation

**ETC Clients (implementation targets):**
- core-geth (`chris-mercer/core-geth`) — Go, forked from go-ethereum
- Fukuii (`chris-mercer/fukuii`) — Scala/JVM, forked from Mantis
- Besu (`chris-mercer/besu`) — Java/JVM, forked from Hyperledger Besu

**Process (per EIP/ECIP):**
1. Read the canonical EIP specification at `eips.ethereum.org`
2. Verify implementation in all 3 ETH production clients (constants, formulas, gas costs, addresses)
3. Verify implementation in all 3 ETC clients against both the spec and ETH implementations
4. Cross-compare ETC clients against each other for consistency
5. Document any discrepancies with severity (consensus-critical vs. cosmetic)

**Catches from this process:**
- **EIP-7825:** All 3 ETC clients had `MaxTxGas = 30,000,000` (from early drafts). ETH clients all use `1 << 24 = 16,777,216` per final spec. Corrected in all 3 ETC clients.
- **EIP-7883:** Fukuii's `PostEIP7883Cost` had 3 formula bugs (wrong `multComplexity`, wrong `adjExpLen` multiplier 8→16, wrong divisor 3→1 and minimum 200→500). Caught by comparing against go-ethereum's `osakaModexpGas()`. Besu wired `PragueGasCalculator` instead of `OsakaGasCalculator` for ModExp — correct implementation exists but wasn't connected.
- **EIP-7951:** Fukuii initially had `P256VerifyGas = 3450` (half the correct 6,900). Caught by cross-referencing core-geth.
- **EIP-7702:** Fukuii initially had `TxAuthTupleGas = 25000` (double the correct 12,500). Caught by cross-referencing core-geth.

This methodology ensures implementation parity across all ETC clients and consistency with ETH production client behavior for shared EIPs.

---

## Known Risks

### Single Production Client
Core-geth is the only production ETC client. Hyperledger Besu and Fukuii exist but are not production-ready. There is no cross-client consensus validation possible at this time. All consensus changes are validated by test coverage and Mordor soak testing only. A consensus bug means a network halt, not a chain split. Cross-client testing becomes relevant when Besu or Fukuii mature enough to implement Olympia.

### EIP-7702 Edge Cases
EIP-7702 was designed and fuzzed for Ethereum's multi-client PoS environment. ETC's PoW context has not undergone cross-client differential fuzzing. Specific risk areas:
- Delegation + SELFDESTRUCT (EIP-6780 restriction mitigates but does not eliminate)
- Delegation chain resolution (enforced one-level deep in `resolveCode`)
- Self-delegation (account delegates to itself)
- Authorization replay across chains with matching chain IDs

### P-256 Performance (EIP-7951)
The P256VERIFY precompile uses Go's `crypto/elliptic` stdlib, not optimized assembly. Under adversarial conditions at ETC's 8M gas limit, this is unlikely to be a practical issue, but lacks the performance margins of assembly implementations used in Ethereum clients.

### ModExp Repricing (EIP-7883)
The gas formula change may cause edge cases where previously-affordable ModExp calls become too expensive, breaking contracts that depend on specific gas costs. Existing ModExp usage on ETC is minimal, reducing practical risk.

### Gas Limit and Fee Market Dynamics
EIP-7935 changes the default gas limit from 8M to **60M**, with an EIP-1559 target of **30M** (limit/2). This is a miner policy default — miners can still adjust it. At 60M, the network needs **1,428+ simple transfers per block** just to start pushing baseFee upward. ETC mainnet currently averages ~30,000 txs/day (~5 txs/block) — **0.2% utilization** at 60M. The `TestSecurityBudgetAnalysis` test simulates era 4 (2.048 ETC/block) and era 5 (1.6384 ETC/block) at this baseline, showing block rewards constitute 99.99% of miner income. Under congestion (2,000 txs/block, 70% utilization), tips rise to ~5% and treasury funding increases ~19×. EIP-7825's 2^24 (16,777,216) per-tx gas cap remains below the block gas limit. Stress profiles are fundamentally different from Ethereum mainnet.

---

## Treasury Contract (ECIP-1112)

> **Demo v0.2** — Pre-Olympia EVM (Shanghai). Pure Solidity, no OpenZeppelin dependency. Deployed on Mordor + ETC mainnet. Not production.

**Repo:** `github.com/olympiadao/olympia-treasury-contract` (branch `demo_v0.2`)
**Contract:** `OlympiaTreasury.sol` — Pure Solidity 0.8.28, immutable executor pattern

**Deployment (Mordor 63 + ETC mainnet 61, identical addresses):**
- Treasury: `0x035b2e3c189B772e52F4C3DA6c45c84A3bB871bf` (CREATE, nonce-based)
- Executor: `0x64624f74F77639CbA268a6c8bEDC2778B707eF9a` (CREATE2, deterministic factory)
- Deployer: `0x7C3311F29e318617fed0833E68D6522948AaE995`

**Architecture:**
- `immutable executor` — single authorized caller, pre-computed CREATE2 address
- `withdraw(to, amount)` — executor only, no admin roles, no upgrade path
- `receive()` — accepts ETC (baseFee revenue credited by consensus)

**Immutability guarantee:** The treasury contract has no mint function. It can only receive baseFee revenue credited by consensus (`Finalize()` in the ethash engine). The executor can only withdraw accumulated funds via `withdraw(to, amount)`, not create new ETC. The total ETC in the treasury is always ≤ the sum of `baseFee * gasUsed` across all post-Olympia blocks.

**Governance contracts** (separate repo: `olympia-governance-contracts`, branch `demo_v0.2`):
- OZ 5.1.0 (required: Shanghai EVM has no `mcopy`, OZ 5.2+ requires Cancun)
- Governor → Timelock → Executor → Treasury pipeline
- 3-layer sanctions defense (ECIP-1119)

**Production note:** Post-Olympia deployment will use OZ 5.6 contracts targeting Cancun EVM. Different bytecode → different CREATE2 addresses for Executor. Treasury CREATE address also changes (different deployer/nonce). Production addresses TBD.

**Tests:** 33 passing (15 unit + 12 security + 6 pre-governance)

---

## Critical Files Modified

| File | Changes |
|------|---------|
| `consensus/ethash/consensus.go` | Treasury basefee credit in Finalize() |
| `params/types/ctypes/configurator_iface.go` | 14 new EIP getter/setter interface methods |
| `params/types/coregeth/chain_config.go` | 14 new FBlock fields + OlympiaTreasuryAddress |
| `params/types/coregeth/chain_config_configurator.go` | 28 getter/setter implementations |
| `params/types/goethereum/goethereum_configurator.go` | 28 stub methods |
| `params/types/genesisT/genesis.go` | 28 delegation methods |
| `params/config_mordor.go` | All Olympia activation blocks + treasury address |
| `params/vars/protocol_params.go` | Constants: BlockRLPSizeCap, P256VerifyGas, MaxTxGas, etc. |
| `core/vm/contracts.go` | bigModExp (7883/7823), p256Verify (7951), PrecompiledContractsForConfig |
| `core/vm/evm.go` | resolveCode/resolveCodeHash for EIP-7702 delegation |
| `core/vm/eips.go` | enable7702 function |
| `core/vm/jump_table.go` | Wire EIP-7702 |
| `core/state_transition.go` | EIP-7702 auth, EIP-7623 floor gas, EIP-7825 cap |
| `core/state_processor.go` | ProcessParentBlockHash (EIP-2935) |
| `core/block_validator.go` | EIP-7934 block size check |
| `core/types/tx_setcode.go` | NEW — SetCodeTx type 0x04 (EIP-7702) |
| `crypto/secp256r1/verifier.go` | NEW — P-256 verification (EIP-7951) |
| `internal/ethapi/api_config.go` | NEW — eth_config RPC (EIP-7910) |
| `miner/worker.go` | EIP-7934 size tracking, EIP-2935 system call |

---

## Next Steps

### Before Mordor Fork (Block 15,800,850)

1. **Build and deploy:** `make geth` with the olympia branch, run on Mordor
2. **Rotate P2P node keys** after security patches: `rm <datadir>/geth/nodekey`
3. **Monitor** for several thousand blocks post-fork:
   - BaseFee appears in block headers
   - Treasury address accumulates basefee revenue
   - Miners receive only tips for Type-2 txs
   - Legacy (Type-0) and access list (Type-1) txs unchanged
   - New opcodes functional (MCOPY, TLOAD/TSTORE, PUSH0)
   - BLS precompiles accessible at addresses 0x0b-0x11 (7 precompiles per final EIP-2537 spec)
   - P256VERIFY accessible at address 0x100

### For ETC Mainnet

1. Confirm Mordor soak successful (no consensus failures, no crashes)
2. Set mainnet activation block in `params/config_classic.go` (target: ~24,751,337)
3. Deploy production treasury contract to ETC mainnet (OZ 5.6, post-Olympia Cancun EVM)
4. Update mainnet treasury address to production deployment
5. Release core-geth v1.13.0
6. Coordinate community upgrade timeline

### Futarchy Governance (Future)

1. Design and implement prediction market-based governance contract
2. Deploy on Mordor, test governance flow
3. Grant WITHDRAWER_ROLE to governance contract
4. Revoke admin direct withdrawal rights
5. Deploy to mainnet

---

## Running Tests

```bash
cd /media/dev/2tb/dev/core-geth

# Full suite
go test ./...

# Key subsystems
go test ./core/forkid/...          # Forkid
go test ./consensus/ethash/...     # Consensus + treasury
go test ./core/vm/...              # EVM opcodes + precompiles
go test ./core/...                 # State processor
go test ./params/...               # Config

# Review-response tests (13 tests)
go test ./core -run "TestOlympiaReorg|TestEIP2935|TestEIP7702|TestTreasuryCumulative|TestTreasuryImmutable" -v
go test ./core/txpool -run "TestEIP7825|TestEIP7623" -v

# Security budget analysis (generates detailed report)
go test ./core -run "TestSecurityBudget" -v

# Live RPC integration tests (41 tests, requires running Mordor node on :8545)
go test -tags live ./tests/live/ -v

# Treasury contract (Foundry)
cd /media/dev/2tb/dev/olympia-treasury-contract
forge test -vv                                          # All 16 tests
forge test --match-path test/StagedEvolution.t.sol -vv  # Staged evolution only
```

---

## Live Integration Tests (tests/live/)

**Build tag:** `//go:build live` — excluded from normal `go test ./...`
**Requires:** Running Mordor node on `:8545` (configurable via `MORDOR_RPC` env)
**Total: 41 tests across 14 files**

### Pre-Fork Tests (run against any synced Mordor node)
| File | Tests | What |
|------|-------|------|
| `chain_basics_test.go` | 8 | Chain ID, genesis hash, block structure, no baseFee pre-fork, ETC mainnet basics |
| `tx_types_test.go` | 2 | No Type-2/4 TXs pre-Olympia, legacy TX decode |
| `treasury_prefork_test.go` | 1 | Treasury balance snapshot |
| `precompiles_prefork_test.go` | 2 | ecrecover works, P256Verify not active |

### Post-Fork Tests (skip gracefully if fork not reached)
| File | Tests | What |
|------|-------|------|
| `eip1559_live_test.go` | 4 | Fork block baseFee, initial=1 gwei, adjustment, Type-2 TX |
| `treasury_live_test.go` | 2 | Treasury balance increases, credit matches baseFee*gasUsed |
| `eip7702_live_test.go` | 2 | Type-4 TX scan, delegation code (0xef0100) prefix |
| `eip2935_live_test.go` | 2 | System contract deployed, parent hash stored |
| `precompiles_live_test.go` | 4 | P256Verify, BLS stubs, ModExp, ecrecover post-fork |
| `eip7934_live_test.go` | 1 | Blocks under 8MB RLP cap |

### PoW/ETChash Mining Tests
| File | Tests | What |
|------|-------|------|
| `pow_mining_test.go` | 9 | Valid PoW fields, difficulty ranges, ECIP-1099 boundary, ECIP-1017 eras, timestamps, gas limit rules, ETC mainnet PoW |

### Cross-Client Tests (optional, skip if clients unavailable)
| File | Tests | What |
|------|-------|------|
| `cross_client_test.go` | 2 | Fork block hash + state root across core-geth/fukuii/besu |
| `mainnet_test.go` | 2 | ETC mainnet era rewards, treasury balance |
