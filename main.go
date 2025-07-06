package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	rootDir string
)

var rootCmd = &cobra.Command{
	Use:   "alec",
	Short: "A directory navigator and script executor",
	Long:  "Navigate directories and execute shell scripts with a beautiful TUI interface",
	Run: func(cmd *cobra.Command, args []string) {
		config := getDefaultConfig()
		if rootDir != "" {
			config.RootDir = rootDir
		}
		
		p := tea.NewProgram(initialModelWithConfig(config))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting program: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.Flags().StringVarP(&rootDir, "dir", "d", "", "Root directory to navigate (default: directory where alec binary is located)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}