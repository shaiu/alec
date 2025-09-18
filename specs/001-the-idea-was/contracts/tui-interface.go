// Contract: Terminal UI Interface
// This file defines the contract for TUI components and state management
// Implementation must pass all associated tests

package contracts

import (
	"context"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewType represents different UI views in the application
type ViewType string

const (
	ViewBrowser  ViewType = "browser"
	ViewExecutor ViewType = "executor"
	ViewHelp     ViewType = "help"
	ViewConfig   ViewType = "config"
)

// ComponentType represents focusable UI components
type ComponentType string

const (
	ComponentSidebar ComponentType = "sidebar"
	ComponentMain    ComponentType = "main"
	ComponentOutput  ComponentType = "output"
	ComponentSearch  ComponentType = "search"
)

// UIState represents the current state of the user interface
type UIState struct {
	CurrentView       ViewType      `json:"current_view"`
	FocusedComponent  ComponentType `json:"focused_component"`
	TerminalWidth     int           `json:"terminal_width"`
	TerminalHeight    int           `json:"terminal_height"`
	SidebarWidth      int           `json:"sidebar_width"`
	SelectedScript    *ScriptInfo   `json:"selected_script,omitempty"`
	SelectedDirectory string        `json:"selected_directory,omitempty"`
	SearchQuery       string        `json:"search_query,omitempty"`
	ShowHidden        bool          `json:"show_hidden"`
}

// NavigationState tracks navigation history and context
type NavigationState struct {
	History     []string `json:"history"`
	CurrentPath string   `json:"current_path"`
	Breadcrumb  []string `json:"breadcrumb"`
}

// TUIManager interface defines the contract for TUI state management
type TUIManager interface {
	// Initialize sets up the TUI with initial state
	// Must handle terminal size detection and component initialization
	Initialize(ctx context.Context) error

	// HandleResize responds to terminal size changes
	// Must recalculate layout and update all components
	HandleResize(width, height int) error

	// UpdateFocus changes the currently focused component
	// Must handle focus transitions and visual indicators
	UpdateFocus(component ComponentType) error

	// NavigateToScript selects a script and updates UI state
	// Must update navigation history and breadcrumb
	NavigateToScript(script ScriptInfo) error

	// NavigateToDirectory selects a directory and updates tree state
	// Must expand directory and update navigation context
	NavigateToDirectory(path string) error

	// UpdateSearch applies search/filter query to script list
	// Must filter both scripts and directories based on query
	UpdateSearch(query string) error

	// GetCurrentState returns current UI state
	// Must provide consistent snapshot of UI state
	GetCurrentState() UIState

	// Shutdown cleans up UI resources and restores terminal
	// Must restore terminal state and cleanup background processes
	Shutdown() error
}

// KeyBinding represents a keyboard shortcut configuration
type KeyBinding struct {
	Key         string `json:"key"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Component   string `json:"component,omitempty"`
}

// ThemeConfig contains visual styling configuration
type ThemeConfig struct {
	Primary        string `json:"primary"`
	Secondary      string `json:"secondary"`
	Background     string `json:"background"`
	Foreground     string `json:"foreground"`
	Border         string `json:"border"`
	Focused        string `json:"focused"`
	Selected       string `json:"selected"`
	Error          string `json:"error"`
	Success        string `json:"success"`
}

// LayoutConfig contains responsive layout settings
type LayoutConfig struct {
	MinTerminalWidth  int     `json:"min_terminal_width"`
	MinTerminalHeight int     `json:"min_terminal_height"`
	SidebarRatio      float64 `json:"sidebar_ratio"`
	MaxSidebarWidth   int     `json:"max_sidebar_width"`
	MinSidebarWidth   int     `json:"min_sidebar_width"`
}

// BubbleTeaModel interface for Bubble Tea integration
type BubbleTeaModel interface {
	tea.Model

	// SetSize updates model dimensions for responsive layout
	SetSize(width, height int)

	// SetFocus updates focus state for component
	SetFocus(focused bool)

	// GetHeight returns current model height
	GetHeight() int

	// GetWidth returns current model width
	GetWidth() int
}

// Message types for Bubble Tea communication
type (
	// ScriptSelectedMsg sent when user selects a script
	ScriptSelectedMsg struct {
		Script ScriptInfo
	}

	// DirectorySelectedMsg sent when user selects a directory
	DirectorySelectedMsg struct {
		Path string
	}

	// SearchUpdatedMsg sent when search query changes
	SearchUpdatedMsg struct {
		Query string
	}

	// FocusChangedMsg sent when component focus changes
	FocusChangedMsg struct {
		Component ComponentType
	}

	// ViewChangedMsg sent when active view changes
	ViewChangedMsg struct {
		View ViewType
	}

	// RefreshRequestedMsg sent when user requests manual refresh
	RefreshRequestedMsg struct{}
)

// Contract Requirements:
// 1. TUI MUST be responsive to terminal size changes
// 2. Focus management MUST provide clear visual indicators
// 3. Navigation MUST maintain history and breadcrumb context
// 4. Search MUST filter both scripts and directories real-time
// 5. Layout MUST use golden ratio for sidebar proportions
// 6. Components MUST handle keyboard input according to conventions
// 7. State changes MUST be communicated via Bubble Tea messages
// 8. UI MUST remain responsive during script execution
// 9. Error states MUST be clearly communicated to user
// 10. Shutdown MUST restore terminal to original state