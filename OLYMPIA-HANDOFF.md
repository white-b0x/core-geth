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
- **ECIP-1112** — Immutable treasury vault (OZ AccessControl, deployed via CREATE2)
- **ECIP-1121** — 13 execution-layer EIPs aligned with Ethereum (PoW-compatible subset)

**Mordor activation:** Block **15,800,850** (~March 28, 2026)
**ETC mainnet target:** Block ~24,751,337 (~mid-June 2026, before Era 6 at 25M)

---

## Commit Log (16 commits on `olympia` branch)

| # | SHA | Description |
|---|-----|-------------|
| 1 | a7741046c | security: fix secp256k1 coordinate validation (CVE-2026-26314) |
| 2 | f8fd797d5 | security: update golang.org/x/crypto and x/net |
| 3 | 477b1c9a6 | feat: add OlympiaTreasuryAddress to chain config |
| 4 | 998706823 | feat: activate EIP-1559 + EIP-3198 on Mordor |
| 5 | aff65a65e | feat: redirect basefee to treasury (ECIP-1111) |
| 6 | 5f79936e7 | feat: activate EIP-5656, 1153, 6780, 2537 on Mordor |
| 7 | a0b6523f8 | feat: EIP-7883 + EIP-7823 (ModExp gas/bounds) |
| 8 | 1eaa81f3e | feat: EIP-7825 (TX gas limit cap 30M) |
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
| 7825 (TX gas limit cap) | PORTED | Low | 30M cap in state_transition + txpool |
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

### New Tests Written

| File | Tests | What |
|------|-------|------|
| `consensus/ethash/consensus_test.go` | Treasury redirect | baseFee*gasUsed credited, zero-gas blocks, fork boundary, miner tips |
| `core/backward_compat_test.go` | Backward compatibility | Pre-fork blocks unchanged, ECIP-1017 era rewards correct |
| `core/eip7934_test.go` | Block size limit | Constant value, normal blocks pass, error defined |
| `core/vm/contracts_test.go` | ModExp + P256 | Gas calc, bounds check, precompile test vectors |
| `core/forkid/forkid_test.go` | Forkid | Updated Mordor creation + gatherForks tests |

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

## Treasury Contract (ECIP-1112)

**Repo:** `github.com/olympiadao/olympia-treasury-contract`
**Contract:** `OlympiaTreasury.sol` — OZ AccessControl v5.6.0, Solidity 0.8.28

**Mordor deployment:**
- Address: `0xCfE1e0ECbff745e6c800fF980178a8dDEf94bEe2`
- Admin: `0x3b0952fB8eAAC74E56E176102eBA70BAB1C81537`
- Deployed via CREATE2 (salt: `keccak256("OLYMPIA_TREASURY_V1")`)

**Roles:**
- `DEFAULT_ADMIN_ROLE` — can grant/revoke roles
- `WITHDRAWER_ROLE` — can call `withdraw(to, amount)`

**Staged governance plan:**
1. Phase 1 (now): Admin EOA controls withdrawals
2. Phase 2: Deploy futarchy DAO, grant it WITHDRAWER_ROLE
3. Phase 3: Revoke admin's WITHDRAWER_ROLE, DAO is sole spender

**Tests:** 10 passing (roles, withdrawal, revocation, events, edge cases)

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
   - BLS precompiles accessible at addresses 0x0a-0x12
   - P256VERIFY accessible at address 0x100

### For ETC Mainnet

1. Confirm Mordor soak successful (no consensus failures, no crashes)
2. Set mainnet activation block in `params/config_classic.go` (target: ~24,751,337)
3. Deploy treasury contract to ETC mainnet via CREATE2 (same salt → deterministic address)
4. Update mainnet treasury address
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

# Treasury contract
cd /media/dev/2tb/dev/olympia-treasury-contract
forge test -vv
```
