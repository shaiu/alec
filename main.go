package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "alec",
	Short: "A directory navigator and script executor",
	Long:  "Navigate directories and execute shell scripts with a beautiful TUI interface",
	Run: func(cmd *cobra.Command, args []string) {
		config := loadConfig()
		
		p := tea.NewProgram(initialModelWithConfig(config))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error starting program: %v", err)
			os.Exit(1)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}