package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cpfiffer/note/internal/api"
	"github.com/cpfiffer/note/internal/util"
)

// PushOptions configures the push operation
type PushOptions struct {
	Force bool
	Dir   string
}

// PushResult contains the results of a push operation
type PushResult struct {
	Created   []string
	Updated   []string
	Conflicts []string
	Unchanged []string
	Skipped   []string
}

// Push uploads local files to Letta
func Push(client *api.Client, state *State, opts PushOptions) (*PushResult, error) {
	result := &PushResult{}

	// Walk the directory to find all .md files
	err := filepath.Walk(opts.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && path != opts.Dir {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-md files
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		// Skip hidden files and conflict files
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		if IsConflictFile(info.Name()) {
			result.Skipped = append(result.Skipped, path)
			return nil
		}

		// Get relative path from sync directory
		relPath, err := filepath.Rel(opts.Dir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		notePath := FileToPath(relPath)

		// Skip note_directory
		if notePath == "/note_directory" {
			result.Skipped = append(result.Skipped, notePath)
			return nil
		}

		// Read local content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		localContent := string(content)
		localHash := util.HashContent(localContent)

		// Check state
		fileState, hasState := state.Files[notePath]

		// Get remote block
		ownerSearch := fmt.Sprintf("owner:%s", state.AgentID)
		blocks, err := client.ListBlocks(ownerSearch)
		if err != nil {
			return fmt.Errorf("failed to list blocks: %w", err)
		}

		var remoteBlock *api.Block
		for i := range blocks {
			if blocks[i].Label == notePath {
				remoteBlock = &blocks[i]
				break
			}
		}

		if remoteBlock != nil {
			remoteHash := util.HashContent(remoteBlock.Value)

			// Check for conflicts
			if hasState && remoteHash != fileState.RemoteHash && localHash != fileState.LocalHash {
				if !opts.Force {
					result.Conflicts = append(result.Conflicts, notePath)
					return nil
				}
			}

			// Check if unchanged
			if localHash == remoteHash {
				result.Unchanged = append(result.Unchanged, notePath)
				state.Files[notePath] = FileState{
					LocalHash:  localHash,
					RemoteHash: remoteHash,
					SyncedAt:   time.Now(),
					BlockID:    remoteBlock.ID,
				}
				return nil
			}

			// Update remote
			_, err := client.UpdateBlock(remoteBlock.ID, localContent)
			if err != nil {
				return fmt.Errorf("failed to update block %s: %w", notePath, err)
			}
			result.Updated = append(result.Updated, notePath)

			state.Files[notePath] = FileState{
				LocalHash:  localHash,
				RemoteHash: localHash, // Remote now matches local
				SyncedAt:   time.Now(),
				BlockID:    remoteBlock.ID,
			}
		} else {
			// Create new block
			description := fmt.Sprintf("owner:%s", state.AgentID)
			block, err := client.CreateBlock(notePath, localContent, description)
			if err != nil {
				return fmt.Errorf("failed to create block %s: %w", notePath, err)
			}
			result.Created = append(result.Created, notePath)

			state.Files[notePath] = FileState{
				LocalHash:  localHash,
				RemoteHash: localHash,
				SyncedAt:   time.Now(),
				BlockID:    block.ID,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
