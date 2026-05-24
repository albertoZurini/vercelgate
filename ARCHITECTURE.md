# Architecture

`vercelgate` is a Go CLI tool that manages multiple Vercel account credentials locally, allowing users to switch between accounts and teams without re-running `vercel login`/`vercel logout`.

---

## Overview

The tool works by reading and writing the same JSON files that the Vercel CLI uses for authentication (`auth.json`, `config.json`). Accounts and their associated teams are stored locally in an SQLite database. When a user switches accounts, the tool overwrites the token in `auth.json`; subsequent `vercel` commands then use that token transparently.

---

## Directory Layout

```
vercelgate/
├── main.go                 # CLI entry point — all cobra commands defined here
├── schema/                 # Ent ORM schema definitions (source of truth)
│   ├── user.go
│   └── team.go
├── gen/ent/                # Auto-generated Ent ORM code (do not edit by hand)
├── cmd/
│   ├── dbmigrate/          # Standalone DB migration binary
│   └── entg/               # Standalone Ent code-generation binary
└── pkg/
    ├── constants/          # Shared constants (DB filename)
    ├── entcfn/             # Schema migration helper
    ├── entdb/              # Singleton DB client
    ├── jsonupdate/         # Thin wrapper around tidwall/sjson for JSON patching
    ├── logger/             # Debug/verbose logging + token masking
    ├── utils/              # File I/O helpers
    ├── vercelapi/          # HTTP client for Vercel REST API
    ├── vercelfn/           # Sync orchestration (API → DB)
    └── vercelutil/         # Vercel config file path resolution + read/write
```

---

## Data Model

Two entities with a one-to-many relationship (User → Teams):

```
User
  id        string  (Vercel user ID, primary key)
  name      string?
  username  string?
  email     string?
  token     string? (Vercel auth token — stored in plaintext)

Team
  id        string  (Vercel team ID, primary key)
  name      string?
  slug      string?
  user_id   string? (foreign key → User.id)
```

Both are defined in `schema/` and the generated CRUD code lives in `gen/ent/`.

---

## Storage

| Store | Path | Purpose |
|-------|------|---------|
| SQLite DB | `<vercel-config-dir>/vercelgate.db` | Persisted user + team records |
| `auth.json` | `<vercel-config-dir>/auth.json` | Active Vercel CLI token (Vercel-owned) |
| `config.json` | `<vercel-config-dir>/config.json` | Active team selection (Vercel-owned) |

The Vercel config directory is resolved using the XDG spec (`pkg/vercelutil/vercelutil.go:GetGlobalPathConfig`). On macOS this is typically `~/.config/com.vercel.cli/`; on Linux it may be `~/.local/share/com.vercel.cli/`.

---

## CLI Commands

| Command | Description |
|---------|-------------|
| `init` | Creates the SQLite schema; safe to run once (no-ops if DB exists) |
| `sync` | Reads active token → calls Vercel API → upserts User + Teams into DB |
| `new` | Deletes `auth.json` so `vercel login` can create a fresh session |
| `switch` | Prompts user selection → writes token to `auth.json`; clears `currentTeam` |
| `switchteam` | Same as `switch` + prompts team selection → writes team ID to `config.json` |
| `reset` | Deletes all User and Team rows from the DB |
| `path` | Prints the resolved Vercel config directory |

Global flags `--debug` and `--verbose` enable layered logging via `pkg/logger`.

---

## Key Flows

### `sync`
```
auth.json ──read──► vercelutil.ParseAuthFile
                          │ token
                          ▼
                   vercelapi.GetUser(token) ──► Vercel API /v2/user
                          │ User
                          ▼
                   entdb.Client().User.Upsert(...)
                          │
                   vercelapi.GetTeams(token) ──► Vercel API /v2/teams
                          │ []Team
                          ▼
                   entdb.Client().Team.Upsert(...) (per team)
```

### `switch`
```
entdb.Client().User.Query().All() ──► promptui.Select
        │ selected User
        ▼
vercelutil.SetAuthToken(user.Token) → overwrites auth.json
        │
        └─ (if switchTeam)
           entdb.Client().Team.Query(userID).All() ──► promptui.Select
                   │ selected Team
                   ▼
           vercelutil.SetCurrentTeam(team.ID) → overwrites config.json
```

---

## External Dependencies

| Package | Version | Role |
|---------|---------|------|
| `entgo.io/ent` | 0.14.1 | ORM + code generation for SQLite |
| `github.com/spf13/cobra` | 1.7.0 | CLI framework |
| `github.com/manifoldco/promptui` | 0.9.0 | Interactive terminal prompts |
| `github.com/mattn/go-sqlite3` | 1.14.24 | CGO SQLite driver |
| `github.com/adrg/xdg` | 0.5.3 | XDG Base Directory spec |
| `github.com/tidwall/sjson` | 1.2.5 | In-place JSON field patching |

---

## Build & Release

- `Taskfile.yml` defines dev tasks: `entg` (regenerate ORM), `migrate` (run migration), `tag`, `release`.
- `.goreleaser.yaml` builds cross-platform binaries (macOS amd64/arm64, Linux amd64/arm64) and publishes a Homebrew tap at `khanakia/vercelgate`.
- The binary embeds a `version` variable injected by goreleaser at build time.
