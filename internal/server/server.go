package server

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/devesh-anand/pebble-ui/internal/db"
	"github.com/devesh-anand/pebble-ui/internal/models"
)

type Server struct {
	pebbleDB        *pebble.DB
	dbPath          string
	substringSearch bool

	// cached total key count for stats
	statsMu       sync.Mutex
	cachedTotal   int
	cachedTotalAt time.Time
}

const statsCacheTTL = 30 * time.Second

func NewServer(database *pebble.DB, path string, substringSearch bool) *Server {
	return &Server{
		pebbleDB:        database,
		dbPath:          path,
		substringSearch: substringSearch,
	}
}

func (s *Server) RegisterHandlers(mux *http.ServeMux) {
	mux.Handle("/api/keys", s.loggingMiddleware(http.HandlerFunc(s.handleListKeys)))
	mux.Handle("/api/key/", s.loggingMiddleware(http.HandlerFunc(s.handleGetKey)))
	mux.Handle("/api/stats", s.loggingMiddleware(http.HandlerFunc(s.handleStats)))
	mux.Handle("/api/config", s.loggingMiddleware(http.HandlerFunc(s.handleConfig)))
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleListKeys(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	mode := r.URL.Query().Get("mode") // "prefix" or "substring"
	if mode == "" {
		mode = "prefix"
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var keys []string
	var total int
	var err error

	if mode == "substring" && !s.substringSearch {
		http.Error(w, "substring search is disabled", http.StatusForbidden)
		return
	}

	if mode == "substring" {
		keys, total, err = db.ListKeysSubstring(s.pebbleDB, query, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		keys, total, err = db.ListKeys(s.pebbleDB, query, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	resp := models.KeyListResponse{
		Keys:   keys,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGetKey(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[len("/api/key/"):]
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}

	val, err := db.GetValue(s.pebbleDB, key)
	if err != nil {
		if err == pebble.ErrNotFound {
			http.Error(w, "key not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	resp := models.ValueResponse{
		Key:      key,
		Value:    string(val),
		ValueHex: hex.EncodeToString(val),
		Size:     len(val),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	total := s.getCachedKeyCount()

	var dbSizeBytes int64
	if s.dbPath != "" {
		var err error
		dbSizeBytes, err = dirSizeBytes(s.dbPath)
		if err != nil {
			log.Printf("Error calculating db size: %v", err)
		}
	}
	resp := models.StatsResponse{
		TotalKeys:   total,
		DBPath:      s.dbPath,
		DBSizeBytes: dbSizeBytes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) getCachedKeyCount() int {
	s.statsMu.Lock()
	defer s.statsMu.Unlock()

	if time.Since(s.cachedTotalAt) < statsCacheTTL {
		return s.cachedTotal
	}

	total, err := db.CountKeys(s.pebbleDB, "")
	if err != nil {
		log.Printf("Error counting keys for stats: %v", err)
		return s.cachedTotal
	}
	s.cachedTotal = total
	s.cachedTotalAt = time.Now()
	return total
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"substring_search": s.substringSearch,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func dirSizeBytes(dir string) (int64, error) {
	var size int64
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		size += info.Size()
		return nil
	})
	return size, err
}
