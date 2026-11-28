# Agent Instructions

These instructions apply to every coding agent working in `chorong-sh/chorong`.

## Core Principles

1. **Simplicity:** Prioritize the simplest solution.
2. **No Duplication:** Reuse existing functionality (`pkg/...`).
3. **TDD:** Practice Test-Driven Development. Understand desired behavior -> Test -> Implement.

## Architecture & Boundary Rules

### 1. Keep the Core Agnostic

- Core domain logic (state, storage, schema, protocol) must remain agnostic to
  whatever connects to it.
- Do not hardcode `herdr`, terminal pane layouts (`HERDR_PANE_ID`), a specific
  LLM harness, or a sibling tool into core packages.
- Every tool must remain usable standalone without requiring companion tools to
  be installed or running.

### 2. Isolate Peer Integrations at the Boundary

- Put peer integrations (`herdr`, `agentbus`, `tmux`, chat bridges, etc.) behind
  a narrow core boundary—such as a `Transport` interface or an executable
  JSON-over-stdio plugin—never woven into core logic.
- Keep convenience profiles, presets, and config defaults for common pairings at
  the configuration/adapter edge only. They must never become core assumptions
  or mandatory dependencies.

### 3. Treat Non-LLM and Traditional CLI Tools as First-Class

- Coordination tools like `decide` and `agentbus` are general-purpose
  primitives, not LLM-only tools.
- Traditional CLI tools, daemons, and scripts must be able to invoke them and
  consume structured input/answers without an LLM in the loop.

### 4. Keep Interfaces Conservative

- Prefer small, general boundaries (transport/adapter interfaces, structured
  JSON or environment contracts) over adding special-case subcommands, flags, or
  branches for individual callers or missing backends.
- When a backend is absent or not configured, degrade cleanly through the same
  interface (e.g., a pull/no-op transport) instead of scattering caller-specific
  checks across core logic.

## Automation & Workflow

- **ALWAYS run `make check`** before requesting a commit.
  - This runs Bazel tests, `nogo` static analysis, formatters, `shellcheck`, and `gitleaks`.
  - If `make check` or `make format` changes files, include them in the commit.
- **Use `make` targets:** `make test`, `make lint`, `make fix`, `make format`, `make install`.
- **Keep standard Go working:** End users never need Bazel. Keep `go.mod` tidy and ensure `GOPACKAGESDRIVER= go test ./...` and `make install` work cleanly.
- **Documentation:** All docs use GitHub Flavored Markdown (`*.md`).

## Git Workflow

- **Keep `main` history flat** (squash-merge PRs, no merge commits).
- **Use rebase** to integrate upstream changes (`git pull --rebase`).
- **Reuse fixed worktree paths** (e.g., `../chorong-wt-1`, `../chorong-wt-2`) so Bazel reuses output bases.

## Versioning & Releases

- **Single module, single tag:** One root `go.mod` (`github.com/chorong-sh/chorong`)
  and one monorepo-wide SemVer tag (`vMAJOR.MINOR.PATCH`).
- **Binary version reporting:** Go binaries report their version via
  `github.com/chorong-sh/chorong/pkg/version`
  (`-ldflags "-X github.com/chorong-sh/chorong/pkg/version.Version=..."` with
  `runtime/debug.ReadBuildInfo()` fallback).
- **Release workflow:** `release-please` maintains a Release PR on `main` from
  Conventional Commits. Squash-merging the Release PR cuts `vX.Y.Z`, builds
  cross-platform binaries via GoReleaser, and updates `chorong-sh/homebrew-tap`.

## Pull Requests

This repository squash-merges with the pull request title and description as
the commit message. Do not leave `<!-- ... -->` HTML comments in PR bodies.

## Temp File Handling

For multi-step workflows, use the gitignored project-local `.tmp/` directory
instead of `$TMPDIR` when a temp file must be referenced across commands.
