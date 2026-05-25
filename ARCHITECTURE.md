# Architecture

`vcx` is a Go CLI tool that manages multiple Vercel account credentials locally, allowing users to switch between accounts without re-running `vercel login`/`vercel logout`.

---

## Overview

The tool works by reading and writing the same JSON files that the Vercel CLI uses for authentication (`auth.json`, `config.json`). Accounts are stored in a single JSON file (`vcx_accounts.json`) as a list of `{name, data}` entries, where `data` is the raw content of `auth.json` at the time the account was synced. When a user switches accounts, the stored `data` is written back verbatim to `auth.json`; subsequent `vercel` commands then use that token transparently.

---

## Directory Layout

```
vcx/
├── main.go                 # CLI entry point — all cobra commands defined here
└── pkg/
    ├── accountstore/       # Read/write vcx_accounts.json
    ├── jsonupdate/         # Thin wrapper around tidwall/sjson for JSON patching
    ├── logger/             # Debug/verbose logging + token masking
    ├── utils/              # File I/O helpers
    ├── vercelapi/          # HTTP client for Vercel REST API
    ├── vercelfn/           # Sync orchestration (API → account store)
    └── vercelutil/         # Vercel config file path resolution + read/write
```

---

## Data Model

Accounts are stored as a JSON array in `vcx_accounts.json`:

```json
[
  {
    "name": "johndoe (john@example.com)",
    "data": { "token": "..." }
  }
]
```

- `name` — display key, formatted as `"{Name or Username} ({Email})"` by `vercelfn.SyncAuthJson`. Used to identify accounts in the switch prompt.
- `data` — raw bytes of `auth.json` at sync time. Schema is intentionally opaque: the full file is stored so future Vercel auth fields are preserved without code changes.

Defined in `pkg/accountstore/accountstore.go` as:

```go
type Account struct {
    Name string          `json:"name"`
    Data json.RawMessage `json:"data"`
}
```

---

## Storage

| Store | Path | Permissions | Purpose |
|-------|------|-------------|---------|
| `vcx_accounts.json` | `<vercel-config-dir>/vcx_accounts.json` | `0600` | All saved accounts |
| `auth.json` | `<vercel-config-dir>/auth.json` | `0600` | Active Vercel CLI token (Vercel-owned) |
| `config.json` | `<vercel-config-dir>/config.json` | `0644` | Active team selection (Vercel-owned) |

The Vercel config directory is resolved via the XDG spec (`pkg/vercelutil/vercelutil.go:GetGlobalPathConfig`). On macOS this is typically `~/.config/com.vercel.cli/`.

---

## CLI Commands

| Command | Description |
|---------|-------------|
| `init` | Creates an empty `vcx_accounts.json`; no-ops if the file already exists |
| `sync` | Reads active `auth.json` → calls Vercel API → upserts account into the JSON store |
| `new` | Deletes `auth.json` so `vercel login` can create a fresh session |
| `switch` | Prompts account selection → restores `auth.json` from stored data; clears `currentTeam` |
| `reset` | Overwrites `vcx_accounts.json` with an empty array |
| `path` | Prints the resolved Vercel config directory |

Global flags `--debug` and `--verbose` enable layered logging via `pkg/logger`.

---

## Key Flows

### `sync`

```
auth.json ──os.ReadFile──► rawBytes
                │
                ├── json.Unmarshal ──► authConfig.Token
                │                           │
                │                    vercelapi.GetUser(token) ──► Vercel API /v2/user
                │                           │ user.Name/Username/Email
                │                           ▼
                │                    name = "{Name or Username} ({Email})"
                │
                └──────────────────► accountstore.AddOrUpdate(name, rawBytes)
                                             │
                                     vcx_accounts.json
```

### `switch`

```
accountstore.All() ──► promptui.Select
        │ selected Account
        ▼
vercelutil.RestoreAuthJson(account.Data) → overwrites auth.json verbatim
        │
vercelutil.DeleteCurrentTeam() → removes currentTeam from config.json
```

---

## Package Responsibilities

| Package | Responsibility |
|---------|---------------|
| `accountstore` | Load/save `vcx_accounts.json`; `AddOrUpdate`, `All`, `Clear` |
| `vercelapi` | `GetUser` — single HTTP call to `/v2/user`; 15 s timeout |
| `vercelfn` | Orchestrates sync: read `auth.json` → call API → store account |
| `vercelutil` | Resolve config dir (XDG); `RestoreAuthJson`; `DeleteCurrentTeam`; `ParseAuthFile` |
| `jsonupdate` | Surgical JSON field update/delete via `tidwall/sjson` |
| `utils` | `OpenFile` — stat-guarded file read |
| `logger` | Conditional `[DEBUG]`/`[VERBOSE]` output; `MaskToken` for safe logging |

---

## External Dependencies

| Package | Version | Role |
|---------|---------|------|
| `github.com/spf13/cobra` | 1.7.0 | CLI framework |
| `github.com/manifoldco/promptui` | 0.9.0 | Interactive terminal prompts |
| `github.com/adrg/xdg` | 0.5.3 | XDG Base Directory spec |
| `github.com/tidwall/sjson` | 1.2.5 | In-place JSON field patching |

No CGO dependencies — the binary is pure Go and can be cross-compiled without a C toolchain.

---

## Build & Release

- `task test` — runs `go test ./...`
- `task release` — runs `goreleaser release --clean --skip=validate`
- `.goreleaser.yaml` builds cross-platform binaries (macOS, Linux) and publishes a Homebrew tap.
- The binary embeds a `version` variable injected by goreleaser at build time (`-X main.version={{.Version}}`).
