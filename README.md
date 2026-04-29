# Insighta CLI

A minimal command-line client for the Insighta platform. Authenticate with GitHub, query and export profile data, and manage your session from the terminal.

## Prerequisites

- Go toolchain (to build from source) or download a release binary.

## Quick install

### Build from source

```bash
git clone <this-repo>
cd insighta-cli
go install .
```

### Download release

Download the latest release binary and place it on your `$PATH`.

## Configuration

By default the CLI targets production. Override with:

```bash
export INSIGHTA_API_URL=http://localhost:8080
```

Credentials are stored in `~/.insighta/credentials.json` (owner read/write only).

## Authentication

- `insighta login` starts a PKCE + GitHub OAuth flow (opens a browser, runs a local callback server).
- Access tokens are short-lived JWTs (3 minutes). Refresh tokens are rotating, single-use, and expire after ~5 minutes. The CLI auto-refreshes access tokens; if the refresh token is expired you must run `insighta login` again.

## Commands (examples)

### Auth

```bash
insighta login    # GitHub OAuth (PKCE)
insighta logout   # Revoke session, clear credentials
insighta whoami   # Show authenticated user
```

### Profiles

```bash
insighta profiles list --country NG --min-age 25 --limit 20
insighta profiles get <id>
insighta profiles search "young males from nigeria"
insighta profiles export --format csv --country NG
```

## Behavior

- Output: formatted tables in terminal.
- Exports: CSV files saved to the current working directory.
- Errors: printed to stderr with clear messages.

## Backend reference

The backend is a lightweight Go service (Go 1.26, chi, SQLite) providing GitHub OAuth, short-lived JWTs, rotating refresh tokens, and profile APIs.

### Routes

Health

```txt
GET /health
```

Auth (unauthenticated)

```txt
GET  /auth/github                # redirect to GitHub (supports PKCE)
GET  /auth/github/callback       # callback (exchange code)
POST /auth/github/callback       # callback (exchange code)
POST /auth/refresh               # rotate tokens (body: {"refresh_token"})
POST /auth/logout                # revoke refresh token
```

Authenticated API (requires `Authorization: Bearer <jwt>` and header `X-API-Version: 1`)

```txt
GET /auth/me                     # current user
```

Profiles (authenticated)

```txt
GET  /api/profiles               # list (filters, pagination)
GET  /api/profiles/search?q=...  # NLP-backed search
GET  /api/profiles/{id}          # single profile
GET  /api/profiles/export?format=csv  # CSV export
POST /api/profiles               # create (admin only)
```

See the backend README for server setup, env vars, rate limits, and migration details.

## Project layout (top-level)

- `cmd/` — cobra commands (root, auth, profiles)
- `internal/credentials` — credential storage
- `internal/client` — HTTP client with auto-refresh
- `internal/display` — table renderer and spinner
- `main.go`

## License

See repository for license and releases.
