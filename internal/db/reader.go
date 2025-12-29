package db

import (
	"bytes"

	"github.com/cockroachdb/pebble"
)

// Open opens a Pebble database.
func Open(path string, readOnly bool) (*pebble.DB, error) {
	opts := &pebble.Options{
		ReadOnly: readOnly,
	}
	return pebble.Open(path, opts)
}

// ListKeys returns a list of keys with pagination (prefix search).
func ListKeys(db *pebble.DB, prefix string, limit, offset int) ([]string, error) {
	var keys []string
	iter, err := db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(prefix),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	count := 0
	for iter.First(); iter.Valid(); iter.Next() {
		if prefix != "" && !bytes.HasPrefix(iter.Key(), []byte(prefix)) {
			break
		}
		if count >= offset {
			keys = append(keys, string(append([]byte(nil), iter.Key()...)))
			if len(keys) >= limit {
				break
			}
		}
		count++
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return keys, nil
}

// ListKeysSubstring returns keys containing the substring (scans all keys).
func ListKeysSubstring(db *pebble.DB, substring string, limit, offset int) ([]string, int, error) {
	var keys []string
	iter, err := db.NewIter(nil)
	if err != nil {
		return nil, 0, err
	}
	defer iter.Close()

	total := 0
	count := 0
	substringBytes := []byte(substring)
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		if substring == "" || bytes.Contains(key, substringBytes) {
			total++
			if count >= offset && len(keys) < limit {
				keys = append(keys, string(append([]byte(nil), key...)))
			}
			count++
		}
	}
	if err := iter.Error(); err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}

// GetValue returns the value for a specific key.
func GetValue(db *pebble.DB, key string) ([]byte, error) {
	val, closer, err := db.Get([]byte(key))
	if err != nil {
		return nil, err
	}
	defer closer.Close()
	return append([]byte(nil), val...), nil
}

// CountKeys counts the number of keys matching a prefix.
func CountKeys(db *pebble.DB, prefix string) (int, error) {
	iter, err := db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(prefix),
	})
	if err != nil {
		return 0, err
	}
	defer iter.Close()

	count := 0
	for iter.First(); iter.Valid(); iter.Next() {
		if prefix != "" && !bytes.HasPrefix(iter.Key(), []byte(prefix)) {
			break
		}
		count++
	}
	if err := iter.Error(); err != nil {
		return 0, err
	}
	return count, nil
}

// ScanRange returns a map of keys and values in a given range.
func ScanRange(db *pebble.DB, startKey, endKey string, limit int) (map[string][]byte, error) {
	results := make(map[string][]byte)
	iter, err := db.NewIter(&pebble.IterOptions{
		LowerBound: []byte(startKey),
		UpperBound: []byte(endKey),
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	for iter.First(); iter.Valid(); iter.Next() {
		results[string(append([]byte(nil), iter.Key()...))] = append([]byte(nil), iter.Value()...)
		if len(results) >= limit {
			break
		}
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return results, nil
}
