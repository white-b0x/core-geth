@AGENTS.md

## Claude-specific notes

`AGENTS.md` above is the tool-agnostic project context and is the authority for
build commands, structure, branching and boundaries. Everything below is
Claude-only and deliberately short.

- **Do not build or run the full test suite casually.** `make all`, `make test`
  and `make test-coregeth` are heavy — a full geth build plus consensus suites.
  Ask before starting one, and never run two at once.
- **`go test` on a narrow package path is the cheap default.** Prefer the
  `-run` filtered invocations in `AGENTS.md` over whole-tree runs.
- Machine-local facts — data directories, node identities, local run scripts —
  belong in `CLAUDE.local.md`, which `.gitignore` holds back. Do not move them
  into this file or `AGENTS.md`; both are public.
