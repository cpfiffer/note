package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const StateFileName = ".note-sync.json"

// FileState tracks the sync state of a single file
type FileState struct {
	LocalHash  string    `json:"local_hash"`
	RemoteHash string    `json:"remote_hash"`
	SyncedAt   time.Time `json:"synced_at"`
	BlockID    string    `json:"block_id,omitempty"`
}

// State represents the overall sync state
type State struct {
	AgentID     string               `json:"agent_id"`
	LettaBaseURL string              `json:"letta_base_url"`
	Files       map[string]FileState `json:"files"`
}

// LoadState loads the sync state from the current directory
func LoadState(dir string) (*State, error) {
	statePath := filepath.Join(dir, StateFileName)
	
	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not a note-sync directory (no %s found). Run 'note-sync init' first", StateFileName)
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	if state.Files == nil {
		state.Files = make(map[string]FileState)
	}

	return &state, nil
}

// SaveState saves the sync state to disk
func (s *State) SaveState(dir string) error {
	statePath := filepath.Join(dir, StateFileName)

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(statePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// InitState creates a new state file
func InitState(dir, agentID, baseURL string) (*State, error) {
	if baseURL == "" {
		baseURL = "https://api.letta.com"
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	state := &State{
		AgentID:     agentID,
		LettaBaseURL: baseURL,
		Files:       make(map[string]FileState),
	}

	if err := state.SaveState(dir); err != nil {
		return nil, err
	}

	return state, nil
}

// PathToFile converts a Letta note path to a local file path
// /projects/webapp -> projects/webapp.md
func PathToFile(notePath string) string {
	// Remove leading slash
	if len(notePath) > 0 && notePath[0] == '/' {
		notePath = notePath[1:]
	}
	return notePath + ".md"
}

// FileToPath converts a local file path to a Letta note path
// projects/webapp.md -> /projects/webapp
func FileToPath(filePath string) string {
	// Remove .md extension
	if len(filePath) > 3 && filePath[len(filePath)-3:] == ".md" {
		filePath = filePath[:len(filePath)-3]
	}
	return "/" + filePath
}

// IsConflictFile checks if a file is a conflict file
func IsConflictFile(filename string) bool {
	return len(filename) > 12 && filename[len(filename)-12:] == ".conflict.md"
}

// ConflictFileName returns the conflict file name for a path
func ConflictFileName(notePath string) string {
	// /projects/webapp -> projects/webapp.conflict.md
	if len(notePath) > 0 && notePath[0] == '/' {
		notePath = notePath[1:]
	}
	return notePath + ".conflict.md"
}
