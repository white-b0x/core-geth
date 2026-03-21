## CoreGeth: Ethereum Classic Execution Client

> A [go-ethereum](https://github.com/ethereum/go-ethereum) fork providing the production Ethereum Classic (ETC) execution client.

CoreGeth is the current production client for the Ethereum Classic network. It supports all ETC hard forks from Frontier through Spiral, and implements the upcoming **Olympia** hard fork (ECIP-1111, ECIP-1112, ECIP-1121).

**Note:** Upstream go-ethereum has deprecated support for Ethereum Classic. Long-term, ETC will migrate to [Fukuii](https://github.com/chris-mercer/fukuii) as the native ETC client. CoreGeth remains the recommended production client during this transition.

## Supported Networks

| Network | Chain ID | Consensus | Status |
|---------|----------|-----------|--------|
| Ethereum Classic (ETC) | 61 | Proof of Work (ETChash) | Production |
| Mordor Testnet | 63 | Proof of Work (ETChash) | Active |
| Private chains | configurable | PoW / PoA | Supported |

### ETC Hard Fork History ([ECIP-1066](https://ecips.ethereumclassic.org/ECIPs/ecip-1066))

| Fork | Block | Mordor | Included EIPs / ECIPs | Spec |
|------|-------|--------|----------------------|------|
| Frontier | 1 | 0 | Genesis | — |
| Frontier Thawing | 200,000 | 0 | Ice Age introduction | — |
| Homestead | 1,150,000 | 0 | [EIP-2](https://eips.ethereum.org/EIPS/eip-2), [EIP-7](https://eips.ethereum.org/EIPS/eip-7), [EIP-8](https://eips.ethereum.org/EIPS/eip-8) | [HFM-606](https://eips.ethereum.org/EIPS/eip-606) |
| Gas Reprice | 2,500,000 | 0 | [EIP-150](https://eips.ethereum.org/EIPS/eip-150) | [ECIP-1015](https://ecips.ethereumclassic.org/ECIPs/ecip-1015) |
| Die Hard | 3,000,000 | 0 | [ECIP-1010](https://ecips.ethereumclassic.org/ECIPs/ecip-1010), [EIP-155](https://eips.ethereum.org/EIPS/eip-155), [EIP-160](https://eips.ethereum.org/EIPS/eip-160) | — |
| Gotham | 5,000,000 | 0 | [ECIP-1017](https://ecips.ethereumclassic.org/ECIPs/ecip-1017), [ECIP-1039](https://ecips.ethereumclassic.org/ECIPs/ecip-1039) | — |
| Defuse Difficulty Bomb | 5,900,000 | 0 | [ECIP-1041](https://ecips.ethereumclassic.org/ECIPs/ecip-1041) | — |
| Atlantis | 8,772,000 | 0 | [EIP-100](https://eips.ethereum.org/EIPS/eip-100), [EIP-140](https://eips.ethereum.org/EIPS/eip-140), [EIP-196](https://eips.ethereum.org/EIPS/eip-196), [EIP-197](https://eips.ethereum.org/EIPS/eip-197), [EIP-198](https://eips.ethereum.org/EIPS/eip-198), [EIP-211](https://eips.ethereum.org/EIPS/eip-211), [EIP-214](https://eips.ethereum.org/EIPS/eip-214), [EIP-658](https://eips.ethereum.org/EIPS/eip-658) | [ECIP-1054](https://ecips.ethereumclassic.org/ECIPs/ecip-1054) |
| Agharta | 9,573,000 | 301,243 | [EIP-145](https://eips.ethereum.org/EIPS/eip-145), [EIP-1014](https://eips.ethereum.org/EIPS/eip-1014), [EIP-1052](https://eips.ethereum.org/EIPS/eip-1052) | [ECIP-1056](https://ecips.ethereumclassic.org/ECIPs/ecip-1056) |
| Phoenix | 10,500,839 | 999,983 | [EIP-152](https://eips.ethereum.org/EIPS/eip-152), [EIP-1108](https://eips.ethereum.org/EIPS/eip-1108), [EIP-1344](https://eips.ethereum.org/EIPS/eip-1344), [EIP-1884](https://eips.ethereum.org/EIPS/eip-1884), [EIP-2028](https://eips.ethereum.org/EIPS/eip-2028), [EIP-2200](https://eips.ethereum.org/EIPS/eip-2200) | [ECIP-1088](https://ecips.ethereumclassic.org/ECIPs/ecip-1088) |
| Thanos | 11,700,000 | 2,520,000 | [ECIP-1099](https://ecips.ethereumclassic.org/ECIPs/ecip-1099) (ETChash, 60K epoch length) | — |
| Magneto | 13,189,133 | 3,985,893 | [EIP-2565](https://eips.ethereum.org/EIPS/eip-2565), [EIP-2718](https://eips.ethereum.org/EIPS/eip-2718), [EIP-2929](https://eips.ethereum.org/EIPS/eip-2929), [EIP-2930](https://eips.ethereum.org/EIPS/eip-2930) | [ECIP-1103](https://ecips.ethereumclassic.org/ECIPs/ecip-1103) |
| Mystique | 14,525,000 | 5,520,000 | [EIP-3529](https://eips.ethereum.org/EIPS/eip-3529), [EIP-3541](https://eips.ethereum.org/EIPS/eip-3541) (partial London, no EIP-1559) | [ECIP-1104](https://ecips.ethereumclassic.org/ECIPs/ecip-1104) |
| Spiral | 19,250,000 | 9,957,000 | [EIP-3651](https://eips.ethereum.org/EIPS/eip-3651), [EIP-3855](https://eips.ethereum.org/EIPS/eip-3855), [EIP-3860](https://eips.ethereum.org/EIPS/eip-3860), [EIP-6049](https://eips.ethereum.org/EIPS/eip-6049) (Shanghai equivalent) | [ECIP-1109](https://ecips.ethereumclassic.org/ECIPs/ecip-1109) |
| **Olympia** | **TBD** | **TBD** | ECIP-1111 + ECIP-1121 (EIP-1559 + treasury + EVM modernization) | [ECIP-1111](https://ecips.ethereumclassic.org/ECIPs/ecip-1111) |

**Gas limit:** 8M (current) converging to 60M post-Olympia (EIP-7935).

## Olympia Hard Fork

The `olympia` branch implements the Olympia upgrade:

- **ECIP-1111:** EIP-1559 dynamic base fee + EIP-3198 BASEFEE opcode. Base fee is redirected to the Olympia Treasury (not burned).
- **ECIP-1112:** Deterministic, immutable Treasury contract receiving all base fee revenue.
- **ECIP-1121:** 13 execution-layer EIPs for EVM modernization (EIP-5656/MCOPY, EIP-1153/transient storage, EIP-6780/SELFDESTRUCT, EIP-2537/BLS12-381, EIP-7951/P256VERIFY, EIP-7702/EOA code delegation, EIP-2935/block hashes in state, and more).

**EIP-7642 (eth/69) is intentionally excluded** — it removes Total Difficulty from the protocol handshake, which ETC requires for Proof-of-Work chain selection.

### Activation

- **Mordor testnet:** Block TBD
- **ETC mainnet:** Block TBD

### Cross-Client Alignment

| Client | Pre-Olympia | Post-Olympia | Role |
|--------|-------------|--------------|------|
| [core-geth](https://github.com/chris-mercer/core-geth) | `pre-olympia` branch | `olympia` branch | Production client |
| [Fukuii](https://github.com/chris-mercer/fukuii) | `alpha` branch | `olympia` branch | Native ETC client (migration target) |
| [Besu](https://github.com/chris-mercer/besu) | `etc` branch | `olympia` branch | Reference/testing client |

## Build
```bash
make geth
```

## Test

```bash
# ETC-specific unit tests
go test ./params/... -run TestETC -v
go test ./core/... -run "TestGasLimit|TestForkCompliance|TestECIP1017|TestTreasury|TestOlympia" -v
go test ./consensus/ethash/... -run "TestDifficultyETC|TestDifficultyECIP" -v

# Live tests (requires running Mordor/ETC node)
go test -tags live ./tests/live_etc/ -v
```

## Mining

CoreGeth supports full Ethash/ETChash Proof-of-Work mining:

```bash
./build/bin/geth --classic --mine --miner.etherbase <address>
```

For testing with fake PoW (no DAG generation):

```bash
./build/bin/geth --classic --mine --miner.etherbase <address> --fakepow
```

## Documentation

- [ETC-HANDOFF.md](./ETC-HANDOFF.md) — Pre-Olympia branch documentation
- [OLYMPIA-HANDOFF.md](./OLYMPIA-HANDOFF.md) — Olympia hard fork documentation
- [CoreGeth docs](https://etclabscore.github.io/core-geth) — General documentation
- [go-ethereum docs](https://geth.ethereum.org/docs/) — Upstream reference

## Security Patches

The `etc` branch includes 5 security patches applied before any feature work:

| CVE | Severity | Component |
|-----|----------|-----------|
| CVE-2026-26314 | High | secp256k1 coordinate validation |
| CVE-2026-26315 | Moderate | ECIES invalid-curve attack |
| CVE-2026-22862 | High | ECIES decrypt length |
| CVE-2025-24883 | Moderate | UnmarshalPubkey curve check |
| x/crypto, x/net | — | Dependency updates |

Nodes should rotate P2P keys after upgrading: `rm <datadir>/geth/nodekey`

## License

The core-geth library (outside `cmd/`) is licensed under [LGPL-3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html).
The core-geth binaries (`cmd/`) are licensed under [GPL-3.0](https://www.gnu.org/licenses/gpl-3.0.en.html).
