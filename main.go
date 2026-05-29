package main

import (
	"bytes"
	"embed"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/axshb/guardian/internal/api"
	"github.com/axshb/guardian/internal/auth"
	"github.com/axshb/guardian/internal/config"
	"github.com/axshb/guardian/internal/store"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

//go:embed dashboard/dist/*
var spaFiles embed.FS

var spaFS fs.FS

func init() {
	var err error
	spaFS, err = fs.Sub(spaFiles, "dashboard/dist")
	if err != nil {
		log.Fatalf("spa fs: %v", err)
	}
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataPath, 0755); err != nil {
		log.Fatalf("mkdir data: %v", err)
	}

	auth.SetSecretPath(filepath.Join(cfg.DataPath, "session.secret"))

	// Open SQLite store
	s, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer s.Close()

	refs := store.NewReferrerManager(s.DB())

	assetHandler := api.NewHandler(s, refs)
	authHandler := auth.NewHandler(cfg.DataPath)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.StripSlashes)

	// Mount auth routes (no auth required for login/setup/config)
	authHandler.RegisterRoutes(r)

	// Mount asset routes with admin middleware
	assetHandler.RegisterRoutes(r, auth.RequireAdmin)

	// Block all crawlers
	r.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("User-agent: *\nDisallow: /\n"))
	})

	// Serve SPA for all other routes
	r.Get("/*", spaHandler)

	addr := ":" + cfg.Port
	log.Printf("guardian listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// spaHandler serves the Vue SPA for any non-API route.
func spaHandler(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/")
	if filePath == "" {
		filePath = "index.html"
	}

	// Try to serve the exact file; http.ServeContent detects Content-Type from filename.
	data, err := fs.ReadFile(spaFS, filePath)
	if err == nil {
		http.ServeContent(w, r, path.Base(filePath), time.Time{}, bytes.NewReader(data))
		return
	}

	// Fallback to index.html for SPA client-side routing
	index, err := spaFS.Open("index.html")
	if err != nil {
		http.Error(w, "SPA not found", http.StatusInternalServerError)
		return
	}
	defer index.Close()
	indexData, _ := io.ReadAll(index)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", time.Time{}, bytes.NewReader(indexData))
}
