# Identity and accounts

## Connection types

| Auth path | Persistence | Display |
|-----------|-------------|---------|
| No credentials | Guest | `guest` or SSH username if provided |
| SSH public key | Fingerprint-scoped KV | `key-<prefix>` unless nickname/account |
| Password (when accounts enabled) | Account-backed | SSH username = account name |

OpenSSH often sends your local login as the SSH username when you omit `user@`. That label is **not** persisted unless you use accounts or nicknames.

## Nicknames

`experience.ReserveNicknameEffect(name)` registers a unique nickname for the current identity. `Identity.String()` prefers nickname, then account name, then guest/key label.

## When anonymous SSH is allowed

With **no** account store, **no** key allow/deny lists, and `RequireAccount` false, clients can connect without a key (“none” auth path).

Enabling **any** of:

- `AllowKeys` / `DenyKeys`
- Password accounts
- `RequireAccount`

disables the open guest path so SSH clients do not stop at “none” before trying keys or passwords.

## Account store precedence

`portico.Serve` resolves stores in order:

1. `Options.AccountBackend` (`identity.AccountStore`)
2. `HashedAccounts` + `DataDir` → bcrypt files (`identity.OpenFileAccountStore`)
3. `Options.AccountStore` (dev plaintext `runtime.AccountStore`)
4. `Options.Accounts.Accounts`

If hashed accounts are requested and the directory cannot be opened, `Serve` returns an error (no silent fallback).

## Production passwords

Use `HashedAccounts: true` with `DataDir`, or set `AccountBackend` to a file-backed store. Do **not** use dev `AccountStore` in production.

`experience.RegisterAccountEffect` requires an enabled account store (registration scaffold).

## Key and account lists

| Option | Purpose |
|--------|---------|
| `AllowKeys` / `DenyKeys` | Fingerprint allow/deny |
| `AllowAccounts` / `DenyAccounts` | Account name allow/deny |
| `RequireAccount` | Reject guests and unlinked keys |

A public key remains a key for persistence until linked to an account.

## CLI account tools

With bcrypt files under `DataDir`:

```bash
portico account-export -data .portico-data
portico account-import -data .portico-data -file accounts.json
```

See [CLI](reference/cli.md).
