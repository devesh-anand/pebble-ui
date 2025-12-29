package server

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/pebble"
	"github.com/devesh-anand/pebble-ui/internal/db"
	"github.com/devesh-anand/pebble-ui/internal/models"
)

type Server struct {
	DB     *pebble.DB
	DBPath string
}

func NewServer(database *pebble.DB, path string) *Server {
	return &Server{
		DB:     database,
		DBPath: path,
	}
}

func (s *Server) RegisterHandlers(mux *http.ServeMux) {
	mux.Handle("/api/keys", s.loggingMiddleware(http.HandlerFunc(s.handleListKeys)))
	mux.Handle("/api/key/", s.loggingMiddleware(http.HandlerFunc(s.handleGetKey)))
	mux.Handle("/api/stats", s.loggingMiddleware(http.HandlerFunc(s.handleStats)))
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

	if mode == "substring" {
		keys, total, err = db.ListKeysSubstring(s.DB, query, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		keys, err = db.ListKeys(s.DB, query, limit, offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		total, err = db.CountKeys(s.DB, query)
		if err != nil {
			log.Printf("Error counting keys: %v", err)
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

	val, err := db.GetValue(s.DB, key)
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
	total, err := db.CountKeys(s.DB, "")
	if err != nil {
		log.Printf("Error counting keys for stats: %v", err)
	}

	dbSizeBytes, err := dirSizeBytes(s.DBPath)
	if err != nil {
		log.Printf("Error calculating db size: %v", err)
	}
	resp := models.StatsResponse{
		TotalKeys:   total,
		DBPath:      s.DBPath,
		DBSizeBytes: dbSizeBytes,
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
