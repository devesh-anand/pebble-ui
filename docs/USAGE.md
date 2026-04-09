# Usage

## CLI

```bash
pebble-ui --db /path/to/pebble
```

Custom host/port:

```bash
pebble-ui --db /path/to/pebble --host 0.0.0.0 --port 9090
```

With authentication:

```bash
pebble-ui --db /path/to/pebble --username admin --password secret
```

Enable Contains search (disabled by default):

```bash
pebble-ui --db /path/to/pebble --substring-search
```

### Flags

- `--db` (required): Path to Pebble DB directory
- `--host` (default: `localhost`): Bind address
- `--port` (default: `8080`): HTTP port
- `--username`: Basic auth username (requires `--password`)
- `--password`: Basic auth password (requires `--username`)
- `--substring-search` (default: `false`): Enable the Contains search mode (scans all keys, CPU-intensive on large DBs)
- `--snapshot` (default: `false`): Create a temporary hard-link snapshot to open a locked/live DB directory
- `--version`: Print version and exit

### When to use `--snapshot`

Use `--snapshot` when the DB is **live** (another process is writing), or you see a `LOCK` file inside the DB dir.

Notes:
- Snapshot uses **hard links**, so the snapshot directory must be on the **same filesystem/partition** as the DB.
- The snapshot skips the `LOCK` file.

## Embed in a Go service

Add the package:

```bash
go get github.com/devesh-anand/pebble-ui@latest
```

Mount at any route using your existing `*pebble.DB`:

```go
import pebbleui "github.com/devesh-anand/pebble-ui"

// Basic — prefix search only, no auth
mux.Handle("/pebble-ui/", http.StripPrefix("/pebble-ui", pebbleui.Handler(db)))

// With all options
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

No snapshots needed when embedding — you're reusing your service's existing DB connection, so there are no lock conflicts.

## UI search modes

The search dropdown supports:

- **Prefix**: fast range scan; matches keys that start with the query
- **Contains** (opt-in): slower; scans all keys and filters in memory (can be expensive on large DBs). Only available when `--substring-search` is passed (CLI) or `WithSubstringSearch()` is used (library).

Search is triggered by pressing the **Refresh** button.

## HTTP API

### GET `/api/config`

Response:

```json
{
  "substring_search": false
}
```

### GET `/api/stats`

Response:

```json
{
  "total_keys": 123,
  "db_path": "/path/to/db",
  "db_size_bytes": 456789
}
```

Note: total key count is cached for 30 seconds.

### GET `/api/keys`

Query params:
- `q` (string): search query
- `mode` (`prefix` | `substring`): search mode (default: `prefix`)
- `offset` (int): pagination offset (default: `0`)
- `limit` (int): max keys returned (default: `50`)

Example:

```text
/api/keys?q=group-iterations/&mode=prefix&offset=0&limit=50
```

Response:

```json
{
  "keys": ["k1", "k2"],
  "total": 2,
  "offset": 0,
  "limit": 50
}
```

Note: `mode=substring` returns `403 Forbidden` if substring search is not enabled.

### GET `/api/key/<key>`

The key must be URL-encoded.

Response:

```json
{
  "key": "some-key",
  "value": "raw-value-string",
  "value_hex": "686578",
  "size": 3
}
```
