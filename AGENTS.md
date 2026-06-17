# Agents

## Build Commands

```sh
go build -ldflags="-s -w" -o bin/craftctl ./cmd/main.go        # Backend (run after frontend build)
cd frontend && npm install && npm run build && cd ..           # Frontend
```

## Test Commands

```sh
go test ./...                                                  # All Go tests
go test ./internal/...                                         # Backend tests
cd frontend && npm test                                        # Frontend tests
```

## Lint / Format

```sh
go vet ./...                                                   # Go vet
cd frontend && npm run lint                                    # Frontend type-check (vue-tsc)
```

## Code Style

- **Go**: stdlib layout (`internal/{config,mc,server,ui}/`). Fiber v2, zap for logging, resty for HTTP.
- **Frontend**: Vue 3 Composition API + TypeScript, Vite. Pinia-like store in `store.ts`, API client in `api.ts`.
- **Imports**: Group stdlib, third-party, internal.
- **Errors**: Return early; wrap with `fmt.Errorf("context: %w", err)`.
- **No commented-out code** or debug logs in commits.
- **Concurrency**: Use `sync.Mutex` for shared state; avoid raw goroutines without lifecycle management.

## Project Structure

- `cmd/main.go` — Entry point, signal handling
- `internal/server/` — HTTP handlers + Fiber setup
- `internal/mc/` — Minecraft process + server.properties
- `internal/config/` — Paths and defaults
- `internal/ui/` — Embedded frontend dist
- `frontend/` — Vue 3 SPA source

## Key Conventions

- Server data lives in `~/craftctl/servers/default/` (or `APPDATA` on Windows).
- All HTTP handlers accept/return JSON.
- WebSocket at `/ws` for console streaming.
- Frontend builds to `internal/ui/dist/` and is embedded via `go:embed`.
- `internal/ui/dist/` is gitignored — run `./build-frontend.sh` before `go build`.
