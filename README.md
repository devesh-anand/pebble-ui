# Pebble UI Viewer

A web-based UI to browse PebbleDB key-value pairs.

## Features

- **Browse keys** with pagination
- **View values** as Raw / JSON / Hex
- **Search modes**
  - **Prefix** (fast; Pebble iterator range scan)
  - **Contains** (slower; scans all keys in-memory)
- **Live DB support** via `--snapshot` (for DBs with a `LOCK` file)

## Install / Build

Build locally:

```bash
go build -o pebble-ui ./cmd/pebble-ui
```

## Usage

Run (read-only):

```bash
./pebble-ui --db /path/to/pebble
```

Run against a **live/locked DB** (recommended for apps like Thanos):

```bash
./pebble-ui --db /path/to/pebble --snapshot
```

Open the UI at `http://localhost:8080`.

### CLI flags

- **`--db`**: Path to Pebble DB directory (required)
- **`--host`**: Bind address (default: `localhost`)
- **`--port`**: HTTP port (default: `8080`)
- **`--snapshot`**: Create a temporary hard-link snapshot to open a locked/live DB
- **`--version`**: Print version and exit

## HTTP API (for debugging)

- **GET `/api/stats`**
- **GET `/api/keys?q=<query>&mode=<prefix|substring>&offset=<n>&limit=<n>`**
- **GET `/api/key/<urlencoded-key>`**

## Troubleshooting

### “0 keys” but you’re sure data exists

- **Use `--snapshot`** if the DB directory contains a `LOCK` file or belongs to a running process.
- If the DB directory contains **no `*.sst` files** and the `*.log` WAL files are tiny, the DB may actually be empty.
- If you see logs like “WAL file … stopped reading at offset: 0; replayed 0 keys”, that can be normal (no pending WAL entries). The data should still appear if SSTs exist.

## Development notes

UI assets are embedded from `cmd/pebble-ui/ui/dist` via `go:embed`.
