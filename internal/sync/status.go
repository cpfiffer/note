package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cpfiffer/note/internal/api"
	"github.com/cpfiffer/note/internal/util"
)

// StatusResult contains the sync status
type StatusResult struct {
	ModifiedLocally  []string
	ModifiedRemotely []string
	Conflicts        []string
	UntrackedLocal   []string
	UntrackedRemote  []string
	Synced           []string
}

// Status computes the sync status
func Status(client *api.Client, state *State, dir string) (*StatusResult, error) {
	result := &StatusResult{}

	// Pattern to filter out legacy UUID paths
	uuidPattern := regexp.MustCompile(`/\[?agent-[a-f0-9-]+\]?/`)

	// Get remote blocks
	ownerSearch := fmt.Sprintf("owner:%s", state.AgentID)
	blocks, err := client.ListBlocks(ownerSearch)
	if err != nil {
		return nil, fmt.Errorf("failed to list blocks: %w", err)
	}

	// Build map of remote blocks
	remoteBlocks := make(map[string]*api.Block)
	for i := range blocks {
		block := &blocks[i]
		if !strings.HasPrefix(block.Label, "/") {
			continue
		}
		if block.Label == "/note_directory" {
			continue
		}
		if uuidPattern.MatchString(block.Label) {
			continue
		}
		remoteBlocks[block.Label] = block
	}

	// Track which paths we've seen
	seenPaths := make(map[string]bool)

	// Check local files
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") || IsConflictFile(info.Name()) {
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		notePath := FileToPath(relPath)

		if notePath == "/note_directory" {
			return nil
		}

		seenPaths[notePath] = true

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		localHash := util.HashContent(string(content))

		fileState, hasState := state.Files[notePath]
		remoteBlock, hasRemote := remoteBlocks[notePath]

		if !hasRemote {
			// Only exists locally
			result.UntrackedLocal = append(result.UntrackedLocal, notePath)
			return nil
		}

		remoteHash := util.HashContent(remoteBlock.Value)

		if !hasState {
			// Not tracked but exists both places
			if localHash != remoteHash {
				result.Conflicts = append(result.Conflicts, notePath)
			} else {
				result.Synced = append(result.Synced, notePath)
			}
			return nil
		}

		localChanged := localHash != fileState.LocalHash
		remoteChanged := remoteHash != fileState.RemoteHash

		if localChanged && remoteChanged {
			result.Conflicts = append(result.Conflicts, notePath)
		} else if localChanged {
			result.ModifiedLocally = append(result.ModifiedLocally, notePath)
		} else if remoteChanged {
			result.ModifiedRemotely = append(result.ModifiedRemotely, notePath)
		} else {
			result.Synced = append(result.Synced, notePath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Check for remote-only files
	for path := range remoteBlocks {
		if !seenPaths[path] {
			result.UntrackedRemote = append(result.UntrackedRemote, path)
		}
	}

	return result, nil
}
