# bear-sync

Syncs Bear notes with openclaw. Two components: **hub** (API server on VPS) and **bridge** (Mac agent that reads Bear SQLite).

## Architecture

- Bear is source-of-truth for user content
- Hub is a read replica with a write queue for openclaw
- Read flow: Bear → bridge → hub → openclaw API
- Write flow: openclaw → hub write_queue → bridge → Bear x-callback-url → ack

## Prerequisites

- Go 1.24+
- Bear.app (for bridge)
- [xcall](https://github.com/nicoulaj/xcall) CLI (for bridge write operations)

## Build

```
make build
```

Binaries are placed in `bin/bear-sync-hub` and `bin/bear-bridge`.

## Hub Setup

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `HUB_HOST` | No | `127.0.0.1` | Listen host (`0.0.0.0` for Docker) |
| `HUB_PORT` | No | `8080` | Listen port |
| `HUB_DB_PATH` | Yes | — | Path to SQLite database file |
| `HUB_OPENCLAW_TOKEN` | Yes | — | Bearer token for openclaw API access |
| `HUB_BRIDGE_TOKEN` | Yes | — | Bearer token for bridge sync access |
| `HUB_ATTACHMENTS_DIR` | No | `attachments` | Directory for attachment file storage |

### Running

```
export HUB_DB_PATH=/opt/bear-sync/data/hub.db
export HUB_OPENCLAW_TOKEN=<token>
export HUB_BRIDGE_TOKEN=<token>
./bin/bear-sync-hub
```

The hub listens on `127.0.0.1:PORT` (localhost only). Use a reverse proxy (e.g. Caddy) for TLS termination.

### Systemd (production)

```
sudo cp deploy/bear-sync-hub.service /etc/systemd/system/
sudo systemctl enable bear-sync-hub
sudo systemctl start bear-sync-hub
```

Create `/opt/bear-sync/.env` with the environment variables above.

### Docker Compose (production)

1. Copy `.env.example` to `.env` and fill in secrets:

```
cp .env.example .env
```

2. Set your domain in `.env`:

```
HUB_OPENCLAW_TOKEN=<token>
HUB_BRIDGE_TOKEN=<token>
DOMAIN=bear-sync.example.com
```

3. Start the stack:

```
docker compose up -d
```

This starts the hub server and Caddy reverse proxy with automatic TLS. The hub is accessible only through Caddy (ports 80/443).

To check status:

```
docker compose ps
curl https://your-domain.com/healthz
```

To update to a new version:

```
docker compose pull
docker compose up -d
```

Data is persisted in Docker named volumes (`hub-data` for SQLite + attachments).

## Bridge Setup

### Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `BRIDGE_HUB_URL` | Yes | — | Hub API URL (e.g. `https://bear-sync.example.com`) |
| `BRIDGE_HUB_TOKEN` | Yes | — | Bearer token matching `HUB_BRIDGE_TOKEN` |
| `BEAR_TOKEN` | Yes | — | Bear app API token (from Bear preferences) |
| `BRIDGE_STATE_PATH` | No | `~/.bear-bridge-state.json` | Path to bridge state file |
| `BEAR_DB_DIR` | No | `~/Library/Group Containers/9K33E3U3T4.net.shinyfrog.bear/Application Data` | Path to Bear Application Data directory |

### Running

```
export BRIDGE_HUB_URL=https://bear-sync.example.com
export BRIDGE_HUB_TOKEN=<token>
export BEAR_TOKEN=<token>
./bin/bear-bridge
```

The bridge runs once per invocation (no daemon mode). Use launchd to schedule periodic runs.

### Launchd (production)

```
cp deploy/com.romancha.bear-bridge.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.romancha.bear-bridge.plist
```

Edit the plist to set your tokens and hub URL. Default interval: every 5 minutes.

## Reverse Proxy

A sample Caddyfile is provided in `deploy/Caddyfile` for systemd setup. For Docker Compose, `deploy/Caddyfile.docker` is used automatically.

The sample Caddyfile uses rate limiting, which requires the [caddy-ratelimit](https://github.com/mholt/caddy-ratelimit) plugin. Build Caddy with this plugin using `xcaddy`:

```
xcaddy build --with github.com/mholt/caddy-ratelimit
```

## Development

```
make test          # run all tests
make test-race     # run tests with race detector
make lint          # run golangci-lint
make fmt           # format code
make tidy          # go mod tidy
```

## CI/CD

GitHub Actions runs automatically:

- **CI** (push/PR to main): lint, test, test with race detector
- **Docker Publish** (push tag `v*`): builds multi-platform image (`linux/amd64`, `linux/arm64`) and pushes to `ghcr.io/romancha/bear-sync-hub`

To publish a new release:

```
git tag v0.1.0
git push origin v0.1.0
```
