# vercelgate Refactor Plan

## Overview

Two-phase refactor:
1. Remove the unused `switchteam` command and its team infrastructure
2. Replace the SQLite + Ent ORM account store with a plain JSON file

---

## Phase 1: Remove `switchteam`

**Goal:** Delete the team-switching command and all code that only exists to support it. The `switch` command is unaffected.

### 1.1 — Remove `switchTeamCmd` from `main.go`

- Delete the `switchTeamCmd` `cobra.Command` definition (lines 117–130)
- Remove `rootCmd.AddCommand(switchTeamCmd)` (line 42)
- Remove the `"github.com/khanakia/vercelgate/gen/ent/team"` import (only used by `promptGetTeam`)
- Remove the `"github.com/khanakia/vercelgate/gen/ent"` import if it becomes unused after removing `promptGetTeam` and its `*ent.Team` return type (verify — `*ent.User` still used in `promptGetUser`)

### 1.2 — Remove `promptGetTeam()` from `main.go`

- Delete the entire `promptGetTeam(userID string) (*ent.Team, error)` function (lines 171–198)
- This also removes the only direct use of `entdb.Client().Team.Query()`

### 1.3 — Simplify `SwitchCmd()` in `main.go`

- Remove the `switchTeam bool` parameter
- Delete the `if switchTeam { ... } else { ... }` block (lines 149–166)
- Replace it with a single unconditional call to `vercelutil.DeleteCurrentTeam()` — when switching accounts, any persisted `currentTeam` in `config.json` is invalid and should be cleared
- Update the call site in `switchCmd` to `SwitchCmd()` (no argument)

### 1.4 — Remove `SetCurrentTeam()` from `pkg/vercelutil/vercelutil.go`

- Delete `SetCurrentTeam(teamID string) error` (lines 50–77) — no longer called by anything
- Keep `DeleteCurrentTeam()` — still called from the simplified `SwitchCmd()`

### 1.5 — Verify

- `go build ./...` passes with no errors
- `vercelgate switch` still prompts for an account and writes the token
- `vercelgate switchteam` no longer exists (`--help` should not list it)

---

## Phase 2: Replace SQLite with a JSON account store

**Goal:** Replace the Ent ORM + SQLite database with a single JSON file at `~/.config/com.vercel.cli/vercelgate_accounts.json`. Each entry stores the account display name and the raw `auth.json` bytes.

**Target JSON shape:**
```json
[
  {
    "name": "johndoe (john@example.com)",
    "data": { "token": "..." }
  }
]
```

`data` is stored as raw JSON — its schema is intentionally opaque so future Vercel auth.json fields are preserved without any code change.

---

### 2.1 — Create `pkg/accountstore/accountstore.go`

New package responsible for all read/write operations on the accounts JSON file.

**Types:**
```go
type Account struct {
    Name string          `json:"name"`
    Data json.RawMessage `json:"data"`
}
```

**Functions to implement:**

| Function | Signature | Description |
|----------|-----------|-------------|
| `FilePath` | `() (string, error)` | Returns `~/.config/com.vercel.cli/vercelgate_accounts.json` using `vercelutil.GetGlobalPathConfig()` |
| `Load` | `() ([]Account, error)` | Reads and unmarshals the file; returns `[]` if file does not exist |
| `Save` | `(accounts []Account) error` | Marshals and writes the file with `json.MarshalIndent` |
| `AddOrUpdate` | `(name string, data json.RawMessage) error` | Loads, upserts by `name`, saves |
| `All` | `() ([]Account, error)` | Alias for `Load` |
| `Clear` | `() error` | Writes an empty `[]` to the file |

No singleton, no global state — functions are stateless and operate on the file directly.

---

### 2.2 — Update `pkg/vercelfn/vercelfn.go` (sync logic)

Replace the Ent upsert calls with an `accountstore.AddOrUpdate` call.

**New flow for `SyncAuthJson()`:**

1. Find and read `auth.json` — **read raw bytes** (`os.ReadFile`), not just the parsed token
2. Parse the token from raw bytes to pass to the Vercel API (still need `vercelutil.ParseAuthFile`)
3. Call `vercelapi.GetUser(token)` to get `username`, `name`, `email` for building the display name
4. Build account name: `fmt.Sprintf("%s (%s)", usernameOrName, email)` — same logic as the current switch prompt
5. Call `accountstore.AddOrUpdate(name, rawAuthJsonBytes)`
6. Remove all Ent client calls and the `vercelapi.GetTeams` call (teams are no longer stored)

**Imports to remove:** `entdb`, `context` (if unused after Ent removal)

---

### 2.3 — Update `SwitchCmd()` in `main.go`

Replace the `promptGetUser()` function (which queries Ent) with one backed by the account store.

**New `promptGetUser()` implementation:**

1. Call `accountstore.All()` to get `[]Account`
2. Build prompt items from `account.Name` (already formatted — no re-formatting needed)
3. Show `promptui.Select`
4. Return the selected `Account`

**Update `SwitchCmd()`:**

- Instead of calling `vercelutil.SetAuthToken(user.Token)`, write `account.Data` directly to `auth.json` — this restores the full original auth.json content, not just the token field
- New helper in `vercelutil`: `RestoreAuthJson(data []byte) error` — overwrites `auth.json` with the given bytes via `os.WriteFile`
- The confirmation print changes to `fmt.Printf("Switched to %s\n", account.Name)` since `account.Name` is already the display string

**Remove:** the old `promptGetUser() (*ent.User, error)` function

---

### 2.4 — Update `initCmd` in `main.go`

Currently `initCmd` checks for the DB file and runs schema migration.

**New behavior:**

- Check if the accounts JSON file exists; if yes, print "already initialized"
- If not, call `accountstore.Clear()` (creates an empty `[]` file) and print success
- Remove all `entcfn.Migrate()` calls and DB path checks

---

### 2.5 — Update `ResetCmd()` in `main.go`

Currently deletes all rows from `users` and `teams` tables.

**New behavior:**

- Call `accountstore.Clear()` to overwrite the accounts file with `[]`
- Remove Ent client calls

---

### 2.6 — Remove `SetAuthToken()` from `pkg/vercelutil/vercelutil.go`

After 2.3, `SetAuthToken` is no longer called (switch now writes the full `account.Data`).

- Delete `SetAuthToken(token string) error`
- Add `RestoreAuthJson(data []byte) error` — simple `os.WriteFile` to `auth.json` path
- Remove the `jsonupdate` import from `vercelutil.go` if `DeleteCurrentTeam` no longer uses it (check — `DeleteCurrentTeam` still uses `jsonupdate.NewJsonUpdate`)

---

### 2.7 — Remove SQLite infrastructure

Once all call sites are updated, delete all SQLite/Ent code:

| What | Action |
|------|--------|
| `pkg/entdb/` | Delete entire package |
| `pkg/entcfn/` | Delete entire package |
| `schema/user.go` | Delete |
| `schema/team.go` | Delete |
| `gen/` | Delete entire generated directory |
| `main.go` import `_ "github.com/mattn/go-sqlite3"` | Remove |
| `main.go` imports `entdb`, `entcfn`, `ent` | Remove |
| `constants.DB_FILE_NAME` | Remove from `pkg/constants/constants.go`; delete file if it becomes empty |

Run `go mod tidy` to remove orphaned dependencies from `go.mod` and `go.sum` (`entgo.io/ent`, `mattn/go-sqlite3`, and any transitive deps).

---

### 2.8 — Verify

- `go build ./...` passes
- `vercelgate init` creates the accounts JSON file
- `vercelgate sync` adds/updates an entry in the accounts JSON file (check file contents manually)
- `vercelgate switch` prompts with account names and restores `auth.json` on selection
- `vercelgate reset` empties the accounts JSON file
- `vercelgate path` still works (unchanged)
- No references to `.db` files or Ent remain in the source tree (`grep -r "sqlite\|ent\.Client\|entdb" .`)

---

## File Change Summary

| File | Phase 1 | Phase 2 |
|------|---------|---------|
| `main.go` | Remove switchTeamCmd, promptGetTeam, simplify SwitchCmd | Replace promptGetUser, update initCmd/ResetCmd, remove Ent imports |
| `pkg/vercelutil/vercelutil.go` | Remove SetCurrentTeam | Remove SetAuthToken, add RestoreAuthJson |
| `pkg/vercelfn/vercelfn.go` | — | Replace Ent upserts with accountstore.AddOrUpdate, remove GetTeams |
| `pkg/accountstore/accountstore.go` | — | **New file** |
| `pkg/entdb/` | — | Delete |
| `pkg/entcfn/` | — | Delete |
| `pkg/constants/constants.go` | — | Remove DB_FILE_NAME |
| `schema/` | — | Delete |
| `gen/` | — | Delete |
| `go.mod` / `go.sum` | — | `go mod tidy` |
