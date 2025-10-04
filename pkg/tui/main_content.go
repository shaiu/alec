package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/your-org/alec/pkg/contracts"
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

func NewMainContentModel() MainContentModel {
	style := MainContentStyle{
		Base: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingTop(1),
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
			BorderForeground(lipgloss.Color("#BD93F9")),
	}

	return MainContentModel{
		style:       style,
		contentView: ContentViewWelcome,
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

	baseStyle := m.style.Base
	if m.focused {
		baseStyle = baseStyle.Copy().Inherit(m.style.Focused)
	}

	// Simplified rendering without width constraints to debug
	return content
}

func (m MainContentModel) renderWelcome() string {
	welcome := "🚀 Alec Script Runner\n\n" +
		"Welcome! Select a script from the sidebar to view details.\n\n" +
		"Navigation:\n" +
		"• ↑/↓ or k/j to navigate scripts\n" +
		"• Enter to execute selected script\n" +
		"• / or Ctrl+F to search & filter scripts\n" +
		"• Esc to exit search mode\n" +
		"• r to refresh script list\n" +
		"• q or Ctrl+C to quit\n\n" +
		"🔍 Search Features:\n" +
		"• Real-time filtering as you type\n" +
		"• Navigate results with ↑/↓ or j/k\n" +
		"• Visual match highlighting\n" +
		"• Contextual search within current folder\n\n" +
		"The selected script will run directly and the application will exit."
	return m.style.Content.Render(welcome)
}

func (m MainContentModel) renderScriptDetails() string {
	if m.selectedScript == nil {
		return m.renderWelcome()
	}

	var content strings.Builder

	// Script header with icon
	icon := m.getScriptIcon(m.selectedScript.Type)
	title := m.style.Title.Render(fmt.Sprintf("%s %s", icon, m.selectedScript.Name))
	content.WriteString(title + "\n\n")

	// Script information
	content.WriteString("📁 " + m.style.Subtitle.Render("Location: ") + m.selectedScript.Path + "\n")
	content.WriteString("🔧 " + m.style.Subtitle.Render("Type: ") + m.selectedScript.Type + "\n")

	// Get file info if available
	if stat, err := os.Stat(m.selectedScript.Path); err == nil {
		content.WriteString("📅 " + m.style.Subtitle.Render("Modified: ") + stat.ModTime().Format("2006-01-02 15:04:05") + "\n")
		content.WriteString("📏 " + m.style.Subtitle.Render("Size: ") + fmt.Sprintf("%d bytes", stat.Size()) + "\n")
	}

	content.WriteString("\n")

	// Try to extract description from script comments
	description := m.extractScriptDescription(m.selectedScript.Path)
	if description != "" {
		content.WriteString("📝 " + m.style.Subtitle.Render("Description:") + "\n")
		content.WriteString(m.style.Content.Render(description) + "\n\n")
	}

	// Execution instructions
	content.WriteString(strings.Repeat("─", 50) + "\n")
	instructions := m.style.Running.Render("⚡ Press Enter to execute this script")
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
					description.WriteString(" ")
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

