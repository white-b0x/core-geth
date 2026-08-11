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

**Inferred from history, not stated by the maintainer — confirm before relying
on it.** The repository carries ~29 topic branches using conventional prefixes
(`security/`, `test/`, `docs/`, `chore/`, `fix/`, `rlp/`) with kebab-case
descriptions, and `main` is the default branch that release tags and upstream
pull requests are cut from.

Work on a topic branch; keep `main` releasable. Pushing is a separate decision
from committing and should be confirmed explicitly — this repository is public
and its pull requests go to a consensus-critical upstream.

**Known merge collision on `AGENTS.md`.** `ethereum/go-ethereum` ships its own
root `AGENTS.md` (commit `406a852ec`, 2026-02-25), already present in this clone
via the `geth` remote, so a merge from `geth/master` will conflict on it.
**This fork's `AGENTS.md` wins** — upstream's describes go-ethereum and names
commands this fork does not have, such as `go run ./build/ci.go check_generate`.
`ethereumclassic/core-geth`, the actual PR target, has no `AGENTS.md`.

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

## Deliberately present — do not tidy

These look like clutter and are not. Do not delete, move, consolidate or
"clean up" any of them:

- **`git.diff`** — a ~6 MB tracked diff artifact at the repository root.
- **Legacy CI files** — `.travis.yml`, `circle.yml`, `appveyor.yml`,
  `Jenkinsfile`, `oss-fuzz.sh`. Inherited from upstream; the live CI is
  `.github/workflows/`.
- **`swarm/` and `integration/`** — near-empty holdover directories.
- **`AUTHORS` and `.mailmap`** — upstream attribution records.
- **`accounts/keystore/`** — this is source code and test vectors, not key
  material, despite the directory name.

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
