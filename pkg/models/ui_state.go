package models

import (
	"fmt"

	"github.com/shaiu/alec/pkg/contracts"
)

// UIState represents the current state of the Terminal User Interface
type UIState struct {
	CurrentView       contracts.ViewType      `json:"current_view"`
	FocusedComponent  contracts.ComponentType `json:"focused_component"`
	TerminalWidth     int                     `json:"terminal_width"`
	TerminalHeight    int                     `json:"terminal_height"`
	SidebarWidth      int                     `json:"sidebar_width"`
	MainWidth         int                     `json:"main_width"`
	SelectedScript    *Script                 `json:"selected_script,omitempty"`
	SelectedDirectory string                  `json:"selected_directory,omitempty"`
	SearchQuery       string                  `json:"search_query,omitempty"`
	ShowHidden        bool                    `json:"show_hidden"`
	NavigationHistory []string                `json:"navigation_history"`
	CurrentExecution  *ExecutionSession       `json:"current_execution,omitempty"`
}

// NewUIState creates a new UI state with defaults
func NewUIState() *UIState {
	return &UIState{
		CurrentView:       contracts.ViewBrowser,
		FocusedComponent:  contracts.ComponentSidebar,
		TerminalWidth:     80,
		TerminalHeight:    24,
		SidebarWidth:      30,
		MainWidth:         50,
		ShowHidden:        false,
		NavigationHistory: make([]string, 0),
	}
}

// UpdateTerminalSize updates terminal dimensions and recalculates layout
func (u *UIState) UpdateTerminalSize(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid terminal dimensions: %dx%d", width, height)
	}

	if width < 40 || height < 10 {
		return fmt.Errorf("terminal too small: %dx%d (minimum 40x10)", width, height)
	}

	u.TerminalWidth = width
	u.TerminalHeight = height

	return u.RecalculateLayout()
}

// RecalculateLayout updates layout dimensions based on current terminal size
func (u *UIState) RecalculateLayout() error {
	if u.TerminalWidth <= 0 {
		return fmt.Errorf("terminal width not set")
	}

	// Golden ratio calculation for sidebar
	goldenRatio := 0.382
	sidebarWidth := int(float64(u.TerminalWidth) * goldenRatio)

	// Apply constraints
	if sidebarWidth < 20 {
		sidebarWidth = 20
	}
	if sidebarWidth > 50 {
		sidebarWidth = 50
	}

	// Ensure we don't exceed terminal width
	if sidebarWidth >= u.TerminalWidth-10 {
		sidebarWidth = u.TerminalWidth - 10
		if sidebarWidth < 20 {
			return fmt.Errorf("terminal too narrow for UI layout")
		}
	}

	u.SidebarWidth = sidebarWidth
	u.MainWidth = u.TerminalWidth - sidebarWidth - 1 // Account for border

	return nil
}

// SetCurrentView changes the active view
func (u *UIState) SetCurrentView(view contracts.ViewType) error {
	switch view {
	case contracts.ViewBrowser, contracts.ViewExecutor, contracts.ViewHelp, contracts.ViewConfig:
		u.CurrentView = view
		return nil
	default:
		return fmt.Errorf("invalid view type: %s", view)
	}
}

// SetFocusedComponent changes the focused component
func (u *UIState) SetFocusedComponent(component contracts.ComponentType) error {
	switch component {
	case contracts.ComponentSidebar, contracts.ComponentMain, contracts.ComponentOutput, contracts.ComponentSearch:
		u.FocusedComponent = component
		return nil
	default:
		return fmt.Errorf("invalid component type: %s", component)
	}
}

// SelectScript updates the selected script and navigation history
func (u *UIState) SelectScript(script *Script) {
	u.SelectedScript = script
	if script != nil {
		u.AddToHistory(script.Path)
		u.SelectedDirectory = script.Path // Directory containing the script
	}
}

// SelectDirectory updates the selected directory
func (u *UIState) SelectDirectory(path string) {
	u.SelectedDirectory = path
	u.AddToHistory(path)
}

// UpdateSearch sets the search query and clears inappropriate selections
func (u *UIState) UpdateSearch(query string) {
	u.SearchQuery = query

	// If search is active, we might need to clear current selections
	// that don't match the search
	if query != "" && u.SelectedScript != nil {
		// Implementation would check if selected script matches search
	}
}

// AddToHistory adds a path to navigation history
func (u *UIState) AddToHistory(path string) {
	// Remove if already exists to avoid duplicates
	for i, existing := range u.NavigationHistory {
		if existing == path {
			u.NavigationHistory = append(u.NavigationHistory[:i], u.NavigationHistory[i+1:]...)
			break
		}
	}

	// Add to front of history
	u.NavigationHistory = append([]string{path}, u.NavigationHistory...)

	// Limit history size
	if len(u.NavigationHistory) > 20 {
		u.NavigationHistory = u.NavigationHistory[:20]
	}
}

// GetBreadcrumbs returns breadcrumb navigation for current location
func (u *UIState) GetBreadcrumbs() []string {
	if u.SelectedDirectory == "" {
		return []string{"Home"}
	}

	// Split directory path into components
	// This would be implemented to create proper breadcrumbs
	return []string{"Home", "Scripts"}
}

// CanGoBack returns true if there's navigation history to go back to
func (u *UIState) CanGoBack() bool {
	return len(u.NavigationHistory) > 1
}

// GoBack navigates to the previous location in history
func (u *UIState) GoBack() bool {
	if !u.CanGoBack() {
		return false
	}

	// Remove current location and go to previous
	if len(u.NavigationHistory) > 1 {
		u.NavigationHistory = u.NavigationHistory[1:]
		u.SelectedDirectory = u.NavigationHistory[0]
		return true
	}

	return false
}

// IsResponsive returns true if terminal is large enough for full UI
func (u *UIState) IsResponsive() bool {
	return u.TerminalWidth >= 80 && u.TerminalHeight >= 24
}

// IsSearchActive returns true if search is currently active
func (u *UIState) IsSearchActive() bool {
	return u.SearchQuery != ""
}

// ClearSearch clears the current search query
func (u *UIState) ClearSearch() {
	u.SearchQuery = ""
}

// StartExecution sets the current execution session
func (u *UIState) StartExecution(session *ExecutionSession) {
	u.CurrentExecution = session
	u.SetCurrentView(contracts.ViewExecutor)
}

// EndExecution clears the current execution session
func (u *UIState) EndExecution() {
	u.CurrentExecution = nil
	u.SetCurrentView(contracts.ViewBrowser)
}

// IsExecuting returns true if there's an active execution
func (u *UIState) IsExecuting() bool {
	return u.CurrentExecution != nil && u.CurrentExecution.IsRunning()
}

// Clone creates a copy of the UI state
func (u *UIState) Clone() *UIState {
	clone := *u

	// Deep copy slices
	if u.NavigationHistory != nil {
		clone.NavigationHistory = make([]string, len(u.NavigationHistory))
		copy(clone.NavigationHistory, u.NavigationHistory)
	}

	// Don't deep copy script/execution as they're references
	return &clone
}

// Reset resets the UI state to defaults while preserving terminal size
func (u *UIState) Reset() {
	width, height := u.TerminalWidth, u.TerminalHeight
	*u = *NewUIState()
	u.TerminalWidth = width
	u.TerminalHeight = height
	u.RecalculateLayout()
}