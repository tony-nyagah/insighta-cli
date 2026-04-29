# Insighta CLI

Command-line interface for the Insighta Labs+ platform.  
Authenticate via GitHub, query profiles, export data, and manage your account — all from the terminal.

---

## Installation

### Option A — build from source

```bash
git clone <this-repo>
cd insighta-cli/stage-3
go install .
```

This places `insighta` in your `$GOPATH/bin`. Make sure that directory is on your `$PATH`.

### Option B — download a pre-built binary

Download the latest release from the [releases page](../../releases) and move it to a directory on your `$PATH`:

```bash
chmod +x insighta
mv insighta /usr/local/bin/
```

### Verify installation

```bash
insighta --help
```

---

## Configuration

By default the CLI talks to the production backend. Override with an environment variable:

```bash
export INSIGHTA_API_URL=http://localhost:8080
```

Credentials are stored at `~/.insighta/credentials.json` (owner read/write only).

---

## Authentication Flow

```
insighta login
     │
     ├─ generates state, code_verifier, code_challenge (S256 / PKCE)
     ├─ starts a local HTTP server on a random port
     ├─ opens browser → backend /auth/github → GitHub OAuth page
     │
     │  (user authenticates on GitHub)
     │
     ├─ GitHub redirects to local callback server with ?code=...
     ├─ CLI validates state (anti-CSRF)
     ├─ CLI sends { code, code_verifier } to backend /auth/github/callback
     │
     └─ backend exchanges code, issues access + refresh tokens
        CLI stores tokens at ~/.insighta/credentials.json
```

### Token Handling

- **Access token** — 3-minute JWT. The CLI proactively refreshes it when fewer than 10 seconds remain before expiry.
- **Refresh token** — 5-minute opaque token. Single-use; each refresh issues a new pair.
- If the refresh token is also expired, the CLI prompts the user to run `insighta login` again.

---

## Commands

### Auth

```bash
insighta login          # Authenticate via GitHub OAuth (PKCE)
insighta logout         # Revoke session and clear stored credentials
insighta whoami         # Show currently authenticated user
```

### Profiles

```bash
# List profiles
insighta profiles list
insighta profiles list --gender male
insighta profiles list --country NG --age-group adult
insighta profiles list --min-age 25 --max-age 40
insighta profiles list --sort-by age --order desc
insighta profiles list --page 2 --limit 20

# Get a single profile
insighta profiles get <id>

# Natural language search
insighta profiles search "young males from nigeria"
insighta profiles search "female adults above 30"

# Create a profile (admin only)
insighta profiles create --name "Harriet Tubman"

# Export to CSV (saved to current directory)
insighta profiles export --format csv
insighta profiles export --format csv --gender male --country NG
```

### Flags

| Flag | Command | Description |
|---|---|---|
| `--gender` | list, export | `male` or `female` |
| `--age-group` | list, export | `child`, `teenager`, `adult`, `senior` |
| `--country` | list, export | ISO 3166-1 alpha-2 code (e.g. `NG`) |
| `--min-age` | list | Minimum age |
| `--max-age` | list | Maximum age |
| `--sort-by` | list | `age`, `gender_probability`, `created_at` |
| `--order` | list | `asc` or `desc` |
| `--page` | list | Page number (default: 1) |
| `--limit` | list | Results per page, max 50 (default: 10) |
| `--name` | create | Full name (required) |
| `--format` | export | Export format (`csv`) |

---

## Output

- Results display as formatted tables in the terminal
- A spinner is shown while requests are in-flight
- Exported CSV files are saved to the current working directory
- All errors print to stderr with a clear message

---

## Project Structure

```
cli/stage-3/
├── cmd/
│   ├── root.go              — cobra root command
│   ├── login.go             — PKCE OAuth login flow
│   ├── logout.go            — logout + whoami
│   └── profiles/
│       ├── profiles.go      — profiles parent command
│       ├── list.go          — list with filters
│       ├── get.go           — single profile
│       ├── search.go        — NLP search
│       ├── create.go        — create (admin)
│       └── export.go        — CSV export
├── internal/
│   ├── credentials/         — read/write ~/.insighta/credentials.json
│   ├── client/              — HTTP client with auto-refresh
│   └── display/             — table renderer + spinner
├── main.go
└── go.mod
```
