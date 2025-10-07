package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiu/alec/pkg/contracts"
	"github.com/shaiu/alec/pkg/services"
)

type SidebarModel struct {
	width   int
	height  int
	focused bool

	directories    []contracts.DirectoryInfo
	scripts        []contracts.ScriptInfo
	selectedIndex  int
	scrollOffset   int
	maxVisibleRows int

	// Navigation state
	currentPath    string
	currentItems   []NavigationItem
	allDirectories []contracts.DirectoryInfo
	allScripts     []contracts.ScriptInfo

	scriptDiscovery contracts.ScriptDiscovery
	configManager   contracts.ConfigManager

	loading bool
	err     error

	// Search functionality
	searchMode      bool
	searchQuery     string
	filteredScripts []contracts.ScriptInfo

	// Debug info
	debugInfo string

	style SidebarStyle
}

type NavigationItem struct {
	Type     NavigationItemType
	Name     string
	Path     string
	Script   *contracts.ScriptInfo
	IsParent bool // ".." item to go up one level
}

type NavigationItemType int

const (
	NavigationItemDirectory NavigationItemType = iota
	NavigationItemScript
)

type SidebarStyle struct {
	Base     lipgloss.Style
	Title    lipgloss.Style
	Item     lipgloss.Style
	Selected lipgloss.Style
	Focused  lipgloss.Style
	Loading  lipgloss.Style
	Error    lipgloss.Style
}

func NewSidebarModel(scriptDiscovery contracts.ScriptDiscovery, configManager contracts.ConfigManager) SidebarModel {
	style := SidebarStyle{
		Base: lipgloss.NewStyle().
			BorderRight(true).
			BorderForeground(lipgloss.Color("#3C3C3C")).
			PaddingLeft(1).
			PaddingRight(1),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")),
		Item: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")),
		Selected: lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(lipgloss.Color("#F8F8F2")),
		Focused: lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("#BD93F9")),
		Loading: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")),
	}

	return SidebarModel{
		scriptDiscovery: scriptDiscovery,
		configManager:   configManager,
		style:           style,
		loading:         true,
	}
}

func (m SidebarModel) Init() tea.Cmd {
	return m.loadScripts()
}

func (m SidebarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		// Handle search mode input
		if m.searchMode {
			switch msg.String() {
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.applyFilter()
				}
				return m, nil
			case "up", "k":
				// Allow navigation in search mode
				if len(m.filteredScripts) > 0 {
					m.selectedIndex = max(0, m.selectedIndex-1)
					m.updateScroll()
					return m, m.sendScriptSelectedMsg()
				}
				return m, nil
			case "down", "j":
				// Allow navigation in search mode
				if len(m.filteredScripts) > 0 {
					maxIndex := len(m.filteredScripts) - 1
					m.selectedIndex = min(maxIndex, m.selectedIndex+1)
					m.updateScroll()
					return m, m.sendScriptSelectedMsg()
				}
				return m, nil
			case "pageup":
				// Page up in search results
				if len(m.filteredScripts) > 0 {
					m.selectedIndex = max(0, m.selectedIndex-m.maxVisibleRows)
					m.updateScroll()
					return m, m.sendScriptSelectedMsg()
				}
				return m, nil
			case "pagedown":
				// Page down in search results
				if len(m.filteredScripts) > 0 {
					maxIndex := len(m.filteredScripts) - 1
					m.selectedIndex = min(maxIndex, m.selectedIndex+m.maxVisibleRows)
					m.updateScroll()
					return m, m.sendScriptSelectedMsg()
				}
				return m, nil
			case "home":
				// Go to first result
				if len(m.filteredScripts) > 0 {
					m.selectedIndex = 0
					m.updateScroll()
					return m, m.sendScriptSelectedMsg()
				}
				return m, nil
			case "end":
				// Go to last result
				if len(m.filteredScripts) > 0 {
					m.selectedIndex = len(m.filteredScripts) - 1
					m.updateScroll()
					return m, m.sendScriptSelectedMsg()
				}
				return m, nil
			default:
				// Add character to search query (only printable characters)
				if len(msg.String()) == 1 && msg.String()[0] >= 32 && msg.String()[0] <= 126 {
					m.searchQuery += msg.String()
					m.applyFilter()
				}
				return m, nil
			}
		}

		// Normal navigation mode
		switch msg.String() {
		case "up", "k":
			if len(m.currentItems) > 0 {
				m.selectedIndex = max(0, m.selectedIndex-1)
				m.updateScroll()
				return m, m.sendScriptSelectedMsg()
			}
		case "down", "j":
			if len(m.currentItems) > 0 {
				maxIndex := len(m.currentItems) - 1
				m.selectedIndex = min(maxIndex, m.selectedIndex+1)
				m.updateScroll()
				return m, m.sendScriptSelectedMsg()
			}
		case "enter":
			if len(m.currentItems) > 0 && m.selectedIndex < len(m.currentItems) {
				selectedItem := m.currentItems[m.selectedIndex]
				if selectedItem.Type == NavigationItemDirectory {
					// Navigate into directory
					if selectedItem.IsParent {
						// Go up one level
						m.navigateUp()
					} else {
						// Go into subdirectory
						m.navigateInto(selectedItem.Path)
					}
					return m, nil
				}
				// If it's a script, let the parent handle execution
			}
		case "r":
			m.loading = true
			cmd := m.loadScripts()
			return m, cmd
		case "/", "ctrl+f":
			m.enterSearchMode()
		case "escape":
			if m.searchMode {
				m.exitSearchMode()
			}
		}

	case ScriptsLoadedMsg:
		m.loading = false
		m.directories = msg.Directories
		m.allDirectories = msg.Directories
		m.allScripts = msg.Scripts
		m.scripts = msg.Scripts
		m.err = nil
		m.selectedIndex = 0
		m.scrollOffset = 0 // Reset scroll to show first item

		// Set initial current path to first directory or root
		if len(msg.Directories) > 0 {
			m.currentPath = msg.Directories[0].Path
		} else {
			m.currentPath = "."
		}

		// Build navigation items for current directory
		m.currentItems = m.buildNavigationItems(m.currentPath)

		// Reset search when loading new scripts
		if m.searchMode {
			m.applyFilter()
		}

	case ScriptsLoadErrorMsg:
		m.loading = false
		m.err = msg.Error
	}

	return m, nil
}

func (m SidebarModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var content strings.Builder

	// Show current path and navigation
	if m.searchMode {
		searchTitle := m.style.Title.Render("🔍 Search & Filter")
		content.WriteString(searchTitle + "\n")

		// Show current search query with visual indicator
		var searchPrompt string
		if m.searchQuery == "" {
			searchPrompt = "Type to filter scripts..."
			content.WriteString(m.style.Loading.Render(searchPrompt) + "\n")
		} else {
			searchPrompt = fmt.Sprintf("Filtering: \"%s\"", m.searchQuery)
			content.WriteString(m.style.Selected.Render(searchPrompt) + "\n")
		}

		// Show filtering status and results within current context
		contextScripts := m.getScriptsInCurrentContext()
		totalContextScripts := len(contextScripts)
		filteredCount := len(m.filteredScripts)

		// Show current directory context
		pathDisplay := filepath.Base(m.currentPath)
		if pathDisplay == "." || pathDisplay == "" {
			pathDisplay = "root"
		}
		contextInfo := fmt.Sprintf("In: %s", pathDisplay)
		content.WriteString(m.style.Loading.Render(contextInfo) + "\n")

		var filterStatus string
		if m.searchQuery == "" {
			filterStatus = fmt.Sprintf("Showing all %d scripts in context", totalContextScripts)
		} else if filteredCount == 0 {
			filterStatus = fmt.Sprintf("No matches for \"%s\" (0 of %d)", m.searchQuery, totalContextScripts)
		} else {
			filterStatus = fmt.Sprintf("Found %d of %d scripts in context", filteredCount, totalContextScripts)
		}

		content.WriteString(m.style.Loading.Render(filterStatus) + "\n")

		// Add search help if query is empty
		if m.searchQuery == "" {
			helpText := "Press Esc to exit search"
			content.WriteString(m.style.Loading.Render(helpText) + "\n")
		}
	} else {
		// Show title
		title := m.style.Title.Render("Scripts")
		content.WriteString(title + "\n\n")
	}

	if m.loading {
		loading := m.style.Loading.Render("Loading scripts...")
		content.WriteString(loading)
	} else if m.err != nil {
		error := m.style.Error.Render(fmt.Sprintf("Error: %s", m.err.Error()))
		content.WriteString(error)
	} else {
		m.renderScripts(&content)
	}

	baseStyle := m.style.Base
	if m.focused {
		baseStyle = baseStyle.Copy().Inherit(m.style.Focused)
	}

	// Force sidebar to fixed width of 35 characters to prevent layout shifts
	const fixedSidebarWidth = 35
	return baseStyle.
		Width(fixedSidebarWidth).
		MaxWidth(fixedSidebarWidth).
		Height(m.height).
		MaxHeight(m.height).
		Render(content.String())
}

func (m *SidebarModel) renderScripts(content *strings.Builder) {
	if m.searchMode {
		// In search mode, show filtered scripts
		if len(m.filteredScripts) == 0 && m.searchQuery != "" {
			// Show "no results" message
			noResultsMsg := "No scripts match your filter"
			content.WriteString(m.style.Error.Render(noResultsMsg) + "\n")
			content.WriteString(m.style.Loading.Render("Try a different search term") + "\n")
			return
		}

		visibleStart := m.scrollOffset
		visibleEnd := min(len(m.filteredScripts), m.scrollOffset+m.maxVisibleRows)

		// Add a subtle separator line
		if m.searchQuery != "" {
			separator := strings.Repeat("─", min(20, m.width-4))
			content.WriteString(m.style.Loading.Render(separator) + "\n")
		}

		for i := visibleStart; i < visibleEnd; i++ {
			script := m.filteredScripts[i]
			line := m.formatSearchScriptLine(script, i == m.selectedIndex)
			content.WriteString(line + "\n")
		}

		if len(m.filteredScripts) > m.maxVisibleRows {
			scrollInfo := fmt.Sprintf("\n%d-%d of %d matches",
				visibleStart+1, visibleEnd, len(m.filteredScripts))
			content.WriteString(m.style.Loading.Render(scrollInfo))
		}
	} else {
		// In navigation mode, show current directory items
		visibleStart := m.scrollOffset
		visibleEnd := min(len(m.currentItems), m.scrollOffset+m.maxVisibleRows)

		for i := visibleStart; i < visibleEnd; i++ {
			item := m.currentItems[i]
			line := m.formatNavigationItemLine(item, i == m.selectedIndex)
			content.WriteString(line + "\n")
		}

		if len(m.currentItems) > m.maxVisibleRows {
			scrollInfo := fmt.Sprintf("\n%d-%d of %d",
				visibleStart+1, visibleEnd, len(m.currentItems))
			content.WriteString(m.style.Loading.Render(scrollInfo))
		}
	}
}

func (m SidebarModel) formatScriptLine(script contracts.ScriptInfo, selected bool) string {
	icon := m.getScriptIcon(script.Type)
	name := script.Name

	// Calculate available space more conservatively
	// Sidebar width (35) - icon (2) - space (1) - padding/borders (4) - ellipsis reserve (3) = ~25
	const fixedSidebarWidth = 35
	maxNameLength := fixedSidebarWidth - 10 // Conservative calculation

	if len(name) > maxNameLength {
		if maxNameLength > 3 {
			name = name[:maxNameLength-3] + "..."
		} else if maxNameLength > 0 {
			name = name[:maxNameLength]
		} else {
			name = "..."
		}
	}

	line := fmt.Sprintf("%s %s", icon, name)

	// Apply max width constraint to prevent overflow
	lineStyle := m.style.Item
	if selected {
		lineStyle = m.style.Selected
	}

	return lineStyle.MaxWidth(fixedSidebarWidth - 2).Render(line)
}

func (m SidebarModel) formatSearchScriptLine(script contracts.ScriptInfo, selected bool) string {
	icon := m.getScriptIcon(script.Type)
	name := script.Name

	// Calculate available space more conservatively
	// Sidebar width (35) - icon (2) - space (1) - padding/borders (4) - ellipsis reserve (3) = ~25
	const fixedSidebarWidth = 35
	maxNameLength := fixedSidebarWidth - 10 // Conservative calculation

	// Truncate BEFORE highlighting to avoid style codes in length calculation
	if len(name) > maxNameLength {
		if maxNameLength > 3 {
			name = name[:maxNameLength-3] + "..."
		} else if maxNameLength > 0 {
			name = name[:maxNameLength]
		} else {
			name = "..."
		}
	}

	// Highlight search matches in the name (after truncation)
	if m.searchQuery != "" && !strings.HasSuffix(name, "...") {
		name = m.highlightSearchMatch(name, m.searchQuery)
	}

	line := fmt.Sprintf("%s %s", icon, name)

	// Apply max width constraint to prevent overflow
	lineStyle := m.style.Item
	if selected {
		lineStyle = m.style.Selected
	}

	return lineStyle.MaxWidth(fixedSidebarWidth - 2).Render(line)
}

func (m SidebarModel) highlightSearchMatch(text, query string) string {
	if query == "" {
		return text
	}

	lowerText := strings.ToLower(text)
	lowerQuery := strings.ToLower(query)

	index := strings.Index(lowerText, lowerQuery)
	if index == -1 {
		return text
	}

	// Create highlighted version
	before := text[:index]
	match := text[index : index+len(query)]
	after := text[index+len(query):]

	// Use a different style for the match (bold/colored)
	highlightStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB86C"))
	highlightedMatch := highlightStyle.Render(match)

	return before + highlightedMatch + after
}

func (m SidebarModel) getScriptIcon(scriptType string) string {
	switch scriptType {
	case "shell":
		return "\ue86f" // Material icon: code
	case "python":
		return "\ue86f" // Material icon: code
	case "node":
		return "\ue86f" // Material icon: code
	default:
		return "\ue873" // Material icon: description
	}
}

func (m SidebarModel) GetSelectedScript() *contracts.ScriptInfo {
	if m.searchMode {
		// In search mode, return selected script from filtered list
		if len(m.filteredScripts) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.filteredScripts) {
			return nil
		}
		return &m.filteredScripts[m.selectedIndex]
	} else {
		// In navigation mode, return script from current items
		if len(m.currentItems) == 0 || m.selectedIndex < 0 || m.selectedIndex >= len(m.currentItems) {
			return nil
		}

		selectedItem := m.currentItems[m.selectedIndex]

		// Only return script if it's a script item (not a directory)
		if selectedItem.Type == NavigationItemScript && selectedItem.Script != nil {
			return selectedItem.Script
		}

		return nil
	}
}

func (m *SidebarModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.updateLayout()
}

func (m *SidebarModel) SetFocused(focused bool) {
	m.focused = focused
}

func (m SidebarModel) IsSearchMode() bool {
	return m.searchMode
}

func (m SidebarModel) GetSearchQuery() string {
	return m.searchQuery
}

// GetCurrentPath returns the current directory path being viewed
func (m SidebarModel) GetCurrentPath() string {
	return m.currentPath
}

// GetSelectionPosition returns the current selection position as "X of Y" format
func (m SidebarModel) GetSelectionPosition() string {
	if m.searchMode {
		if len(m.filteredScripts) == 0 {
			return ""
		}
		return fmt.Sprintf("%d of %d", m.selectedIndex+1, len(m.filteredScripts))
	} else {
		// For navigation mode, handle ".." item specially
		totalContentItems := m.GetCurrentItemCount()
		if totalContentItems == 0 {
			// If no content items but we have navigation items (like "..")
			if len(m.currentItems) > 0 {
				return "nav"
			}
			return ""
		}

		// Check if we're on the ".." item by examining the selected item directly
		if len(m.currentItems) > 0 && m.selectedIndex < len(m.currentItems) {
			selectedItem := m.currentItems[m.selectedIndex]
			if selectedItem.IsParent && selectedItem.Name == ".." {
				// Use consistent length with padding for visual stability
				sampleLength := len(fmt.Sprintf("1 of %d", totalContentItems))
				padding := sampleLength - 1 // -1 for the "↑" character
				return "↑" + strings.Repeat(" ", padding)
			}
		}

		// Calculate position for content items (adjusting for ".." if present)
		adjustedIndex := m.selectedIndex
		if len(m.currentItems) > 0 && m.currentItems[0].Name == ".." {
			adjustedIndex--
		}

		return fmt.Sprintf("%d of %d", adjustedIndex+1, totalContentItems)
	}
}

// IsLoading returns whether the sidebar is currently loading scripts
func (m SidebarModel) IsLoading() bool {
	return m.loading
}

// GetDebugInfo returns debug information about item counting
func (m SidebarModel) GetDebugInfo() string {
	return m.debugInfo
}

func (m *SidebarModel) updateLayout() {
	// Account for content that takes up space:
	// - Title line: "📁 Scripts" (1 line)
	// - Empty line after title: (1 line)
	// - Potential scroll info at bottom: (1 line)
	// - Extra buffer for borders/padding: (2 lines)
	// Total overhead: 5 lines
	usableHeight := m.height - 5
	m.maxVisibleRows = max(1, usableHeight) // Ensure at least 1 row is visible
	m.updateScroll()
}

func (m *SidebarModel) updateScroll() {
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	} else if m.selectedIndex >= m.scrollOffset+m.maxVisibleRows {
		m.scrollOffset = m.selectedIndex - m.maxVisibleRows + 1
	}
}

func (m SidebarModel) loadScripts() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Load script directories from configuration
		var scriptDirs []string
		var config *contracts.AppConfig
		if m.configManager != nil {
			var err error
			config, err = m.configManager.LoadConfig()
			if err == nil && len(config.ScriptDirectories) > 0 {
				scriptDirs = config.ScriptDirectories
			}
		}

		// Fallback to default directories if config not available
		if len(scriptDirs) == 0 {
			scriptDirs = []string{".", "./scripts", "./test-scripts", "~/scripts", "~/bin"}
		}

		// Create a new discovery service with current script directories as allowed dirs
		// This ensures the security validator allows scripts from the configured directories
		var discoveryService contracts.ScriptDiscovery
		if config != nil {
			discoveryService = services.NewScriptDiscoveryService(scriptDirs, config.ScriptExtensions)
		} else {
			// Fallback extensions if config not available
			defaultExtensions := map[string]string{
				".sh": "shell",
				".py": "python",
				".js": "node",
			}
			discoveryService = services.NewScriptDiscoveryService(scriptDirs, defaultExtensions)
		}

		directories, err := discoveryService.ScanDirectories(ctx, scriptDirs)
		if err != nil {
			return ScriptsLoadErrorMsg{Error: err}
		}

		// Collect all scripts from directory tree (including subdirectories)
		var allScripts []contracts.ScriptInfo
		for _, dir := range directories {
			allScripts = append(allScripts, m.collectAllScriptsFromDirectory(dir)...)
		}

		return ScriptsLoadedMsg{
			Directories: directories,
			Scripts:     allScripts,
		}
	})
}

// Search functionality methods

// getCurrentScripts returns the currently displayed scripts (filtered or all)
func (m SidebarModel) getCurrentScripts() []contracts.ScriptInfo {
	if m.searchMode && m.filteredScripts != nil {
		return m.filteredScripts
	}
	return m.scripts
}

// EnterSearchMode activates search mode
func (m *SidebarModel) EnterSearchMode() tea.Cmd {
	m.searchMode = true
	m.searchQuery = ""
	m.applyFilter() // Apply contextual filter from the start
	m.selectedIndex = 0
	m.scrollOffset = 0
	return nil
}

// ExitSearchMode deactivates search mode and returns to normal view
func (m *SidebarModel) ExitSearchMode() tea.Cmd {
	m.searchMode = false
	m.searchQuery = ""
	m.filteredScripts = nil
	m.selectedIndex = 0
	m.scrollOffset = 0
	return nil
}

// enterSearchMode is the internal method used by keyboard handling
func (m *SidebarModel) enterSearchMode() {
	m.searchMode = true
	m.searchQuery = ""
	m.applyFilter() // Apply contextual filter from the start
	m.selectedIndex = 0
	m.scrollOffset = 0
}

// exitSearchMode is the internal method used by keyboard handling
func (m *SidebarModel) exitSearchMode() {
	m.searchMode = false
	m.searchQuery = ""
	m.filteredScripts = nil
	m.selectedIndex = 0
	m.scrollOffset = 0

	// Rebuild navigation items for current directory
	if m.currentPath == "" && len(m.allDirectories) > 0 {
		// If no current path is set, default to first directory
		m.currentPath = m.allDirectories[0].Path
	}
	if m.currentPath != "" {
		m.currentItems = m.buildNavigationItems(m.currentPath)
	}
}

// applyFilter filters scripts based on current search query within the current directory context
func (m *SidebarModel) applyFilter() {
	// Get scripts that are in the current directory context
	currentScripts := m.getScriptsInCurrentContext()

	if m.searchQuery == "" {
		m.filteredScripts = currentScripts
	} else {
		m.filteredScripts = make([]contracts.ScriptInfo, 0)
		query := strings.ToLower(m.searchQuery)

		for _, script := range currentScripts {
			// Search in script name and path
			if strings.Contains(strings.ToLower(script.Name), query) ||
				strings.Contains(strings.ToLower(script.Path), query) ||
				strings.Contains(strings.ToLower(script.Type), query) {
				m.filteredScripts = append(m.filteredScripts, script)
			}
		}
	}

	// Reset selection to first item after filtering
	m.selectedIndex = 0
	m.scrollOffset = 0
}

// getScriptsInCurrentContext returns scripts that are within the current directory context
func (m SidebarModel) getScriptsInCurrentContext() []contracts.ScriptInfo {
	if m.currentPath == "" {
		// If no current path is set, return all scripts (fallback)
		return m.allScripts
	}

	var contextScripts []contracts.ScriptInfo

	// Include scripts in current directory and all subdirectories
	for _, script := range m.allScripts {
		scriptDir := filepath.Dir(script.Path)

		// Check if script is in current directory or a subdirectory
		if scriptDir == m.currentPath || strings.HasPrefix(scriptDir, m.currentPath+string(filepath.Separator)) {
			contextScripts = append(contextScripts, script)
		}
	}

	return contextScripts
}

// RefreshScripts triggers a script reload
func (m *SidebarModel) RefreshScripts() tea.Cmd {
	m.loading = true
	return m.loadScripts()
}

// HandleSizeChange handles terminal size changes for responsive layout
func (m *SidebarModel) HandleSizeChange(width, height int) tea.Cmd {
	m.SetSize(width, height)
	return nil
}

// ProcessMessage handles messages for enhanced component communication
func (m *SidebarModel) ProcessMessage(msg tea.Msg) tea.Cmd {
	// Handle any sidebar-specific messages here
	switch msg.(type) {
	case ScriptsLoadedMsg, ScriptsLoadErrorMsg:
		_, cmd := m.Update(msg)
		return cmd
	}
	return nil
}

// GetFilteredScriptCount returns the number of filtered scripts in search mode
func (m SidebarModel) GetFilteredScriptCount() int {
	return len(m.filteredScripts)
}

// GetContextScriptCount returns the number of scripts in the current context
func (m SidebarModel) GetContextScriptCount() int {
	return len(m.getScriptsInCurrentContext())
}

// GetCurrentItemCount returns the number of content items (excluding navigation helpers)
func (m *SidebarModel) GetCurrentItemCount() int {
	count := 0
	for _, item := range m.currentItems {
		// Skip ".." navigation item specifically
		if item.Name == ".." {
			continue
		}
		count++
	}
	return count
}

// GetTotalNavigationItems returns the total number of items including navigation helpers
func (m SidebarModel) GetTotalNavigationItems() int {
	return len(m.currentItems)
}

type ScriptsLoadedMsg struct {
	Directories []contracts.DirectoryInfo
	Scripts     []contracts.ScriptInfo
}

type ScriptsLoadErrorMsg struct {
	Error error
}

type ScriptSelectedMsg struct {
	Script contracts.ScriptInfo
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m SidebarModel) sendScriptSelectedMsg() tea.Cmd {
	selectedScript := m.GetSelectedScript()
	if selectedScript == nil {
		return nil
	}
	return tea.Cmd(func() tea.Msg {
		return ScriptSelectedMsg{
			Script: *selectedScript,
		}
	})
}

// Navigation methods

func (m *SidebarModel) navigateInto(path string) {
	m.currentPath = path
	m.currentItems = m.buildNavigationItems(path)
	m.selectedIndex = 0
	m.scrollOffset = 0
}

func (m *SidebarModel) navigateUp() {
	parentPath := filepath.Dir(m.currentPath)
	// Don't go above the root directories
	for _, rootDir := range m.allDirectories {
		if m.currentPath == rootDir.Path {
			return // Already at root level
		}
	}
	m.currentPath = parentPath
	m.currentItems = m.buildNavigationItems(parentPath)
	m.selectedIndex = 0
	m.scrollOffset = 0
}

func (m SidebarModel) buildNavigationItems(currentPath string) []NavigationItem {
	var items []NavigationItem

	// Check if we're not at root level - add ".." item
	isAtRoot := false
	for _, rootDir := range m.allDirectories {
		if currentPath == rootDir.Path {
			isAtRoot = true
			break
		}
	}

	if !isAtRoot {
		items = append(items, NavigationItem{
			Type:     NavigationItemDirectory,
			Name:     "..",
			Path:     filepath.Dir(currentPath),
			IsParent: true,
		})
	}

	// Find all directories and scripts at current path
	scriptsByDir := make(map[string][]contracts.ScriptInfo)
	allDirs := make(map[string]bool)

	// Group all scripts by their directory
	for _, script := range m.allScripts {
		scriptDir := filepath.Dir(script.Path)
		scriptsByDir[scriptDir] = append(scriptsByDir[scriptDir], script)

		// Track directory hierarchy
		tempDir := scriptDir
		for tempDir != "." && tempDir != "/" {
			allDirs[tempDir] = true
			tempDir = filepath.Dir(tempDir)
		}
	}

	// Find subdirectories of current path
	subdirs := make(map[string]bool)
	for dir := range allDirs {
		if filepath.Dir(dir) == currentPath {
			subdirs[dir] = true
		}
	}

	// Add subdirectories
	var subdirNames []string
	for subdir := range subdirs {
		subdirNames = append(subdirNames, subdir)
	}
	sort.Strings(subdirNames)

	for _, subdir := range subdirNames {
		items = append(items, NavigationItem{
			Type:     NavigationItemDirectory,
			Name:     filepath.Base(subdir),
			Path:     subdir,
			IsParent: false,
		})
	}

	// Add scripts in current directory
	if scripts, exists := scriptsByDir[currentPath]; exists {
		// Sort scripts by name
		sort.Slice(scripts, func(i, j int) bool {
			return scripts[i].Name < scripts[j].Name
		})

		for i, script := range scripts {
			items = append(items, NavigationItem{
				Type:     NavigationItemScript,
				Name:     script.Name,
				Path:     script.Path,
				Script:   &scripts[i],
				IsParent: false,
			})
		}
	}

	return items
}

func (m SidebarModel) formatNavigationItemLine(item NavigationItem, selected bool) string {
	var icon string
	var name string

	if item.Type == NavigationItemDirectory {
		if item.IsParent {
			icon = "⬆️"
			name = ".."
		} else {
			icon = "\ue2c7" // Material icon: folder
			name = item.Name
		}
	} else {
		icon = m.getScriptIcon(item.Script.Type)
		name = item.Name
	}

	// Calculate available space more conservatively
	// Sidebar width (35) - icon (2) - space (1) - padding/borders (4) - ellipsis reserve (3) = ~25
	const fixedSidebarWidth = 35
	maxNameLength := fixedSidebarWidth - 10 // Conservative calculation

	if len(name) > maxNameLength {
		if maxNameLength > 3 {
			name = name[:maxNameLength-3] + "..."
		} else if maxNameLength > 0 {
			name = name[:maxNameLength]
		} else {
			name = "..."
		}
	}

	line := fmt.Sprintf("%s %s", icon, name)

	// Apply max width constraint to prevent overflow
	lineStyle := m.style.Item
	if selected {
		lineStyle = m.style.Selected
	}

	return lineStyle.MaxWidth(fixedSidebarWidth - 2).Render(line)
}

// collectAllScriptsFromDirectory recursively collects all scripts from a directory tree
func (m SidebarModel) collectAllScriptsFromDirectory(dir contracts.DirectoryInfo) []contracts.ScriptInfo {
	var allScripts []contracts.ScriptInfo

	// Add scripts from this directory
	allScripts = append(allScripts, dir.Scripts...)

	// Recursively add scripts from subdirectories
	for _, child := range dir.Children {
		allScripts = append(allScripts, m.collectAllScriptsFromDirectory(child)...)
	}

	return allScripts
}
