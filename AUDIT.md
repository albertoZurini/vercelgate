# Security & Code Audit

Audit date: 2026-05-24  
Audited ref: `main` (3909183)

Severity labels: **Critical** · **High** · **Medium** · **Low** · **Info**

---

## Findings

### 1. Nil pointer dereference in file helpers — **High**

**Files:** `pkg/utils/utils.go:12`, `pkg/utils/utils.go:29`

Both `OpenFile` and `IsFileExists` call `os.Stat` and then unconditionally call `info.IsDir()`:

```go
info, err := os.Stat(filepath)
if errors.Is(err, os.ErrNotExist) || info.IsDir() {
```

If `os.Stat` returns any error *other* than `ErrNotExist` (e.g. a permissions error), `info` is `nil` and `info.IsDir()` panics. The fix is to check `err != nil` first:

```go
if err != nil {
    return nil, err
}
if info.IsDir() {
    return nil, fmt.Errorf("%s is a directory", filepath)
}
```

---

### 2. Auth file written world-readable — **High**

**File:** `pkg/vercelutil/vercelutil.go:41`

```go
err = os.WriteFile(filePath, []byte(jsonupd.String()), 0644)
```

`auth.json` contains the Vercel auth token and is written with permission `0644`, making it readable by every user on the system. Token files should use `0600`:

```go
err = os.WriteFile(filePath, []byte(jsonupd.String()), 0600)
```

---

### 3. Auth token stored in plaintext SQLite — **Medium**

**File:** `schema/user.go:21`

```go
field.String("token").Optional(),
```

The Vercel auth token is persisted unencrypted in `vercelgate.db`. Anyone who can read that file (same user, shared machine, backup leaks) obtains valid credentials. The database should either encrypt the token field or use the OS keychain (e.g. via `github.com/zalando/go-keyring`).

---

### 4. `DeleteCurrentTeam` error silently dropped — **Medium**

**File:** `main.go:165`

```go
vercelutil.DeleteCurrentTeam()
```

The return value is ignored. If removing `currentTeam` from `config.json` fails, the user silently ends up with a mismatched state: the token was already switched to the new account, but the old team ID is still active in `config.json`. The Vercel CLI will then make API calls to a team that may not belong to the new account.

```go
if err := vercelutil.DeleteCurrentTeam(); err != nil {
    return err
}
```

---

### 5. No HTTP timeout on Vercel API calls — **Medium**

**File:** `pkg/vercelapi/vercelapi.go:24`, `:72`

```go
client := &http.Client{}
```

Both `GetUser` and `GetTeams` create HTTP clients with no timeout. A slow or unresponsive Vercel API will hang `vercelgate sync` indefinitely. Add a timeout:

```go
client := &http.Client{Timeout: 15 * time.Second}
```

---

### 6. `GetTeams` does not paginate — **Medium**

**File:** `pkg/vercelapi/vercelapi.go:62-108`

The Vercel `/v2/teams` API returns a `pagination` object with a `next` cursor. The code only fetches the first page and discards the cursor. Users who belong to more than one page of teams (rare, but possible for large organisations) will have their later teams silently missing.

---

### 7. Silently wrong DB path when config dir is missing — **Medium**

**File:** `pkg/entdb/entdb.go:33`

```go
globalPath, _ := vercelutil.GetGlobalPathConfig()
path := globalPath + "/vercelgate.db"
```

If `GetGlobalPathConfig()` returns an error (e.g. Vercel CLI was never installed), `globalPath` is an empty string and the resulting path is `/vercelgate.db` — a file at the filesystem root. This will either fail silently or create a DB in an unexpected location. The error should be surfaced:

```go
globalPath, err := vercelutil.GetGlobalPathConfig()
if err != nil {
    log.Fatalf("vercel config directory not found: %v", err)
}
```

---

### 8. `Migrate()` calls `log.Fatalf` before `return err` — **Low**

**File:** `pkg/entcfn/entcfn.go:21-22`

```go
log.Fatalf("failed creating schema resources: %v", err)
return err  // unreachable
```

`log.Fatalf` calls `os.Exit(1)`, so `return err` is dead code. The function signature promises to return an error to the caller, but it never does — it terminates the process. This prevents any caller from handling the error gracefully. Replace with `return fmt.Errorf(...)` and let the caller (in `main.go`) call `log.Fatal`.

---

### 9. Duplicate error output — **Low**

**Files:** `pkg/vercelutil/vercelutil.go:33,44`, `pkg/vercelapi/vercelapi.go:28,35`

Several functions print the error with `fmt.Println(err.Error())` *and* return it. The caller then also prints it (via `log.Fatal`), resulting in the same message appearing twice. Remove the intermediate `fmt.Println` calls and let only the top-level handler display the error.

---

### 10. `Migrate()` closes singleton DB client — **Low**

**File:** `pkg/entcfn/entcfn.go:25`

```go
client.Close()
```

`entcfn.Migrate()` obtains the singleton client from `entdb.Client()`, then closes it. The `entdb` package variable is still non-nil after this, so subsequent calls to `entdb.Client()` return a closed handle. In practice the `init` command exits immediately after `Migrate()`, so this does not cause a runtime failure today, but it is a latent bug that would surface if any code were added after the migration step.

---

### 11. `SetCurrentTeam` and `DeleteCurrentTeam` write inconsistent JSON formatting — **Low**

**Files:** `pkg/vercelutil/vercelutil.go:70`, `:99`

`SetCurrentTeam` writes `jsonupd.Pretty()` (indented), while `DeleteCurrentTeam` writes `jsonupd.String()` (minified). This means the formatting of `config.json` changes depending on which operation ran last. Both should use the same format; `Pretty()` is preferred since Vercel CLI writes indented JSON.

---

### 12. Typo in `jsonupdate.Deleete` — **Info**

**Files:** `pkg/jsonupdate/jsonupdate.go:31`, `pkg/vercelutil/vercelutil.go:97`

The method is named `Deleete` (double `e`) in both the definition and the call site, so the code compiles and runs correctly. However, it is a public method with a misspelled name that would break any future caller relying on the expected name `Delete`. Rename to `Delete` in both files.

---

### 13. Typo in root command description — **Info**

**File:** `main.go:55`

```go
Long: `You can swithc between multiple accounts...`,
```

`swithc` → `switch`

---

### 14. `switch` command short description is ungrammatical — **Info**

**File:** `main.go:104`

```go
Short: "Switch between account",
```

Should be "Switch between accounts".

---

### 15. Module path does not match repository owner — **Info**

**File:** `go.mod:1`

```
module github.com/khanakia/vercelgate
```

The git remote and commit author are `albertoZurini`, not `khanakia`. This mismatch suggests the project was cloned/forked but the module path was not updated. While this works locally, it means `go get github.com/albertoZurini/vercelgate` would not resolve correctly, and import paths in the source embed the original author's namespace.

---

## Summary Table

| # | Severity | File | Issue |
|---|----------|------|-------|
| 1 | **High** | `pkg/utils/utils.go:12,29` | Nil dereference on non-ErrNotExist stat error |
| 2 | **High** | `pkg/vercelutil/vercelutil.go:41` | Auth file written with world-readable permissions (0644) |
| 3 | **Medium** | `schema/user.go:21` | Auth token stored in plaintext SQLite |
| 4 | **Medium** | `main.go:165` | Unchecked error from `DeleteCurrentTeam` |
| 5 | **Medium** | `pkg/vercelapi/vercelapi.go:24,72` | No HTTP timeout — sync can hang indefinitely |
| 6 | **Medium** | `pkg/vercelapi/vercelapi.go:62` | `GetTeams` fetches first page only; no pagination |
| 7 | **Medium** | `pkg/entdb/entdb.go:33` | Silently wrong DB path when Vercel config dir missing |
| 8 | **Low** | `pkg/entcfn/entcfn.go:21` | `log.Fatalf` makes `return err` unreachable |
| 9 | **Low** | `vercelutil`, `vercelapi` | Errors printed twice (in function and at call site) |
| 10 | **Low** | `pkg/entcfn/entcfn.go:25` | `Migrate()` closes singleton DB client |
| 11 | **Low** | `pkg/vercelutil/vercelutil.go:70,99` | Inconsistent JSON formatting (Pretty vs String) |
| 12 | **Info** | `pkg/jsonupdate/jsonupdate.go:31` | Misspelled public method name `Deleete` |
| 13 | **Info** | `main.go:55` | Typo: `swithc` |
| 14 | **Info** | `main.go:104` | Grammar: "between account" → "between accounts" |
| 15 | **Info** | `go.mod:1` | Module path references original fork owner |
