package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/devesh-anand/pebble-ui/internal/db"
	"github.com/devesh-anand/pebble-ui/internal/server"
)

//go:embed ui/dist/*
var uiAssets embed.FS

var (
	version = "dev"
)

func main() {
	dbPath := flag.String("db", "", "Path to PebbleDB directory (required)")
	port := flag.Int("port", 8080, "HTTP server port")
	host := flag.String("host", "localhost", "Bind address")
	showVersion := flag.Bool("version", false, "Show version and exit")
	snapshot := flag.Bool("snapshot", false, "Create a temporary hard-link snapshot to open a locked/live DB")
	flag.Parse()

	if *showVersion {
		fmt.Printf("pebble-ui version %s\n", version)
		os.Exit(0)
	}

	if *dbPath == "" {
		fmt.Println("Error: --db flag is required")
		flag.Usage()
		os.Exit(1)
	}

	finalDbPath := *dbPath
	if *snapshot {
		tmpDir, err := createSnapshot(*dbPath)
		if err != nil {
			log.Fatalf("Failed to create snapshot: %v", err)
		}
		finalDbPath = tmpDir
		defer os.RemoveAll(tmpDir)
		fmt.Printf("Created temporary snapshot at %s\n", tmpDir)
	}

	database, err := db.Open(finalDbPath, !*snapshot)
	if err != nil {
		log.Fatalf("Failed to open PebbleDB: %v", err)
	}
	defer database.Close()

	srv := server.NewServer(database, *dbPath)
	mux := http.NewServeMux()
	srv.RegisterHandlers(mux)

	// Serve UI assets
	uiContent, err := fs.Sub(uiAssets, "ui/dist")
	if err != nil {
		log.Fatalf("Failed to sub ui assets: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(uiContent)))

	addr := fmt.Sprintf("%s:%d", *host, *port)
	httpSrv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		fmt.Printf("Pebble UI Viewer starting on http://%s\n", addr)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	if err := httpSrv.Close(); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}

func createSnapshot(src string) (string, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "pebble-snapshot-*")
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue // Pebble usually has a flat structure, skip subdirs for simplicity
		}
		if entry.Name() == "LOCK" {
			continue // Do not link the LOCK file
		}

		oldPath := filepath.Join(src, entry.Name())
		newPath := filepath.Join(tmpDir, entry.Name())

		if err := os.Link(oldPath, newPath); err != nil {
			// If hard link fails (e.g. cross-device), we'd need to copy, 
			// but for large DBs on same device, Link is what we want.
			os.RemoveAll(tmpDir)
			return "", fmt.Errorf("failed to link %s: %v (hint: snapshot must be on same partition as DB)", entry.Name(), err)
		}
	}

	return tmpDir, nil
}
