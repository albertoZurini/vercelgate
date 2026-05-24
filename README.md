## vercelgate

vercelgate is a command-line tool designed to streamline the process of managing multiple Vercel client accounts. It eliminates the need for repetitive login and logout actions, allowing users to switch between accounts and teams seamlessly.

## WHY

As developers, we often encounter situations where clients provide access to their Vercel accounts without subscribing to a Team plan. The Vercel CLI restricts usage to one account at a time, requiring you to log out from one account before logging into another. This can be a significant inconvenience, especially since the Vercel CLI does not natively support multi-account management, likely encouraging users to opt for a paid Team plan.

vercelgate offers a practical solution by enabling users to switch between multiple personal Vercel hobby plan accounts without the need to upgrade to a Vercel Pro Team plan.

## Installation

```sh
go install github.com/khanakia/vercelgate@main
```

### Install with Homebrew:

```sh
brew tap khanakia/vercelgate
brew install vercelgate
```

## Ubuntu

For ubuntu vercel configuration is stored in `~/.local/share/com.vercel.cli/` instead of `~/.config/com.vercel.cli/`

Make sure you create a Symlink: `mkdir -p ~/.config/com.vercel.cli && ln -s ~/.local/share/com.vercel.cli/* ~/.config/com.vercel.cli/`

---

## Getting started

### Step 1 — Initialize (once)

```bash
vercelgate init
```

This sets up the local database. Only needs to be done once.

### Step 2 — Sync your accounts

#### Already logged in to Vercel

If you are already logged in via the Vercel CLI, sync your current session immediately — no logout needed:

```bash
vercelgate sync
```

#### Adding a second (or third) account

`vercelgate new` clears Vercel's local auth file so you can log in as a different user. Your already-synced accounts are **not** affected — they remain stored in vercelgate's database.

```bash
vercelgate new    # clears current Vercel session
vercel login      # log in as the new account
vercelgate sync   # save the new account to vercelgate
```

Repeat this for every additional account you want to manage.

> [!NOTE]
> `vercelgate new` does **not** log you out from the Vercel platform. It only clears the local credential file so the Vercel CLI can accept a new `vercel login`.

### Step 3 — Switch between accounts

```bash
vercelgate switch
```

Use the arrow keys to pick which account to activate. The selected account's token is written back into Vercel's local auth file, so all subsequent `vercel` commands run under that account.

To also switch the active team at the same time:

```bash
vercelgate switchteam
```

---

## Commands

### `vercelgate init`

Initializes vercelgate for first-time use. Sets up the local SQLite database.

### `vercelgate sync`

Reads the token from the current Vercel CLI session and saves (or updates) that account in vercelgate's database, together with all its teams.

### `vercelgate new`

Clears the current Vercel CLI session so you can run `vercel login` to authenticate as a different account. Run `vercelgate sync` afterwards to add the new account.

### `vercelgate switch`

Displays all synced accounts and lets you select one to activate.

### `vercelgate switchteam`

Displays all synced accounts and their teams, and lets you select both an account and a team to activate.

### `vercelgate reset`

Removes all synced accounts from vercelgate's database. Does not affect the Vercel CLI itself.

### `vercelgate path`

Prints the Vercel global configuration directory that vercelgate is reading from. Useful for troubleshooting.

### `vercelgate --version`

Prints the installed version of vercelgate.

### `vercelgate --debug`

Enables detailed logging of every operation (paths resolved, API calls made, DB writes).

### `vercelgate --debug --verbose`

Enables all debug output plus per-function argument logging. Tokens are masked in the output.

---

## Example: adding a second account

```sh
# Already logged in as jane@example.com — sync her first
➜ vercelgate sync
synced successfully

# Now add a second account
➜ vercelgate new
you can now add new account using `vercel login` and then run `vercelgate sync` again

➜ vercel login
Vercel CLI 34.0.0
? Log in to Vercel Continue with Email
? Enter your email address: dummy@gmail.com
We sent an email to dummy@gmail.com. Please follow the steps provided inside it and make sure the security code matches Eager Bornean Orang-utan.
> Success! Email authentication complete for dummy@gmail.com

➜ vercelgate sync
synced successfully

# Switch between them at any time
➜ vercelgate switch
Use the arrow keys to navigate: ↓ ↑ → ←
? Select Account:
  ▸ Jane Doe (jane@example.com)
    dummy (dummy@gmail.com)

Switched to user Jane Doe
```

---

## Features

- **Simple account switching** — switch between Vercel accounts without logging out
- **Multiple accounts** — manage as many personal / hobby plan accounts as you need
- **Team management** — switch the active team alongside the account
- **Debug logging** — `--debug` and `--debug --verbose` flags for troubleshooting
- **Configuration path** — `vercelgate path` shows where config is being read from

## TODO

- [ ] Auto detect Ubuntu config file path
