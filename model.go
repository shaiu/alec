package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	items       []FileItem
	cursor      int
	currentPath string
	config      *Config
	breadcrumb  []string
}

func initialModel() model {
	config := getDefaultConfig()
	rootDir := config.GetRootDir()
	items, _ := readDirectory(rootDir)
	
	return model{
		items:       items,
		cursor:      0,
		currentPath: rootDir,
		config:      config,
		breadcrumb:  []string{filepath.Base(rootDir)},
	}
}

func initialModelWithConfig(config *Config) model {
	rootDir := config.GetRootDir()
	items, _ := readDirectory(rootDir)
	
	return model{
		items:       items,
		cursor:      0,
		currentPath: rootDir,
		config:      config,
		breadcrumb:  []string{filepath.Base(rootDir)},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items) {
				m.cursor++
			}
		case "enter", " ":
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if item.IsScript {
					return m, tea.ExecProcess(exec.Command("bash", item.Path), func(err error) tea.Msg {
						return scriptExecuted{err: err}
					})
				} else if item.IsDir {
					items, err := readDirectory(item.Path)
					if err == nil {
						m.items = items
						m.cursor = 0
						m.currentPath = item.Path
						m.breadcrumb = append(m.breadcrumb, item.Name)
					}
				}
			} else if m.cursor == len(m.items) {
				return m, tea.Quit
			}
		case "backspace", "h":
			if len(m.breadcrumb) > 1 {
				m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
				parentPath := filepath.Dir(m.currentPath)
				items, err := readDirectory(parentPath)
				if err == nil {
					m.items = items
					m.cursor = 0
					m.currentPath = parentPath
				}
			}
		}
	case scriptExecuted:
		return m, tea.Quit
	}
	return m, nil
}

type scriptExecuted struct {
	err error
}

func (m model) View() string {
	var headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingTop(1).
		PaddingLeft(2).
		PaddingRight(2).
		Width(50)

	var dirStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	var scriptStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)

	s := headerStyle.Render("Alec CLI - Directory Navigator") + "\n\n"
	s += fmt.Sprintf("Path: %s\n", strings.Join(m.breadcrumb, " > "))
	s += "\n"

	for i, item := range m.items {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		var itemName string
		if item.IsDir {
			itemName = dirStyle.Render(fmt.Sprintf("📁 %s/", item.Name))
		} else if item.IsScript {
			itemName = scriptStyle.Render(fmt.Sprintf("🚀 %s", item.Name))
		} else {
			itemName = item.Name
		}

		s += fmt.Sprintf("%s %s\n", cursor, itemName)
	}

	if len(m.items) == 0 {
		s += "  (empty directory)\n"
	}

	s += "\n"
	if m.cursor == len(m.items) {
		s += "> Exit\n"
	} else {
		s += "  Exit\n"
	}

	s += "\n"
	s += "Controls: ↑/↓ or j/k to navigate, Enter to select, Backspace/h to go back, q to quit\n"
	return s
}