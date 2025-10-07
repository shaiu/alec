package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/shaiu/alec/pkg/services"
	"github.com/shaiu/alec/pkg/tui"
)

func createTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Start interactive Terminal UI",
		Long: `Start the interactive Terminal User Interface for browsing and executing scripts.

The TUI provides a visual interface with:
- Script tree navigation in the sidebar
- Script details and execution output in the main area
- Responsive layout that adapts to terminal size
- Real-time output streaming during script execution

Use arrow keys to navigate, Enter to execute scripts, Tab to switch panels,
and 'q' to quit.`,
		RunE: runTUI,
	}

	return cmd
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Create service registry
	registry, err := services.NewServiceRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// Create TUI manager directly
	tuiManager := tui.NewTUIManager(registry)

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	// Start the TUI
	fmt.Println("Starting Alec Script Runner TUI...")
	fmt.Println("Use 'q' or Ctrl+C to quit")

	err = tuiManager.Start(ctx)
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	fmt.Println("TUI shutdown complete")
	return nil
}

