# Pebble UI Viewer

A web-based UI to browse PebbleDB key-value pairs. Use it as a **standalone binary** or **embed it in your Go service** as a package.

## Features

- **Browse keys** with pagination
- **View values** as Raw / JSON / Hex
- **Search modes**
  - **Prefix** (fast; Pebble iterator range scan)
  - **Contains** (opt-in; scans all keys in-memory — disabled by default to protect large DBs)
- **Live DB support** via `--snapshot` (for DBs with a `LOCK` file)
- **HTTP Basic Auth** support
- **Embeddable** — mount the UI at any route in your existing Go service

## Install

### Install with Go (recommended)

```bash
go install github.com/devesh-anand/pebble-ui/cmd/pebble-ui@latest
```

This installs the `pebble-ui` binary into:
- `$(go env GOBIN)` (if set), otherwise
- `$(go env GOPATH)/bin` (commonly `~/go/bin`)

If `pebble-ui` is not found after install, add `~/go/bin` to your `PATH`.

## Standalone Usage

Run (read-only):

```bash
pebble-ui --db /path/to/pebble
```

Run on a different port / bind address:

```bash
pebble-ui --db /path/to/pebble --host 0.0.0.0 --port 9090
```

Run against a **live/locked DB** (recommended for apps like Thanos):

```bash
pebble-ui --db /path/to/pebble --snapshot
```

Run with **authentication**:

```bash
pebble-ui --db /path/to/pebble --username admin --password secret
```

Enable **Contains search** (disabled by default to avoid CPU spikes on large DBs):

```bash
pebble-ui --db /path/to/pebble --substring-search
```

Open the UI at `http://localhost:8080`.

### CLI flags

- **`--db`**: Path to Pebble DB directory (required)
- **`--host`**: Bind address (default: `localhost`)
- **`--port`**: HTTP port (default: `8080`)
- **`--username`**: Basic auth username (requires `--password`)
- **`--password`**: Basic auth password (requires `--username`)
- **`--substring-search`**: Enable the Contains search mode (scans all keys, CPU-intensive on large DBs)
- **`--snapshot`**: Create a temporary hard-link snapshot to open a locked/live DB
- **`--version`**: Print version and exit

## Embed in your Go service

Add the package to your service:

```bash
go get github.com/devesh-anand/pebble-ui@latest
```

Mount the handler on any route using your existing `*pebble.DB`:

```go
import pebbleui "github.com/devesh-anand/pebble-ui"

// Basic usage — no auth, prefix search only
mux.Handle("/pebble-ui/", http.StripPrefix("/pebble-ui", pebbleui.Handler(db)))

// With auth and all options
mux.Handle("/pebble-ui/", http.StripPrefix("/pebble-ui",
    pebbleui.Handler(db,
        pebbleui.WithBasicAuth("admin", "secret"),
        pebbleui.WithSubstringSearch(),
        pebbleui.WithDBPath("/path/to/db"),
    ),
))
```

### Options

| Option | Description |
|--------|-------------|
| `WithBasicAuth(user, pass)` | Enable HTTP Basic Authentication |
| `WithSubstringSearch()` | Enable the Contains search mode (CPU-intensive on large DBs) |
| `WithDBPath(path)` | Show the DB path and disk size in the UI stats bar |

Since you're reusing your service's existing `*pebble.DB` connection, there are no lock conflicts — no need for snapshots.

## Build from source (alternative)

```bash
go build -o pebble-ui ./cmd/pebble-ui
```

Then run:

```bash
./pebble-ui --db /path/to/pebble
```

For a live/locked DB:

```bash
./pebble-ui --db /path/to/pebble --snapshot
```

## HTTP API (for debugging)

- **GET `/api/config`** — feature flags (e.g. `substring_search`)
- **GET `/api/stats`** — key count, DB path, disk size
- **GET `/api/keys?q=<query>&mode=<prefix|substring>&offset=<n>&limit=<n>`** — paginated key listing
- **GET `/api/key/<urlencoded-key>`** — fetch a single key's value

## Troubleshooting

### "0 keys" but you're sure data exists

- **Use `--snapshot`** if the DB directory contains a `LOCK` file or belongs to a running process.
- If the DB directory contains **no `*.sst` files** and the `*.log` WAL files are tiny, the DB may actually be empty.
- If you see logs like "WAL file … stopped reading at offset: 0; replayed 0 keys", that can be normal (no pending WAL entries). The data should still appear if SSTs exist.

## Development notes

UI assets are embedded from `ui/dist` via `go:embed`.
