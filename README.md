# gh-relay

> Share a private GitHub repo with anyone, no collaborator invite, no paid seat, no cleanup.

[![Release](https://github.com/soub4i/gh-relay/actions/workflows/release.yaml/badge.svg)](https://github.com/soub4i/gh-relay/actions/workflows/release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/soub4i/gh-relay)](https://goreportcard.com/report/github.com/soub4i/gh-relay)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

Adding a contractor or auditor as a GitHub collaborator means IT tickets, legal paperwork, a paid seat, and a permission that lingers long after the review is done. Most teams end up emailing zip files or screensharing instead, both worse.

**gh-relay** fixes this in one command. Run it on your machine, share a temporary URL, and your guest gets a read-only browser view of the codebase. When you press `Ctrl+C`, the link is dead, zero cleanup, zero lingering access.

```
$ gh-relay share --token ghp_... --repo my-org/private-app --expire 1h

   Token valid
   Repository: my-org/private-app (private)
   Found 12 branch(es)
   Tunnel active

    Share this URL with your guest:
    https://shiny-apple-92.trycloudflare.com

    Session expires in: 1h
  Press Ctrl+C to end the session immediately.
```

---

## How it works

```
┌─────────┐      HTTPS tunnel       ┌──────────────────────┐     GitHub API
│  Guest  │ ◄────────────────────── │  gh-relay            │ ──────────────►
│ browser │                         │  (your machine)      │  (your PAT)
└─────────┘  sees code, never token └──────────────────────┘
```

1. **You run** `gh-relay share` with your GitHub token and repo name.
2. **A local proxy** starts on your machine and fetches files from the GitHub API using your token.
3. **A secure tunnel** (Cloudflare or ngrok) exposes the proxy via a temporary public URL.
4. **Your guest opens the URL** and gets a read-only file browser, no GitHub account required.
5. **You press `Ctrl+C`** (or `--expire` elapses) and the tunnel closes instantly. The URL is dead.

Your token never leaves your machine. The guest can't push, clone, or access anything you didn't share.

---

## Installation

### Homebrew *(coming soon)*

```bash
brew install soub4i/tap/gh-relay
```

### Download a binary

Grab the latest release for your platform from the [Releases page](https://github.com/soub4i/gh-relay/releases).

```bash
# macOS (Apple Silicon)
curl -L https://github.com/soub4i/gh-relay/releases/latest/download/gh-relay_darwin_arm64.tar.gz | tar xz
sudo mv gh-relay /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/soub4i/gh-relay/releases/latest/download/gh-relay_darwin_amd64.tar.gz | tar xz
sudo mv gh-relay /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/soub4i/gh-relay/releases/latest/download/gh-relay_linux_amd64.tar.gz | tar xz
sudo mv gh-relay /usr/local/bin/

# Windows (amd64) - run in PowerShell
Invoke-WebRequest -Uri https://github.com/soub4i/gh-relay/releases/latest/download/gh-relay_windows_amd64.zip -OutFile gh-relay.zip
Expand-Archive gh-relay.zip
```

### Build from source

Requires **Go 1.22+**.

```bash
git clone https://github.com/soub4i/gh-relay
cd gh-relay
go build -o gh-relay .
sudo mv gh-relay /usr/local/bin/
```

### Tunnel prerequisite

You need at least one tunnel provider installed:

| Provider | Install |
|---|---|
| **Cloudflare** *(recommended, free, no account needed)* | [developers.cloudflare.com → Downloads](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) |
| **ngrok** | [ngrok.com/download](https://ngrok.com/download) |

---

## Usage

### Share a repo

```bash
gh-relay share \
  --token ghp_YourPersonalAccessToken \
  --repo my-org/private-app
```

### Set an expiry

```bash
gh-relay share \
  --token ghp_... \
  --repo my-org/private-app \
  --expire 30m
```

### Share a specific branch

```bash
gh-relay share \
  --token ghp_... \
  --repo my-org/private-app \
  --branch feature/new-auth
```

### Use ngrok instead of Cloudflare

```bash
gh-relay share \
  --token ghp_... \
  --repo my-org/private-app \
  --tunnel ngrok
```

### Local only (no tunnel, useful for testing)

```bash
gh-relay share \
  --token ghp_... \
  --repo my-org/private-app \
  --tunnel none \
  --port 8080
# Open http://localhost:8080 in your browser
```

### Enable audit logging

```bash
gh-relay share \
  --token ghp_... \
  --repo my-org/private-app \
  --audit
```

Logs guest activity to the terminal and prints a summary on exit:

```
[audit] Guest viewed: src/main.go (from 105.190.183.127)
[audit] GET /api/commits (from 105.190.183.127)

  SESSION AUDIT SUMMARY
  Files viewed  : 5 (3 unique)
  Total requests: 12
  Unique IPs    : 1
  Duration      : 4m32s
```

### All flags

| Flag | Default | Description |
|---|---|---|
| `--token` | *(required)* | GitHub PAT with `repo` or `public_repo` scope |
| `--repo` | *(required)* | Target repository in `owner/repo` format |
| `--branch` | `main` | Branch to share |
| `--port` | `8080` | Local port for the proxy server |
| `--expire` | unlimited | Auto-close after this duration (`30m`, `1h`, `2h30m`) |
| `--tunnel` | `cloudflare` | Tunnel provider: `cloudflare`, `ngrok`, or `none` |
| `--audit` | `false` | Log guest activity and print a session summary on exit |

---

## Security

gh-relay is designed from the ground up to share as little as possible.

| Property | How it's enforced |
|---|---|
| **Token never leaves your machine** | All GitHub API calls are made server-side. The guest only receives a short-lived session . |
| **Read-only by design** | The proxy only registers `GET` handlers. `POST`, `PATCH`, `DELETE` return `405` before any session check. |
| **Nothing written to disk** | Files are fetched on demand and streamed directly to the guest. No `git clone`, no temp files. |
| **Instant teardown** | `Ctrl+C` or `--expire` kills the tunnel, shuts the server, and invalidates all session cookies simultaneously. |
| **No external dependencies in the browser** | The file browser SPA is fully self-contained. No third-party scripts, no CDN calls from the guest's browser. |
| **Zero trust** | The guest can only see what you share. They can't navigate to other repos, access your GitHub account, or do anything outside the API calls you allow. |
| **No dependencies on third-party services** | The core functionality relies only on GitHub and your chosen tunnel provider. No analytics, no databases, no external APIs. |

### GitHub token scopes

For **fine-grained PATs** (recommended):
- `Contents: Read-only`
- `Metadata: Read-only`

For **classic PATs**:
- `repo` | for private repositories
- `public_repo` | for public repositories only

> **Tip:** Create a dedicated PAT for gh-relay sessions with the minimum required scopes and a short expiry matching your longest expected review session.

---

## Guest experience

The guest opens the URL in any browser, no GitHub account, no sign-in, no extension required. They see:

- A **file tree** with folder expand/collapse and a live filter
- A **syntax-highlighted code viewer** for all common languages (Go, Python, Rust, TypeScript, JS, YAML, JSON, and more)
- A **branch switcher** to explore other branches
- A **commit history** panel showing recent commits on the active branch

They cannot clone, push, download a zip, or navigate outside the repository you shared.

---

## Project layout

```
├── cmd
│   ├── root.go
│   └── share.go // CLI command definitions and flag parsing
├── go.mod
├── internal
│   ├── github
│   │   ├── client.go // GitHub API client and handlers
│   │   ├── types.go
│   │   └── utils.go
│   ├── logo
│   │   └── logo.go
│   ├── server
│   │   ├── server.go // HTTP server and handlers
│   │   ├── spa.go
│   │   ├── types.go
│   │   └── utils.go
│   ├── session
│   │   ├── manager.go 
│   │   ├── types.go
│   │   └── utils.go
│   ├── tunnel
│   │   ├── cf.go // Cloudflare tunnel support
│   │   ├── ngrok.go // ngrok tunnel support
│   │   └── tunnel.go // Tunnel interface and factory
│   └── version
│       └── version.go
├── LICENSE
├── main.go
└── README.md
```

Zero external Go dependencies.

---

## Contributing

Contributions are welcome. Please open an issue before starting work on a large change so we can discuss the approach first.

```bash
# Run tests
go test ./...

# Run with race detector
go test -race ./...

# Vet
go vet ./...

# Build
go build -o gh-relay .
```

---

## Roadmap

- [x] File browsing, syntax highlighting, Cloudflare tunnel
- [x] Branch switcher, commit history
- [] Markdown rendering for README files
- [ ] System keychain integration for token storage
- [ ] Optional one-time password protection for the shared URL
- [ ] `gh` CLI extension (`gh relay share ...`)

---

## License

[MIT](LICENSE) -  © 2025 soub4i

---

> gh-relay is intended for internal collaboration and peer review. Always ensure that sharing code via a relay complies with your organisation's data-handling policies and Terms of Service