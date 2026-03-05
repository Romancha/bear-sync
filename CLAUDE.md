# salmon

Monorepo with two Go binaries for syncing Bear notes with external consumers.

## Project Structure

- cmd/hub/ — API server (runs on VPS)
- cmd/bridge/ — Mac agent that reads Bear SQLite and syncs with hub
- internal/models/ — shared data models (Note, Tag, Attachment, Backlink, WriteQueueItem)
- internal/mapper/ — Bear SQLite → Hub model mapping
- internal/beardb/ — Bear SQLite reader (bridge only)
- internal/hubclient/ — HTTP client for hub API (bridge only)
- internal/store/ — SQLite store for hub (hub only)
- internal/api/ — HTTP handlers with chi router (hub only); Swagger UI at /api/docs/
- internal/api/docs/ — generated OpenAPI spec (swag init, committed to repo)
- internal/xcallback/ — Bear x-callback-url executor via bear-xcall CLI (bridge only)
- internal/ipc/ — Unix socket IPC server for daemon mode (bridge only)
- tools/bear-xcall/ — Swift CLI source for bear-xcall .app bundle (macOS only, bridge dependency)
- tools/salmon-run-app/ — SwiftUI menu bar app (macOS 14+, Xcode project, wraps bridge daemon)
- tools/create-dmg.sh — script to create .dmg disk image for SalmonRun.app distribution
- deploy/ — deployment configs (systemd unit, Caddyfile)
- docs/ — consumer-facing documentation (API quick start guide)
- testdata/ — test fixtures (test Bear SQLite)

## Commands

- make build — build both binaries to bin/ (accepts VERSION=vX.Y.Z for bridge version injection, default: dev)
- make test — run all tests
- make test-coverage — run tests with coverage report
- make test-race — run tests with race detector
- make lint — run golangci-lint
- make fmt — format code (gofumpt + goimports)
- make tidy — go mod tidy
- make build-xcall — build bear-xcall Swift CLI .app bundle (macOS only)
- make generate — run go generate (moq)
- make swagger — generate Swagger docs (swag init)
- make test-xcall — run bear-xcall manual tests (macOS + Bear required)
- make build-app — build SalmonRun menu bar .app bundle (macOS only)
- make test-app — run SalmonRun Swift tests (macOS only)
- make dmg — create SalmonRun .dmg disk image (macOS only)

## After Making Changes

Run these checks before committing (in order):

1. `make fmt` — format code
2. `make lint` — run linter, fix all warnings
3. `make test` — ensure all tests pass
4. `make test-race` — ensure no data races
5. `make test-coverage` — ensure coverage does not decrease compared to the main branch
6. `make tidy` — update go.mod/go.sum if dependencies changed
7. If interfaces changed: `make generate` — regenerate mocks
8. If swag annotations changed: `make swagger` — regenerate OpenAPI spec

## Code Patterns

- Interface-first design for testability (Store, BearDB, HubClient, XCallback)
- Mocks via moq with //go:generate directives
- Structured logging with log/slog
- context.Context in all external operations
- Error wrapping: fmt.Errorf("message: %w", err)
- Configuration via environment variables (no config files)
- Tests with github.com/stretchr/testify (assert/require); Swift tests with XCTest (tools/salmon-run-app/)
- Line length limit: 140 characters
- SQLite via modernc.org/sqlite (pure Go, no CGO)

## Database

- Hub uses SQLite with WAL mode, busy_timeout=5000, foreign_keys=ON
- Dual-ID pattern: hub UUID (PK) + bear_id (Bear's UUID, nullable, UNIQUE)
- FTS5 for full-text search on notes (title + body)

## Sync Architecture

- Bear is source-of-truth for user content
- Hub is read replica + write queue for external consumers
- Write actions: create, update, add_tag, trash, add_file, archive, rename_tag, delete_tag
- Write flow: consumer → hub write_queue → bridge lease → bear-xcall to Bear → ack
- Read flow: Bear → bridge delta export → hub sync/push → consumer API
- Delivery: effectively-once (consumer→hub), at-least-once (hub→bridge), duplicate-safe (bridge apply)

## Auth

- Multiple consumer Bearer tokens (api/* scope) configured via `SALMON_HUB_CONSUMER_TOKENS`, plus one bridge token (sync/* scope)
- Each consumer is identified by name and authenticated with its own token
- Write queue items are attributed to the originating consumer via `consumer_id`
- Encrypted notes are read-only (403 for write operations)
- All mutating consumer API requests (POST/PUT/DELETE) require an `Idempotency-Key` header; missing header returns HTTP 400

## Note sync_status lifecycle

- `synced`: normal state; Bear delta pushes overwrite hub fields freely
- `pending_to_bear`: a consumer has enqueued a write; hub will NOT overwrite `title`/`body` from Bear delta pushes while in this state
- `conflict`: set when a Bear push arrives for a `pending_to_bear` note with a newer `modified_at`; bridge creates a `[Conflict] Title` note in Bear instead of applying the queue item
- Transitions: `synced` → `pending_to_bear` (on enqueue) → `synced` (on ack with "applied") or `conflict` (on conflicting Bear push)

## Database

- Bear SQLite uses Core Data epoch timestamps (float64 seconds since 2001-01-01)
- Conversion: `unix_ts = core_data_ts + 978307200` (defined as `mapper.CoreDataEpochOffset`)

## Bridge Daemon Mode

- `--daemon` flag: runs continuous sync loop instead of one-shot (used by SalmonRun.app)
- `--version` flag: prints version and exits
- `SALMON_SYNC_INTERVAL`: sync interval in seconds for daemon mode (default: 300)
- `SALMON_IPC_SOCKET`: Unix socket path for IPC (default: `~/.salmon.sock`)

## IPC (Daemon Mode)

- Unix socket at `~/.salmon.sock` (configurable via `SALMON_IPC_SOCKET`)
- JSON-based newline-delimited request/response protocol
- Commands: status, sync_now, logs, queue_status, quit
- Stats tracked via `ipc.StatsTracker` (notes, tags, queue items, last sync duration)
- Structured status events emitted to stdout (sync_start, sync_progress, sync_complete, sync_error)

## Bridge State

- State file: `~/.salmon-state.json` (path overridable via `SALMON_STATE_PATH`)
- `last_sync_at`: Core Data epoch (float64, NOT Unix epoch) — used as `>= lastSyncAt` delta read threshold
- `junction_full_scan_counter`: incremented every cycle; triggers full junction table scan every 12 cycles (`junctionFullScanInterval`)
- `known_*_ids`: Bear UUIDs seen on last sync; diffed against current Bear DB to produce `deleted_*_ids` in push requests
- Absent state file → initial sync (full export in 50-note batches); state written atomically (write to `.tmp` then `rename`)

## CI/CD

- `.github/workflows/ci.yml` — lint, test, test-race on push/PR
- `.github/workflows/docker-publish.yml` — builds and pushes hub Docker image on v* tag
- `.github/workflows/release-bridge.yml` — builds, signs, notarizes, and publishes SalmonRun.app as .dmg disk images on v* tag (arm64 + amd64)
