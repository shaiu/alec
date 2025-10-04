package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/your-org/alec/pkg/contracts"
	"github.com/your-org/alec/pkg/services"
)


type RootModel struct {
	width  int
	height int

	sidebar     SidebarModel
	mainContent MainContentModel
	header      HeaderModel
	footer      FooterModel

	registry *services.ServiceRegistry

	quitting bool
}

type Component int

const (
	ComponentSidebar Component = iota
	ComponentMainContent
)

type ScriptExecutionErrorMsg struct {
	Error error
}

type ScriptExecutionCompleteMsg struct {
	SessionID string
	ExitCode  int
}

func NewRootModel(registry *services.ServiceRegistry) *RootModel {
	sidebar := NewSidebarModel(registry.GetScriptDiscovery(), registry.GetConfigManager())
	mainContent := NewMainContentModel()

	// Set initial focus state
	sidebar.SetFocused(true)
	mainContent.SetFocused(false)

	return &RootModel{
		registry:    registry,
		sidebar:     sidebar,
		mainContent: mainContent,
		header:      NewHeaderModel(),
		footer:      NewFooterModel(),
	}
}

func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.sidebar.Init(),
		m.mainContent.Init(),
		m.header.Init(),
		m.footer.Init(),
	)
}

func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		oldWidth, oldHeight := m.width, m.height
		m.width = msg.Width
		m.height = msg.Height

		// Detect significant size changes that require layout adjustment
		sizeChanged := m.hasSizeChanged(oldWidth, oldHeight, msg.Width, msg.Height)

		// Always update layout, but add responsive behavior for size changes
		m.updateLayout()

		// If terminal became too small, show warning in footer
		if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
			m.footer.ShowWarning("Terminal too small - some features may be limited")
		} else if sizeChanged {
			// Clear any previous size warnings
			m.footer.ClearWarning()
		}

		// Propagate size changes to all components
		var cmd tea.Cmd
		var model tea.Model

		model, cmd = m.sidebar.Update(msg)
		m.sidebar = model.(SidebarModel)
		cmds = append(cmds, cmd)

		model, cmd = m.mainContent.Update(msg)
		m.mainContent = model.(MainContentModel)
		cmds = append(cmds, cmd)

		model, cmd = m.header.Update(msg)
		m.header = model.(HeaderModel)
		cmds = append(cmds, cmd)

		model, cmd = m.footer.Update(msg)
		m.footer = model.(FooterModel)
		cmds = append(cmds, cmd)

		// If we had a significant size change, trigger a refresh
		if sizeChanged {
			refreshCmd := m.handleSizeChangeRefresh()
			if refreshCmd != nil {
				cmds = append(cmds, refreshCmd)
			}
		}

	case tea.KeyMsg:
		// Handle escape key using KeyType instead of string comparison
		if msg.Type == tea.KeyEsc {
			// If sidebar is in search mode, exit search mode directly
			if m.sidebar.IsSearchMode() {
				cmd := m.sidebar.ExitSearchMode()
				cmds = append(cmds, cmd)
				// Reset footer and header to normal mode
				m.footer.ShowHelp(false)
				m.header.ClearStatus()
			}
			// Always consume escape key to prevent other handling
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+r":
			// Refresh script list
			cmd := m.sidebar.RefreshScripts()
			cmds = append(cmds, cmd)
		case "ctrl+f", "/":
			// Enter search mode
			cmd := m.sidebar.EnterSearchMode()
			cmds = append(cmds, cmd)
			// Update footer and header for search mode
			m.footer.ShowHelp(true)
			m.footer.SetHelpText("Type to filter • ↑/↓/j/k navigate • Enter execute • Esc exit search")
			m.header.SetStatus("🔍 Search Mode")
		case "enter":
			selectedScript := m.sidebar.GetSelectedScript()
			if selectedScript != nil {
				// If in search mode, exit search mode first, then execute
				if m.sidebar.IsSearchMode() {
					cmd := m.sidebar.ExitSearchMode()
					cmds = append(cmds, cmd)
					m.footer.ShowHelp(false)
					m.header.ClearStatus()
				}
				// Execute script and return command to handle execution
				return m, m.executeScript(*selectedScript)
			} else {
				// No script selected, pass Enter key to sidebar for directory navigation
				var cmd tea.Cmd
				var model tea.Model
				model, cmd = m.sidebar.Update(msg)
				m.sidebar = model.(SidebarModel)
				cmds = append(cmds, cmd)
			}
		case "f1", "h", "?":
			// Show help
			m.showHelp()
		default:
			// Pass all other keys to sidebar (always focused)
			// This includes: "up", "k", "down", "j", "pageup", "pagedown", "home", "end", etc.
			var cmd tea.Cmd
			var model tea.Model
			model, cmd = m.sidebar.Update(msg)
			m.sidebar = model.(SidebarModel)
			cmds = append(cmds, cmd)
			// Don't return early - let footer update happen below
		}

	case ScriptSelectedMsg:
		// Forward script selection to main content
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.mainContent.Update(msg)
		m.mainContent = model.(MainContentModel)
		cmds = append(cmds, cmd)

	case ScriptExecutionErrorMsg:
		// Handle script execution errors (don't exit)
		m.footer.ShowError("Script execution failed: " + msg.Error.Error())

	case ScriptExecutionCompleteMsg:
		// Script completed, application will exit
		m.quitting = true
		return m, tea.Quit



	default:
		var cmd tea.Cmd
		var model tea.Model

		model, cmd = m.sidebar.Update(msg)
		m.sidebar = model.(SidebarModel)
		cmds = append(cmds, cmd)

		model, cmd = m.mainContent.Update(msg)
		m.mainContent = model.(MainContentModel)
		cmds = append(cmds, cmd)

		model, cmd = m.header.Update(msg)
		m.header = model.(HeaderModel)
		cmds = append(cmds, cmd)

		model, cmd = m.footer.Update(msg)
		m.footer = model.(FooterModel)
		cmds = append(cmds, cmd)
	}

	// Update footer with current script count
	m.updateFooterScriptCount()

	return m, tea.Batch(cmds...)
}

func (m *RootModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	header := m.header.View()
	footer := m.footer.View()

	// contentHeight := m.height - lipgloss.Height(header) - lipgloss.Height(footer)

	sidebar := m.sidebar.View()
	mainContent := m.mainContent.View()

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		mainContent,
	)

	app := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)

	// Apply padding using lipgloss margin instead of padding to avoid clipping
	containerStyle := lipgloss.NewStyle().
		Margin(1, 2) // 1 line top/bottom, 2 chars left/right margin

	return containerStyle.Render(app)
}

// Layout constants for responsive design
const (
	MinTerminalWidth     = 80
	MinTerminalHeight    = 24
	MaxSidebarWidth      = 20  // Very narrow - max 20 characters
	MinSidebarWidth      = 12  // Very narrow minimum
	DefaultSidebarRatio  = 0.12 // Only 12% of screen width
	HeaderHeight         = 3
	FooterHeight         = 3
	MinContentHeight     = 10
)

func (m *RootModel) updateLayout() {
	// Check minimum terminal size requirements
	if m.width < MinTerminalWidth || m.height < MinTerminalHeight {
		m.handleSmallTerminal()
		return
	}

	// Account for margin (1 line top + 1 line bottom = 2 lines total)
	// and component heights when calculating available content height
	sidebarWidth := 24 // Wide enough for script names, narrow enough for main content

	mainContentWidth := m.width - sidebarWidth
	contentHeight := m.height - HeaderHeight - FooterHeight - 2 // -2 for top/bottom margin

	// Ensure minimum content height
	if contentHeight < MinContentHeight {
		// Reduce header/footer height for very small terminals
		adjustedHeaderHeight := max(1, HeaderHeight-2)
		adjustedFooterHeight := max(1, FooterHeight-2)
		contentHeight = m.height - adjustedHeaderHeight - adjustedFooterHeight

		m.header.SetSize(m.width, adjustedHeaderHeight)
		m.footer.SetSize(m.width, adjustedFooterHeight)
	} else {
		m.header.SetSize(m.width, HeaderHeight)
		m.footer.SetSize(m.width, FooterHeight)
	}

	// Update component sizes with responsive calculations
	m.sidebar.SetSize(sidebarWidth, contentHeight)
	m.mainContent.SetSize(mainContentWidth, contentHeight)
}

// handleSmallTerminal manages layout for terminals below minimum size
func (m *RootModel) handleSmallTerminal() {
	// For very small terminals, use a simplified single-column layout
	if m.width < 60 {
		// Hide sidebar in extremely narrow terminals
		m.sidebar.SetSize(0, m.height-6) // -4 for header/footer, -2 for margin
		m.mainContent.SetSize(m.width, m.height-6) // -4 for header/footer, -2 for margin
		m.header.SetSize(m.width, 2)
		m.footer.SetSize(m.width, 2)
	} else {
		// Use minimum sizes for small but usable terminals
		sidebarWidth := MinSidebarWidth
		mainContentWidth := m.width - sidebarWidth
		contentHeight := max(MinContentHeight, m.height-6) // -4 for header/footer, -2 for margin

		m.sidebar.SetSize(sidebarWidth, contentHeight)
		m.mainContent.SetSize(mainContentWidth, contentHeight)
		m.header.SetSize(m.width, 2)
		m.footer.SetSize(m.width, 2)
	}
}

// getLayoutInfo returns current layout information for debugging
func (m *RootModel) getLayoutInfo() map[string]interface{} {
	return map[string]interface{}{
		"terminal_width":    m.width,
		"terminal_height":   m.height,
		"sidebar_width":     int(float64(m.width) * DefaultSidebarRatio),
		"main_width":        m.width - int(float64(m.width)*DefaultSidebarRatio),
		"content_height":    m.height - HeaderHeight - FooterHeight,
		"is_small_terminal": m.width < MinTerminalWidth || m.height < MinTerminalHeight,
	}
}

// hasSizeChanged determines if terminal size change is significant enough to trigger responsive behavior
func (m *RootModel) hasSizeChanged(oldWidth, oldHeight, newWidth, newHeight int) bool {
	// Consider it a significant change if:
	// 1. Width changed by more than 10 characters
	// 2. Height changed by more than 5 lines
	// 3. Crossed minimum size thresholds
	// 4. Size change affects layout (sidebar width calculation changes significantly)

	widthChange := abs(newWidth - oldWidth)
	heightChange := abs(newHeight - oldHeight)

	// Check for significant absolute changes
	if widthChange > 10 || heightChange > 5 {
		return true
	}

	// Check for crossing minimum size thresholds
	wasSmall := oldWidth < MinTerminalWidth || oldHeight < MinTerminalHeight
	isSmall := newWidth < MinTerminalWidth || newHeight < MinTerminalHeight
	if wasSmall != isSmall {
		return true
	}

	// Check if sidebar width calculation changes significantly
	oldSidebarWidth := int(float64(oldWidth) * DefaultSidebarRatio)
	newSidebarWidth := int(float64(newWidth) * DefaultSidebarRatio)
	if abs(newSidebarWidth-oldSidebarWidth) > 5 {
		return true
	}

	return false
}

// handleSizeChangeRefresh triggers appropriate refreshes when terminal size changes significantly
func (m *RootModel) handleSizeChangeRefresh() tea.Cmd {
	// When terminal size changes significantly, we may need to:
	// 1. Refresh script list display to fit new sidebar width
	// 2. Adjust any ongoing search/filter operations
	// 3. Update content view to utilize new space

	var cmds []tea.Cmd

	// Trigger sidebar refresh to adjust to new width
	if cmd := m.sidebar.HandleSizeChange(m.width, m.height); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Trigger main content refresh to adjust to new dimensions
	if cmd := m.mainContent.HandleSizeChange(m.width, m.height); cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Return batch command if we have any commands to execute
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}

	return nil
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}


func (m *RootModel) showHelp() {
	// This would show a help overlay or switch to help view
	// For now, we'll update the footer with help info
	m.footer.ShowHelp(true)
}

// Enhanced message routing for better component communication
func (m *RootModel) routeMessage(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	// Route messages to all components and collect commands
	if cmd := m.sidebar.ProcessMessage(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.mainContent.ProcessMessage(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.header.ProcessMessage(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	if cmd := m.footer.ProcessMessage(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return cmds
}

// executeScript executes a script and returns a command that will trigger application exit
func (m *RootModel) executeScript(script contracts.ScriptInfo) tea.Cmd {
	return tea.ExecProcess(m.buildScriptCommand(script), func(err error) tea.Msg {
		if err != nil {
			return ScriptExecutionErrorMsg{Error: err}
		}
		// Script completed successfully, exit the application
		return ScriptExecutionCompleteMsg{
			SessionID: "",
			ExitCode:  0,
		}
	})
}

// buildScriptCommand creates the appropriate command to execute a script
func (m *RootModel) buildScriptCommand(script contracts.ScriptInfo) *exec.Cmd {
	switch script.Type {
	case "shell":
		return exec.Command("bash", script.Path)
	case "python":
		return exec.Command("python3", script.Path)
	case "node":
		return exec.Command("node", script.Path)
	default:
		// Try to execute directly if it's executable
		return exec.Command(script.Path)
	}
}

// updateFooterScriptCount updates the footer with current script count and other information
func (m *RootModel) updateFooterScriptCount() {
	// Update script count
	var countText string
	if m.sidebar.IsSearchMode() {
		// In search mode, show filtered count
		filteredCount := m.sidebar.GetFilteredScriptCount()
		totalCount := m.sidebar.GetContextScriptCount()

		if m.sidebar.GetSearchQuery() == "" {
			countText = fmt.Sprintf("%d scripts", totalCount)
		} else {
			countText = fmt.Sprintf("%d of %d scripts", filteredCount, totalCount)
		}
	} else {
		// In navigation mode, show current directory content count (excluding '..' navigation)
		itemCount := m.sidebar.GetCurrentItemCount()
		countText = fmt.Sprintf("%d items", itemCount)
	}
	m.footer.SetScriptCount(countText)

	// Update current path
	currentPath := m.sidebar.GetCurrentPath()
	if currentPath != "" {
		// Clean up path display - show relative path or just folder name
		pathDisplay := filepath.Base(currentPath)
		if pathDisplay == "." || pathDisplay == "" {
			pathDisplay = "root"
		}
		m.footer.SetCurrentPath(pathDisplay)
	} else {
		m.footer.SetCurrentPath("")
	}

	// Update selection position
	position := m.sidebar.GetSelectionPosition()
	m.footer.SetPosition(position)

	// Update loading status
	m.footer.SetLoading(m.sidebar.IsLoading())
}

