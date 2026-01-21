package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/cpfiffer/note/internal/api"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	client    *api.Client
	templates *template.Template
	mux       *http.ServeMux
}

func NewServer() (*Server, error) {
	client, err := api.NewClient("")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	s := &Server{
		client:    client,
		templates: tmpl,
		mux:       http.NewServeMux(),
	}

	s.setupRoutes()
	return s, nil
}

func (s *Server) setupRoutes() {
	// Static files
	s.mux.Handle("/static/", http.FileServer(http.FS(staticFS)))

	// Pages
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/agents", s.handleAgents)
	s.mux.HandleFunc("/notes", s.handleNotes)

	// API endpoints for htmx
	s.mux.HandleFunc("/api/agents", s.handleAPIAgents)
	s.mux.HandleFunc("/api/notes", s.handleAPINotes)
	s.mux.HandleFunc("/api/note/", s.handleAPINote)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.templates.ExecuteTemplate(w, "index.html", nil)
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	agents, err := s.client.ListAgents(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.templates.ExecuteTemplate(w, "agents.html", map[string]interface{}{
		"Agents": agents,
		"Search": search,
	})
}

func (s *Server) handleNotes(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		http.Redirect(w, r, "/agents", http.StatusFound)
		return
	}

	ownerSearch := fmt.Sprintf("owner:%s", agentID)
	blocks, err := s.client.ListBlocks(ownerSearch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Filter to note paths only
	var notes []api.Block
	for _, b := range blocks {
		if strings.HasPrefix(b.Label, "/") && b.Label != "/note_directory" {
			notes = append(notes, b)
		}
	}

	s.templates.ExecuteTemplate(w, "notes.html", map[string]interface{}{
		"AgentID": agentID,
		"Notes":   notes,
	})
}

func (s *Server) handleAPIAgents(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	agents, err := s.client.ListAgents(search)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	s.templates.ExecuteTemplate(w, "agent-list.html", agents)
}

func (s *Server) handleAPINotes(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent")
	if agentID == "" {
		http.Error(w, "agent required", http.StatusBadRequest)
		return
	}

	ownerSearch := fmt.Sprintf("owner:%s", agentID)
	blocks, err := s.client.ListBlocks(ownerSearch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var notes []api.Block
	for _, b := range blocks {
		if strings.HasPrefix(b.Label, "/") && b.Label != "/note_directory" {
			notes = append(notes, b)
		}
	}

	w.Header().Set("Content-Type", "text/html")
	s.templates.ExecuteTemplate(w, "note-list.html", notes)
}

func (s *Server) handleAPINote(w http.ResponseWriter, r *http.Request) {
	// Extract path from URL: /api/note/{agent}/{path...}
	path := strings.TrimPrefix(r.URL.Path, "/api/note/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	agentID := parts[0]
	notePath := "/" + parts[1]

	ownerSearch := fmt.Sprintf("owner:%s", agentID)

	switch r.Method {
	case "GET":
		blocks, err := s.client.ListBlocks(ownerSearch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var note *api.Block
		for _, b := range blocks {
			if b.Label == notePath {
				note = &b
				break
			}
		}

		if note == nil {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(note)

	case "PUT":
		var body struct {
			Value string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		blocks, err := s.client.ListBlocks(ownerSearch)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var blockID string
		for _, b := range blocks {
			if b.Label == notePath {
				blockID = b.ID
				break
			}
		}

		if blockID == "" {
			http.Error(w, "note not found", http.StatusNotFound)
			return
		}

		_, err = s.client.UpdateBlock(blockID, body.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) Start(addr string) error {
	fmt.Printf("Starting server at http://%s\n", addr)
	return http.ListenAndServe(addr, s)
}
