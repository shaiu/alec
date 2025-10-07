package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shaiu/alec/pkg/icon"
)

type FooterModel struct {
	width  int
	height int

	helpText    string
	status      string
	scriptCount string
	currentPath string
	position    string
	loading     bool

	style FooterStyle
}

type FooterStyle struct {
	Base        lipgloss.Style
	Help        lipgloss.Style
	Status      lipgloss.Style
	ScriptCount lipgloss.Style
	Path        lipgloss.Style
	Position    lipgloss.Style
	Loading     lipgloss.Style
	Border      lipgloss.Style
}

func NewFooterModel() FooterModel {
	style := FooterStyle{
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6272A4")),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true),
		ScriptCount: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9")).
			Bold(true),
		Path: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Italic(true),
		Position: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")),
		Loading: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C")).
			Bold(true),
		Border: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")),
	}

	return FooterModel{
		helpText:    fmt.Sprintf("%s/%s navigate %s Enter execute %s / search %s r refresh %s q quit",
			icon.Current.ArrowUp, icon.Current.ArrowDown, icon.Current.Separator,
			icon.Current.Separator, icon.Current.Separator, icon.Current.Separator),
		status:      "Ready",
		scriptCount: "",
		currentPath: "",
		position:    "",
		loading:     false,
		style:       style,
	}
}

func (m FooterModel) Init() tea.Cmd {
	return nil
}

func (m FooterModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height


	case ScriptsLoadedMsg:
		m.status = "Scripts Loaded"

	case ScriptsLoadErrorMsg:
		m.status = "Error Loading Scripts"
	}

	return m, nil
}

func (m FooterModel) View() string {
	if m.width == 0 {
		return ""
	}

	// Left section: Help text
	help := m.style.Help.Render(m.helpText)

	// Middle section: Path and position info
	var middleSection []string
	if m.currentPath != "" {
		pathText := m.style.Path.Render("📁 " + m.currentPath)
		middleSection = append(middleSection, pathText)
	}
	if m.position != "" {
		posText := m.style.Position.Render(m.position)
		middleSection = append(middleSection, posText)
	}

	middle := ""
	if len(middleSection) > 0 {
		middle = lipgloss.JoinHorizontal(lipgloss.Left, middleSection...)
		if len(middleSection) > 1 {
			middle = strings.Join(middleSection, " • ")
		}
	}

	// Right section: Script count, loading, and status
	var rightElements []string
	if m.loading {
		loadingText := m.style.Loading.Render("⏳ Loading...")
		rightElements = append(rightElements, loadingText)
	}
	if m.scriptCount != "" {
		scriptCount := m.style.ScriptCount.Render(m.scriptCount)
		rightElements = append(rightElements, scriptCount)
	}

	status := m.style.Status.Render(m.status)
	rightElements = append(rightElements, status)

	rightSide := lipgloss.JoinHorizontal(lipgloss.Left, strings.Join(rightElements, "  "))

	// Calculate padding
	usedWidth := lipgloss.Width(help) + lipgloss.Width(middle) + lipgloss.Width(rightSide) + 8 // 8 for spacing
	availableWidth := m.width - usedWidth

	var footerContent string
	if middle != "" && availableWidth > 0 {
		// Three sections: help | middle | right
		leftPadding := strings.Repeat(" ", max(1, availableWidth/2))
		rightPadding := strings.Repeat(" ", max(1, availableWidth-len(leftPadding)))

		footerContent = lipgloss.JoinHorizontal(
			lipgloss.Left,
			help,
			leftPadding,
			middle,
			rightPadding,
			rightSide,
		)
	} else {
		// Two sections: help | right
		padding := strings.Repeat(" ", max(0, m.width-lipgloss.Width(help)-lipgloss.Width(rightSide)-4))
		footerContent = lipgloss.JoinHorizontal(
			lipgloss.Left,
			help,
			padding,
			rightSide,
		)
	}

	// Render with border - no need for separate border line
	// Account for borders (2 chars for left and right)
	return m.style.Base.
		Width(m.width - 2).
		Render(footerContent)
}

func (m *FooterModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *FooterModel) SetHelpText(help string) {
	m.helpText = help
}

func (m *FooterModel) SetStatus(status string) {
	m.status = status
}

func (m *FooterModel) SetScriptCount(count string) {
	m.scriptCount = count
}

func (m *FooterModel) SetCurrentPath(path string) {
	m.currentPath = path
}

func (m *FooterModel) SetPosition(pos string) {
	m.position = pos
}

func (m *FooterModel) SetLoading(loading bool) {
	m.loading = loading
}

// ShowWarning displays a warning message in the footer
func (m *FooterModel) ShowWarning(warning string) {
	m.status = "⚠️ " + warning
}

// ClearWarning clears any warning message and returns to ready status
func (m *FooterModel) ClearWarning() {
	m.status = "Ready"
}

// ShowError displays an error message in the footer
func (m *FooterModel) ShowError(error string) {
	m.status = "❌ " + error
}

// ShowHelp toggles the display of extended help information
func (m *FooterModel) ShowHelp(show bool) {
	if show {
		m.helpText = fmt.Sprintf("Type to filter %s %s/%s navigate %s Enter execute %s Esc exit search %s r refresh %s q quit",
			icon.Current.Separator, icon.Current.ArrowUp, icon.Current.ArrowDown,
			icon.Current.Separator, icon.Current.Separator, icon.Current.Separator, icon.Current.Separator)
	} else {
		m.helpText = fmt.Sprintf("%s/%s navigate %s Enter execute %s / search %s r refresh %s q quit",
			icon.Current.ArrowUp, icon.Current.ArrowDown, icon.Current.Separator,
			icon.Current.Separator, icon.Current.Separator, icon.Current.Separator)
	}
}

// ProcessMessage handles messages for enhanced component communication
func (m *FooterModel) ProcessMessage(msg tea.Msg) tea.Cmd {
	// Handle any footer-specific messages here
	switch msg.(type) {
	case ScriptsLoadedMsg, ScriptsLoadErrorMsg:
		_, cmd := m.Update(msg)
		return cmd
	}
	return nil
}

