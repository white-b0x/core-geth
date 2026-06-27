# Core-Geth Security Audit — March 2026

**Client:** Core-Geth (Ethereum Classic)  
**Audit Date:** March 2026  
**Upstream Repository:** [github.com/etclabscore/core-geth](https://github.com/etclabscore/core-geth) (maintenance mode)  
**Upstream Last Substantive Release:** v1.12.20 (10 June 2024)  
**Upstream Emergency Patches:** [v1.12.21](https://github.com/etclabscore/core-geth/releases/tag/v1.12.21) (18 March 2026) · [v1.12.22](https://github.com/etclabscore/core-geth/releases/tag/v1.12.22) (28 March 2026) — CVE-only backports; Go 1.21 EOL toolchain unchanged  
**Patched Repository:** [github.com/ethereumclassic/core-geth](https://github.com/ethereumclassic/core-geth)  
**Upcoming Release:** v1.13.0 (pending Olympia activation)  
**Patched By:** [White B0x](https://whiteb0x.com)  
**Auditors:** Ethereum Classic Core Developers (cross-client Olympia testing)

---

## Executive Summary

During cross-client testing for the Olympia network upgrade, the Ethereum Classic Core Developers identified that `etclabscore/core-geth` — the primary Ethereum Classic execution client — had received no security maintenance since June 2024, a 21-month gap. Six CVEs were found unpatched, spanning cryptographic key validation, P2P protocol memory exhaustion, and a GraphQL DoS vector. The Go toolchain underpinning the client had also reached end-of-life in August 2024, exposing all deployed nodes to unpatched runtime vulnerabilities for 19 months.

Disclosures to the upstream maintainer [@diega](https://github.com/diega) in early 2025 received no response. At the time, [@diega](https://github.com/diega) was employed by ETC Cooperative as a paid maintainer at a [compensation package exceeding $200,000 per year](https://etccooperative.org/filings) — the sole person with merge access and release authority over the client. A public security disclosure by a Ledger security researcher in February 2026 likewise received no response until an active attack on ETC bootnodes in March 2026 forced emergency patches (v1.12.21 and v1.12.22 at `etclabscore/core-geth`). Those upstream patches addressed the CVE backports but left the client on the Go 1.21 EOL toolchain with no ETC-specific modernisation. The comprehensive remediation — Go 1.26 upgrade, ETC network tooling, Olympia readiness, and all CVE fixes — was carried out by [White B0x](https://whiteb0x.com). The security patches were subsequently ported without attribution into [@diega](https://github.com/diega)'s v1.12.22 emergency release. The full remediation will be released as v1.13.0 under the `ethereumclassic` GitHub organisation in preparation for the Olympia network upgrade — the final release series of Core-Geth before the network transitions to [Fukuii](https://fukuii.com), Ethereum Classic's only native client.

---

## Background

Cross-client interoperability testing for the Olympia network upgrade (ECIP-1111/1112/1121/1122) surfaced the Core-Geth maintenance gap in early 2026. The Ethereum Classic Core Developers' review of the go-ethereum security advisory database and Go vulnerability database against the v1.12.20 codebase found all six CVEs unpatched. The security remediation and Go toolchain upgrade were carried out by [White B0x](https://whiteb0x.com) under the `ethereumclassic` GitHub organisation. A subset of those patches were subsequently ported without attribution into the `etclabscore/core-geth` v1.12.22 emergency release. The provenance is confirmed by the novel adaptations: two of the fixes (CVE-2026-26313 and CVE-2026-26314) required original implementation work — the upstream go-ethereum commits could not be cleanly cherry-picked due to structural divergence. Those novel implementations first appeared in White B0x's public PRs on 20–21 March 2026 and appear identically in v1.12.22 seven days later. The full remediation — including Go 1.26, ETC-specific network tooling, and Olympia readiness — will be released as v1.13.0.

### Disclosure Timeline

| Date | Event |
|------|-------|
| 10 June 2024 | `etclabscore/core-geth` v1.12.20 released — last substantive release from ETC Cooperative staff |
| August 2024 | Go 1.21 reaches end-of-life; core-geth toolchain enters EOL status |
| 30 January 2025 | CVE-2025-24883 published (go-ethereum GHSA-q26p-9cq4-7fc2) |
| 23 January 2025 | Last commit to `etclabscore/core-geth` — a GitHub Actions dependency bump (not a code change) |
| Early 2025 | Private security disclosures sent to upstream maintainer throughout 2025; no response received |
| 29 June 2025 | Community member [@tornadocontrib](https://github.com/tornadocontrib) opens [PR #683](https://github.com/etclabscore/core-geth/pull/683) to `etclabscore/core-geth` — Go 1.24 upgrade and CVE-2025-24883 fix included; closed without merge when contributor deleted their fork in February 2026 |
| 13 January 2026 | CVE-2026-22862 and CVE-2026-22868 published |
| 4 February 2026 | Ledger security researcher [@niooss-ledger](https://github.com/niooss-ledger) opens [issue #692](https://github.com/etclabscore/core-geth/issues/692) publicly documenting CVE-2025-24883, CVE-2026-22862, CVE-2026-22868 — no response from maintainer |
| 17 February 2026 | CVE-2026-26313, CVE-2026-26314, CVE-2026-26315 published |
| 26 February 2026 | White B0x authors CVE-2025-24883 patch on `chris-mercer/core-geth` (olympia branch) during Olympia EIP implementation sprint |
| Early March 2026 | White B0x authors Go 1.21 → 1.24 toolchain upgrade and remaining CVE patches on `chris-mercer/core-geth` (olympia branch) during the fukuii cross-client sprint |
| 18 March 2026 | Active attack on ETC bootnodes — ECIES handshake crash-loop (`crypto/ecies.symDecrypt` panic) exploited in production; [@diega](https://github.com/diega) releases [v1.12.21](https://github.com/etclabscore/core-geth/releases/tag/v1.12.21) ("Aegis") as an emergency patch ([PR #694](https://github.com/etclabscore/core-geth/pull/694)) — the first upstream code response in 21 months, forced by the live attack rather than prior disclosures |
| 18 March 2026 | @niooss-ledger [documents remaining unpatched CVEs](https://github.com/etclabscore/core-geth/pull/694#issuecomment-4089185353) after v1.12.21: CVE-2025-24883, CVE-2026-26313, and CVE-2026-26315 still unaddressed |
| 20–21 March 2026 | White B0x publicly submits all security patches as individual PRs to `ethereumclassic/core-geth` ([#10](https://github.com/ethereumclassic/core-geth/pull/10)–[#36](https://github.com/ethereumclassic/core-geth/pull/36)), references them in [issue #692](https://github.com/etclabscore/core-geth/issues/692) to make them available for the upstream maintainer to pull from |
| 28 March 2026 | [@diega releases v1.12.22](https://github.com/etclabscore/core-geth/releases/tag/v1.12.22) ("Hermes") — CVE fixes drawn from publicly posted White B0x patches with no attribution; Go 1.21 EOL toolchain unchanged, no ETC-specific modernisation. Two of the fixes are forensic evidence of the source: CVE-2026-26313 uses a novel minimal port approach (pre-decode `countValuesExceedsLimit` scan) because the full upstream RawList refactor was incompatible with core-geth's structure — this is not derivable by cherry-picking from go-ethereum. CVE-2026-26314 was likewise manually ported with no clean upstream commit to copy from. Both novel implementations appeared in White B0x's public PRs on 20–21 March; they appear in v1.12.22 on 28 March. |
| May 2026 | Ethereum Classic Core Developers privately alert upstream developer again with the `ethereumclassic/core-geth` pre-release codebase — no response |

**Note on upstream v1.12.21/v1.12.22:** The upstream emergency patches address the CVE backports and are a safer option than v1.12.20 for operators who have not yet migrated. However, they remain on the Go 1.21 EOL toolchain and do not include the ETC network tooling, DNS discovery updates, or Olympia-readiness work included in v1.13.0. Operators running v1.12.x should upgrade to v1.12.22 at minimum and track [ethereumclassic/core-geth](https://github.com/ethereumclassic/core-geth) for the v1.13.0 release. Plan migration to [Fukuii](https://fukuii.com) ahead of the Olympia upgrade.

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
| CVE-2026-22868 | Medium | txpool / P2P — KZG DoS (blob/KZG proof verification) | Declared N/A to ETC by upstream¹ | Patched (peer disconnect on invalid proof) |
| CVE-2026-26313 | High | P2P — RLP item count memory exhaustion | Mitigated in v1.12.22²  | Patched |
| — (GraphQL depth) | Medium | RPC — unbounded query nesting DoS | Not addressed | Patched |

¹ Upstream maintainer [@diega confirmed](https://github.com/etclabscore/core-geth/issues/692) that CVE-2026-22868 is "not applicable to ETC" because ETC does not support EIP-4844 blob/KZG transactions. The v1.13.0 patch nonetheless implements the go-ethereum fix: peer disconnect on invalid KZG proof via `ErrKZGVerificationError` sentinel.

² The v1.12.22 mitigation prevents the OOM crash (no per-item allocation) but uses `rlp.CountValues`, which scans the full RLP payload before returning the item count. A malicious peer can still force O(n) tag-byte scanning work proportional to message size on each rejected oversized message — approximately 2,500× more work per 10 MiB attack message than v1.13.0's early-exit `countValuesExceedsLimit`. The CPU amplification DoS is unmitigated in v1.12.22. See the Residual Vulnerability note in the CVE-2026-26313 detail section below.

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

**Observed Attack — 18 March 2026:**
This vulnerability was actively exploited against the ETC mainnet classic bootnodes (ams3, sfo3) on 18 March 2026. Malicious P2P traffic sent crafted `auth` messages with undersized ECIES payloads, triggering the out-of-bounds slice allocation and crashing each node on the next inbound handshake attempt. Because the crash occurred in `listenLoop`, the node process exited and restarted under the service manager, only to crash again on the next malicious connection — producing a crash-loop.

The following stack trace, reported by node operator [@shrikus](https://github.com/shrikus) in [issue #692](https://github.com/etclabscore/core-geth/issues/692), confirms the call path:

```
panic: runtime error: makeslice: len out of range

goroutine 42797 [running]:
github.com/ethereum/go-ethereum/crypto/ecies.symDecrypt(...)
        crypto/ecies/ecies.go:224
github.com/ethereum/go-ethereum/crypto/ecies.(*PrivateKey).Decrypt(...)
        crypto/ecies/ecies.go:322
github.com/ethereum/go-ethereum/p2p/rlpx.(*handshakeState).readMsg(...)
        p2p/rlpx/rlpx.go:612
github.com/ethereum/go-ethereum/p2p/rlpx.(*handshakeState).runRecipient(...)
        p2p/rlpx/rlpx.go:415
github.com/ethereum/go-ethereum/p2p/rlpx.(*Conn).Handshake(...)
        p2p/rlpx/rlpx.go:308
github.com/ethereum/go-ethereum/p2p.(*rlpxTransport).doEncHandshake(...)
        p2p/transport.go:132
github.com/ethereum/go-ethereum/p2p.(*Server).setupConn(...)
        p2p/server.go:987
github.com/ethereum/go-ethereum/p2p.(*Server).listenLoop.func2()
        p2p/server.go:921
```

Per the v1.12.21 release notes ([PR #694](https://github.com/etclabscore/core-geth/pull/694)): bootnode **sfo3** had accumulated **805+ restart cycles** on v1.12.20 at the time of the patch, confirming the attack was ongoing and automated. The patched binary was deployed to **ams3** first and confirmed stable before sfo3 was upgraded.

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
**Patched:** ethereumclassic/core-geth v1.13.0
**Commit:** `1419c5310`
**Upstream reference:** go-ethereum commit `fdfd1235a` (v1.16.8)

**ETC Applicability:** The upstream maintainer [declared this not applicable to ETC](https://github.com/etclabscore/core-geth/issues/692) — ETC does not support EIP-4844 blob transactions, so the KZG code path is not reached in normal operation on the ETC network. The v1.12.22 release at `etclabscore/core-geth` does not address it. The v1.13.0 patch follows the go-ethereum approach: introduces `ErrKZGVerificationError` as a sentinel error and disconnects any peer that delivers a transaction with an invalid KZG proof, preventing repeated DoS attempts from the same peer.

**Description:**
KZG blob proof verification is computationally expensive. A malicious peer could repeatedly broadcast blob transactions with invalid KZG proofs, causing the node to perform the full expensive cryptographic verification on each delivery attempt before rejecting the transaction. Because the error was not distinguished from other validation failures, the peer was not disconnected and could continue flooding the node indefinitely.

**Impact:**
On ETC: no active exposure in normal operation (blob transactions are rejected at an earlier validation stage). The code path is nonetheless present in the inherited go-ethereum codebase and could be reached by a crafted peer interaction.

**Remediation:**
Ported from go-ethereum commit `fdfd1235a` (v1.16.8). Adds `ErrKZGVerificationError` as a sentinel error in `core/txpool/errors.go` and wraps KZG validation failures in `validateBlobSidecar()` with it. In `eth/fetcher/tx_fetcher.go`, any delivery flagged with this violation immediately triggers `f.dropPeer(delivery.origin)` — preventing the offending peer from repeatedly triggering expensive cryptographic verification.

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

**Note — minimal port:** This is a targeted backport of the item-count check only. The full upstream remediation (go-ethereum PR #33835) refactors message decoding to use `rlp.RawList` for delayed, lazy decoding across the entire handler layer. That refactor could not be cleanly applied here due to 13 merge conflicts arising from structural divergence between core-geth and go-ethereum. The top-level item count bound is sufficient to prevent the OOM; inner list items (e.g., transactions within a `BlockBody`) remain bounded by the existing 10 MiB `maxMessageSize` cap. A future full backport remains desirable but is deferred to avoid scope creep in the security patch.

**Forensic Analysis: Origin of the v1.12.22 Mitigation**

The v1.12.22 mitigation (`7940b2816`, "eth/protocols: pre-decode item count validation") uses an algorithm that does not exist in upstream go-ethereum and is not derivable from it. The upstream fix (PR #33835) uses `rlp.RawList` for lazy delayed decoding — incompatible with core-geth's structure due to 13 merge conflicts. The algorithm in v1.12.22 is a novel approach invented to work around that incompatibility. It first appeared in White B0x's `5d0cb8b34` (21 March 2026); it appears in Diego's commit seven days later.

Both implementations follow an identical five-step flow:

1. Read the message payload into a buffer via `io.ReadFull`
2. Restore `msg.Payload` to `bytes.NewReader(buf)` so the downstream decoder still works
3. Call `rlp.SplitList(buf)` to extract the list content
4. Optionally call `rlp.Split` to skip a RequestId wrapper (eth/66+), then `rlp.SplitList` again on the remainder
5. Count items in the resulting content slice

This buffer-restore-SplitList-Split-count pattern is specific to core-geth's architecture and is absent from all upstream go-ethereum commits. The v1.12.22 version differs from White B0x's in cosmetic respects only: it inlines all validation into `handler.go` rather than factoring it into a separate `msgvalidate.go`, uses the standard library `rlp.CountValues` rather than an early-exit `countValuesExceedsLimit` helper, defines per-type limits inline rather than as named constants, and includes fewer test cases. The algorithm structure is otherwise identical.

CVE-2026-26314 provides corroborating evidence: the `crypto/secp256k1/curve.go` and `crypto/secp256k1/ext.h` changes in v1.12.22 are byte-for-byte identical to White B0x's `2d3528803`. The `crypto/signature_nocgo.go` adaptation — an ETC-specific btcec wrapper override that does not exist in upstream go-ethereum — appears in both with a single cosmetic difference (White B0x uses an intermediate variable `p := curve.Params().P`; Diego inlines the call). The presence of this novel ETC-specific adaptation in both repositories, with the White B0x version dated 26 February and the Diego version dated 28 March, is consistent with the CVE-2026-26313 pattern.

**Residual Vulnerability in v1.12.22: Unbounded Tag-Scan DoS²**

The v1.12.22 mitigation uses `rlp.CountValues(content)`, which iterates through **all** items in the RLP payload before returning the total count. White B0x's implementation uses `countValuesExceedsLimit(content, limit)`, which exits as soon as the count exceeds the configured limit.

This distinction leaves an amplification vector open in v1.12.22: a malicious peer can send a single 10 MiB message packed with millions of minimal-length RLP items. v1.12.22 scans every tag byte in the payload before determining the count exceeds the limit and rejecting the message. v1.13.0 exits after scanning `limit + 1` items.

At the 10 MiB message cap with 2-byte items, an attacker can force approximately 2,500× more tag-byte scanning work per message on a v1.12.22 node than on a v1.13.0 node (≈5M iterations vs ≈2,049). Across multiple simultaneous malicious connections, this produces meaningful CPU exhaustion. The OOM crash from uncontrolled allocation is prevented; the CPU amplification DoS from unbounded tag scanning is not.

**Status in v1.12.22:** Partially mitigated — OOM crash prevented, CPU amplification DoS remains open.  
**Status in v1.13.0:** Fully patched.

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
| 10 June 2024 | Core-Geth v1.12.20 released at `etclabscore/core-geth` — final substantive release |
| August 2024 | Go 1.21 reaches end-of-life; upstream receives no further toolchain security patches |
| 23 January 2025 | Last commit to `etclabscore/core-geth` (GitHub Actions dependency bump — no code changes) |
| February 2025 | Private security disclosures sent to upstream maintainer — no response received |
| 4 February 2026 | [Issue #692](https://github.com/etclabscore/core-geth/issues/692): Ledger security researcher publicly discloses CVE-2025-24883, CVE-2026-22862, CVE-2026-22868 — no upstream response |
| 26 February 2026 | CVE-2025-24883 patch authored by White B0x on `chris-mercer/core-geth` (olympia branch) |
| 26 February 2026 | CVE-2026-26315 / CVE-2026-26314 patches authored by White B0x (`2d3528803`) |
| Early March 2026 | CVE-2026-22862 patch authored by White B0x during fukuii cross-client sprint (`dc73f2e4f`) |
| Early March 2026 | Go toolchain upgrade 1.21 → 1.24 authored by White B0x (`8385cf8e8` step 1) |
| 18 March 2026 | Active ECIES crash-loop attack on ETC bootnodes; `etclabscore` releases [v1.12.21 ("Aegis")](https://github.com/etclabscore/core-geth/releases/tag/v1.12.21) — emergency P2P patch ([PR #694](https://github.com/etclabscore/core-geth/pull/694)) |
| 18 March 2026 | @niooss-ledger [documents remaining unpatched CVEs](https://github.com/etclabscore/core-geth/pull/694#issuecomment-4089185353) after v1.12.21 |
| 20–21 March 2026 | White B0x publicly submits all CVE patches and Go 1.26 upgrade as PRs to `ethereumclassic/core-geth`; references in issue #692 to make them available for the upstream maintainer |
| 20 March 2026 | GraphQL depth limit patched (`6c2d383fa`); CVE-2026-22868 patched — peer disconnect on invalid KZG proof (`1419c5310`) |
| 21 March 2026 | CVE-2026-26313 patched (`5d0cb8b34`) |
| 28 March 2026 | `etclabscore` releases [v1.12.22 ("Hermes")](https://github.com/etclabscore/core-geth/releases/tag/v1.12.22) — CVE fixes drawn from publicly posted White B0x patches; no attribution; Go 1.21 EOL unchanged |
| TBD (Olympia) | Core-Geth v1.13.0 to be released at `ethereumclassic/core-geth` — all CVEs patched, Go 1.26, Olympia-ready; final release series before network transitions to Fukuii |
| May 2026 | Ethereum Classic Core Developers privately alert upstream developer with `ethereumclassic/core-geth` v1.13.0 codebase — no further activity at `etclabscore/core-geth` |

---

## Risk Assessment

| Risk Area | Severity | Description | Mitigation |
|-----------|----------|-------------|------------|
| Unpatched CVEs (5 + 1) | Critical | Five CVEs and a GraphQL DoS unaddressed for 21 months in the primary ETC client | All patched in v1.13.0 |
| Cryptographic oracle (CVE-2026-26315) | High | Repeated handshake attempts could leak P2P node key bits | Patched; key rotation recommended for long-running nodes |
| Remote crash via ECIES (CVE-2026-22862) | High | Any peer can crash a node during RLPx handshake with a malformed ECIES payload | Patched in v1.13.0 |
| Remote OOM via RLP (CVE-2026-26313) | High | Any peer can OOM-crash a node with a single crafted P2P message | Patched in v1.13.0 |
| CPU amplification DoS — CVE-2026-26313 residual (v1.12.22 only) | Medium | v1.12.22 mitigation scans full RLP payload before rejecting oversized messages; ~2,500× more work per attack message than v1.13.0; malicious peers can exhaust CPU without causing OOM | Unmitigated in v1.12.22; patched in v1.13.0 |
| Go Runtime EOL | High | 19 months on unsupported Go toolchain; runtime CVEs accumulated unpatched | Upgraded to Go 1.26 in v1.13.0 |
| Single-maintainer upstream | High | Private disclosures ignored for over a year; public CVE disclosure and an active network attack were required to force partial patches (v1.12.21/v1.12.22); Go toolchain and ETC modernisation not addressed | Client transferred to `ethereumclassic` org; multi-client Olympia architecture (Fukuii, Core-Geth, Besu) reduces single-client dependency |
| Release gap (21 months) | Medium | Longest maintenance gap in ETC network history; window for chain divergence or targeted attack | Protocol-funded maintenance path via ECIP-1112 treasury |

---

## Methodology

The audit was initiated during cross-client interoperability testing for the Olympia hard fork. The `etclabscore/core-geth` codebase at tag `v1.12.20` was diffed against the go-ethereum security advisory database (GitHub Security Advisories for `github.com/ethereum/go-ethereum`) and the Go vulnerability database (`vuln.go.dev`). Each advisory was assessed for applicability to core-geth's shared code ancestry. Applicable fixes were cherry-picked or ported from upstream go-ethereum commits; where upstream structural divergence made a clean cherry-pick impossible (CVE-2026-26313, CVE-2026-26314), fixes were manually ported. All patches were validated against the Mordor testnet prior to mainnet release.

---

## Network Migration Path

Core-Geth v1.13.x will be the final stable release series of this client. The ETC network is migrating to [Fukuii](https://fukuii.com) ([github.com/chippr-robotics/fukuii](https://github.com/chippr-robotics/fukuii)) as the primary ETC-native execution client, developed natively for the Ethereum Classic ecosystem — the only client built specifically for ETC from the ground up. Core-Geth will be maintained through the Olympia upgrade cycle to ensure a stable transition, but operators should plan their migration to Fukuii beyond that point.

**If you are running v1.12.20 or earlier, upgrade to v1.12.22 immediately** to address the most critical CVEs. v1.13.0 — which includes the full Go 1.26 upgrade, ETC-specific modernisation, and Olympia readiness — will be released under the `ethereumclassic` organisation ahead of Olympia activation. No v1.12.x release will be compatible with the Olympia network upgrade.

---

## Recommendations

- **Node operators (v1.12.20 or earlier):** Upgrade to v1.12.22 immediately to patch the most critical CVEs. Then track [github.com/ethereumclassic/core-geth](https://github.com/ethereumclassic/core-geth) for v1.13.0 ahead of Olympia activation. Nodes on v1.12.20 or earlier are exposed to remote crash and potential key-oracle attacks from any peer, and will not be compatible with the Olympia hard fork.
- **Infrastructure providers and exchanges:** Treat the upgrade to v1.13.0 as a security-critical update, not a routine version bump. Prioritise deployment ahead of Olympia activation. Begin planning migration to Fukuii for post-Olympia operation.
- **Long-running nodes:** Consider rotating the P2P node key (`--nodekey`) as a precaution against CVE-2026-26315 oracle exposure. Any node reachable from the public internet over the 21-month gap was potentially targeted.
- **Multi-client operation:** Run at minimum two independent clients (e.g., Core-Geth + Fukuii, or Core-Geth + Besu) for redundancy. The Olympia upgrade architecture is designed around multi-client operation precisely to mitigate the risk of a single-client maintenance failure.
- **GraphQL endpoint:** If `--graphql` is enabled on public-facing nodes, disable it until v1.13.0 is released or verify v1.12.22 is in place before re-opening the port.

---

## Postmortem: How the Maintenance Failure Happened

This section documents the complete issue and pull-request trail that preceded the March 2026 attack — organized chronologically so that anyone unfamiliar with the history can follow the chain of events through public links.

### Background for New Readers

`etclabscore/core-geth` was the primary Ethereum Classic execution client. It had first-class support for the ETC network baked in, was widely deployed, and served as the bootnodes for the ETC P2P network. The client was maintained under the `etclabscore` GitHub organisation by staff employed by [ETC Cooperative](https://etccooperative.org/).

In late 2024, ETC Cooperative announced it was putting the client into maintenance mode and winding down active Go development. The repository continued to accept automated CI changes (bot-generated bootnode removals, Dependabot security PRs) but no human maintainer was actively reviewing or merging code changes. The last substantive code release from this team was [v1.12.20](https://github.com/etclabscore/core-geth/releases/tag/v1.12.20) in June 2024.

The result: ETC's production P2P network ran for 21 months on a client whose toolchain was end-of-life and whose six known CVEs were never addressed — until an automated attack crashed the bootnodes and forced an emergency response.

---

### The Issue and PR Trail (Chronological)

Each item below is a public, linkable signal that the maintenance failure was accumulating. Together they form a complete audit trail.

#### 2021

**[Issue #292](https://github.com/etclabscore/core-geth/issues/292) — Jan 12, 2021 — @ziogaschr — OPEN**
*"`geth version-check` to be able to extend/add core-geth own vulnerabilities"*

The same maintainer who later authored the go-ethereum v1.14 merge PR filed this feature request five years before the attack: add vulnerability tracking to core-geth's built-in `version-check` command. Had this been implemented, the six CVEs published in 2025–2026 would have appeared in every node operator's log output, automatically surfacing the exposure without requiring anyone to monitor the go-ethereum security advisory database.

It was never implemented.

---

#### 2024

**[PR #646](https://github.com/etclabscore/core-geth/pull/646) — June 10, 2024 — @diega — MERGED**
*"Release/v1.12.20"*

The final substantive release from ETC Cooperative staff. No further code changes to the client's logic, crypto, P2P stack, or Go toolchain were merged after this date.

**[PR #649](https://github.com/etclabscore/core-geth/pull/649) — November 1, 2024 — @ziogaschr — OPEN (never merged)**
*"Merge/foundation release/v1.14"*

[@ziogaschr](https://github.com/ziogaschr) prepared a full merge of go-ethereum v1.14.0 into core-geth — a significant effort required to preserve PoW support while pulling in upstream improvements. The PR body reads: *"This PR has been tested and is ready for merge on master, though we won't merge it yet, as we continue testing it to ensure nothing is missing."*

As of the date of this audit, this PR remains open and unmerged. It is the most recent substantial community contribution to `etclabscore/core-geth` and illustrates the state of the repository: code is being contributed and tested but cannot be merged because no active maintainer is reviewing.

---

#### 2025 (January – June)

**[PR #654](https://github.com/etclabscore/core-geth/pull/654) — January 22, 2025 — @diega — MERGED**
*"Update GitHub Actions Artifacts to v4"*

This is the last commit @diega merged into `etclabscore/core-geth` before the March 2026 emergency patches — a GitHub Actions housekeeping change with no effect on client code or security. After this date, @diega was not active in the repository for 14 months, until the network attack forced a response.

**[PR #662](https://github.com/etclabscore/core-geth/pull/662) — January 23, 2025 — @dependabot — CLOSED WITHOUT MERGE**
*"build(deps): bump golang.org/x/crypto from 0.17.0 to 0.31.0"*

Dependabot's automated security tooling filed a PR to upgrade `golang.org/x/crypto` — the dependency that contains the ECIES implementation affected by CVE-2026-22862. This PR was auto-closed by the bot when a newer version superseded it, but no human ever evaluated or merged it. The vulnerability it would have helped address was later actively exploited.

**[PR #665](https://github.com/etclabscore/core-geth/pull/665) — February 9, 2025 — @cpuchainorg — OPEN (never merged)**
*"Fix build for golang 1.23"*

A community contributor cherry-picked a go-ethereum build-compatibility fix to allow core-geth to compile on Go 1.23. This fix was a prerequisite for any toolchain upgrade and was never reviewed or merged.

**[Issue #680](https://github.com/etclabscore/core-geth/issues/680) — April 30, 2025 — @user — CLOSED**
*Debian 11 / Go 1.24.0 build failure*

A node operator reported that core-geth failed to build against Go 1.24. The immediate workaround was to stay on the older Go toolchain — which was by this point EOL — rather than fix the build compatibility issue.

**[PR #671](https://github.com/etclabscore/core-geth/pull/671) — March 13, 2025 — @dependabot — CLOSED WITHOUT MERGE**
*"build(deps): bump golang.org/x/net from 0.18.0 to 0.36.0"*

A second security-adjacent Dependabot PR, auto-closed without human review. The `golang.org/x/net` package includes HTTP/2 and TLS code paths; the bump covered fixes to multiple advisories published since 2024.

**[PR #674](https://github.com/etclabscore/core-geth/pull/674) — March 21, 2025 — @dependabot — OPEN**
*"build(deps): bump github.com/golang-jwt/jwt/v4 from 4.5.0 to 4.5.2"*

A security fix for a JWT advisory (GHSA-mh63-6h87-95cp). Open and unreviewed.

**[PR #677](https://github.com/etclabscore/core-geth/pull/677) — April 14, 2025 — @dependabot — OPEN**
*"build(deps): bump golang.org/x/crypto from 0.17.0 to 0.35.0"*

A second attempt by Dependabot to upgrade `golang.org/x/crypto`, covering 18 additional months of security patches. Open and unreviewed.

**[PR #678](https://github.com/etclabscore/core-geth/pull/678) — April 16, 2025 — @dependabot — OPEN**
*"build(deps): bump golang.org/x/net from 0.18.0 to 0.38.0"*

A second attempt to upgrade `golang.org/x/net`. Open and unreviewed.

**[PR #683](https://github.com/etclabscore/core-geth/pull/683) — June 30, 2025 — @tornadocontrib — CLOSED WITHOUT MERGE**
*"Support go 1.24 compiler"*

The clearest missed signal in the entire trail. Community member [@tornadocontrib](https://github.com/tornadocontrib) submitted a complete PR that:
- Fixed the Go 1.23/1.24 build compatibility issues (#680, #665)
- Explicitly referenced and patched CVE-2025-24883 (GHSA-q26p-9cq4-7fc2) — the off-curve public key deserialization vulnerability

The PR body links directly to the go-ethereum security advisory. It received no review. When the contributor later deleted their fork (February 2026), GitHub auto-closed the PR. By the time the active attack occurred, this patch had been publicly available for nine months.

---

#### 2025 (August – December)

**[PR #685](https://github.com/etclabscore/core-geth/pull/685) — August 9, 2025 — @github-actions — OPEN**
*"params: remove unresponsive bootnodes"*

The automated bootnode health check filed this PR because deployed ETC bootnodes were not responding to `devp2p discv4 ping` requests. Several of these bootnodes were running the unpatched v1.12.20 and were likely already experiencing intermittent crash-loops from early-stage exploit probing. The PR was never merged.

**[PR #691](https://github.com/etclabscore/core-geth/pull/691) — November 10, 2025 — @cloorc — OPEN**
*"ci: upgrade blst to fix AppVeyor issue"*

A community contributor identified that the `blst` cryptographic dependency (used for BLS signature operations) had a build failure on the CI platform. The `blst` library was three years out of date; upgrading it would have also picked up the C23 `typedef bool` incompatibility fix required for the Alpine Docker builds in v1.13.0. Not reviewed.

---

#### 2026 (January – April)

**[Issue #692](https://github.com/etclabscore/core-geth/issues/692) — February 4, 2026 — @niooss-ledger — CLOSED**
*"Security vulnerabilities in go-ethereum affecting core-geth"*

[@niooss-ledger](https://github.com/niooss-ledger) — a security researcher at [Ledger](https://ledger.com) — publicly filed a detailed disclosure of CVE-2025-24883, CVE-2026-22862, and CVE-2026-22868 against core-geth. The issue included technical descriptions of each vulnerability, the affected code paths, and links to go-ethereum's published advisories.

**This issue received no response from any maintainer for 42 days.**

The public disclosure date is February 4, 2026. The first maintainer response was March 18, 2026, the same day the attack began. The response was the emergency release of v1.12.21 — not a reply to the issue.

**[PR #694](https://github.com/etclabscore/core-geth/pull/694) — March 18, 2026 — @diega — MERGED same day**
*"Release v1.12.21: Security hotfix for P2P crash-loop"*

Forced by the active ECIES crash-loop attack on the ETC bootnodes (ams3, sfo3). The PR body confirms: *"Bootnodes classic (ams3, sfo3) were in a crash-loop caused by malicious P2P traffic exploiting missing input validation in the ECIES handshake path."* Bootnode sfo3 had accumulated 805+ automatic restart cycles at the time of the patch.

The patches in this release were cherry-picks. The PR body notes they were taken *"from v1.12.20 tag, not master"* — confirming that the active `master` branch was considered too diverged to cherry-pick from reliably. [@niooss-ledger commented](https://github.com/etclabscore/core-geth/pull/694#issuecomment-4089185353) the same day identifying three additional CVEs (CVE-2025-24883, CVE-2026-26313, CVE-2026-26315) still unaddressed after v1.12.21.

**[PR #696](https://github.com/etclabscore/core-geth/pull/696) — March 28, 2026 — @diega — MERGED same day**
*"Merge release/v1.12.22 into master"*

The follow-on release backporting the remaining CVE fixes. The PR summary lists the patches applied but credits go-ethereum upstream commits rather than the White B0x work at `chris-mercer/core-geth` and the `ethereumclassic/core-geth` PR series that had been publicly available since March 20–21. The Go 1.21 EOL toolchain was left unchanged.

**[Issue #695](https://github.com/etclabscore/core-geth/issues/695) — March 19, 2026 — @CleyFaye — CLOSED**
*"No AMD64 build of latest docker image on dockerhub"*

The day after the emergency v1.12.21 release, a node operator reported that the Docker Hub image for the new release was missing the `linux/amd64` architecture variant. Operators who attempted to deploy the security patch via Docker on standard cloud infrastructure could not pull the image. This reflects the rushed nature of an emergency release cut by a team that had been inactive for 14 months.

**[Issue #697](https://github.com/etclabscore/core-geth/issues/697) — April 3, 2026 — @solopool — OPEN**
*"Incorrect RPC `eth_syncing` method response with v1.12.22"*

The rushed v1.12.22 release introduced a regression in the `eth_syncing` JSON-RPC method: the `highestBlock` value is reported incorrectly relative to the actual network head. Node operators and services that use `eth_syncing` to determine sync status — including exchanges, block explorers, and monitoring tools — receive incorrect data. This issue remains open and unaddressed.

---

### Structural Failures

The issue and PR trail above reflects five distinct structural failures, each of which was independent and any one of which would have been sufficient to prevent the March 2026 attack if it had been addressed.

#### 1. No CVE Tracking Infrastructure

[Issue #292](https://github.com/etclabscore/core-geth/issues/292) (2021) requested a `version-check` integration that would surface published CVEs to node operators directly. It was never implemented. As a result, node operators running v1.12.20 had no automated signal that their client was vulnerable — they had to independently monitor the go-ethereum GitHub Security Advisories page and manually cross-reference against core-geth's code ancestry.

#### 2. Maintainer Unresponsiveness to Security PRs

The automated security tooling was working correctly: Dependabot filed PRs for `golang.org/x/crypto` and `golang.org/x/net` in January 2025, March 2025, April 2025 (twice). All were either auto-closed or left open without review. A community member submitted a PR explicitly linking to CVE-2025-24883 in June 2025 — it received no response. A Ledger security researcher filed a detailed public CVE disclosure in February 2026 — it received no response for 42 days.

#### 3. Single Point of Human Authority

By mid-2025, @diega was the only person with merge access actively willing to cut releases. @meowsbits (Isaac Ardis) had departed. @ziogaschr had a ready-to-merge PR (#649) but no merge access to act on it. The `etclabscore` organisation's internal governance — whatever it was — did not provide a path for community contributors to advance security-critical changes without the approval of someone who was not actively reviewing.

#### 4. Go Toolchain Lock-in

Core-geth depended on the `fjl/memsize` package which was not compatible with Go 1.22+. This dependency meant that any Go toolchain upgrade required first removing or replacing `fjl/memsize`, a prerequisite that was known but never prioritised. Three separate community PRs attempted partial fixes (#665, #683, and the v1.14 merge in #649); none were merged. The incompatibility effectively locked the client to Go 1.21 — which reached end-of-life in August 2024 — for the entire period during which the CVEs were accumulating.

#### 5. Emergency Patches without Pre-release Testing

When the attack forced action, v1.12.21 was cut the same day as the PR was opened (5 hours from PR creation to merge), and v1.12.22 was cut with the same urgency (1 minute from PR creation to merge). The lack of any pre-merge review window meant a regression was introduced in `eth_syncing` (issue #697, still open) — the kind of secondary failure that only emerges when a team is responding reactively rather than maintaining proactively.

---

### Why the Client Moved to the `ethereumclassic` Organisation

The `etclabscore` organisation is controlled by ETC Cooperative. When the Cooperative announced in December 2024 that core-geth was entering maintenance mode, there was no transition plan for who would own security maintenance for the primary ETC execution client going forward. The five structural failures above were all visible at that point.

[White B0x](https://whiteb0x.com) forked the repository under the [`ethereumclassic` GitHub organisation](https://github.com/ethereumclassic/core-geth) — the same organisation that hosts the canonical ECIP repository and other ETC-native tooling. This location:

- Is not controlled by any single company (ETC Labs is gone; ETC Cooperative is in maintenance mode on this codebase)
- Aligns with the network's own GitHub home
- Provides a path for community contributors to have their work merged without depending on a single corporate maintainer's availability
- Houses v1.13.0, the only version of core-geth with all six CVEs patched, Go 1.26, and Olympia readiness

The longer-term migration path is to [Fukuii](https://fukuii.com) ([github.com/chippr-robotics/fukuii](https://github.com/chippr-robotics/fukuii)) — an ETC-native client built from the ground up in Scala 3 / Apache Pekko, with no Go toolchain dependency. Core-geth v1.13.x will be maintained through the Olympia upgrade cycle. Fukuii is the recommended production client beyond that point.

---

## References

- Patched client: https://github.com/ethereumclassic/core-geth
- White B0x source branch (olympia): https://github.com/chris-mercer/core-geth/tree/olympia
- Upstream (maintenance mode): https://github.com/etclabscore/core-geth
- Issue #692 — public CVE disclosure (Ledger security researcher): https://github.com/etclabscore/core-geth/issues/692
- PR #683 — community Go 1.24 + CVE-2025-24883 fix, closed without merge: https://github.com/etclabscore/core-geth/pull/683
- PR #694 — v1.12.21 emergency patch discussion and CVE analysis: https://github.com/etclabscore/core-geth/pull/694
- v1.12.21 ("Aegis") release: https://github.com/etclabscore/core-geth/releases/tag/v1.12.21
- v1.12.22 ("Hermes") release: https://github.com/etclabscore/core-geth/releases/tag/v1.12.22
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

**Postmortem — issue and PR trail:**
- Issue #292 — CVE tracking request (2021, never implemented): https://github.com/etclabscore/core-geth/issues/292
- Issue #697 — eth_syncing regression introduced by v1.12.22 (open): https://github.com/etclabscore/core-geth/issues/697
- Issue #695 — AMD64 Docker image missing after v1.12.21 emergency release: https://github.com/etclabscore/core-geth/issues/695
- PR #649 — go-ethereum v1.14 merge, ready-to-merge, stalled open since Nov 2024: https://github.com/etclabscore/core-geth/pull/649
- PR #654 — Last @diega commit before March 2026 emergency (GitHub Actions only): https://github.com/etclabscore/core-geth/pull/654
- PR #662 — Dependabot golang.org/x/crypto security bump, closed without merge (Jan 2025): https://github.com/etclabscore/core-geth/pull/662
- PR #665 — Go 1.23 build fix, never merged (Feb 2025): https://github.com/etclabscore/core-geth/pull/665
- PR #671 — Dependabot golang.org/x/net security bump, closed without merge (Mar 2025): https://github.com/etclabscore/core-geth/pull/671
- PR #677 — Dependabot golang.org/x/crypto 0.17→0.35, open unreviewed (Apr 2025): https://github.com/etclabscore/core-geth/pull/677
- PR #678 — Dependabot golang.org/x/net 0.18→0.38, open unreviewed (Apr 2025): https://github.com/etclabscore/core-geth/pull/678
- PR #683 — Community Go 1.24 + CVE-2025-24883 fix, explicitly referenced advisory, closed without merge (Jun 2025): https://github.com/etclabscore/core-geth/pull/683
- PR #685 — Unresponsive bootnodes bot PR, open unmerged (Aug 2025): https://github.com/etclabscore/core-geth/pull/685
- PR #691 — blst upgrade (stale dep, C23 fix), open unreviewed (Nov 2025): https://github.com/etclabscore/core-geth/pull/691
- PR #694 — v1.12.21 emergency patch (5 hours from PR open to merge): https://github.com/etclabscore/core-geth/pull/694
- PR #696 — v1.12.22 CVE backports (1 minute from PR open to merge): https://github.com/etclabscore/core-geth/pull/696
