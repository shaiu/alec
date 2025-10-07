package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BreadcrumbModel struct {
	width  int
	height int

	breadcrumbs string

	style BreadcrumbStyle
}

type BreadcrumbStyle struct {
	Base   lipgloss.Style
	Text   lipgloss.Style
	Border lipgloss.Style
}

func NewBreadcrumbModel() BreadcrumbModel {
	style := BreadcrumbStyle{
		Base: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(0, 2),
		Text: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")),
		Border: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")),
	}

	return BreadcrumbModel{
		style:       style,
		breadcrumbs: "",
	}
}

func (m BreadcrumbModel) Init() tea.Cmd {
	return nil
}

func (m BreadcrumbModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m BreadcrumbModel) View() string {
	if m.width == 0 {
		return ""
	}

	var content string
	if m.breadcrumbs != "" {
		content = m.style.Text.Render(m.breadcrumbs)
	} else {
		content = m.style.Text.Render("📁 Scripts")
	}

	breadcrumbBar := m.style.Base.Width(m.width).Render(content)
	border := m.style.Border.Width(m.width).Render(strings.Repeat("─", m.width))

	return breadcrumbBar + "\n" + border
}

func (m *BreadcrumbModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *BreadcrumbModel) SetBreadcrumbs(breadcrumbs string) {
	m.breadcrumbs = breadcrumbs
}

func (m *BreadcrumbModel) ClearBreadcrumbs() {
	m.breadcrumbs = ""
}

// ProcessMessage handles messages for enhanced component communication
func (m *BreadcrumbModel) ProcessMessage(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case tea.WindowSizeMsg:
		_, cmd := m.Update(msg)
		return cmd
	}
	return nil
}
