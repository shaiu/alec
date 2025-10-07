package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiu/alec/pkg/contracts"
	"github.com/shaiu/alec/pkg/icon"
)

type ContentView int

const (
	ContentViewScriptDetails ContentView = iota
	ContentViewWelcome
)

type MainContentModel struct {
	width   int
	height  int
	focused bool

	contentView ContentView

	selectedScript *contracts.ScriptInfo

	configManager contracts.ConfigManager

	style MainContentStyle
}

type MainContentStyle struct {
	Base       lipgloss.Style
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Content    lipgloss.Style
	Output     lipgloss.Style
	OutputLine lipgloss.Style
	Error      lipgloss.Style
	Success    lipgloss.Style
	Running    lipgloss.Style
	Focused    lipgloss.Style
}

func NewMainContentModel(configManager contracts.ConfigManager) MainContentModel {
	style := MainContentStyle{
		Base: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6272A4")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")),
		Subtitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")),
		Content: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")),
		Output: lipgloss.NewStyle().
			Background(lipgloss.Color("#282A36")).
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(1),
		OutputLine: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")),
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")),
		Running: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true),
		Focused: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#BD93F9")),
	}

	return MainContentModel{
		style:         style,
		contentView:   ContentViewWelcome,
		configManager: configManager,
	}
}

func (m MainContentModel) Init() tea.Cmd {
	return nil
}

func (m MainContentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case ScriptSelectedMsg:
		m.selectedScript = &msg.Script
		m.contentView = ContentViewScriptDetails
	}

	return m, nil
}

func (m MainContentModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var content string

	switch m.contentView {
	case ContentViewScriptDetails:
		content = m.renderScriptDetails()
	case ContentViewWelcome:
		content = m.renderWelcome()
	}

	// Don't truncate - renderScriptDetails already handles sizing appropriately
	// to ensure execution instructions are visible

	baseStyle := m.style.Base
	if m.focused {
		baseStyle = m.style.Focused
	}

	// Account for borders (2 chars for left and right borders)
	// Render without width constraint to allow horizontal scrolling
	return baseStyle.
		Width(m.width - 2).
		MaxHeight(m.height).
		Render(content)
}

func (m MainContentModel) renderWelcome() string {
	welcome := fmt.Sprintf("%s Alec Script Runner\n\n", icon.Current.Execute) +
		"Welcome! Select a script from the sidebar to view details.\n\n" +
		"Navigation:\n" +
		fmt.Sprintf("%s %s/%s or k/j to navigate scripts\n", icon.Current.Bullet, icon.Current.ArrowUp, icon.Current.ArrowDown) +
		fmt.Sprintf("%s Enter to execute selected script\n", icon.Current.Bullet) +
		fmt.Sprintf("%s / or Ctrl+F to search & filter scripts\n", icon.Current.Bullet) +
		fmt.Sprintf("%s Esc to exit search mode\n", icon.Current.Bullet) +
		fmt.Sprintf("%s r to refresh script list\n", icon.Current.Bullet) +
		fmt.Sprintf("%s q or Ctrl+C to quit\n\n", icon.Current.Bullet) +
		fmt.Sprintf("%s Search Features:\n", icon.Current.Search) +
		fmt.Sprintf("%s Real-time filtering as you type\n", icon.Current.Bullet) +
		fmt.Sprintf("%s Navigate results with %s/%s or j/k\n", icon.Current.Bullet, icon.Current.ArrowUp, icon.Current.ArrowDown) +
		fmt.Sprintf("%s Visual match highlighting\n", icon.Current.Bullet) +
		fmt.Sprintf("%s Contextual search within current folder\n\n", icon.Current.Bullet) +
		"The selected script will run directly and the application will exit."
	return m.style.Content.Render(welcome)
}

func (m MainContentModel) renderScriptDetails() string {
	if m.selectedScript == nil {
		return m.renderWelcome()
	}

	var content strings.Builder

	// Script header with icon
	scriptIcon := m.getScriptIcon(m.selectedScript.Type)
	title := m.style.Title.Render(fmt.Sprintf("%s %s", scriptIcon, m.selectedScript.Name))
	content.WriteString(title + "\n\n")

	// Script information
	content.WriteString(icon.Current.Bullet + " " + m.style.Subtitle.Render("Type: ") + m.selectedScript.Type + "\n")

	// Show interpreter if available from metadata
	if m.selectedScript.Metadata != nil && m.selectedScript.Metadata.Interpreter != "" {
		content.WriteString(icon.Current.Bullet + " " + m.style.Subtitle.Render("Interpreter: ") + m.selectedScript.Metadata.Interpreter + "\n")
	}

	// Get file info if available
	if stat, err := os.Stat(m.selectedScript.Path); err == nil {
		content.WriteString(icon.Current.Bullet + " " + m.style.Subtitle.Render("Modified: ") + stat.ModTime().Format("2006-01-02 15:04:05") + "\n")
	}

	content.WriteString("\n")

	// Display description from metadata (preferred) or fallback to old method
	description := ""
	if m.selectedScript.Metadata != nil && m.selectedScript.Metadata.Description != "" {
		description = m.selectedScript.Metadata.Description
	} else {
		// Fallback to old extraction method for scripts without metadata
		description = m.extractScriptDescription(m.selectedScript.Path)
	}

	if description != "" {
		content.WriteString(icon.Current.Bullet + " " + m.style.Subtitle.Render("Description:") + "\n")
		content.WriteString(m.style.Content.Render(description) + "\n\n")
	}

	// Display script preview if metadata is available
	if m.selectedScript.Metadata != nil && m.selectedScript.Metadata.FullContent != "" {
		content.WriteString(strings.Repeat("─", 50) + "\n")

		previewTitle := "Script Preview"
		if m.selectedScript.Metadata.IsTruncated {
			previewTitle = fmt.Sprintf("Script Preview (showing %d of %d lines)",
				m.selectedScript.Metadata.PreviewLines,
				m.selectedScript.Metadata.LineCount)
		} else {
			previewTitle = "Full Script"
		}

		content.WriteString(icon.Current.Bullet + " " + m.style.Subtitle.Render(previewTitle) + "\n\n")

		// Limit preview to a conservative maximum to ensure execution instructions are always visible
		// Be very aggressive here because lipgloss padding/styling adds significant height
		maxPreviewLines := 10

		// If we have more height available, allow more preview lines
		// But never exceed half the available height to ensure bottom is visible
		if m.height > 30 {
			maxPreviewLines = (m.height / 2) - 10
		}

		if maxPreviewLines < 5 {
			maxPreviewLines = 5
		}

		previewContent := m.selectedScript.Metadata.FullContent
		previewLines := strings.Split(previewContent, "\n")
		isTruncatedForDisplay := false

		if len(previewLines) > maxPreviewLines {
			previewLines = previewLines[:maxPreviewLines]
			previewContent = strings.Join(previewLines, "\n")
			isTruncatedForDisplay = true
		}

		// Render the script content with subtle styling
		previewStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#1E1E1E")).
			Padding(1).
			MarginBottom(1)

		content.WriteString(previewStyle.Render(previewContent) + "\n")

		if isTruncatedForDisplay || m.selectedScript.Metadata.IsTruncated {
			truncateNote := m.style.Subtitle.Render("... (script continues)")
			content.WriteString(truncateNote + "\n")
		}
		content.WriteString("\n")
	}

	// Execution instructions
	content.WriteString(strings.Repeat("─", 50) + "\n")
	instructions := m.style.Running.Render(fmt.Sprintf("%s Press Enter to execute this script", icon.Current.Lightning))
	content.WriteString(instructions + "\n")
	content.WriteString(m.style.Content.Render("The script will run and the application will exit."))

	return content.String()
}

func (m MainContentModel) getScriptIcon(scriptType string) string {
	switch scriptType {
	case "shell":
		return "🐚"
	case "python":
		return "🐍"
	case "node":
		return "📦"
	default:
		return "📄"
	}
}

// getDisplayPath computes a user-friendly relative path from configured script directories
func (m MainContentModel) getDisplayPath(fullPath string) string {
	// Get configured script directories
	config, err := m.configManager.LoadConfig()
	if err == nil && config != nil && len(config.ScriptDirectories) > 0 {
		// Try to find which script directory this file belongs to
		for _, scriptDir := range config.ScriptDirectories {
			// Expand home directory if needed
			expandedDir := scriptDir
			if strings.HasPrefix(scriptDir, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					expandedDir = filepath.Join(homeDir, scriptDir[2:])
				}
			}

			// Clean and make absolute
			expandedDir, err := filepath.Abs(expandedDir)
			if err != nil {
				continue
			}

			// Check if the script is under this directory
			if strings.HasPrefix(fullPath, expandedDir) {
				relPath, err := filepath.Rel(expandedDir, fullPath)
				if err == nil {
					return relPath
				}
			}
		}
	}

	// Fallback: return just the filename and parent directory
	dir := filepath.Dir(fullPath)
	parentDir := filepath.Base(dir)
	filename := filepath.Base(fullPath)
	return filepath.Join(parentDir, filename)
}

func (m MainContentModel) extractScriptDescription(scriptPath string) string {
	file, err := os.Open(scriptPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var description strings.Builder
	buf := make([]byte, 2048) // Read first 2KB to look for comments
	n, err := file.Read(buf)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(buf[:n]), "\n")
	for i, line := range lines {
		// Skip shebang line
		if i == 0 && strings.HasPrefix(line, "#!") {
			continue
		}

		// Look for comment lines at the beginning
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			// Extract comment content
			comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if comment != "" {
				if description.Len() > 0 {
					description.WriteString("\n")
				}
				description.WriteString(comment)
			}
		} else if trimmed != "" {
			// Stop at first non-comment, non-empty line
			break
		}
	}

	result := description.String()
	if len(result) > 200 {
		return result[:200] + "..."
	}
	return result
}

func (m *MainContentModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *MainContentModel) SetFocused(focused bool) {
	m.focused = focused
}

// HandleSizeChange handles terminal size changes for responsive layout
func (m *MainContentModel) HandleSizeChange(width, height int) tea.Cmd {
	m.SetSize(width, height)
	return nil
}

// ProcessMessage handles messages for enhanced component communication
func (m *MainContentModel) ProcessMessage(msg tea.Msg) tea.Cmd {
	// Handle any main content-specific messages here
	switch msg.(type) {
	case ScriptSelectedMsg:
		_, cmd := m.Update(msg)
		return cmd
	}
	return nil
}

// truncateToHeight truncates content to fit within the specified height
// This prevents content overflow that would push the layout off screen
func (m MainContentModel) truncateToHeight(content string, maxHeight int) string {
	if maxHeight <= 0 {
		return ""
	}

	lines := strings.Split(content, "\n")

	// If content fits, return as-is
	if len(lines) <= maxHeight {
		return content
	}

	// Reserve space at bottom for critical execution instructions (last 3 lines)
	// These typically include the separator, "Press Enter" message, and description
	const reservedBottomLines = 3
	const truncationIndicatorLines = 1

	// Calculate how many lines we can show from the top
	availableTopLines := maxHeight - reservedBottomLines - truncationIndicatorLines
	if availableTopLines < 1 {
		availableTopLines = 1
	}

	// Take top lines
	result := make([]string, 0, maxHeight)
	result = append(result, lines[:availableTopLines]...)

	// Add truncation indicator
	indicator := m.style.Subtitle.Render("... (content truncated, scroll to see more)")
	result = append(result, indicator)

	// Add bottom lines (execution instructions)
	bottomStartIndex := len(lines) - reservedBottomLines
	if bottomStartIndex > availableTopLines {
		result = append(result, lines[bottomStartIndex:]...)
	}

	return strings.Join(result, "\n")
}
