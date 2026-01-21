package main

import (
	"fmt"
	"os"

	"github.com/cpfiffer/note/internal/api"
	"github.com/cpfiffer/note/internal/sync"
	"github.com/cpfiffer/note/internal/web"
	"github.com/spf13/cobra"
)

var (
	forceFlag  bool
	dirFlag    string
	nameFlag   string
	portFlag   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "note-sync",
		Short: "Sync Letta notes with local markdown files",
		Long:  `A CLI tool to bidirectionally sync Letta memory blocks with a local folder of markdown files.`,
	}

	// Init command
	initCmd := &cobra.Command{
		Use:   "init <agent_id>",
		Short: "Initialize a sync directory for an agent",
		Args:  cobra.ExactArgs(1),
		RunE:  runInit,
	}
	initCmd.Flags().StringVar(&dirFlag, "dir", ".", "Directory to initialize")
	rootCmd.AddCommand(initCmd)

	// Pull command
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Download notes from Letta to local files",
		RunE:  runPull,
	}
	pullCmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite local changes without conflict check")
	rootCmd.AddCommand(pullCmd)

	// Push command
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Upload local files to Letta",
		RunE:  runPush,
	}
	pushCmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite remote changes without conflict check")
	rootCmd.AddCommand(pushCmd)

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		RunE:  runStatus,
	}
	rootCmd.AddCommand(statusCmd)

	// Agents command
	agentsCmd := &cobra.Command{
		Use:   "agents",
		Short: "List available agents",
		RunE:  runAgents,
	}
	agentsCmd.Flags().StringVar(&nameFlag, "name", "", "Search agents by name")
	rootCmd.AddCommand(agentsCmd)

	// Serve command (web UI)
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start web UI for browsing and editing notes",
		RunE:  runServe,
	}
	serveCmd.Flags().StringVar(&portFlag, "port", "8080", "Port to listen on")
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	agentID := args[0]

	// Check if already initialized
	if _, err := sync.LoadState(dirFlag); err == nil {
		return fmt.Errorf("directory already initialized. Delete %s to reinitialize", sync.StateFileName)
	}

	// Verify API key works
	client, err := api.NewClient("")
	if err != nil {
		return err
	}

	// Try to list blocks to verify agent exists
	ownerSearch := fmt.Sprintf("owner:%s", agentID)
	_, err = client.ListBlocks(ownerSearch)
	if err != nil {
		return fmt.Errorf("failed to connect to Letta API: %w", err)
	}

	state, err := sync.InitState(dirFlag, agentID, "")
	if err != nil {
		return err
	}

	fmt.Printf("Initialized note-sync for agent %s\n", state.AgentID)
	fmt.Printf("State file: %s/%s\n", dirFlag, sync.StateFileName)
	fmt.Println("\nNext steps:")
	fmt.Println("  note-sync pull   # Download notes from Letta")
	fmt.Println("  note-sync push   # Upload local changes")
	fmt.Println("  note-sync status # Check sync status")

	return nil
}

func runPull(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	state, err := sync.LoadState(dir)
	if err != nil {
		return err
	}

	client, err := api.NewClient(state.LettaBaseURL)
	if err != nil {
		return err
	}

	opts := sync.PullOptions{
		Force: forceFlag,
		Dir:   dir,
	}

	result, err := sync.Pull(client, state, opts)
	if err != nil {
		return err
	}

	// Save updated state
	if err := state.SaveState(dir); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Print results
	if len(result.Created) > 0 {
		fmt.Println("Created:")
		for _, p := range result.Created {
			fmt.Printf("  + %s\n", p)
		}
	}
	if len(result.Updated) > 0 {
		fmt.Println("Updated:")
		for _, p := range result.Updated {
			fmt.Printf("  ~ %s\n", p)
		}
	}
	if len(result.Conflicts) > 0 {
		fmt.Println("Conflicts (see .conflict.md files):")
		for _, p := range result.Conflicts {
			fmt.Printf("  ! %s\n", p)
		}
	}
	if len(result.Created) == 0 && len(result.Updated) == 0 && len(result.Conflicts) == 0 {
		fmt.Println("Already up to date.")
	}

	return nil
}

func runPush(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	state, err := sync.LoadState(dir)
	if err != nil {
		return err
	}

	client, err := api.NewClient(state.LettaBaseURL)
	if err != nil {
		return err
	}

	opts := sync.PushOptions{
		Force: forceFlag,
		Dir:   dir,
	}

	result, err := sync.Push(client, state, opts)
	if err != nil {
		return err
	}

	// Save updated state
	if err := state.SaveState(dir); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Print results
	if len(result.Created) > 0 {
		fmt.Println("Created:")
		for _, p := range result.Created {
			fmt.Printf("  + %s\n", p)
		}
	}
	if len(result.Updated) > 0 {
		fmt.Println("Updated:")
		for _, p := range result.Updated {
			fmt.Printf("  ~ %s\n", p)
		}
	}
	if len(result.Conflicts) > 0 {
		fmt.Println("Conflicts (pull first to see remote changes):")
		for _, p := range result.Conflicts {
			fmt.Printf("  ! %s\n", p)
		}
	}
	if len(result.Created) == 0 && len(result.Updated) == 0 && len(result.Conflicts) == 0 {
		fmt.Println("Nothing to push.")
	}

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	state, err := sync.LoadState(dir)
	if err != nil {
		return err
	}

	client, err := api.NewClient(state.LettaBaseURL)
	if err != nil {
		return err
	}

	result, err := sync.Status(client, state, dir)
	if err != nil {
		return err
	}

	hasChanges := false

	if len(result.ModifiedLocally) > 0 {
		hasChanges = true
		fmt.Println("Modified locally (push to upload):")
		for _, p := range result.ModifiedLocally {
			fmt.Printf("  M %s\n", p)
		}
	}

	if len(result.ModifiedRemotely) > 0 {
		hasChanges = true
		fmt.Println("Modified remotely (pull to download):")
		for _, p := range result.ModifiedRemotely {
			fmt.Printf("  M %s\n", p)
		}
	}

	if len(result.Conflicts) > 0 {
		hasChanges = true
		fmt.Println("Conflicts (both changed):")
		for _, p := range result.Conflicts {
			fmt.Printf("  ! %s\n", p)
		}
	}

	if len(result.UntrackedLocal) > 0 {
		hasChanges = true
		fmt.Println("Untracked local (push to create):")
		for _, p := range result.UntrackedLocal {
			fmt.Printf("  ? %s\n", p)
		}
	}

	if len(result.UntrackedRemote) > 0 {
		hasChanges = true
		fmt.Println("Untracked remote (pull to download):")
		for _, p := range result.UntrackedRemote {
			fmt.Printf("  ? %s\n", p)
		}
	}

	if !hasChanges {
		fmt.Printf("Everything in sync. (%d files)\n", len(result.Synced))
	}

	return nil
}

func runAgents(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient("")
	if err != nil {
		return err
	}

	agents, err := client.ListAgents(nameFlag)
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	if len(agents) == 0 {
		if nameFlag != "" {
			fmt.Printf("No agents found matching name '%s'\n", nameFlag)
		} else {
			fmt.Println("No agents found")
		}
		return nil
	}

	fmt.Printf("%-40s  %s\n", "ID", "NAME")
	fmt.Printf("%-40s  %s\n", "----", "----")
	for _, agent := range agents {
		fmt.Printf("%-40s  %s\n", agent.ID, agent.Name)
	}

	return nil
}

func runServe(cmd *cobra.Command, args []string) error {
	server, err := web.NewServer()
	if err != nil {
		return err
	}

	addr := "localhost:" + portFlag
	return server.Start(addr)
}
