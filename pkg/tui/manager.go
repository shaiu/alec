package tui

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/your-org/alec/pkg/contracts"
	"github.com/your-org/alec/pkg/services"
)

type TUIManager struct {
	registry *services.ServiceRegistry
	program  *tea.Program
}

func NewTUIManager(registry *services.ServiceRegistry) *TUIManager {
	return &TUIManager{
		registry: registry,
	}
}

func (tm *TUIManager) Start(ctx context.Context) error {
	model := NewRootModel(tm.registry)

	tm.program = tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	go func() {
		<-ctx.Done()
		if tm.program != nil {
			tm.program.Quit()
		}
	}()

	_, err := tm.program.Run()
	return err
}

func (tm *TUIManager) Stop() {
	if tm.program != nil {
		tm.program.Quit()
	}
}

type TUIConfig struct {
	Title           string
	ShowLineNumbers bool
	EnableMouse     bool
	Theme           ThemeConfig
}

type ThemeConfig struct {
	Primary    string
	Secondary  string
	Background string
	Foreground string
	Border     string
	Focused    string
	Selected   string
	Error      string
	Success    string
}

func DefaultTUIConfig() *TUIConfig {
	return &TUIConfig{
		Title:           "Alec Script Runner",
		ShowLineNumbers: true,
		EnableMouse:     true,
		Theme: ThemeConfig{
			Primary:    "#BD93F9",
			Secondary:  "#6272A4",
			Background: "#282A36",
			Foreground: "#F8F8F2",
			Border:     "#44475A",
			Focused:    "#BD93F9",
			Selected:   "#44475A",
			Error:      "#FF5555",
			Success:    "#50FA7B",
		},
	}
}

type TUIService struct {
	config   *TUIConfig
	registry *services.ServiceRegistry
}

func NewTUIService(registry *services.ServiceRegistry, config *TUIConfig) contracts.TUIManager {
	if config == nil {
		config = DefaultTUIConfig()
	}

	return &TUIService{
		config:   config,
		registry: registry,
	}
}

func (ts *TUIService) StartInteractiveMode(ctx context.Context) error {
	manager := NewTUIManager(ts.registry)
	return manager.Start(ctx)
}

func (ts *TUIService) Initialize(ctx context.Context) error {
	return nil
}

func (ts *TUIService) HandleResize(width, height int) error {
	return nil
}

func (ts *TUIService) UpdateFocus(component contracts.ComponentType) error {
	return nil
}

func (ts *TUIService) NavigateToScript(script contracts.ScriptInfo) error {
	return nil
}

func (ts *TUIService) NavigateToDirectory(path string) error {
	return nil
}

func (ts *TUIService) UpdateSearch(query string) error {
	return nil
}


func (ts *TUIService) Shutdown() error {
	return nil
}

func (ts *TUIService) GetCurrentState() contracts.UIState {
	return contracts.UIState{
		CurrentView:      contracts.ViewBrowser,
		FocusedComponent: contracts.ComponentSidebar,
		TerminalWidth:    80,
		TerminalHeight:   24,
		SelectedScript:   nil,
	}
}


func (ts *TUIService) ExecuteSelectedScript(ctx context.Context) (string, error) {
	scriptExecutor := ts.registry.GetScriptExecutor()

	state := ts.GetCurrentState()

	if state.SelectedScript == nil {
		return "", fmt.Errorf("no script selected")
	}

	return scriptExecutor.ExecuteScript(ctx, *state.SelectedScript)
}

func (ts *TUIService) RefreshScriptList(ctx context.Context) error {
	scriptDiscovery := ts.registry.GetScriptDiscovery()

	_, err := scriptDiscovery.ScanDirectories(ctx, []string{"."}) // Simplified for now
	if err != nil {
		return fmt.Errorf("failed to scan directories: %w", err)
	}

	return nil
}

func (ts *TUIService) HandleKeyInput(ctx context.Context, key string) error {
	switch key {
	case "q", "ctrl+c":
		return fmt.Errorf("quit requested")
	case "r":
		return ts.RefreshScriptList(ctx)
	case "enter":
		_, err := ts.ExecuteSelectedScript(ctx)
		return err
	default:
		return nil
	}
}