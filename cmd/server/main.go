package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/qurt-quiz/qurt-quiz/internal/hub"
	"github.com/qurt-quiz/qurt-quiz/internal/pack"
	memstore "github.com/qurt-quiz/qurt-quiz/internal/store/memory"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "listen address")
	packDir := flag.String("packs", envOr("PACK_DIR", "packs"), "directory containing quiz pack JSON files")
	staticDir := flag.String("static", envOr("STATIC_DIR", ""), "directory of built frontend files (empty = skip)")
	flag.Parse()

	absPackDir, err := filepath.Abs(*packDir)
	if err != nil {
		log.Fatalf("resolving pack dir: %v", err)
	}
	if _, err := os.Stat(absPackDir); err != nil {
		log.Fatalf("pack dir %q not found: %v", absPackDir, err)
	}

	store := memstore.New()
	h := hub.New(store, absPackDir)

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", h.ServeWS)

	mux.HandleFunc("/api/packs", func(w http.ResponseWriter, r *http.Request) {
		packs, err := pack.ListAvailable(absPackDir)
		if err != nil {
			http.Error(w, "failed to list packs", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"packs": packs})
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	if *staticDir != "" {
		absStatic, _ := filepath.Abs(*staticDir)
		fs := http.FileServer(http.Dir(absStatic))
		mux.Handle("/", spaHandler{fs: fs, root: absStatic})
	}

	log.Printf("qurt-quiz server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, corsMiddleware(mux)); err != nil {
		log.Fatal(err)
	}
}

// corsMiddleware adds permissive CORS headers for local development.
// In production, restrict Access-Control-Allow-Origin to the actual frontend origin.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// spaHandler serves static files and falls back to index.html for SPA routing.
type spaHandler struct {
	fs   http.Handler
	root string
}

func (s spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(s.root, r.URL.Path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(s.root, "index.html"))
		return
	}
	s.fs.ServeHTTP(w, r)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
