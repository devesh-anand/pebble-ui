// Package pebbleui provides an embeddable web UI for browsing PebbleDB databases.
//
// Use Handler to mount the UI at any route in your HTTP server:
//
//	mux.Handle("/pebble-ui/", http.StripPrefix("/pebble-ui", pebbleui.Handler(db)))
//
// Options can be passed to enable authentication, substring search, and more:
//
//	pebbleui.Handler(db, pebbleui.WithBasicAuth("admin", "secret"), pebbleui.WithSubstringSearch())
package pebbleui

import (
	"crypto/subtle"
	"embed"
	"io/fs"
	"net/http"

	"github.com/cockroachdb/pebble"
	"github.com/devesh-anand/pebble-ui/internal/server"
)

//go:embed ui/dist/*
var uiAssets embed.FS

// Option configures the handler returned by Handler.
type Option func(*options)

type options struct {
	dbPath          string
	username        string
	password        string
	substringSearch bool
}

// WithDBPath displays the database path and disk size in the UI stats bar.
func WithDBPath(path string) Option {
	return func(o *options) {
		o.dbPath = path
	}
}

// WithSubstringSearch enables the "Contains" search mode in the UI.
// This scans all keys and can be CPU-intensive on large databases.
func WithSubstringSearch() Option {
	return func(o *options) {
		o.substringSearch = true
	}
}

// WithBasicAuth enables HTTP Basic Authentication for the UI and API.
func WithBasicAuth(username, password string) Option {
	return func(o *options) {
		o.username = username
		o.password = password
	}
}

// Handler returns an http.Handler that serves the Pebble UI and its API.
// Mount it at any path in your router:
//
//	mux.Handle("/pebble-ui/", http.StripPrefix("/pebble-ui", pebbleui.Handler(db)))
//	mux.Handle("/pebble-ui/", http.StripPrefix("/pebble-ui", pebbleui.Handler(db, pebbleui.WithBasicAuth("admin", "secret"))))
func Handler(db *pebble.DB, opts ...Option) http.Handler {
	var cfg options
	for _, o := range opts {
		o(&cfg)
	}

	mux := http.NewServeMux()

	srv := server.NewServer(db, cfg.dbPath, cfg.substringSearch)
	srv.RegisterHandlers(mux)

	uiContent, err := fs.Sub(uiAssets, "ui/dist")
	if err != nil {
		panic("pebbleui: failed to load embedded UI assets: " + err.Error())
	}
	mux.Handle("/", http.FileServer(http.FS(uiContent)))

	if cfg.username != "" && cfg.password != "" {
		return basicAuth(mux, cfg.username, cfg.password)
	}
	return mux
}

func basicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="pebble-ui"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
