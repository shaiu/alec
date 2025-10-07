package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type HeaderModel struct {
	width  int
	height int

	title   string
	version string
	status  string

	style HeaderStyle
}

type HeaderStyle struct {
	Base    lipgloss.Style
	Title   lipgloss.Style
	Version lipgloss.Style
	Status  lipgloss.Style
	Border  lipgloss.Style
}

func NewHeaderModel() HeaderModel {
	style := HeaderStyle{
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(0, 2),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BD93F9")),
		Version: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true),
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C")).
			Bold(true),
		Border: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")),
	}

	return HeaderModel{
		title:   "Alec Script Runner",
		version: "v1.0.0",
		style:   style,
	}
}

func (m HeaderModel) Init() tea.Cmd {
	return nil
}

func (m HeaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m HeaderModel) View() string {
	if m.width == 0 {
		return ""
	}

	title := m.style.Title.Render(m.title)
	version := m.style.Version.Render(m.version)

	leftSide := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		" ",
		version,
	)

	var headerContent string
	if m.status != "" {
		status := m.style.Status.Render(m.status)
		leftWidth := lipgloss.Width(leftSide)
		statusWidth := lipgloss.Width(status)
		totalUsed := leftWidth + statusWidth + 4

		var padding string
		if m.width > totalUsed {
			padding = strings.Repeat(" ", m.width-totalUsed)
		} else {
			padding = "  "
		}

		headerContent = lipgloss.JoinHorizontal(
			lipgloss.Left,
			leftSide,
			padding,
			status,
		)
	} else {
		leftWidth := lipgloss.Width(leftSide)
		if m.width > leftWidth+4 {
			padding := strings.Repeat(" ", m.width-leftWidth-4)
			headerContent = leftSide + padding
		} else {
			headerContent = leftSide
		}
	}

	header := m.style.Base.Width(m.width).Render(headerContent)
	border := m.style.Border.Width(m.width).Render(strings.Repeat("─", m.width))

	return header + "\n" + border
}

func (m *HeaderModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *HeaderModel) SetTitle(title string) {
	m.title = title
}

func (m *HeaderModel) SetVersion(version string) {
	m.version = version
}

func (m *HeaderModel) SetStatus(status string) {
	m.status = status
}

func (m *HeaderModel) ClearStatus() {
	m.status = ""
}

// ProcessMessage handles messages for enhanced component communication
func (m *HeaderModel) ProcessMessage(msg tea.Msg) tea.Cmd {
	// Handle any header-specific messages here
	switch msg.(type) {
	case tea.WindowSizeMsg:
		_, cmd := m.Update(msg)
		return cmd
	}
	return nil
}

