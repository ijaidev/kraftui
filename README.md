# KraftUI

Local console for [Kraft](https://unikraft.org) unikernels.

KraftUI is a single Go binary that talks to a supported `kraft` CLI on your
machine and serves a read-only web UI for machines, networks, volumes, and
packages. The UI is a Next.js app, statically exported and embedded into the
binary.

The HTTP API is documented in [`openapi/openapi.yaml`](openapi/openapi.yaml)
and is the contract between the Go server and the frontend.

## Requirements

- [Go](https://go.dev/dl/) 1.25.5
- [Node.js](https://nodejs.org/) 20+ and [pnpm](https://pnpm.io/)
- [just](https://github.com/casey/just)
- [Kraft](https://unikraft.org/docs/cli) **v0.12.15** on `PATH` (exact version)

KraftUI refuses to start if `kraft version` does not report `v0.12.15`. Point
`--kraft-binary` at another install if `kraft` is not on `PATH`.

## Quick start

Build the frontend, embed it, compile the binary, and run the server:

```bash
just run
```

The process listens on `127.0.0.1` only. With the default `--port 0` it tries
`5200` through `5204` and uses the first free port. Open the URL printed at
startup, typically <http://localhost:5200>.

Build without starting:

```bash
just build          # frontend export + Go binary → bin/kraftui
just build-go       # Go binary only (needs an existing embed in internal/ui/dist)
just frontend-build # Next.js static export → internal/ui/dist
```

Print the application version (`dev` unless overridden at link time):

```bash
./bin/kraftui --version
```

Override the version when compiling:

```bash
go build -ldflags "-X github.com/ijaidev/kraftui/config.version=0.1.0" -o bin/kraftui ./cmd/kraftui
```

## Development

Split the stack while iterating on the UI.

Terminal 1 — Go API (rebuilds the binary, uses the last embedded export):

```bash
just dev
# or
just dev-debug
```

Terminal 2 — Next.js dev server on <http://localhost:3000>, with `/api/*`
rewritten to the Go process:

```bash
just frontend-dev
```

Both sides honor `KRAFTUI_PORT` (default `5200`). Set it before starting if the
auto-selected port is not `5200`:

```bash
KRAFTUI_PORT=5201 just dev
KRAFTUI_PORT=5201 just frontend-dev
```

Other recipes:

| Recipe | What it does |
|---|---|
| `just` / `just list` | List recipes |
| `just deps` | `pnpm install` in `frontend/` |
| `just fmt` | `go fmt ./...` |
| `just openapi-generate` | Regenerate Go models/server and the frontend Fetch SDK |
| `just openapi-check` | Regenerate and fail if generated files drifted |
| `just clean` | Remove `bin/`, Next build output, and the embed dist |

Frontend lint:

```bash
cd frontend && pnpm lint
```

Go tests:

```bash
go test ./...
```

## Configuration

Precedence: command-line flags, then environment variables, then defaults.
Empty environment values are ignored.

| Flag | Environment | Default | Meaning |
|---|---|---|---|
| `--port` | `KRAFTUI_PORT` | `0` | Listen port. `0` tries `5200`–`5204` |
| `--log-type` | `KRAFTUI_LOG_TYPE` | `fancy` | `basic`, `fancy`, or `json` |
| `--log-level` | `KRAFTUI_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `--quiet` | `KRAFTUI_SUPPRESS_LOGS` | `false` | Discard all log output |
| `--kraft-binary` | `KRAFTUI_KRAFT_BINARY` | `kraft` | Path to the Kraft CLI |
| `--kraft-timeout` | `KRAFTUI_KRAFT_TIMEOUT` | `15s` | Timeout for one Kraft command |
| `--version` | — | — | Print version and exit |
| `--help` | — | — | Print help and exit |

```bash
./bin/kraftui --port 5200 --log-level debug --kraft-timeout 30s
```

## HTTP API

All endpoints are read-only. The server validates requests against the
embedded OpenAPI spec and rejects undocumented query parameters.

| Method | Path | Kraft command |
|---|---|---|
| `GET` | `/api/health` | `kraft version` (checked at startup) |
| `GET` | `/api/v1/machines` | `kraft ps --output json` |
| `GET` | `/api/v1/packages` | `kraft pkg list --local --output json` |
| `GET` | `/api/v1/networks` | `kraft net list --output json` |
| `GET` | `/api/v1/volumes` | `kraft vol ls --output json` |

Useful query parameters (see the spec for the full set):

- machines: `all` (default `true`), `platform`, `architecture`, `long`
- packages: `limit` (1–100, default 50), `kind` (`all` / `app` / `lib` / `core`), `platform`, `architecture`
- networks and volumes: `long`

`GET /api/health` returns `{ "status": "ok", "version": "...", "kraftVersion": "..." }`.
Kraft command failures are `502` with `{ "code": "kraft_command_failed", "message": "..." }`.
A missing or incompatible CLI is `503` / `kraft_unavailable`.

Everything that is not `/api/` is the embedded Next.js export. Clean URLs such
as `/machines` map to `machines.html`.

## Console

The dashboard at `/` previews up to four rows per resource and links **View
all** to the full lists:

| Page | Resource |
|---|---|
| `/` | Overview |
| `/machines` | All machines |
| `/networks` | All networks |
| `/volumes` | All volumes |
| `/packages` | Local packages (up to 100) |

The shell follows the system theme. The header shows Kraft and KraftUI
versions when `/api/health` is healthy.

## Layout

```
cmd/kraftui/          CLI entrypoint (go-arg)
config/               Runtime settings and supported Kraft version
log/                  Process-wide slog logger
openapi/openapi.yaml  Source of truth for the HTTP API
oapi-codegen.yaml     Go codegen config
internal/
  api/                Generated models, strict server, embedded spec
  kraft/              Kraft CLI runner and JSON list parsers
  server/             HTTP server, OpenAPI validation, handlers
  ui/                 Embedded Next.js export (internal/ui/dist)
frontend/             Next.js 16 app (React 19, Tailwind v4, shadcn)
```

The frontend Fetch client lives in `frontend/src/lib/api` and is generated by
[`@hey-api/openapi-ts`](https://github.com/hey-api/openapi-ts) from the same
YAML. Do not edit generated files by hand; run `just openapi-generate`.

Frontend stack: Next.js App Router, React Compiler, IBM Plex Sans/Mono,
shadcn (Base UI, Vega style, Tabler icons), `next-themes`. Production builds
set `STATIC_EXPORT=1` so Next emits a static `out/` tree that `just frontend-build`
copies into the Go embed.
