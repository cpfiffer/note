package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cpfiffer/note/internal/api"
	"github.com/cpfiffer/note/internal/util"
)

// PullOptions configures the pull operation
type PullOptions struct {
	Force bool
	Dir   string
}

// PullResult contains the results of a pull operation
type PullResult struct {
	Created   []string
	Updated   []string
	Conflicts []string
	Unchanged []string
	Skipped   []string
}

// Pull downloads notes from Letta to local files
func Pull(client *api.Client, state *State, opts PullOptions) (*PullResult, error) {
	result := &PullResult{}

	// Pattern to filter out legacy UUID paths and note_directory
	uuidPattern := regexp.MustCompile(`/\[?agent-[a-f0-9-]+\]?/`)

	// Fetch all blocks for this agent
	ownerSearch := fmt.Sprintf("owner:%s", state.AgentID)
	blocks, err := client.ListBlocks(ownerSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocks: %w", err)
	}

	for _, block := range blocks {
		// Skip non-path labels
		if !strings.HasPrefix(block.Label, "/") {
			continue
		}

		// Skip note_directory (auto-generated)
		if block.Label == "/note_directory" {
			result.Skipped = append(result.Skipped, block.Label)
			continue
		}

		// Skip legacy UUID paths
		if uuidPattern.MatchString(block.Label) {
			result.Skipped = append(result.Skipped, block.Label)
			continue
		}

		localPath := filepath.Join(opts.Dir, PathToFile(block.Label))
		remoteHash := util.HashContent(block.Value)

		// Check if file exists locally
		localContent, err := os.ReadFile(localPath)
		localExists := err == nil

		if localExists {
			localHash := util.HashContent(string(localContent))
			fileState, hasState := state.Files[block.Label]

			if hasState && localHash != fileState.LocalHash && remoteHash != fileState.RemoteHash {
				// Both changed - conflict!
				if !opts.Force {
					// Write conflict file with remote content
					conflictPath := filepath.Join(opts.Dir, ConflictFileName(block.Label))
					if err := writeFile(conflictPath, block.Value); err != nil {
						return nil, fmt.Errorf("failed to write conflict file: %w", err)
					}
					result.Conflicts = append(result.Conflicts, block.Label)
					continue
				}
			}

			if remoteHash == localHash {
				// No changes
				result.Unchanged = append(result.Unchanged, block.Label)
				// Update state in case it was missing
				state.Files[block.Label] = FileState{
					LocalHash:  localHash,
					RemoteHash: remoteHash,
					SyncedAt:   time.Now(),
					BlockID:    block.ID,
				}
				continue
			}

			// Remote changed (or force mode)
			if err := writeFile(localPath, block.Value); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", localPath, err)
			}
			result.Updated = append(result.Updated, block.Label)
		} else {
			// New file
			if err := writeFile(localPath, block.Value); err != nil {
				return nil, fmt.Errorf("failed to write file %s: %w", localPath, err)
			}
			result.Created = append(result.Created, block.Label)
		}

		// Update state
		state.Files[block.Label] = FileState{
			LocalHash:  remoteHash, // Local now matches remote
			RemoteHash: remoteHash,
			SyncedAt:   time.Now(),
			BlockID:    block.ID,
		}
	}

	return result, nil
}

// writeFile writes content to a file, creating directories as needed
func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}
