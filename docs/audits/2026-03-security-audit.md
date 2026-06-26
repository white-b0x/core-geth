# Core-Geth Security Audit — March 2026

**Client:** Core-Geth (Ethereum Classic)
**Audit Date:** March 2026
**Upstream Repository:** [github.com/etclabscore/core-geth](https://github.com/etclabscore/core-geth) (maintenance mode)
**Upstream Last Substantive Release:** v1.12.20 (10 June 2024)
**Upstream Emergency Patches:** [v1.12.21](https://github.com/etclabscore/core-geth/releases/tag/v1.12.21) (18 March 2026) · [v1.12.22](https://github.com/etclabscore/core-geth/releases/tag/v1.12.22) (28 March 2026) — CVE-only backports; Go 1.21 EOL toolchain unchanged
**Patched Repository:** [github.com/ethereumclassic/core-geth](https://github.com/ethereumclassic/core-geth)
**Patched Release:** v1.13.0
**Patched By:** [White B0x](https://whiteb0x.com)
**Auditors:** ETC Core Development Team (cross-client Olympia testing)

---

## Executive Summary

During cross-client testing for the Olympia network upgrade, the ETC Core team identified that `etclabscore/core-geth` — the primary Ethereum Classic execution client — had received no security maintenance since June 2024, a 21-month gap. Six CVEs were found unpatched, spanning cryptographic key validation, P2P protocol memory exhaustion, and a GraphQL DoS vector. The Go toolchain underpinning the client had also reached end-of-life in August 2024, exposing all deployed nodes to unpatched runtime vulnerabilities for 19 months.

Disclosures to the upstream maintainer in early 2025 received no response. A public security disclosure by a Ledger security researcher in February 2026 likewise received no response until an active attack on ETC bootnodes in March 2026 forced emergency patches (v1.12.21 and v1.12.22 at `etclabscore/core-geth`). Those upstream patches addressed the CVE backports but left the client on the Go 1.21 EOL toolchain with no ETC-specific modernisation. The comprehensive remediation — Go 1.26 upgrade, ETC network tooling, Olympia readiness, and all CVE fixes — was carried out by [White B0x](https://whiteb0x.com) and released as v1.13.0 under the `ethereumclassic` GitHub organisation.

---

## Background

Cross-client interoperability testing for the Olympia hard fork (ECIP-1111/1112/1121/1122) surfaced the maintenance gap in early 2026. The ETC core team's review of the go-ethereum security advisory database and Go vulnerability database against the v1.12.20 codebase found all six CVEs unpatched. The security remediation, Go toolchain upgrade, and v1.13.0 release were carried out by [White B0x](https://whiteb0x.com) under the `ethereumclassic` GitHub organisation.

### Disclosure Timeline

| Date | Event |
|------|-------|
| 10 June 2024 | `etclabscore/core-geth` v1.12.20 released — last substantive release from ETC Cooperative staff |
| August 2024 | Go 1.21 reaches end-of-life; core-geth toolchain enters EOL status |
| 30 January 2025 | CVE-2025-24883 published (go-ethereum GHSA-q26p-9cq4-7fc2) |
| 23 January 2025 | Last commit to `etclabscore/core-geth` — a GitHub Actions dependency bump (not a code change) |
| Early 2025 | Private security disclosures sent to upstream maintainer; no response received |
| 13 January 2026 | CVE-2026-22862 and CVE-2026-22868 published |
| 4 February 2026 | Ledger security researcher [@niooss-ledger](https://github.com/niooss-ledger) opens [issue #692](https://github.com/etclabscore/core-geth/issues/692) publicly documenting CVE-2025-24883, CVE-2026-22862, CVE-2026-22868 — no response from maintainer |
| 17 February 2026 | CVE-2026-26313, CVE-2026-26314, CVE-2026-26315 published |
| March 2026 | White B0x completes comprehensive remediation; `ethereumclassic/core-geth` codebase prepared with Go 1.26 upgrade and all CVE fixes |
| 18 March 2026 | Active attack on ETC bootnodes — ECIES handshake crash-loop (`crypto/ecies.symDecrypt` panic) being exploited in production; ETC Cooperative developer [@diega](https://github.com/diega) releases [v1.12.21](https://github.com/etclabscore/core-geth/releases/tag/v1.12.21) ("Aegis") as an emergency patch ([PR #694](https://github.com/etclabscore/core-geth/pull/694)) — the first upstream code response in 21 months, forced by the live attack rather than the prior disclosures |
| 18 March 2026 | @niooss-ledger [documents remaining unpatched CVEs](https://github.com/etclabscore/core-geth/pull/694#issuecomment-4089185353) after v1.12.21: CVE-2025-24883, CVE-2026-26313, and CVE-2026-26315 still unaddressed |
| 28 March 2026 | [@diega releases v1.12.22](https://github.com/etclabscore/core-geth/releases/tag/v1.12.22) ("Hermes") backporting the remaining CVE fixes to the v1.12 codebase — Go 1.21 EOL toolchain unchanged, no ETC-specific modernisation |
| May 2026 | ETC core team privately alerts upstream developer again, sharing the `ethereumclassic/core-geth` v1.13.0 patched codebase — no response |

**Note on upstream v1.12.21/v1.12.22:** The upstream emergency patches address the CVE backports and are a safer option for operators who have not yet migrated. However, they remain on the Go 1.21 EOL toolchain and do not include the ETC network tooling, DNS discovery updates, or Olympia-readiness work included in v1.13.0. Operators running v1.12.x should upgrade to [ethereumclassic/core-geth v1.13.0](https://github.com/ethereumclassic/core-geth) and plan migration to [Fukuii](https://fukuii.com) ahead of the Olympia upgrade.

### Prior Maintainers

Core-Geth is a fork of [multi-geth](https://github.com/multi-geth/multi-geth), originally created and maintained by **Wei Tang** ([@sorpaas](https://github.com/sorpaas)). Multi-geth was the first multi-network go-ethereum fork with first-class ETC support, and its architecture is the direct ancestor of core-geth's chain configuration system.

The core-geth fork was then developed by ETC Labs until they left the ETC ecosystem in 2022. ETC Cooperative-paid staff maintained the client through the Spiral hard fork up until announcing maintenance mode for the client in December 2024:

- **Isaac Ardis** ([@meowsbits](https://github.com/meowsbits)) — primary architect and long-term maintainer
- **Diego López León** ([@diega](https://github.com/diega)) — release manager; cut the v1.12.20 release
- **Chris Ziogas** ([@ziogaschr](https://github.com/ziogaschr)) — contributor and maintainer

---

## Vulnerability Summary

| CVE | Severity | Component | Upstream (etclabscore) | v1.13.0 (ethereumclassic) |
|-----|----------|-----------|------------------------|--------------------------|
| CVE-2025-24883 | High | crypto — secp256k1 key deserialization | Backported in v1.12.22 | Patched |
| CVE-2026-22862 | High | crypto/ecies — ECIES decrypt length check | Backported in v1.12.21 | Patched |
| CVE-2026-26315 | High | crypto/ecies + secp256k1 — ECIES GenerateShared / IsOnCurve | Backported in v1.12.22 | Patched |
| CVE-2026-26314 | High | crypto/secp256k1 — coordinate field boundary bypass | Backported in v1.12.21 | Patched |
| CVE-2026-22868 | Medium | txpool / P2P — KZG DoS (blob/KZG proof verification) | Declared N/A to ETC by upstream¹ | Code path removed |
| CVE-2026-26313 | High | P2P — RLP item count memory exhaustion | Mitigated in v1.12.22 | Patched |
| — (GraphQL depth) | Medium | RPC — unbounded query nesting DoS | Not addressed | Patched |

¹ Upstream maintainer [@diega confirmed](https://github.com/etclabscore/core-geth/issues/692) that CVE-2026-22868 is "not applicable to ETC" because ETC does not support EIP-4844 blob/KZG transactions. The v1.13.0 remediation removes the KZG code path entirely rather than patching it.

---

## Vulnerability Details

### CVE-2025-24883 — Off-Curve Public Key in UnmarshalPubkey

**Severity:** High
**GHSA:** GHSA-q26p-9cq4-7fc2
**Component:** `crypto/crypto.go` — `UnmarshalPubkey()`
**Affected:** etclabscore/core-geth ≤ v1.12.20
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `8e40b7e41`
**Upstream reference:** go-ethereum PR #31100 / commit `159fb1a1d`

**Description:**
`UnmarshalPubkey()` decoded a 65-byte uncompressed public key into `(x, y)` field elements but did not verify that the resulting point lies on the secp256k1 curve. A malicious peer could supply an off-curve point that passes unmarshaling without error, then produces invalid or undefined results in all subsequent ECDSA or ECDH operations that consume the deserialized key.

**Impact:**
Any code path that deserializes an untrusted public key — including P2P node identity validation and RLPx handshake processing — could be supplied an off-curve point. Downstream cryptographic operations on the invalid key can produce incorrect signatures, incorrect shared secrets, or panics depending on the caller.

**Fix:**
Added an `IsOnCurve(x, y)` check immediately after coordinate extraction in `UnmarshalPubkey()`. Off-curve points now return `errInvalidPubkey` before any downstream use.

```go
if !S256().IsOnCurve(x, y) {
    return nil, errInvalidPubkey
}
```

---

### CVE-2026-22862 — ECIES Decrypt Ciphertext Length Undercheck

**Severity:** High
**GHSA:** GHSA-mr7q-c9w9-wh4h
**Component:** `crypto/ecies/ecies.go` — `Decrypt()`
**Affected:** etclabscore/core-geth ≤ v1.12.20
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `dc73f2e4f`
**Upstream reference:** go-ethereum commit `638741b08`

**Description:**
The ECIES `Decrypt()` function validated ciphertext minimum length using `rLen + hLen + 1`, where the `+ 1` accounts for only one byte beyond the point and HMAC fields. The correct minimum is `rLen + hLen + params.BlockSize` (AES block size = 16 bytes for the default ECIES parameters). The off-by-fifteen gap allows a ciphertext between 2 and 15 bytes shorter than a valid AES block to pass the length guard, after which array indexing proceeds into out-of-bounds memory.

**Impact:**
A malicious peer sending a crafted RLPx `auth` message with an undersized ECIES payload can trigger an out-of-bounds read during handshake processing, crashing the node (remote DoS). The RLPx handshake accepts unauthenticated ECIES ciphertexts from any connecting peer, so no prior authentication is required.

**Fix:**
Changed the minimum length check from `rLen + hLen + 1` to `rLen + hLen + params.BlockSize`.

```go
// Before:
if len(c) < (rLen + hLen + 1) {
// After:
if len(c) < (rLen + hLen + params.BlockSize) {
```

---

### CVE-2026-26315 — ECIES GenerateShared Accepts Unvalidated Public Key

**Severity:** High
**GHSA:** GHSA-m6j8-rg6r-7mv8
**Component:** `crypto/ecies/ecies.go` — `GenerateShared()`
**Affected:** etclabscore/core-geth ≤ v1.12.20
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `2d3528803`
**Upstream reference:** go-ethereum commit `46bee92f9`

**Description:**
The RLPx handshake uses ECIES decryption on unauthenticated input from the network. `GenerateShared()` — called during ECDH shared-secret derivation — accepted a `*PublicKey` without verifying it lies on the curve. An ephemeral public key with `X == nil`, `Y == nil`, or coordinates not satisfying the secp256k1 curve equation would proceed into ECDH multiplication and fail only at MAC verification. The MAC failure reveals to the attacker whether the faulty key survived ECDH, which can be used as an oracle to leak bits of the node's static P2P private key across multiple handshake attempts.

**Impact:**
A remote attacker making repeated unauthenticated RLPx handshake attempts with crafted ephemeral keys can potentially recover bits of the target node's P2P private key through timing or MAC-oracle analysis. No authentication or prior connection is needed.

**Fix:**
Added an explicit nil and `IsOnCurve` guard at the start of `GenerateShared()`:

```go
if pub.X == nil || pub.Y == nil || !pub.Curve.IsOnCurve(pub.X, pub.Y) {
    return nil, ErrInvalidPublicKey
}
```

---

### CVE-2026-26314 — secp256k1 IsOnCurve Field Boundary Bypass

**Severity:** High
**GHSA:** GHSA-2gjw-fg97-vg3r
**Component:** `crypto/secp256k1/curve.go` — `IsOnCurve()`; `crypto/secp256k1/ext.h` — `secp256k1_ext_scalar_mul()`; `crypto/signature_nocgo.go` — `btCurve.IsOnCurve()`
**Affected:** etclabscore/core-geth ≤ v1.12.20
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `2d3528803` (bundled with CVE-2026-26315)
**Upstream reference:** go-ethereum commit `895a8597c`

**Description:**
`IsOnCurve()` verified the curve equation `y² ≡ x³ + b (mod P)` but did not first verify that the coordinates `x` and `y` are within the field, i.e., strictly less than the curve prime `P`. Due to modular arithmetic, coordinates equal to or greater than `P` may still satisfy the curve equation when reduced, but they represent invalid (non-canonical) points. Additionally, the C-level `secp256k1_ext_scalar_mul` function did not check the return value of `secp256k1_fe_set_b32`, which returns 0 when a coordinate is out-of-field. A crafted out-of-field coordinate could therefore bypass `IsOnCurve` and proceed into scalar multiplication, producing undefined or attacker-influenced results.

**Impact:**
An attacker supplying points with out-of-field coordinates that still satisfy the naive curve check can pass a gate that is supposed to reject invalid public keys, potentially leading to node crash or, in a consensus context, divergent state computation.

**Fix:**
Added field-boundary checks before the curve equation test in both the Go and C implementations:

```go
// Go
if x.Cmp(BitCurve.P) >= 0 || y.Cmp(BitCurve.P) >= 0 {
    return false
}
```

```c
// C
if (!secp256k1_fe_set_b32(&feX, point) ||
    !secp256k1_fe_set_b32(&feY, point+32)) {
    return 0;
}
```

---

### CVE-2026-22868 — KZG Blob Proof Verification DoS

**Severity:** Medium
**Component:** `core/txpool/validation.go` — `validateBlobSidecar()`; `eth/fetcher/tx_fetcher.go` — `Enqueue()`
**Affected:** etclabscore/core-geth ≤ v1.12.20 (code present but inactive on ETC)
**Remediated:** ethereumclassic/core-geth v1.13.0 — code path removed
**Commit:** `1419c5310`
**Upstream reference:** go-ethereum commit `fdfd1235a` (v1.16.8)

**ETC Applicability:** The upstream maintainer [declared this not applicable to ETC](https://github.com/etclabscore/core-geth/issues/692) — ETC does not support EIP-4844 blob transactions, so the KZG code path is never reached on the ETC network. The v1.12.22 release at `etclabscore/core-geth` does not address it. The v1.13.0 remediation removes the dead code path entirely, eliminating any surface area regardless of how future ETC forks evolve.

**Description:**
KZG blob proof verification is computationally expensive. A malicious peer could repeatedly broadcast blob transactions with invalid KZG proofs, causing the node to perform the full expensive cryptographic verification on each delivery attempt before rejecting the transaction. Because the error was not distinguished from other validation failures, the peer was not disconnected and could continue flooding the node indefinitely.

**Impact:**
On ETC: no active exposure (blob transactions rejected at an earlier validation stage). Included for completeness and to remove dead code from the ETC client surface.

**Remediation:**
The KZG validation code path (`validateBlobSidecar`, `ErrKZGVerificationError` sentinel) was removed from the ETC client in v1.13.0. On Ethereum mainnet, the upstream fix introduces a sentinel error that signals peer disconnection on KZG proof failure — that approach applies to Ethereum, not ETC.

```go
case errors.Is(err, txpool.ErrKZGVerificationError):
    violation = err
// ...
if delivery.violation != nil {
    f.dropPeer(delivery.origin)
}
```

---

### CVE-2026-26313 — P2P RLP Item Count Memory Exhaustion

**Severity:** High
**GHSA:** GHSA-689v-6xwf-5jf3
**Component:** `eth/protocols/eth/handler.go`; new file `eth/protocols/eth/msgvalidate.go`; `eth/protocols/snap/handler.go`
**Affected:** etclabscore/core-geth ≤ v1.12.20
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `5d0cb8b34`
**Upstream reference:** go-ethereum PR #33835

**Description:**
The P2P message handler validated message size against a 10 MiB cap (`maxMessageSize`) before decoding, but did not validate the number of items declared in the RLP list header. A malicious peer could craft a valid RLP list header claiming millions of tiny items within the 10 MiB budget. When `msg.Decode` ran, it would allocate a pointer or struct object for each declared item before any further validation, causing out-of-memory crashes proportional to the declared item count rather than the actual payload size.

The attack requires only a valid peer connection at the devp2p handshake level — no authenticated session is needed to send crafted message payloads.

**Impact:**
Remote OOM crash of any reachable Core-Geth node. An attacker with a single peer connection could crash the node by sending one crafted message on the eth or snap protocol sub-channel.

**Fix:**
Pre-decode item count validation using `countValuesExceedsLimit()` — a zero-allocation scan of only RLP tag bytes — before `msg.Decode` runs. Limits are applied per message type:

- Response messages (BlockHeaders, BlockBodies, Receipts, PooledTransactions, AccountRange, StorageRanges, ByteCodes, TrieNodes): 2048 items
- Transaction broadcasts (TransactionsMsg, NewBlockHashesMsg): 32768 items

Messages with unknown codes pass through to the normal decoder unchanged. The validator exits early as soon as the limit is exceeded, avoiding full iteration over attack payloads containing millions of tiny items.

---

### GraphQL Query Depth DoS

**Severity:** Medium
**GHSA (dependency):** GHSA-mh3m-8c74-74xh (graphql-go `MaxDepth` non-functional in v1.3.0)
**Component:** `graphql/service.go`; `go.mod` — `graphql-go v1.3.0 → v1.9.0`
**Affected:** etclabscore/core-geth ≤ v1.12.20
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `6c2d383fa`

**Description:**
The GraphQL endpoint (`--graphql` flag) had no query complexity or depth limit. Deeply nested queries — for example, a query recursively nesting block references — would cause the server to perform unbounded recursive schema traversal, exhausting CPU and memory on the serving node. A secondary issue is that the `graphql-go` dependency was pinned at v1.3.0, which contained a bug where the `MaxDepth` option did not function correctly even when set; this was fixed upstream in v1.9.0.

**Impact:**
Any peer or client with access to the GraphQL endpoint can crash or heavily degrade a running node with a single deeply nested query. The GraphQL endpoint is off by default but is commonly enabled on infrastructure nodes.

**Fix:**
Added a `MaxDepth(20)` limit to the GraphQL schema parser and bumped `graphql-go` from v1.3.0 to v1.9.0 to ensure the depth limit is actually enforced:

```go
const maxQueryDepth = 20

s, err := graphql.ParseSchema(schema, &q, graphql.MaxDepth(maxQueryDepth))
```

---

## Go Toolchain End-of-Life

**Issue:** Upstream Core-Geth v1.12.20 was built and shipped on Go 1.21, which reached end-of-life in August 2024. From that date onward, vulnerabilities in the Go standard library — including `net/http`, `crypto/tls`, and `encoding/json` — received no patches from the Go team for this toolchain version, and all binaries compiled against it remained exposed.

**Impact:** Node operators running the upstream binary were exposed to Go runtime security issues for a minimum of 19 months (August 2024 through March 2026). The Go vulnerability database lists multiple advisories against Go 1.21 in this period, including issues in the TLS stack and HTTP/2 server.

**Fix:** Core-Geth v1.13.0 builds on Go 1.26 (current stable as of March 2026). The upgrade proceeded in two steps — 1.21 → 1.24 (removing the incompatible `fjl/memsize` dependency and fixing `go vet` format string errors introduced by 1.24's stricter checks), then 1.24 → 1.26 (updating all `golang.org/x/` dependencies for compatibility). The `blst` cryptography dependency was simultaneously upgraded from v0.3.11 to v0.3.16 to fix a C23 `typedef bool` incompatibility with the Alpine-based Docker build environment.

**Commit:** `8385cf8e8`

---

## Release Timeline

| Date | Event |
|------|-------|
| 10 June 2024 | Core-Geth v1.12.20 released at `etclabscore/core-geth` — final upstream release |
| August 2024 | Go 1.21 reaches end-of-life; upstream receives no further toolchain security patches |
| 23 January 2025 | Last upstream commit (GitHub Actions dependency bump — no code changes) |
| February 2025 | Security disclosures sent to upstream maintainer — no response received |
| February 2026 | CVE-2025-24883 and CVE-2026-22862 patched (`8e40b7e41`, `dc73f2e4f`) |
| February 2026 | CVE-2026-26315 / CVE-2026-26314 patched (`2d3528803`) |
| 4 March 2026 | Go toolchain upgraded 1.21 → 1.24 → 1.26 (`8385cf8e8`) |
| 20 March 2026 | GraphQL depth limit and CVE-2026-22868 patched (`6c2d383fa`, `1419c5310`) |
| 21 March 2026 | CVE-2026-26313 patched (`5d0cb8b34`) |
| March 2026 | Core-Geth v1.13.0 released at `ethereumclassic/core-geth` — all CVEs patched |

---

## Risk Assessment

| Risk Area | Severity | Description | Mitigation |
|-----------|----------|-------------|------------|
| Unpatched CVEs (5 + 1) | Critical | Five CVEs and a GraphQL DoS unaddressed for 21 months in the primary ETC client | All patched in v1.13.0 |
| Cryptographic oracle (CVE-2026-26315) | High | Repeated handshake attempts could leak P2P node key bits | Patched; key rotation recommended for long-running nodes |
| Remote crash via ECIES (CVE-2026-22862) | High | Any peer can crash a node during RLPx handshake with a malformed ECIES payload | Patched in v1.13.0 |
| Remote OOM via RLP (CVE-2026-26313) | High | Any peer can OOM-crash a node with a single crafted P2P message | Patched in v1.13.0 |
| Go Runtime EOL | High | 19 months on unsupported Go toolchain; runtime CVEs accumulated unpatched | Upgraded to Go 1.26 in v1.13.0 |
| Single-maintainer upstream | High | No security response to disclosures; organisation effectively unmaintained | Client transferred to `ethereumclassic` org; multi-client Olympia architecture (Fukuii, Core-Geth, Besu) reduces single-client dependency |
| Release gap (21 months) | Medium | Longest maintenance gap in ETC network history; window for chain divergence or targeted attack | Protocol-funded maintenance path via ECIP-1112 treasury |

---

## Methodology

The audit was initiated during cross-client interoperability testing for the Olympia hard fork. The `etclabscore/core-geth` codebase at tag `v1.12.20` was diffed against the go-ethereum security advisory database (GitHub Security Advisories for `github.com/ethereum/go-ethereum`) and the Go vulnerability database (`vuln.go.dev`). Each advisory was assessed for applicability to core-geth's shared code ancestry. Applicable fixes were cherry-picked or ported from upstream go-ethereum commits; where upstream structural divergence made a clean cherry-pick impossible (CVE-2026-26313, CVE-2026-26314), fixes were manually ported. All patches were validated against the Mordor testnet prior to mainnet release.

---

## Network Migration Path

Core-Geth v1.13.x is the final stable release series of this client. The ETC network is migrating to [Fukuii](https://fukuii.com) ([github.com/chippr-robotics/fukuii](https://github.com/chippr-robotics/fukuii)) as the primary ETC-native execution client, developed natively for the Ethereum Classic ecosystem. Core-Geth will continue to be maintained for the Olympia upgrade cycle to ensure a stable transition, but operators should plan their migration to Fukuii beyond that point.

**If you are running any v1.12.x release, upgrade to v1.13.0 immediately.** The v1.12 series is affected by all six vulnerabilities documented in this audit and is not supported for the Olympia network upgrade.

---

## Recommendations

- **Node operators (v1.12.x):** Update to Core-Geth v1.13.0 from [github.com/ethereumclassic/core-geth](https://github.com/ethereumclassic/core-geth) immediately. Nodes running v1.12.20 or earlier are exposed to remote crash and potential key-oracle attacks from any peer, and will not be compatible with the Olympia hard fork.
- **Infrastructure providers and exchanges:** Treat the upgrade to v1.13.0 as a security-critical patch, not a routine version bump. Prioritise before Olympia activation. Begin planning migration to Fukuii for post-Olympia operation.
- **Long-running nodes:** Consider rotating the P2P node key (`--nodekey`) as a precaution against CVE-2026-26315 oracle exposure. Any node reachable from the public internet over the 21-month gap was potentially targeted.
- **Multi-client operation:** Run at minimum two independent clients (e.g., Core-Geth v1.13.0 + Fukuii, or Core-Geth + Besu) for redundancy. The Olympia upgrade architecture is designed around multi-client operation precisely to mitigate the risk of a single-client maintenance failure.
- **GraphQL endpoint:** If `--graphql` is enabled on public-facing nodes, verify the upgrade to v1.13.0 is in place before re-opening the port.

---

## References

- Patched client: https://github.com/ethereumclassic/core-geth
- Upstream (abandoned): https://github.com/etclabscore/core-geth
- Go vulnerability database: https://vuln.go.dev
- ECIP-1111 (Olympia base): https://ecips.ethereumclassic.org/ECIPs/ecip-1111
- ECIP-1112 (treasury funding): https://ecips.ethereumclassic.org/ECIPs/ecip-1112
- ECIP-1121 (Olympia — EVM updates): https://ecips.ethereumclassic.org/ECIPs/ecip-1121
- ECIP-1122 (Olympia — network upgrades): https://ecips.ethereumclassic.org/ECIPs/ecip-1122
- go-ethereum security advisories: https://github.com/ethereum/go-ethereum/security/advisories
- CVE-2025-24883 (GHSA-q26p-9cq4-7fc2): https://github.com/ethereum/go-ethereum/security/advisories/GHSA-q26p-9cq4-7fc2
- CVE-2026-22862 (GHSA-mr7q-c9w9-wh4h): https://github.com/ethereum/go-ethereum/security/advisories/GHSA-mr7q-c9w9-wh4h
- CVE-2026-26315 (GHSA-m6j8-rg6r-7mv8): https://github.com/ethereum/go-ethereum/security/advisories/GHSA-m6j8-rg6r-7mv8
- CVE-2026-26314 (GHSA-2gjw-fg97-vg3r): https://github.com/ethereum/go-ethereum/security/advisories/GHSA-2gjw-fg97-vg3r
- CVE-2026-26313 (GHSA-689v-6xwf-5jf3): https://github.com/ethereum/go-ethereum/security/advisories/GHSA-689v-6xwf-5jf3
- graphql-go GHSA-mh3m-8c74-74xh: https://github.com/graph-gophers/graphql-go/security/advisories/GHSA-mh3m-8c74-74xh
