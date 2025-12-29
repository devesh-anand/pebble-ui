# Usage

## CLI

```bash
./pebble-ui --db /path/to/pebble
```

### Flags

- `--db` (required): Path to Pebble DB directory
- `--host` (default: `localhost`): Bind address
- `--port` (default: `8080`): HTTP port
- `--snapshot` (default: `false`): Create a temporary hard-link snapshot to open a locked/live DB directory
- `--version`: Print version and exit

### When to use `--snapshot`

Use `--snapshot` when the DB is **live** (another process is writing), or you see a `LOCK` file inside the DB dir.

Notes:
- Snapshot uses **hard links**, so the snapshot directory must be on the **same filesystem/partition** as the DB.
- The snapshot skips the `LOCK` file.

## UI search modes

The search dropdown supports:

- **Prefix**: fast range scan; matches keys that start with the query
- **Contains**: slower; scans all keys and filters in memory (can be expensive on large DBs)

## HTTP API

### GET `/api/stats`

Response:

```json
{
  "total_keys": 123,
  "db_path": "/path/to/db",
  "db_size_bytes": 456789
}
```

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


