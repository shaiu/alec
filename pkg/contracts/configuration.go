// Contract: Configuration Management Interface
// This file defines the contract for application configuration
// Implementation must pass all associated tests

package contracts

import (
	"time"
)

// AppConfig represents the complete application configuration
type AppConfig struct {
	ScriptDirectories []string               `mapstructure:"script_dirs" json:"script_dirs"`
	ScriptExtensions  map[string]string      `mapstructure:"extensions" json:"extensions"`
	Execution         ExecutionConfig        `mapstructure:"execution" json:"execution"`
	UI                UIConfig               `mapstructure:"ui" json:"ui"`
	Security          SecurityPolicy         `mapstructure:"security" json:"security"`
	Logging           LoggingConfig          `mapstructure:"logging" json:"logging"`
	KeyBindings       map[string]KeyBinding  `mapstructure:"key_bindings" json:"key_bindings"`
}

// UIConfig contains user interface configuration
type UIConfig struct {
	Theme            ThemeConfig   `mapstructure:"theme" json:"theme"`
	Layout           LayoutConfig  `mapstructure:"layout" json:"layout"`
	ShowHidden       bool          `mapstructure:"show_hidden" json:"show_hidden"`
	DefaultView      ViewType      `mapstructure:"default_view" json:"default_view"`
	RefreshOnFocus   bool          `mapstructure:"refresh_on_focus" json:"refresh_on_focus"`
	ConfirmOnExecute bool          `mapstructure:"confirm_on_execute" json:"confirm_on_execute"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level" json:"level"`
	File       string `mapstructure:"file" json:"file,omitempty"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age"`
}

// ConfigManager interface defines the contract for configuration management
type ConfigManager interface {
	// LoadConfig loads configuration from file and environment
	// Must handle missing config file gracefully with defaults
	LoadConfig() (*AppConfig, error)

	// SaveConfig writes current configuration to file
	// Must create directories and handle file permissions
	SaveConfig(config *AppConfig) error

	// GetDefaultConfig returns configuration with default values
	// Must provide sensible defaults for all required fields
	GetDefaultConfig() *AppConfig

	// ValidateConfig checks configuration for errors and conflicts
	// Must verify paths exist and permissions are correct
	ValidateConfig(config *AppConfig) error

	// WatchConfig monitors configuration file for changes
	// Returns channel of configuration updates
	WatchConfig() (<-chan *AppConfig, error)

	// GetConfigPath returns the current configuration file path
	// Must follow OS-specific config directory conventions
	GetConfigPath() string

	// MergeConfig combines configuration from multiple sources
	// Priority: CLI flags > env vars > config file > defaults
	MergeConfig(configs ...*AppConfig) *AppConfig
}

// Default configuration values
var DefaultConfig = &AppConfig{
	ScriptDirectories: []string{"./scripts", "~/.local/bin"},
	ScriptExtensions: map[string]string{
		".sh":  "shell",
		".bash": "shell",
		".py":  "python",
		".js":  "node",
		".rb":  "ruby",
		".pl":  "perl",
	},
	Execution: ExecutionConfig{
		Timeout:       5 * time.Minute,
		MaxOutputSize: 1000,
		Shell:         "", // Auto-detect
		WorkingDir:    "", // Use script directory
	},
	UI: UIConfig{
		Theme: ThemeConfig{
			Primary:    "#7D56F4",
			Secondary:  "#EE6FF8",
			Background: "#1A1A1A",
			Foreground: "#FAFAFA",
			Border:     "#444444",
			Focused:    "#00FF00",
			Selected:   "#FFFF00",
			Error:      "#FF0000",
			Success:    "#00FF00",
		},
		Layout: LayoutConfig{
			MinTerminalWidth:  80,
			MinTerminalHeight: 24,
			SidebarRatio:      0.382, // Golden ratio
			MaxSidebarWidth:   50,
			MinSidebarWidth:   20,
		},
		ShowHidden:       false,
		DefaultView:      ViewBrowser,
		RefreshOnFocus:   true,
		ConfirmOnExecute: false,
	},
	Security: SecurityPolicy{
		AllowedDirectories: []string{}, // Set from ScriptDirectories
		AllowedExtensions:  []string{".sh", ".bash", ".py", ".js", ".rb", ".pl"},
		MaxExecutionTime:   10 * time.Minute,
		MaxOutputSize:      10000,
		RestrictedCommands: []string{"rm", "sudo", "su", "chmod", "chown"},
	},
	Logging: LoggingConfig{
		Level:      "info",
		File:       "", // Stdout only
		MaxSize:    10, // MB
		MaxBackups: 3,
		MaxAge:     30, // Days
	},
	KeyBindings: map[string]KeyBinding{
		"quit":         {Key: "q", Action: "quit", Description: "Quit application"},
		"help":         {Key: "?", Action: "help", Description: "Show help"},
		"refresh":      {Key: "r", Action: "refresh", Description: "Refresh script list"},
		"execute":      {Key: "enter", Action: "execute", Description: "Execute selected script"},
		"search":       {Key: "/", Action: "search", Description: "Search scripts"},
		"focus_next":   {Key: "tab", Action: "focus_next", Description: "Next component"},
		"focus_prev":   {Key: "shift+tab", Action: "focus_prev", Description: "Previous component"},
		"nav_up":       {Key: "up", Action: "nav_up", Description: "Navigate up"},
		"nav_down":     {Key: "down", Action: "nav_down", Description: "Navigate down"},
		"nav_left":     {Key: "left", Action: "nav_left", Description: "Navigate left"},
		"nav_right":    {Key: "right", Action: "nav_right", Description: "Navigate right"},
		"expand":       {Key: "space", Action: "expand", Description: "Expand/collapse directory"},
		"back":         {Key: "esc", Action: "back", Description: "Go back"},
	},
}

// Configuration file locations (OS-specific)
var ConfigLocations = struct {
	Linux   []string
	MacOS   []string
	Windows []string
}{
	Linux:   []string{"~/.config/alec/alec.yaml", "~/.alec.yaml", "/etc/alec/alec.yaml"},
	MacOS:   []string{"~/Library/Application Support/alec/alec.yaml", "~/.alec.yaml"},
	Windows: []string{"%APPDATA%/alec/alec.yaml", "%USERPROFILE%/.alec.yaml"},
}

// Contract Requirements:
// 1. Configuration MUST support YAML, JSON, and TOML formats
// 2. Environment variables MUST override config file values
// 3. CLI flags MUST override environment variables
// 4. Missing config file MUST use defaults (not error)
// 5. Invalid config values MUST provide clear error messages
// 6. Config validation MUST check file/directory existence
// 7. Security settings MUST be enforced (no bypassing)
// 8. Config changes MUST trigger appropriate system updates
// 9. Default paths MUST follow OS conventions
// 10. Config file MUST be created with secure permissions (0600)