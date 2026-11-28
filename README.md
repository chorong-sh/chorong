# Chorong (초롱)

**Composable CLI tools for a human-in-the-loop coding agent factory.**

*Chorong* is a traditional Korean silk lantern—it lights up when your automated
line needs a human eye.

Most AI coding tools ask you to buy into a monolithic framework. Chorong takes
the opposite approach: small, fast, standalone binaries that work independently
in any repository and snap together over narrow contracts (CLI, JSON, Git,
`gh`) to run a full human-in-the-loop software factory.

- **Use one tool or the whole stack.** Start with `decide` as an approval inbox,
  `docsync` as a doc-freshness check, or `gh-throttle` to protect your GitHub
  rate limit—no companion tools required.
- **Human in the loop by design.** Autonomous agents queue work, message each
  other, and open pull requests, while explicit gates (`decide`, `factory gate`,
  `agentbus lease`) keep you in control.
- **Harness- and language-agnostic.** Works alongside Claude Code, Codex,
  Antigravity, scripts, or cron jobs—and with any codebase.

## How the Pieces Fit Together

```text
  Jira / GitHub Issues
          │
          ▼
       factory ──────► picks ready work, checks gates, dispatches agents
          │
          ▼
     agent panes  ◄──► agentbus   (agent-to-agent messages & checkout leases)
          │       ◄──► decide     (human-in-the-loop questions & approvals)
          │       ◄──► orgctl     (declarative team spec vs live tabs/jobs)
          ▼
    Pull Requests ◄──► prsync     (tracks open PRs alongside agent tabs)
                  ◄──► gh-nudge   (gh-approve, gh-merge, gh-throttle, gh-ask)
                  ◄──► docsync / genfiles  (repo guardrails)
```

## Quick Start

### Homebrew (macOS & Linux)

Install individual tools:

```sh
brew install chorong-sh/tap/decide
brew install chorong-sh/tap/factory
```

Or install the full suite:

```sh
brew install chorong-sh/tap/chorong
```

### Prebuilt Binaries

Download macOS and Linux binaries directly from
[GitHub Releases](https://github.com/chorong-sh/chorong/releases).

### From Source (Go 1.26+, no Bazel required)

```sh
go install github.com/chorong-sh/chorong/decide/cmd/decide@latest
go install github.com/chorong-sh/chorong/factory/cmd/factory@latest
```

## Tools

### Factory & Human-in-the-Loop Coordination

| Tool | Purpose | Docs |
| :--- | :--- | :--- |
| `factory` | Mirror Jira/GitHub issues, rank ready work, enforce start gates, dispatch & settle coding agents | [factory/README.md](factory/README.md) |
| `decide` | Human-in-the-loop inbox (TUI, Web/PWA, CLI, optional Emacs UI) for agent questions and approvals | [decide/README.md](decide/README.md) |
| `agentbus` | Local message bus (`agentbus` + `agentbusd`) and checkout leases between coding agents | [agentbus/README.md](agentbus/README.md) |
| `prsync` | Match your open pull requests to live agent tabs | [prsync/README.md](prsync/README.md) |
| `orgctl` | Diff a declarative agent-team spec against live tabs, names, and scheduled jobs | [orgctl/README.md](orgctl/README.md) |

### GitHub Workflow

| Tool | Purpose | Docs |
| :--- | :--- | :--- |
| `gh-nudge` | PR workflow suite: `gh-nudge`, `gh-approve`, `gh-merge`, `gh-throttle`, `gh-codeowners`, `gh-slack`, `gh-storage` | [gh-nudge/README.md](gh-nudge/README.md) |
| `gh-ask` | Hand a question about a pull-request commit range to any command | [gh-ask/README.md](gh-ask/README.md) |

### Repo Guardrails

| Tool | Purpose | Docs |
| :--- | :--- | :--- |
| `docsync` | List documentation files likely made stale by a code change | [docsync/README.md](docsync/README.md) |
| `genfiles` | Verify generated files and their regeneration commands stay in sync | [genfiles/README.md](genfiles/README.md) |

## Development

End users never need Bazel. Contributors can run tests with standard Go
(`go test ./...`) or use the hermetic Bazel + Make workflow:

```sh
make check    # tests, static analysis, formatting, secret & shellcheck scans
make format   # format all files
make install  # install all binaries locally
```

See [AGENTS.md](AGENTS.md) for architecture boundaries and contribution rules.

## License

Apache-2.0 — see [LICENSE](LICENSE).
