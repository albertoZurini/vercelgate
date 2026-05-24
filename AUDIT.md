# Security & Code Audit

Audit date: 2026-05-24  
Audited ref: `main` (c604d22)

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

### 2. No HTTP timeout on Vercel API calls — **Medium**

**File:** `pkg/vercelapi/vercelapi.go:24`

`GetUser` creates an HTTP client with no timeout:

```go
client := &http.Client{}
```

A slow or unresponsive Vercel API will hang `vercelgate sync` indefinitely. Add a timeout:

```go
client := &http.Client{Timeout: 15 * time.Second}
```

---

### 3. Duplicate error output — **Medium**

**Files:** `pkg/vercelutil/vercelutil.go:46,56`, `pkg/vercelapi/vercelapi.go:28,35`

Several functions print the error with `fmt.Println(err.Error())` *and* return it. The caller then also prints it (via `log.Fatal`), resulting in the same message appearing twice. Remove the intermediate `fmt.Println` calls and let only the top-level handler display the error.

In `vercelutil.DeleteCurrentTeam`:
```go
fileBytes, err := utils.OpenFile(filePath)
if err != nil {
    fmt.Println(err.Error()) // remove this line
    return err
}
```

In `vercelapi.GetUser` and `GetTeams`:
```go
if err != nil {
    fmt.Println(err) // remove this line
    return nil, err
}
```

---

### 4. `SyncAuthJson` reads `auth.json` twice — **Low**

**File:** `pkg/vercelfn/vercelfn.go:23,28`

The function calls `os.ReadFile(authJsonFile)` directly to capture the raw bytes, then calls `vercelutil.ParseAuthFile(authJsonFile)` which also calls `os.ReadFile` internally. The file is read twice for no reason:

```go
rawBytes, err := os.ReadFile(authJsonFile)   // read #1
...
authConfig, err := vercelutil.ParseAuthFile(authJsonFile)  // read #2 inside
```

Fix: parse `rawBytes` directly instead of calling `ParseAuthFile`:

```go
rawBytes, err := os.ReadFile(authJsonFile)
if err != nil {
    return err
}
var authConfig vercelutil.AuthConfig
if err := json.Unmarshal(rawBytes, &authConfig); err != nil {
    return err
}
```

---

### 5. `GetTeams` and related types are dead code — **Low**

**File:** `pkg/vercelapi/vercelapi.go:62-150`

`GetTeams()`, `Team`, `GetTeamsResponse`, and `Pagination` are exported but nothing in the codebase calls them after the `switchteam` removal. Dead exported code adds maintenance surface and misleads future readers into thinking teams are still a first-class concept. Remove them, or move to an internal `_archive` if there is any intention to reuse later.

---

### 6. Misspelled public method `Deleete` — **Low**

**Files:** `pkg/jsonupdate/jsonupdate.go:31`, `pkg/vercelutil/vercelutil.go:52`

The method is named `Deleete` (double `e`) in both the definition and the call site, so the code compiles and runs correctly. However, it is a public method with a misspelled name that would break any future caller relying on the expected name `Delete`. Rename to `Delete` in both files.

---

### 7. `utils.IsFileExists` is dead code — **Info**

**File:** `pkg/utils/utils.go:26`

`IsFileExists` is not called from anywhere in the current codebase. It was previously used to check for the SQLite DB file in `initCmd`, but `initCmd` now uses `os.Stat` directly. The function can be removed, along with the duplicate `os.ReadFile` call inside it (checking existence by reading the entire file is wasteful).

---

### 8. `DeleteCurrentTeam` reformats `config.json` from pretty to minified — **Info**

**File:** `pkg/vercelutil/vercelutil.go:54`

```go
err = os.WriteFile(filePath, []byte(jsonupd.String()), 0644)
```

The Vercel CLI writes `config.json` with indented formatting. `jsonupd.String()` produces minified (single-line) JSON. When vercelgate removes `currentTeam`, the file silently changes format, creating noisy diffs and confusing users who inspect the file. Use `jsonupd.Pretty()` to preserve the expected formatting.

---

### 9. Typo in root command description — **Info**

**File:** `main.go:45`

```go
Long: `You can swithc between multiple accounts...`,
```

`swithc` → `switch`

---

### 10. `switch` command short description is ungrammatical — **Info**

**File:** `main.go:89`

```go
Short: "Switch between account",
```

Should be "Switch between accounts".

---

### 11. Module path does not match repository owner — **Info**

**File:** `go.mod:1`

```
module github.com/khanakia/vercelgate
```

The git remote and commit author are `albertoZurini`, not `khanakia`. This mismatch suggests the project was forked but the module path was not updated. While this works locally, `go get github.com/albertoZurini/vercelgate` would not resolve correctly, and import paths in the source embed the original author's namespace.

---

## Resolved Since Previous Audit (3909183)

| # | Was | Resolution |
|---|-----|------------|
| 2 | Auth file written world-readable (0644) | `RestoreAuthJson` now uses `0600`; `SetAuthToken` removed |
| 3 | Auth token stored in plaintext SQLite | SQLite removed; token stored in `vercelgate_accounts.json` at `0600` |
| 4 | `DeleteCurrentTeam` error silently dropped | `main.go:119` now checks and propagates the error |
| 6 | `GetTeams` does not paginate | Function is now unreachable (see finding 5 above) |
| 7 | Silently wrong DB path when config dir missing | `pkg/entdb` removed entirely |
| 8 | `Migrate()` calls `log.Fatalf` before `return err` | `pkg/entcfn` removed entirely |
| 10 | `Migrate()` closes singleton DB client | `pkg/entcfn` removed entirely |
| 11 | Inconsistent JSON formatting (Pretty vs String) | `SetCurrentTeam` removed; residual minification issue noted in finding 8 |

---

## Summary Table

| # | Severity | File | Issue |
|---|----------|------|-------|
| 1 | **High** | `pkg/utils/utils.go:12,29` | Nil dereference on non-ErrNotExist stat error |
| 2 | **Medium** | `pkg/vercelapi/vercelapi.go:24` | No HTTP timeout — sync can hang indefinitely |
| 3 | **Medium** | `vercelutil:46,56` · `vercelapi:28,35` | Errors printed twice (in function and at call site) |
| 4 | **Low** | `pkg/vercelfn/vercelfn.go:23,28` | `auth.json` read twice during sync |
| 5 | **Low** | `pkg/vercelapi/vercelapi.go:62` | `GetTeams`, `Team`, `Pagination` are dead exported code |
| 6 | **Low** | `pkg/jsonupdate/jsonupdate.go:31` | Misspelled public method name `Deleete` |
| 7 | **Info** | `pkg/utils/utils.go:26` | `IsFileExists` is dead code |
| 8 | **Info** | `pkg/vercelutil/vercelutil.go:54` | `DeleteCurrentTeam` minifies `config.json` formatting |
| 9 | **Info** | `main.go:45` | Typo: `swithc` |
| 10 | **Info** | `main.go:89` | Grammar: "between account" → "between accounts" |
| 11 | **Info** | `go.mod:1` | Module path references original fork owner |
