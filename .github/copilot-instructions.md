# GitHub Copilot Instructions — CoreGeth

<!--
  SELF-CONTAINED, not a delta pointing at AGENTS.md. Chosen because this is a
  public repository whose contributors' toolchains are unknown, and several
  Copilot surfaces (github.com Chat, VS Code code review, Visual Studio,
  and the Chat and code-review surfaces of JetBrains, Eclipse and Xcode) do not
  read AGENTS.md at all. On those surfaces this file is the only instruction the
  model sees, so a thin pointer would leave it with nothing.

  Cost of that choice: this file duplicates AGENTS.md. Keep the two in sync when
  either changes. Where a surface reads both, both are supplied to the model, so
  the rule that matters is that they must not contradict each other.
-->

## Project

CoreGeth is the Ethereum Classic (ETC) execution client, a fork of
[go-ethereum](https://github.com/ethereum/go-ethereum). It supports every ETC
hard fork from Frontier through Spiral and implements the **Olympia** upgrade
(ECIP-1111, ECIP-1112, ECIP-1121).

The client is in **maintenance mode and is scheduled for sunset after Olympia**;
`v1.13.x` is the final stable release series. See `README.md` before proposing
forward-looking work.

**This repository is a fork and sends pull requests upstream.** Keep diffs
minimal and focused. Unrelated refactors, renames and reformatting get upstream
PRs rejected.

## Stack

Read from this repository's own manifests, never from an external
"current version" list.

| Thing | Value | Source |
|---|---|---|
| Language | Go 1.26 | `go` directive in `go.mod` |
| Module path | `github.com/ethereum/go-ethereum` | `go.mod` |
| CI Go version | 1.26 | `.github/workflows/test-linux.yml` |
| Docker builder | `golang:1.26-alpine` | `Dockerfile` |
| Build orchestration | `Makefile` → `build/ci.go` | `Makefile` |
| Linter | golangci-lint via `build/ci.go lint` | `.golangci.yml` |

There is no `toolchain` directive in `go.mod`.

## Commands

Every command below exists in this repository's `Makefile`, `build/ci.go`,
`README.md` or `.github/workflows/`. Do not invent others.

```bash
make geth                                      # build geth → ./build/bin/geth
make all                                       # build every executable under cmd/
go run build/ci.go install -static ./cmd/geth  # static build (what the Dockerfile runs)

make test                                      # depends on `all`; ci.go test -timeout 20m
go run build/ci.go test -short                 # faster feedback while iterating
make test-coregeth                             # the core-geth-specific suite CI runs

make lint                                      # → go run build/ci.go lint

go test -tags live ./tests/live_etc/ -v        # live-network tests; needs a running node
```

`build/ci.go test` accepts `-short`, `-race`, `-coverage`, `-v`, `-timeout`,
`-arch`, `-cc`, `-dlgo`, `-cachedir`.

`.golangci.yml` sets `disable-all: true` and enables **thirteen** linters:
`goimports`, `gosimple`, `govet`, `ineffassign`, `misspell`, `unconvert`,
`typecheck`, `unused`, `staticcheck`, `bidichk`, `durationcheck`,
`exportloopref`, `whitespace`. Read that file rather than this list; its
`issues.exclude-rules` block carries `SA1019` suppressions that exist only
because `staticcheck` is enabled.

## Commands that do NOT exist here

Do not call these. Each was checked against this tree.

- `make evm`, `make evmc`, `make android`, `make ios`, `make geth-cross` — named
  in the `Makefile`'s `.PHONY` line but with **no rule**. `make -n evm` prints
  `Nothing to be done for 'evm'`.
- `go run build/ci.go check_generate` — **not in this fork.** The subcommands
  `build/ci.go` defines are `install`, `test`, `lint`, `archive`, `docker`,
  `debsrc`, `nsis`, `purge`, `sanitycheck`. Use the `go generate` route in
  `.github/workflows/go-generate-check.yml` instead.
- `run-classic.sh` — not at the repository root. Only `run-mordor.sh` is.
- There is **no `.editorconfig` and no pre-commit hook config.** Match
  surrounding style by hand; `goimports` runs only under `make lint`.
- `evmc/` and `core/genesis_alloc.go` do not exist, despite stale `skip-files:`
  references in `.golangci.yml`.
- `tests/evm-benchmarks` is an uninitialized submodule in a fresh checkout. Run
  `git submodule update --init --recursive` before suites that need fixtures.

## Generated code

31 `gen_*.go` files are generated and 25 files carry `go:generate` directives.
CI fails if they are stale (`.github/workflows/go-generate-check.yml`).
Regenerate with `go generate ./...` after `make devtools`; `solc` and `protoc`
are required. Never hand-edit a `gen_*.go`.

## Key files

| File | Purpose |
|---|---|
| `params/config_classic.go` | ETC mainnet fork blocks, chain ID 61, ECBP1100 MESS |
| `params/config_mordor.go` | Mordor testnet fork blocks, chain ID 63 |
| `consensus/ethash/ethash.go` | ETChash PoW consensus |
| `core/vm/contracts.go` | precompile registry |
| `core/vm/opcodes.go` | EVM opcode definitions |
| `internal/build/gotool.go` | hard-codes `CGO_CFLAGS=-O2 -g -D__BLST_PORTABLE__ -std=gnu11`, overriding any Docker-set `CGO_CFLAGS` |

## Branching

**Commit directly to `main`.** Maintainer decision, 2026-08-11, recorded so it
is settled once rather than re-decided per commit. This is a solo-maintained
repository: nothing else pulls `main`, and no CI, deploy or release triggers
from it. Do not open a topic branch for routine work on the assumption that a
fork of a consensus-critical client must always branch — that is ceremony here,
not protection.

The repository does carry topic branches using conventional prefixes
(`security/`, `test/`, `docs/`, `chore/`, `fix/`, `rlp/`) with kebab-case
descriptions. Use one when you actually want review before something lands, or
when `main` must stay releasable through a long-running change. `main` is the
default branch that release tags and upstream pull requests are cut from.

**Pushing is unaffected and remains a separate decision from committing.** It
must be confirmed explicitly, because this repository is public and pull
requests from it go to a consensus-critical upstream. The policy above governs
where commits land, not whether they leave the machine.

**Known merge collision on `AGENTS.md`.** `ethereum/go-ethereum` ships its own
root `AGENTS.md` (commit `406a852ec`, 2026-02-25), already present in this clone
via the `geth` remote, so a merge from `geth/master` will conflict on it.
**This fork's `AGENTS.md` wins** — upstream's describes go-ethereum and names
commands this fork does not have, such as `go run ./build/ci.go check_generate`.
`ethereumclassic/core-geth`, the actual PR target, has no `AGENTS.md`.

## Conformance-oracle contract

A downstream conformance-test suite uses this repository as a **conformance
oracle**: it generates consensus fixtures by running this client's real engine,
then scores them against deliberately wrong builds. Five requirements follow.
Every figure below was measured in a clone on 2026-08-25 — re-measure rather
than trusting one.

1. **`upstream` is the oracle branch; `main` is the ETC overlay and never is.**
   Verify the peg before generating: `git rev-list --left-right --count
   upstream/master...upstream` must print `0	0`. `upstream` is ETC's own client
   master, which is itself an overlay of go-ethereum — enough for an *ETC* rule,
   not for a rule merely inherited from upstream Ethereum (Clique is the live
   example), which still owes a byte-identity proof against real go-ethereum.
2. **Do not use the `geth` remote for that proof — it is a shallow fetch.**
   `git merge-base upstream geth/master` exits 1; there is no merge base. File
   content diffs against it are valid; anything needing history returns a
   confidently wrong or empty answer. Use a full-history `ethereum/go-ethereum`
   clone and confirm it is not shallow first.
3. **The generator seam is a `_test.go` file placed inside the consensus
   package**, which is what gives it the unexported identifiers a generator must
   not reimplement — proven by execution against `consensus/clique/`, and the
   same seam exists for `consensus/ethash/`. The generator is added, run, and
   removed: **the clone ends at zero modified and zero untracked**
   (`git status --porcelain | wc -l` must be 0). A dirty oracle clone silently
   poisons every later read of it, so keep the generator's source in the
   consuming suite's own tooling directory.
4. **Source-changed plus compiles is necessary and NOT sufficient** when scoring
   a wrong build. A patch that changed one line and compiled was still
   behaviorally inert, producing a clean pass indistinguishable from a coverage
   gap. Add a third check — the mutation must be reachable and distinguishing
   for the inputs the suite actually uses — and pair every NOT-CAUGHT with a
   known-CAUGHT control before reporting it. One full patch → build → test →
   revert cycle on `consensus/clique/` measured ~1.3 s, so a wrong-build matrix
   is seconds; that says nothing about `make test` or `make all`.
5. **Two clients sharing a root commit are one opinion, not two.** This fork's
   first-parent root is `5db3335dc`, go-ethereum's own, and `go.mod` still reads
   `module github.com/ethereum/go-ethereum` — so neither the directory name nor
   the module path reveals this is an ETC client. `ethereumclassic/core-geth`,
   `etclabscore/core-geth`, `multi-geth` and `ethereumproject/go-ethereum` all
   share it; treat agreement among them as **one** data point. `besu` and
   `nethermind` are genuinely independent.

Reference-clone locations are machine-local and are deliberately not in this
file, `AGENTS.md` or `CLAUDE.md`.

## Code style

- `gofmt` and `goimports`; `goimports` is enforced by `make lint`.
- Match the surrounding file — most of this tree is upstream go-ethereum code.
- Every source file carries a GPL/LGPL header. Preserve it and follow the
  existing form when adding a file.

## Security

`SECURITY.md` at the root carries the disclosure policy and PGP key.

`docs/audits/2026-03-security-audit.md` documents the March 2026 remediation of
six CVEs plus a GraphQL depth DoS. Read it before touching `crypto/`, `p2p/`,
`rlp/` or `graphql/` — those are where the patched issues lived.

Never commit node keys, keystores, mnemonics or JWT secrets. `.gitignore` and
`.dockerignore` both cover them; verify with
`git check-ignore --no-index -q -- <path>` rather than by reading either file.

### Dependency updates — configured, and deliberately off

**Version updates stay OFF. Decision taken 2026-08-21**, recorded so it is
settled once rather than re-decided per session.

`.github/dependabot.yml` carries a single `gomod` entry with
`open-pull-requests-limit: 0` — GitHub's documented way to disable version
updates. The entry names the ecosystem that *would* be used if they were turned
on. There is no `cooldown:` block, and that absence is correct rather than an
oversight: a cooldown gates nothing at a zero limit, and an inert setting reads
as a control that is operating. Raising the limit and adding a cooldown are one
change, not two.

**The reason is this client's lifecycle, not a judgment about supply-chain
risk.** CoreGeth is in maintenance mode and scheduled for sunset after Olympia.
A standing weekly pull-request queue buys version currency, which is worth its
cost only where someone triages it.

**What covers the surface instead — run by hand, not on a schedule.** Go's own
toolchain, which does not depend on that file: `go list -m -u all` for module
retractions, `govulncheck` for known advisories against resolved versions. A
retraction is a prompt to check reachability and advisories, not evidence of
exposure — the common case is maintainer hygiene.

**Naming those tools is not the same as running them, and for months it was not.**
`govulncheck` was first actually executed against this repository on 2026-08-25,
having been credited as the coverage since 2026-08-21 while absent from every
reachable path. **Nothing here schedules either tool.** The sentence above names
an instrument, not a control that is operating — the same distinction the absent
`cooldown:` block draws. State what runs; do not let a tool's existence read as
coverage.

```bash
go list -m -u all | grep -i retract        # module retractions; surfaces nothing unasked
govulncheck ./...                          # source mode: module + stdlib reachability
govulncheck -mode binary build/bin/geth    # binary mode: what you actually run
```

Exit `3` = vulnerabilities found, `0` = clean, `127` = tool absent. Anything else
is did-not-run, never clean — a missing or stale scanner is silent exactly like a
healthy one.

**Security updates are a separate repository setting that this file cannot turn
on or off, and a zero limit does not suppress them.** Measured against the
GitHub API on 2026-08-21, both were off for this repository. **Re-check that,
do not re-read it:** anyone with admin access can flip either, so the repository
can become outward-facing with nobody having edited `dependabot.yml`.

## Deliberately present — do not tidy

These look like clutter and are not. Do not delete, move, consolidate or
"clean up" any of them:

- **`git.diff`** — a ~6 MB tracked diff artifact at the repository root.
- **Legacy CI files** — `.travis.yml`, `circle.yml`, `appveyor.yml`,
  `Jenkinsfile`, `oss-fuzz.sh`. The live CI is `.github/workflows/`. **"Inherited
  from upstream" is provenance, not justification, and is wrong for two of
  them.** go-ethereum deleted `.travis.yml` (2025-06-26), `circle.yml`
  (2026-01-17) and `appveyor.yml` (2026-05-10) — but `ethereumclassic/core-geth`,
  where releases land, still carries all of them. `Jenkinsfile` appears in **no**
  go-ethereum commit ever; it is ETC's own. `oss-fuzz.sh` is **not legacy** — it
  is in go-ethereum master today and is Google's OSS-Fuzz entry point. Do not
  sweep these as a group.
- **`build/checksums.txt`'s `ppa-builder` pin (Go 1.19.6) is knowingly
  insufficient — do not bump it.** Two Go pins serve different paths.
  `version:golang` feeds `DownloadGo` for `-dlgo`, a **binary** download that
  `release-packages.yml` runs for every ARM/arm64 target — live and
  release-critical, at 1.26.6. `version:ppa-builder` feeds
  `downloadGoBootstrapSources`, the compiler that builds Go **from source** on
  the Launchpad path, reached only from `debsrc` (invoked only by `.travis.yml`
  and `build/bot/ppa-build.sh`, neither live). Go 1.26 requires a Go 1.24
  compiler (go.dev/doc/install/source: 1.N needs 1.M where M = N-2 rounded down
  to even), so 1.19.6 cannot bootstrap it — and could not bootstrap the previous
  1.22.1 pin either, which needed 1.20. The file's own comment anticipates this:
  the remedy is a recursive bootstrap chain, which is upstream's design decision.
  Left as-is on a path nothing here runs.
- **`swarm/` and `integration/`** — near-empty holdover directories.
- **`AUTHORS` and `.mailmap`** — upstream attribution records.
- **`accounts/keystore/`** — this is source code and test vectors, not key
  material, despite the directory name.

### Adjudicated `govulncheck` findings — reported forever, and expected

**Six advisories are reported permanently and have been adjudicated. Do not
re-adjudicate them, and do not try to close them with a version bump — neither
set can be.**

**Binary mode, five `github.com/ethereum/go-ethereum` advisories** — `GO-2026-4314`,
`GO-2026-4315`, `GO-2026-4507`, `GO-2026-4508`, `GO-2026-4511`. Already backported
into this fork, each traceable to a core-geth commit ancestral to the build
revision. They recur **structurally**: the fork keeps go-ethereum's module path, so
its pseudo-version sorts below upstream's fixed tags, and `govulncheck` matches
symbol *names* while a backport adds guards inside the same function. Note the RLP
fix is **not** in `rlp/` — it is `countValuesExceedsLimit` in
`eth/protocols/{eth,snap}/msgvalidate.go`, so a grep scoped to `rlp/` reads as
absence.

**Source mode, `GO-2026-5932`** — `golang.org/x/crypto/openpgp`, `Fixed in: N/A`,
unmaintained by design. Measured both directions: 0 openpgp packages in
`./cmd/geth`, 6 in `./internal/build`. **Not in the shipped binary** — but not dead
either: `.github/workflows/release-packages.yml` → `build/archive-signing.sh` →
`ci.go archive -signer` → `build.PGPSignFile`, so it runs on every signed release.
**Accepted** because the only thing openpgp parses is the signing key from a
CI-secret environment variable, signing this project's own archive — no
attacker-supplied input reaches the parser. Removing it breaks release signing; a
build tag cannot help, since the release path is what needs it. **Re-check if
anything here ever verifies third-party signatures or parses user-supplied keys.**

**The scanner is verified before the set is compared.** Three `govulncheck`
binaries exist on this machine, and an older one reports a **smaller** set — which
would surface as "adjudicated advisories GONE", a confident wrong answer pointing
at this repository instead of at `PATH`. A shadowed or mismatched scanner fails as
did-not-run, never as set-changed. Parse its version anchored on the `Scanner:`
line: `govulncheck -version` prints the **Go toolchain** version first, so a bare
`\d+\.\d+\.\d+` match returns the wrong number.

**The rule for both sets: compare the exact set of advisory IDs; any difference in
either direction is a finding.** A new ID is unadjudicated. A *missing* ID is also a
finding — the adjudication went stale, or the artifact is not the one adjudicated.
**Never suppress by module or category**; that hides the next real advisory behind
the known ones. Since findings are expected, a `govulncheck` exit of `0` is itself a
change, not a pass, and any exit other than `3` is did-not-run, never clean.

## Licensing — do not change

There is **no `LICENSE` file**, and that is correct. The project uses
go-ethereum's split licensing:

- `COPYING` — GPL-3.0, for the binaries under `cmd/`
- `COPYING.LESSER` — LGPL-3.0, for the library (everything outside `cmd/`)

Do not add a `LICENSE` file, do not consolidate the two, and do not act on a
community-standards checker flagging the absence.

## Ask before changing

- **Consensus-critical behavior** — `consensus/`, `core/` block processing,
  `core/vm/` opcode and gas semantics, fork-block tables in `params/`. A wrong
  value here splits the chain.
- **Chain configuration** — `params/config_classic.go`,
  `params/config_mordor.go`.
- **`internal/build/gotool.go` `CGO_CFLAGS`** — breaks the blst/c-kzg build in
  ways that surface only at link time.
- **Dependencies** — `go.mod`, `go.sum`, especially `blst`, `c-kzg`,
  `gnark-crypto`, `btcec`, `graphql-go`.
- **Submodule pins** — `.gitmodules`, `tests/testdata*`. These are consensus
  test vectors.
- **CI and release config** — `.github/workflows/`, `Dockerfile`,
  `Dockerfile.alltools`, `build/ci.go`.
- **`README.md`'s maintenance-mode and migration warnings** — a coordinated
  public statement, not editorial copy.

## Never

- Break wire-protocol compatibility with the ETC network. **EIP-7642 (eth/69) is
  deliberately excluded** because it removes Total Difficulty from the handshake,
  which ETC needs for proof-of-work chain selection. Do not restore it.
- Commit private keys, node keys, keystores, JWT secrets or mnemonics.
- Remove or skip tests to make a build pass.
- Refactor, rename or reformat code unrelated to the task.
- Put machine-specific paths, data directories or node identities in this file,
  `AGENTS.md` or `CLAUDE.md` — all three are public.

## Response style

- No pleasantries. Code first; explain only if asked.
- Concise bullets over paragraphs.
- Do not repeat the prompt back.
