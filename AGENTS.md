# AGENTS.md — CoreGeth (Ethereum Classic execution client)

Context for AI coding agents working in this repository. Human-facing overview is
in `README.md`; this file carries the build, test and boundary detail an agent
needs and a README should not.

## What this project is

CoreGeth is the Ethereum Classic (ETC) execution client, a fork of
[go-ethereum](https://github.com/ethereum/go-ethereum). It supports every ETC
hard fork from Frontier through Spiral and implements the **Olympia** upgrade
(ECIP-1111, ECIP-1112, ECIP-1121).

The client is in **maintenance mode and is scheduled for sunset after Olympia**;
`v1.13.x` is the final stable release series. See the warning at the top of
`README.md` before proposing any forward-looking work.

**This repository is a fork.** Its `origin` is a fork of
`ethereumclassic/core-geth`, which is where the Olympia release lands, and it
also carries `ethereum/go-ethereum` as a second upstream remote. Changes may be
sent upstream as pull requests, so keep diffs minimal and focused — see
Boundaries.

## Stack and versions

Read from this repository's own manifests, not from any external "current
version" list.

| Thing | Value | Source |
|---|---|---|
| Language | Go 1.26 | `go` directive in `go.mod` |
| Module path | `github.com/ethereum/go-ethereum` | `go.mod` (unchanged from upstream) |
| CI Go version | 1.26 | `.github/workflows/test-linux.yml` |
| Docker builder | `golang:1.26-alpine` | `Dockerfile` |
| Docker runtime | `alpine:latest` | `Dockerfile` |
| Build orchestration | `Makefile` → `build/ci.go` | `Makefile` |
| Linter | golangci-lint via `build/ci.go lint` | `.golangci.yml`, `Makefile` |

There is **no `toolchain` directive** in `go.mod` — only `go 1.26`.

## Commands

Every command below was read out of this repository's `Makefile`,
`build/ci.go`, `README.md`, or `.github/workflows/`. Do not invent others.

### Build

```bash
make geth                                      # → go run build/ci.go install ./cmd/geth
make all                                       # build every executable under cmd/
go run build/ci.go install -static ./cmd/geth  # static build (what the Dockerfile runs)
```

`make geth` writes the binary to `./build/bin/geth` (`GOBIN = ./build/bin`).

### Test

```bash
make test                                      # depends on `all`; runs build/ci.go test -timeout 20m
go run build/ci.go test -short                 # faster feedback while iterating
make test-coregeth                             # the core-geth-specific suite CI runs

# Narrower go test invocations documented in README.md:
go test ./params/... -run TestETC -v
go test ./core/... -run "TestGasLimit|TestForkCompliance|TestECIP1017|TestTreasury|TestOlympia" -v
go test ./consensus/ethash/... -run "TestDifficultyETC|TestDifficultyECIP" -v

# Consensus-oracle generator seam — see "Conformance-oracle contract" below:
go test ./consensus/clique/ -count=1
```

`build/ci.go test` accepts `-short`, `-race`, `-coverage`, `-v`, `-timeout`,
`-arch`, `-cc`, `-dlgo`, `-cachedir`.

**Live-network tests are build-tagged and require a running node.** They do not
run by default:

```bash
go test -tags live ./tests/live_etc/ -v        # //go:build live
```

### Lint

```bash
make lint                                      # → go run build/ci.go lint
```

`.golangci.yml` sets `disable-all: true` and enables **thirteen** linters:
`goimports`, `gosimple`, `govet`, `ineffassign`, `misspell`, `unconvert`,
`typecheck`, `unused`, `staticcheck`, `bidichk`, `durationcheck`,
`exportloopref`, `whitespace`. Read that file rather than this list — several
more are present but commented out, and its `issues.exclude-rules` block
carries `SA1019` deprecation suppressions that only make sense because
`staticcheck` is on.

### Generated code

31 `gen_*.go` files are generated, and 24 files carry `go:generate` directives.
CI enforces that they are current in `.github/workflows/go-generate-check.yml`,
which runs `make devtools` then `go generate` over `go list ./...` (excluding
`trezor`) and fails if the tree changes. Regenerate with `go generate ./...`
after installing `make devtools`; `solc` and `protoc` are required.

### Docs

```bash
pip install -r requirements-mkdocs.txt
make mkdocs-serve                              # → build/mkdocs-serve.sh
make docs-generate                             # regenerate JSON-RPC docs from OpenRPC
```

### Run a node

`run-mordor.sh` at the repository root starts a Mordor testnet node. It expects
`geth` on `PATH` and takes extra flags through `"$@"`. Read it before running
it — it hard-codes a data directory and binds RPC to `0.0.0.0`.

## Commands that do NOT exist here — do not call them

Absence is load-bearing. Each of these was checked in this tree:

- **`make evm`, `make evmc`, `make android`, `make ios`, `make geth-cross`** are
  listed in the `Makefile`'s `.PHONY` line but have **no rule**. `make -n evm`
  prints `Nothing to be done for 'evm'`. They are dead names inherited from
  upstream, not build targets.
- **`go run build/ci.go check_generate` does not exist in this fork.** The
  subcommands `build/ci.go` actually defines are `install`, `test`, `lint`,
  `archive`, `docker`, `debsrc`, `nsis`, `purge`, `sanitycheck`. Newer
  go-ethereum has `check_generate`; this fork does not. Use the CI workflow's
  `go generate` route instead.
- **`run-classic.sh` is not at the repository root.** Only `run-mordor.sh` is.
- **There is no `LICENSE` file** — see Licensing below.
- **There is no `.editorconfig` and no pre-commit hook configuration.** Match
  surrounding style by hand and rely on `goimports` via `make lint`; there is no
  local formatting gate that runs on its own.
- **`evmc/` and `core/genesis_alloc.go` do not exist**, although `.golangci.yml`
  still lists both under `skip-files:`. Stale upstream references.
- **`tests/evm-benchmarks` is an uninitialized submodule** in a fresh checkout.
  `tests/testdata` and `tests/testdata-etc` are populated. Run
  `git submodule update --init --recursive` before any suite that needs them.

## Repository structure

```
cmd/geth/          CLI entry point
core/              block processing, state, genesis
  vm/              EVM — interpreter, opcodes, precompiles
consensus/ethash/  ETChash proof-of-work (ETC difficulty rules)
params/            chain configs and protocol versions
eth/               wire protocol, sync
p2p/               networking and discovery
rpc/ graphql/      RPC transports
accounts/ signer/  account management and signing
crypto/            cryptographic primitives
trie/ triedb/      state trie
build/             ci.go build orchestration and helper scripts
internal/build/    build toolchain (gotool.go)
tests/             consensus test harness + submodule fixtures
  live_etc/        live-network tests, `live` build tag
docs/              mkdocs site, including docs/audits/
```

### Key files for ETC work

| File | Purpose |
|---|---|
| `params/config_classic.go` | ETC mainnet fork blocks, chain ID 61, ECBP1100 MESS |
| `params/config_mordor.go` | Mordor testnet fork blocks, chain ID 63 |
| `consensus/ethash/ethash.go` | ETChash PoW consensus |
| `core/vm/contracts.go` | precompile registry |
| `core/vm/opcodes.go` | EVM opcode definitions |
| `internal/build/gotool.go` | build flags — hard-codes `CGO_CFLAGS=-O2 -g -D__BLST_PORTABLE__ -std=gnu11`, which **overrides any `CGO_CFLAGS` set in the Docker environment**. `-std=gnu11` is what stops C23-aware compilers rejecting blst's legacy typedefs |

## Branching

**Commit directly to `main`.** Maintainer decision, 2026-08-11, recorded here so
it is settled once rather than re-decided per commit. This is a solo-maintained
repository: nothing else pulls `main`, and no CI, deploy or release triggers from
it. Do not open a topic branch for routine work on the assumption that a fork of
a consensus-critical client must always branch — that is ceremony here, not
protection.

The repository does carry topic branches using conventional prefixes
(`security/`, `test/`, `docs/`, `chore/`, `fix/`, `rlp/`) with kebab-case
descriptions. Use one when you actually want review before something lands, or
when `main` must stay releasable through a long-running change. `main` is the
default branch that release tags and upstream pull requests are cut from.

**Pushing is unaffected and remains a separate decision from committing.** It
must be confirmed explicitly, because this repository is public and pull requests
from it go to a consensus-critical upstream. The policy above governs where
commits land, not whether they leave the machine.

### Known merge collision: `AGENTS.md` itself

**`ethereum/go-ethereum` ships its own root `AGENTS.md`** — commit `406a852ec`,
2026-02-25, *"AGENTS.md: add AGENTS.md (#33890)"*. That commit is **already in
this clone**, reachable through the `geth` remote (`geth/master`) though not
from `main` or `origin/main`. So a future merge from `geth/master` will collide
on this file.

**Resolution, decided in advance: this fork's `AGENTS.md` wins.** Upstream's
file describes go-ethereum — it tells you to run `go run ./build/ci.go
check_generate`, a subcommand this fork does not have — so taking it would
hand an agent commands that fail here. Do not silently drop either side per the
usual merge rule; this is the documented exception, and the reason is that the
two files describe different programs.

Worth porting from upstream's version if it changes: its `gofmt`/`goimports`
pre-commit checklist. Do not port its command list wholesale.

`ethereumclassic/core-geth`, the repository pull requests are actually sent to,
has **no** `AGENTS.md`, so nothing collides in that direction.

## Conformance-oracle contract

A downstream conformance-test suite uses this repository as a **conformance
oracle** — it generates consensus fixtures by running this client's real engine,
then scores those fixtures against deliberately wrong builds. That role imposes
five requirements that the generic repo-wiring standard does not cover. Everything below was
measured in this clone on 2026-08-25; re-measure rather than trusting a figure.

### 1. Which branch is the oracle

**The `upstream` branch is the oracle. `main` is the ETC overlay and is never
one.**

| ref | is | measured 2026-08-25 |
|---|---|---|
| `main` | this fork's ETC overlay | 72 ahead, 0 behind `upstream` |
| `upstream` (local) | pegged to `ethereumclassic/core-geth` `master` | `4185df450`, 0/0 against both `upstream/master` and `origin/upstream` |
| `geth/master` | `ethereum/go-ethereum` | **shallow in this clone — see below** |

The peg is exact in both directions, so the convention already holds here and
needs no repair. Verify before generating:

```bash
git rev-list --left-right --count upstream/master...upstream   # must be 0	0
```

**`upstream` is ETC's own client master, which is itself an overlay of
go-ethereum.** That is sufficient for an *ETC* rule, and it is **not** sufficient
for a rule this client merely inherits from upstream Ethereum — Clique being the
live example. For those, `upstream` is not an independent reference and a
byte-identity proof against real go-ethereum is still owed.

**Do not use the `geth` remote for that proof.** It is a shallow fetch:

```bash
cat .git/shallow                      # 62ac0e05b, a grafted root
git rev-list --count geth/master      # 523, against upstream's 19481
git merge-base upstream geth/master   # exit 1 — NO merge base exists
```

File-content diffs against `geth/master` are still valid; anything requiring
history — merge-base, ancestry, "which go-ethereum commit did this sync from" —
silently returns a wrong or empty answer. Use a full-history `ethereum/go-ethereum`
clone instead, and confirm it is full (`.git/shallow` absent) and clean before
reading it. This machine's copy and the register recording which reference clones
track upstream and which are deliberately frozen are named in `CLAUDE.local.md`;
they are machine-local and do not belong in this file.

### 2. The generator seam

A fixture generator is a `_test.go` file placed **inside** the consensus package,
which is what gives it the unexported identifiers a generator must not
reimplement.

```bash
ORACLE_SEAM_OUT=/path/out.json \
  go test ./consensus/clique/ -run TestGenerateOracleSeamProof -count=1 -v
```

Proven by execution on 2026-08-25, not by description. A throwaway generator in
`package clique` reached `extraVanity`, `extraSeal`, `nonceAuthVote`,
`nonceDropVote`, `diffInTurn`, `diffNoTurn` and `errRecentlySigned`, drove the
unexported `(*Clique).snapshot()` and `ecrecover()` against a real
`core.NewBlockChain`, round-tripped a seal through `SealHash` back to the signing
address, and wrote JSON to the env-given path.

The same seam exists for the other consensus packages — `consensus/ethash/` for
ETChash and the ETC difficulty rules.

### 3. The clean-clone invariant

**The generator file is added, run, and removed, and the clone ends at zero
modified and zero untracked.** A dirty oracle clone silently poisons every later
read of it. Keep the generator's source in the consuming suite's own tooling
directory so the run is reproducible without this clone ever holding it. Verify by effect
after every run:

```bash
git status --porcelain | wc -l        # must be 0
```

### 4. Measured cost of one wrong-build cycle

Go on this machine, warm build cache:

| step | measured |
|---|---|
| `go build ./consensus/clique/` | ~0.2 s |
| `go test ./consensus/clique/ -count=1` | ~1.0 s |
| full patch → build → test → revert | **~1.3 s** |

A six-defect wrong-build matrix is therefore seconds, not an afternoon, and needs
no resource budgeting under `rules/resource-management.md`. This is a
package-scoped figure and says nothing about `make test` or `make all`, which are
a full client build plus consensus suites — see the warning in `CLAUDE.md`.

**Source-changed plus compiles is necessary and NOT sufficient.** The handoff
this section implements requires both checks, and on 2026-08-25 a patch passed
both and was still behaviorally inert: swapping `header.GasLimit` and
`header.GasUsed` in `encodeSigHeader` changed one line, compiled, and moved
nothing, because both fields are zero in `TestSealHash`'s vector. It produced a
clean pass indistinguishable from a coverage gap. **Add a third check: the
mutation must be reachable and distinguishing for the inputs the suite actually
uses**, and pair every NOT-CAUGHT with a known-CAUGHT control before reporting it.
Dropping `header.MixDigest` from the same list changes the RLP list length, so it
moves the hash even with all-zero fields, and it correctly fails `TestSealHash`.

**Measured coverage gap, with that control in place:** mutating `diffInTurn` from
2 to 3 changed source, compiled, and **passed** `go test ./consensus/clique/`.
Every reference to `diffInTurn` in the package's tests is the constant itself
(`clique_test.go:66`, `clique_test.go:84`, `snapshot_test.go:449`), so subject and
assertion move together. **This client's own clique tests are not a safety net for
that rule** — which is the case for the external fixtures, not an argument against
them.

### 5. Independence

**Two clients sharing a root commit are one opinion, not two.**

| property | value |
|---|---|
| first-parent root commit | `5db3335dc` (2013-12-26, "Initial commit") — **go-ethereum's own** |
| module identity | `module github.com/ethereum/go-ethereum` — unchanged from upstream |

```bash
git rev-list --first-parent --max-parents=0 HEAD
head -1 go.mod
```

`git rev-list --max-parents=0 HEAD` reports **six** roots here, from side branches
merged in over the years; the `--first-parent` form above is the one that answers
lineage. Neither the directory name nor `go.mod` reveals that this is an ETC
client, which is exactly why the shared root gets counted twice.

**Already known to share `5db3335dc`:** `ethereumclassic/core-geth`,
`etclabscore/core-geth`, `multi-geth` and `ethereumproject/go-ethereum`. Treat any
agreement among them as **one** data point. `besu` and `nethermind` are genuinely
independent implementations.

## Code style

- Standard Go formatting: `gofmt` and `goimports`. `goimports` is enforced by
  `make lint`.
- Match the surrounding file. Much of this tree is upstream go-ethereum code
  under its own conventions; ETC-specific additions live alongside it.
- Every source file carries a GPL/LGPL header. Preserve it, and follow the
  existing header form when adding a file.

## Security

`SECURITY.md` at the root carries the disclosure policy and PGP key.

`docs/audits/2026-03-security-audit.md` documents the March 2026 remediation of
six CVEs plus a GraphQL depth DoS. Read it before touching `crypto/`, `p2p/`,
`rlp/` or `graphql/` — those are where the patched issues lived.

Never commit node keys, keystores, mnemonics or JWT secrets. `.gitignore`
covers `keystore/`, `*.keystore`, `nodekey`, `jwt.hex`, `jwtsecret`,
`wallet.json`, `mnemonic.txt`, `*.key`, `*.pem` and the `.env` family; verify
with `git check-ignore --no-index -q -- <path>` rather than by reading it.

### Dependency updates — configured, and deliberately off

**Decision: version updates stay OFF. Taken 2026-08-21, recorded here so it is
settled once rather than re-decided per session.**

`.github/dependabot.yml` carries a single `gomod` entry with
`open-pull-requests-limit: 0`. That is GitHub's documented way to disable version
updates; the entry names the ecosystem that *would* be used if they were turned
on. There is no `cooldown:` block, and its absence is correct rather than an
oversight — a cooldown gates nothing at a zero limit, and an inert setting reads
as a control that is operating. Raising the limit and adding a cooldown are one
change, not two.

**The reason is this client's lifecycle, not a judgment about supply-chain risk.**
CoreGeth is in maintenance mode and scheduled for sunset after Olympia. A standing
weekly pull-request queue buys version currency, which is worth its cost only
where someone triages it.

**What covers the surface instead — and it is run by hand, not on a schedule.**
Go's own toolchain, which does not depend on this file: `go list -m -u all` for
module retractions, `govulncheck` for known advisories against resolved versions.
Treat a retraction as a prompt to check reachability and advisories, not as
evidence of exposure; the common case is maintainer hygiene.

**Naming those two tools is not the same as running them, and for months it was
not.** `govulncheck` was first actually executed against this repository on
2026-08-25, having been credited as the coverage since 2026-08-21 while being
absent from every reachable path. **Nothing in this repository schedules either
tool**, so the sentence above describes an instrument, not a control that is
operating — the same distinction the `cooldown:` note above draws, and the reason
that note exists. State what runs; do not let a tool's existence read as coverage.

**How to actually run them**, from the repository root:

```bash
go list -m -u all | grep -i retract        # module retractions; pull-based, surfaces nothing unasked
govulncheck ./...                          # source mode: module + stdlib reachability
govulncheck -mode binary build/bin/geth    # binary mode: what you actually run
```

Exit `3` means vulnerabilities found, `0` clean, `127` the tool is absent. Treat
any other code as did-not-run, never as clean — a scanner that is missing or stale
reports the same silence as a healthy one.

**Security updates are a separate repository setting that this file cannot turn
on or off, and a zero limit does not suppress them.** Measured against the GitHub
API on 2026-08-21, both are off for this repository: `automated-security-fixes`
returned `{"enabled":false,"paused":false}` and `vulnerability-alerts` returned
HTTP 404.

**Re-check that, do not re-read it.** This is a standing condition, not a closed
item. The repository is quiet because of those two settings — not because of the
limit in the config. Either can be flipped by anyone with admin access, and
Dependabot's per-ecosystem support grows over time, so the repository can become
outward-facing with nobody having edited `dependabot.yml`. Query the API rather
than trusting this paragraph.

### Adjudicated `govulncheck` findings — reported forever, and that is expected

**Six advisories are reported permanently and have been adjudicated. Do not
re-adjudicate them from scratch, and do not try to close them with a version
bump — neither set can be.** Evidence per advisory is in the response document
this repository keeps outside the tracked tree; the disposition and the rule are
here because this file is what the next agent reads.

**Binary mode — five `github.com/ethereum/go-ethereum` advisories:** `GO-2026-4314`,
`GO-2026-4315`, `GO-2026-4507`, `GO-2026-4508`, `GO-2026-4511`. All five are already
backported into this fork, each traceable to a core-geth commit ancestral to the
build revision. They recur **structurally**: the fork keeps go-ethereum's module
path, so its pseudo-version always sorts below upstream's fixed tags, and
`govulncheck` matches symbol *names* while a backport adds guards inside the same
function. Verified in this tree — the ECIES point checks in `crypto/ecies/ecies.go`,
the field-boundary check in `crypto/secp256k1/curve.go`, and `countValuesExceedsLimit`
in `eth/protocols/{eth,snap}/msgvalidate.go`. **The RLP fix is not in `rlp/`** — a grep
scoped there finds nothing and reads as absence.

**Source mode — `GO-2026-5932`, `golang.org/x/crypto/openpgp`, `Fixed in: N/A`.**
Unmaintained and unsafe by design, so no version closes it. Scoped, measured both
directions: `go list -deps ./cmd/geth | grep -c openpgp` is **0**;
`go list -deps ./internal/build | grep -c openpgp` is **6**. It is **not in the
shipped binary**.

**It is not dead code either, and the disposition rests on why it is acceptable
rather than on it being unused.** `.github/workflows/release-packages.yml` runs
`build/archive-signing.sh`, which reaches `ci.go archive -signer` and
`build.PGPSignFile`. So it executes on every signed release. **Accepted** because
the only thing openpgp *parses* is the signing key read from a CI-secret
environment variable, and it signs this project's own archive — no
attacker-supplied input reaches the parser, and anyone who can set that variable
already controls the release. Removing it would break release signing; a build tag
cannot help, because the release path is what needs it.

**Re-check condition, and this is the part that expires:** the disposition holds
only while no externally-supplied PGP material reaches that path. If anything here
ever verifies third-party signatures or parses user-supplied keys, it must be
re-decided.

**The rule for both sets: compare the exact set of advisory IDs, and treat any
difference in either direction as a finding.** A new ID is an unadjudicated
advisory. A *missing* ID is equally a finding — it means the adjudication went
stale or the artifact is not the one that was adjudicated. **Never suppress by
module or by category**; that hides the next real advisory behind the known ones,
which is worse than the noise it removes. Because findings are expected here, a
`govulncheck` exit of `0` is itself a change, not a pass — and any exit other than
`3` is did-not-run, never clean.

A checker implementing exactly that comparison lives with this repository's other
machine-local scripts; it is deliberately not tracked, because it hard-codes paths
for one machine. Its contract: exit `0` set matches, `1` set changed, `2`
did-not-run. It was calibrated against a deliberately wrong expected set before
being trusted.

**It verifies the scanner before it compares the set**, because an exact-set gate
is only as good as the tool producing the set. Three `govulncheck` binaries exist
on this machine — the pinned one on `PATH` and two stranded older copies — and an
older scanner reports a **smaller** set, which would surface as "adjudicated
advisories GONE": a confident wrong answer pointing at this repository instead of
at `PATH`. So a shadowed or mismatched scanner fails as **did-not-run**, never as
set-changed. The pinned version is read from the machine's own pin file rather
than restated here, and the version is parsed anchored on the `Scanner:` line —
`govulncheck -version` prints the **Go toolchain** version first, so a bare
`\d+\.\d+\.\d+` match returns the wrong number and would move on every Go bump.

## Licensing — do not change

There is **no file named `LICENSE`**, and that is correct, not an omission. The
project uses go-ethereum's split licensing:

- `COPYING` — GNU **GPL-3.0**, applying to the binaries under `cmd/`
- `COPYING.LESSER` — GNU **LGPL-3.0**, applying to the library (everything
  outside `cmd/`)

Per-file headers state which applies. Do not add a `LICENSE` file, do not
consolidate the two, and do not "fix" a community-standards checker that flags
the absence — the split is deliberate and inherited from upstream.

## Boundaries

### Never without asking

- **Consensus-critical behavior.** `consensus/`, `core/` block processing,
  `core/vm/` opcode and gas semantics, and the fork-block tables in `params/`. A
  wrong value here splits the chain.
- **Chain configuration.** `params/config_classic.go` and
  `params/config_mordor.go`. Fork blocks, chain IDs and MESS activation are
  network-wide facts, not local settings.
- **`internal/build/gotool.go` `CGO_CFLAGS`.** Changing it breaks the blst/c-kzg
  build in ways that surface only at link time.
- **Dependency changes** — `go.mod`, `go.sum`, and especially `blst`, `c-kzg`,
  `gnark-crypto`, `btcec`, `graphql-go`. Route through the repository's
  dependency-review process.
- **Submodule pins** in `.gitmodules` / `tests/testdata*`. These are consensus
  test vectors; moving them changes what "passing" means.
- **Generated files** (`gen_*.go`). Regenerate them; do not hand-edit.
- **CI and release configuration** — `.github/workflows/`, `Dockerfile`,
  `Dockerfile.alltools`, `build/ci.go`.
- **`README.md`'s maintenance-mode and migration warnings.** They are a
  coordinated public statement, not editorial copy.

### Never

- Break wire-protocol compatibility with the ETC network. **EIP-7642 (eth/69) is
  deliberately excluded** because it removes Total Difficulty from the handshake,
  which ETC needs for proof-of-work chain selection. Do not "restore" it.
- Commit private keys, node keys or mnemonics.
- Remove or skip tests to make a build pass.
- Refactor, rename or reformat code unrelated to the task. Diffs from this
  repository may become upstream pull requests, and unrelated churn gets them
  rejected.

### Deliberately present, do not tidy

- **`git.diff`** — a ~6 MB tracked diff artifact at the repository root.
- **Legacy CI files** — `.travis.yml`, `circle.yml`, `appveyor.yml`,
  `Jenkinsfile`, `oss-fuzz.sh`. The live CI is `.github/workflows/`. **"Inherited
  from upstream" is provenance, not justification, and it is wrong for two of
  them.** Measured against a full-history go-ethereum clone, 2026-08-25:
  go-ethereum deleted `.travis.yml` (2025-06-26), `circle.yml` (2026-01-17,
  #33616) and `appveyor.yml` (2026-05-10, #34720), so deleting those three here
  would *converge* with it — while `ethereumclassic/core-geth`, where releases
  land, still carries all of them. `Jenkinsfile` appears in **no** go-ethereum
  commit ever; it is ETC's own. And `oss-fuzz.sh` is **not legacy at all** —
  it is present in go-ethereum master today and is Google's OSS-Fuzz build
  entry point. Do not sweep these as a group.
- **`swarm/` and `integration/`** — near-empty holdover directories.
- **`AUTHORS` and `.mailmap`** — upstream attribution records.
- **`accounts/keystore/`** — 32 tracked source files and test vectors, not key
  material, despite the directory name. `.gitignore`'s `/keystore/` entry is
  anchored specifically so it does not shadow them.

- **`build/checksums.txt`'s `ppa-builder` pin (Go 1.19.6) is knowingly
  insufficient — do not "fix" it by bumping, and do not re-derive this.** The
  file has **two** Go pins serving different paths, and conflating them inverts
  which one matters:
  - **`version:golang`** is read by `DownloadGo` (`internal/build/gotool.go`) for
    `-dlgo`, which downloads a **binary** release. `.github/workflows/release-packages.yml`
    runs `ci.go install -dlgo` for every ARM and arm64 target, so this pin is
    **live and release-critical**. It is at 1.26.6.
  - **`version:ppa-builder`** is read only by `downloadGoBootstrapSources`
    (`build/ci.go`), the compiler used to build Go **from source** on the
    Launchpad path. That path is reached only from `debsrc`, invoked only by
    `.travis.yml` and `build/bot/ppa-build.sh` — neither of which the live CI
    runs. It is at 1.19.6.

  Per <https://go.dev/doc/install/source>, Go 1.22–1.23 require a Go 1.20
  compiler, and *"going forward, Go 1.N will require a Go 1.M compiler, where M
  is N-2 rounded down to an even number"* — so **Go 1.26 requires Go 1.24**.
  1.19.6 cannot bootstrap it. **It could not bootstrap the previous 1.22.1 pin
  either**, which needed 1.20, so this predates the 1.26.6 bump rather than being
  caused by it.

  The file's own comment anticipates exactly this: *"If it ever becomes
  insufficient, we need to switch over to a recursive builder to jump across
  supported versions."* That clause has fired. A single bump cannot fix it —
  reaching 1.24 from 1.19.6 needs a chain — and the remedy is upstream's design
  decision, not this fork's. Left as-is deliberately, on a path nothing here
  runs.

### Environment

Machine-specific facts — data directories, local run scripts, node identities —
do not belong in this file, because it is public. Keep them in `CLAUDE.local.md`
or under `.local/`, both of which `.gitignore` holds back.

`.claude/settings.json` is tracked and travels to clones. It pre-approves only
this repository's cheap, documented commands: the narrow `-run`-filtered `go
test` invocations above, the `consensus/clique` package test and
`go build ./consensus/...` that the oracle contract's generator seam and
wrong-build cycle need (measured ~1.0 s and ~0.2 s), `go vet`, `gofmt -l`,
`make lint`, and `git submodule status`.

**The heavy targets are omitted deliberately.** `make all`, `make test`,
`make test-coregeth` and `make geth` are a full client build plus consensus
suites, and this machine runs one heavy task at a time. Leaving them unlisted
keeps the approval prompt as the gate on starting one. Do not "complete" the list
by adding them.
