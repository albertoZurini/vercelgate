# vcx

Switch between multiple Vercel accounts without re-logging in — like `kubectx` but for Vercel.

## Why

As developers, we often work with clients who share their Vercel accounts instead of subscribing to a Team plan. The Vercel CLI only supports one active account at a time, requiring a full logout/login cycle to switch. vcx eliminates that friction by storing each account's credentials locally and swapping them instantly.

## Installation

```sh
go install github.com/albertoZurini/vcx@main
```

### Homebrew

```sh
brew tap albertoZurini/vcx
brew install vcx
```

## Ubuntu

On Ubuntu, Vercel stores its config in `~/.local/share/com.vercel.cli/` instead of `~/.config/com.vercel.cli/`.

Create a symlink so vcx can find it:

```sh
mkdir -p ~/.config/com.vercel.cli && ln -s ~/.local/share/com.vercel.cli/* ~/.config/com.vercel.cli/
```

---

## Getting started

### Step 1 — Initialize (once)

```bash
vcx init
```

Creates `vcx_accounts.json` in your Vercel config directory.

### Step 2 — Sync your accounts

**Already logged in to Vercel:**

```bash
vcx sync
```

**Adding a second (or third) account:**

```bash
vcx new        # clears current Vercel session
vercel login   # log in as the new account
vcx sync       # save the new account to vcx
```

> [!NOTE]
> `vcx new` does **not** log you out from the Vercel platform. It only clears the local credential file so the Vercel CLI can accept a new `vercel login`.

### Step 3 — Switch between accounts

```bash
vcx switch
```

Use the arrow keys to pick an account. The selected account's full `auth.json` is restored, so all subsequent `vercel` commands run under that account.

---

## Commands

| Command | Description |
|---------|-------------|
| `vcx init` | Create the accounts file (run once) |
| `vcx sync` | Save the currently logged-in Vercel account to vcx |
| `vcx switch` | Interactively select an account to activate |
| `vcx new` | Clear the current session so you can `vercel login` as someone else |
| `vcx reset` | Wipe all stored accounts |
| `vcx path` | Print the Vercel config directory vcx is reading from |
| `vcx --version` | Print the installed version |
| `vcx --debug` | Enable detailed logging |
| `vcx --debug --verbose` | Debug + per-function argument logging (tokens masked) |

---

## Example: adding a second account

```sh
# Already logged in as jane@example.com — sync her first
➜ vcx sync
synced successfully

# Add a second account
➜ vcx new
you can now add new account using `vercel login` and then run `vcx sync` again

➜ vercel login
> Success! Email authentication complete for dummy@gmail.com

➜ vcx sync
synced successfully

# Switch between them at any time
➜ vcx switch
? Select Account:
  ▸ Jane Doe (jane@example.com)
    dummy (dummy@gmail.com)

Switched to Jane Doe
```

---

## Credits

vcx is a fork of [vercelgate](https://github.com/khanakia/vercelgate) by [@khanakia](https://github.com/khanakia), which provided the original boilerplate for multi-account Vercel credential management. vcx replaces the SQLite/Ent ORM backend with a plain JSON file and drops the team-switching command.

---

## TODO

- [ ] Auto-detect Ubuntu config file path
