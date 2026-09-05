package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"mediaplayer/backend/internal/source"
)

type SourceManager interface {
	State() source.SourceState
	Select(ctx context.Context, req source.SelectRequest) error
	SetVolume(volume int) error
	Subscribe() (chan source.SourceState, func())
	GetAdapters() map[string]source.Adapter
	GetCurrentAdapter() source.Adapter
}

type Server struct {
	httpServer *http.Server
	manager    SourceManager
	upgrader   websocket.Upgrader
	sources    []string
	drStations []string
}

func NewServer(listenAddr string, manager SourceManager, frontendFS http.FileSystem) *Server {
	s := &Server{
		manager: manager,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		sources: []string{"dr-radio", "plexamp", "bluetooth"},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/state", s.handleState)
	mux.HandleFunc("/api/source/list", s.listSources)
	mux.HandleFunc("/api/source/select", s.handleSelect)
	mux.HandleFunc("/api/player/volume", s.handleVolume)
	mux.HandleFunc("/api/stations", s.fetchStations)
	mux.HandleFunc("/ws", s.handleWebSocket)

	// 3. SvelteKit Frontend og SPA Fallback handler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Hvis anmodningen starter med /api/ men ikke matchede endpoints ovenfor, giv en 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		cleanedPath := filepath.Clean(r.URL.Path)

		// Tjek om filen eksisterer i SvelteKit-buildet (f.eks. /favicon.png eller js-filer)
		file, err := frontendFS.Open(r.URL.Path)
		if err == nil {
			if err == nil {
				file.Close()

				// Fiks manglende styling: Gennemtving korrekte MIME-types baseret på filendelse
				ext := filepath.Ext(cleanedPath)
				switch ext {
				case ".css":
					w.Header().Set("Content-Type", "text/css; charset=utf-8")
				case ".js":
					w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
				case ".svg":
					w.Header().Set("Content-Type", "image/svg+xml")
				case ".png":
					w.Header().Set("Content-Type", "image/png")
				case ".json":
					w.Header().Set("Content-Type", "application/json")
				}

				// Server den fundne statiske fil (CSS, JS, osv.)
				http.FileServer(frontendFS).ServeHTTP(w, r)
				return
			}
		}

		// Hvis filen ikke findes, server index.html (SPA Fallback til SvelteKit routeren)
		indexFile, err := frontendFS.Open("index.html")
		if err != nil {
			http.Error(w, "Frontend index.html not found", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		stat, err := indexFile.Stat()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
	})

	s.httpServer = &http.Server{
		Addr:              listenAddr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

func (s *Server) serveIndexHTML(w http.ResponseWriter, r *http.Request, frontendFS http.FileSystem) {
	indexFile, err := frontendFS.Open("index.html")
	if err != nil {
		http.Error(w, "Frontend index.html not found", http.StatusInternalServerError)
		return
	}
	defer indexFile.Close()

	stat, err := indexFile.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	log.Printf("Server started!")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, s.manager.State())
}

func (s *Server) listSources(w http.ResponseWriter, r *http.Request) {
	adapters := s.manager.GetAdapters()
	var sourcesList []any

	for name, adapter := range adapters {
		// Tjek om adapteren har stationer (f.eks. ved at tjekke om GetStations returnerer noget)
		stations := adapter.GetStations()

		if len(stations) > 0 {
			sourcesList = append(sourcesList, map[string]any{
				"name":     name,
				"stations": stations,
			})
		} else {
			// Hvis der ikke er stationer, tilføjer vi bare navnet som en streng (eller et simpelt objekt)
			sourcesList = append(sourcesList, name)
		}
	}

	responsePayload := map[string]any{
		"sources": sourcesList,
	}

	writeJSON(w, http.StatusOK, responsePayload)
}

func (s *Server) fetchStations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stations := s.manager.GetCurrentAdapter().GetStations()
	writeJSON(w, http.StatusOK, stations)
}

func (s *Server) handleSelect(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req source.SelectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	fmt.Println(req)
	if req.Source == "" {
		http.Error(w, "source is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if err := s.manager.Select(ctx, req); err != nil {
		log.Printf("FEJL i s.manager.Select for kilde '%s': %v", req.Source, err)
		http.Error(w, err.Error(), statusForSelectionError(err))
		return
	}
	writeJSON(w, http.StatusOK, s.manager.State())
}

type volumeRequest struct {
	Volume int `json:"volume"`
}

func (s *Server) handleVolume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req volumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Volume < 0 || req.Volume > 100 {
		http.Error(w, "volume must be between 0 and 100", http.StatusBadRequest)
		return
	}

	if err := s.manager.SetVolume(req.Volume); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, s.manager.State())
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	stateCh, unsubscribe := s.manager.Subscribe()
	defer unsubscribe()

	for state := range stateCh {
		if err := conn.WriteJSON(state); err != nil {
			return
		}
	}
}

func statusForSelectionError(err error) int {
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
