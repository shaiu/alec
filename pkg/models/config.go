package models

import (
	"fmt"
	"path/filepath"
	"time"
)

// AppConfig represents the complete application configuration
// This implements the contracts.AppConfig interface
type AppConfig struct {
	ScriptDirectories []string                   `mapstructure:"script_dirs" json:"script_dirs" yaml:"script_dirs"`
	ScriptExtensions  map[string]string          `mapstructure:"extensions" json:"extensions" yaml:"extensions"`
	Execution         ExecutionConfig            `mapstructure:"execution" json:"execution" yaml:"execution"`
	UI                UIConfig                   `mapstructure:"ui" json:"ui" yaml:"ui"`
	Security          SecurityConfig             `mapstructure:"security" json:"security" yaml:"security"`
	Logging           LoggingConfig              `mapstructure:"logging" json:"logging" yaml:"logging"`
}

// ExecutionConfig contains execution-related configuration
type ExecutionConfig struct {
	Timeout       time.Duration `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	MaxOutputSize int           `mapstructure:"max_output_size" json:"max_output_size" yaml:"max_output_size"`
	Shell         string        `mapstructure:"shell" json:"shell" yaml:"shell"`
	WorkingDir    string        `mapstructure:"working_dir" json:"working_dir" yaml:"working_dir"`
}

// UIConfig contains user interface configuration
type UIConfig struct {
	ShowHidden       bool          `mapstructure:"show_hidden" json:"show_hidden" yaml:"show_hidden"`
	RefreshOnFocus   bool          `mapstructure:"refresh_on_focus" json:"refresh_on_focus" yaml:"refresh_on_focus"`
	ConfirmOnExecute bool          `mapstructure:"confirm_on_execute" json:"confirm_on_execute" yaml:"confirm_on_execute"`
	Theme            ThemeConfig   `mapstructure:"theme" json:"theme" yaml:"theme"`
	Layout           LayoutConfig  `mapstructure:"layout" json:"layout" yaml:"layout"`
}

// ThemeConfig contains visual styling configuration
type ThemeConfig struct {
	Primary    string `mapstructure:"primary" json:"primary" yaml:"primary"`
	Secondary  string `mapstructure:"secondary" json:"secondary" yaml:"secondary"`
	Background string `mapstructure:"background" json:"background" yaml:"background"`
	Foreground string `mapstructure:"foreground" json:"foreground" yaml:"foreground"`
	Border     string `mapstructure:"border" json:"border" yaml:"border"`
	Focused    string `mapstructure:"focused" json:"focused" yaml:"focused"`
	Selected   string `mapstructure:"selected" json:"selected" yaml:"selected"`
	Error      string `mapstructure:"error" json:"error" yaml:"error"`
	Success    string `mapstructure:"success" json:"success" yaml:"success"`
}

// LayoutConfig contains responsive layout settings
type LayoutConfig struct {
	MinTerminalWidth  int     `mapstructure:"min_terminal_width" json:"min_terminal_width" yaml:"min_terminal_width"`
	MinTerminalHeight int     `mapstructure:"min_terminal_height" json:"min_terminal_height" yaml:"min_terminal_height"`
	SidebarRatio      float64 `mapstructure:"sidebar_ratio" json:"sidebar_ratio" yaml:"sidebar_ratio"`
	MaxSidebarWidth   int     `mapstructure:"max_sidebar_width" json:"max_sidebar_width" yaml:"max_sidebar_width"`
	MinSidebarWidth   int     `mapstructure:"min_sidebar_width" json:"min_sidebar_width" yaml:"min_sidebar_width"`
}

// SecurityConfig contains security policy settings
type SecurityConfig struct {
	AllowedDirectories []string      `mapstructure:"allowed_directories" json:"allowed_directories" yaml:"allowed_directories"`
	AllowedExtensions  []string      `mapstructure:"allowed_extensions" json:"allowed_extensions" yaml:"allowed_extensions"`
	MaxExecutionTime   time.Duration `mapstructure:"max_execution_time" json:"max_execution_time" yaml:"max_execution_time"`
	MaxOutputSize      int           `mapstructure:"max_output_size" json:"max_output_size" yaml:"max_output_size"`
	RestrictedCommands []string      `mapstructure:"restricted_commands" json:"restricted_commands" yaml:"restricted_commands"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level      string `mapstructure:"level" json:"level" yaml:"level"`
	File       string `mapstructure:"file" json:"file" yaml:"file"`
	MaxSize    int    `mapstructure:"max_size" json:"max_size" yaml:"max_size"`
	MaxBackups int    `mapstructure:"max_backups" json:"max_backups" yaml:"max_backups"`
	MaxAge     int    `mapstructure:"max_age" json:"max_age" yaml:"max_age"`
}

// NewDefaultConfig returns configuration with default values
func NewDefaultConfig() *AppConfig {
	return &AppConfig{
		ScriptDirectories: []string{"./scripts", "~/.local/bin"},
		ScriptExtensions: map[string]string{
			".sh":   "shell",
			".bash": "shell",
			".py":   "python",
			".js":   "node",
			".rb":   "ruby",
			".pl":   "perl",
		},
		Execution: ExecutionConfig{
			Timeout:       5 * time.Minute,
			MaxOutputSize: 1000,
			Shell:         "", // Auto-detect
			WorkingDir:    "", // Use script directory
		},
		UI: UIConfig{
			ShowHidden:       false,
			RefreshOnFocus:   true,
			ConfirmOnExecute: false,
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
		},
		Security: SecurityConfig{
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
	}
}

// Validate checks the configuration for errors and conflicts
func (c *AppConfig) Validate() error {
	if len(c.ScriptDirectories) == 0 {
		return fmt.Errorf("at least one script directory must be configured")
	}

	// Validate script directories
	for _, dir := range c.ScriptDirectories {
		if dir == "" {
			return fmt.Errorf("script directory cannot be empty")
		}

		// Expand tilde if present
		if dir[0] == '~' {
			// This would be handled by the actual config loader
		}

		cleanDir := filepath.Clean(dir)
		if !filepath.IsAbs(cleanDir) && !filepath.IsLocal(cleanDir) {
			return fmt.Errorf("invalid script directory: %s", dir)
		}
	}

	// Validate execution config
	if c.Execution.Timeout <= 0 {
		return fmt.Errorf("execution timeout must be positive")
	}

	if c.Execution.MaxOutputSize <= 0 {
		return fmt.Errorf("max output size must be positive")
	}

	// Validate UI config
	if c.UI.Layout.SidebarRatio <= 0 || c.UI.Layout.SidebarRatio >= 1 {
		return fmt.Errorf("sidebar ratio must be between 0 and 1")
	}

	if c.UI.Layout.MinTerminalWidth < 40 {
		return fmt.Errorf("minimum terminal width must be at least 40")
	}

	if c.UI.Layout.MinTerminalHeight < 10 {
		return fmt.Errorf("minimum terminal height must be at least 10")
	}

	// Validate security config
	if c.Security.MaxExecutionTime <= 0 {
		return fmt.Errorf("max execution time must be positive")
	}

	if c.Security.MaxOutputSize <= 0 {
		return fmt.Errorf("security max output size must be positive")
	}

	return nil
}

// Merge combines this configuration with another, with the other taking precedence
func (c *AppConfig) Merge(other *AppConfig) *AppConfig {
	if other == nil {
		return c.Clone()
	}

	merged := c.Clone()

	// Merge script directories (other takes precedence if not empty)
	if len(other.ScriptDirectories) > 0 {
		merged.ScriptDirectories = make([]string, len(other.ScriptDirectories))
		copy(merged.ScriptDirectories, other.ScriptDirectories)
	}

	// Merge script extensions (combine both, other takes precedence for conflicts)
	if len(other.ScriptExtensions) > 0 {
		for ext, scriptType := range other.ScriptExtensions {
			merged.ScriptExtensions[ext] = scriptType
		}
	}

	// Merge execution config
	if other.Execution.Timeout > 0 {
		merged.Execution.Timeout = other.Execution.Timeout
	}
	if other.Execution.MaxOutputSize > 0 {
		merged.Execution.MaxOutputSize = other.Execution.MaxOutputSize
	}
	if other.Execution.Shell != "" {
		merged.Execution.Shell = other.Execution.Shell
	}
	if other.Execution.WorkingDir != "" {
		merged.Execution.WorkingDir = other.Execution.WorkingDir
	}

	// Merge other configs...
	// (Implementation would continue for all fields)

	return merged
}

// Clone creates a deep copy of the configuration
func (c *AppConfig) Clone() *AppConfig {
	clone := *c

	// Deep copy slices and maps
	clone.ScriptDirectories = make([]string, len(c.ScriptDirectories))
	copy(clone.ScriptDirectories, c.ScriptDirectories)

	clone.ScriptExtensions = make(map[string]string)
	for k, v := range c.ScriptExtensions {
		clone.ScriptExtensions[k] = v
	}

	clone.Security.AllowedDirectories = make([]string, len(c.Security.AllowedDirectories))
	copy(clone.Security.AllowedDirectories, c.Security.AllowedDirectories)

	clone.Security.AllowedExtensions = make([]string, len(c.Security.AllowedExtensions))
	copy(clone.Security.AllowedExtensions, c.Security.AllowedExtensions)

	clone.Security.RestrictedCommands = make([]string, len(c.Security.RestrictedCommands))
	copy(clone.Security.RestrictedCommands, c.Security.RestrictedCommands)

	return &clone
}