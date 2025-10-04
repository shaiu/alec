package services

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
	"github.com/your-org/alec/pkg/contracts"
	"github.com/your-org/alec/pkg/models"
)

// ConfigManagerService implements the ConfigManager contract
type ConfigManagerService struct {
	configPath string
	viper      *viper.Viper
}

// NewConfigManagerService creates a new configuration manager
func NewConfigManagerService() *ConfigManagerService {
	v := viper.New()
	v.SetConfigName("alec")
	v.SetConfigType("yaml")

	// Set up config search paths
	configPath := getConfigPath()
	configDir := filepath.Dir(configPath)
	v.AddConfigPath(configDir)
	v.AddConfigPath(".")

	// Environment variables
	v.SetEnvPrefix("ALEC")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v)

	return &ConfigManagerService{
		configPath: configPath,
		viper:      v,
	}
}

// LoadConfig loads configuration from file and environment
func (cm *ConfigManagerService) LoadConfig() (*contracts.AppConfig, error) {
	// Try to read config file (don't fail if it doesn't exist or is corrupted)
	var config models.AppConfig

	if err := cm.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file exists but is corrupted, log warning and use defaults
			fmt.Printf("Warning: Config file corrupted, using defaults: %v\n", err)
		}
		// Use default configuration
		defaultConfig := models.NewDefaultConfig()
		config = *defaultConfig
	} else {
		// Try to unmarshal config file
		if err := cm.viper.Unmarshal(&config); err != nil {
			// Unmarshal failed, use defaults
			fmt.Printf("Warning: Failed to parse config, using defaults: %v\n", err)
			defaultConfig := models.NewDefaultConfig()
			config = *defaultConfig
		}
	}

	// Ensure we have at least default values for critical fields
	if len(config.ScriptDirectories) == 0 {
		config.ScriptDirectories = []string{"./scripts", "~/.local/bin"}
	}
	if len(config.ScriptExtensions) == 0 {
		config.ScriptExtensions = map[string]string{
			".sh":   "shell",
			".bash": "shell",
			".py":   "python",
			".js":   "node",
			".rb":   "ruby",
			".pl":   "perl",
		}
	}

	// Convert to contract interface
	appConfig := &contracts.AppConfig{
		ScriptDirectories: config.ScriptDirectories,
		ScriptExtensions:  config.ScriptExtensions,
		Execution:         convertExecutionConfig(config.Execution),
		UI:                convertUIConfig(config.UI),
		Security:          convertSecurityConfig(config.Security),
		Logging:           convertLoggingConfig(config.Logging),
	}

	return appConfig, nil
}

// SaveConfig writes current configuration to file
func (cm *ConfigManagerService) SaveConfig(config *contracts.AppConfig) error {
	// Ensure config directory exists
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert from contract to models
	modelConfig := &models.AppConfig{
		ScriptDirectories: config.ScriptDirectories,
		ScriptExtensions:  config.ScriptExtensions,
		Execution:         convertFromExecutionConfig(config.Execution),
		UI:                convertFromUIConfig(config.UI),
		Security:          convertFromSecurityConfig(config.Security),
		Logging:           convertFromLoggingConfig(config.Logging),
	}

	// Set all config values in viper
	cm.viper.Set("script_dirs", modelConfig.ScriptDirectories)
	cm.viper.Set("extensions", modelConfig.ScriptExtensions)
	cm.viper.Set("execution", modelConfig.Execution)
	cm.viper.Set("ui", modelConfig.UI)
	cm.viper.Set("security", modelConfig.Security)
	cm.viper.Set("logging", modelConfig.Logging)

	// Write config file
	if err := cm.viper.WriteConfigAs(cm.configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Set secure permissions
	if err := os.Chmod(cm.configPath, 0600); err != nil {
		return fmt.Errorf("failed to set config file permissions: %w", err)
	}

	return nil
}

// GetDefaultConfig returns configuration with default values
func (cm *ConfigManagerService) GetDefaultConfig() *contracts.AppConfig {
	defaultModel := models.NewDefaultConfig()

	return &contracts.AppConfig{
		ScriptDirectories: defaultModel.ScriptDirectories,
		ScriptExtensions:  defaultModel.ScriptExtensions,
		Execution:         convertExecutionConfig(defaultModel.Execution),
		UI:                convertUIConfig(defaultModel.UI),
		Security:          convertSecurityConfig(defaultModel.Security),
		Logging:           convertLoggingConfig(defaultModel.Logging),
	}
}

// ValidateConfig checks configuration for errors and conflicts
func (cm *ConfigManagerService) ValidateConfig(config *contracts.AppConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if len(config.ScriptDirectories) == 0 {
		return fmt.Errorf("at least one script directory must be configured")
	}

	// Validate script directories exist and are accessible
	for _, dir := range config.ScriptDirectories {
		if dir == "" {
			return fmt.Errorf("script directory cannot be empty")
		}

		// Expand home directory
		if strings.HasPrefix(dir, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cannot expand home directory in path %s: %w", dir, err)
			}
			dir = filepath.Join(home, dir[2:])
		}

		// Check if directory exists (but don't fail if it doesn't - might be created later)
		if _, err := os.Stat(dir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("cannot access script directory %s: %w", dir, err)
		}
	}

	// Validate execution config
	if config.Execution.Timeout <= 0 {
		return fmt.Errorf("execution timeout must be positive")
	}

	if config.Execution.MaxOutputSize <= 0 {
		return fmt.Errorf("max output size must be positive")
	}

	// Validate security config
	if config.Security.MaxExecutionTime <= 0 {
		return fmt.Errorf("max execution time must be positive")
	}

	if config.Security.MaxOutputSize <= 0 {
		return fmt.Errorf("security max output size must be positive")
	}

	return nil
}

// WatchConfig monitors configuration file for changes
func (cm *ConfigManagerService) WatchConfig() (<-chan *contracts.AppConfig, error) {
	// This would use fsnotify in a real implementation
	ch := make(chan *contracts.AppConfig)

	// For now, just return a channel that never sends
	go func() {
		// Would watch for file changes and reload config
		close(ch)
	}()

	return ch, nil
}

// GetConfigPath returns the current configuration file path
func (cm *ConfigManagerService) GetConfigPath() string {
	return cm.configPath
}

// MergeConfig combines configuration from multiple sources
func (cm *ConfigManagerService) MergeConfig(configs ...*contracts.AppConfig) *contracts.AppConfig {
	if len(configs) == 0 {
		return cm.GetDefaultConfig()
	}

	result := cm.GetDefaultConfig()

	for _, config := range configs {
		if config == nil {
			continue
		}

		// Merge each field, with later configs taking precedence
		if len(config.ScriptDirectories) > 0 {
			result.ScriptDirectories = make([]string, len(config.ScriptDirectories))
			copy(result.ScriptDirectories, config.ScriptDirectories)
		}

		if len(config.ScriptExtensions) > 0 {
			if result.ScriptExtensions == nil {
				result.ScriptExtensions = make(map[string]string)
			}
			for k, v := range config.ScriptExtensions {
				result.ScriptExtensions[k] = v
			}
		}

		// Merge other fields...
		if config.Execution.Timeout > 0 {
			result.Execution.Timeout = config.Execution.Timeout
		}
		if config.Execution.MaxOutputSize > 0 {
			result.Execution.MaxOutputSize = config.Execution.MaxOutputSize
		}
		if config.Execution.Shell != "" {
			result.Execution.Shell = config.Execution.Shell
		}
	}

	return result
}

// Helper functions

// getConfigPath returns the OS-appropriate config file path
func getConfigPath() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = os.Getenv("USERPROFILE")
		}
		return filepath.Join(appData, "alec", "alec.yaml")
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "alec", "alec.yaml")
	default: // Linux and other Unix-like
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, _ := os.UserHomeDir()
			configHome = filepath.Join(home, ".config")
		}
		return filepath.Join(configHome, "alec", "alec.yaml")
	}
}

// setDefaults sets default values in viper
func setDefaults(v *viper.Viper) {
	defaults := models.NewDefaultConfig()

	v.SetDefault("script_dirs", defaults.ScriptDirectories)
	v.SetDefault("extensions", defaults.ScriptExtensions)
	v.SetDefault("execution.timeout", defaults.Execution.Timeout)
	v.SetDefault("execution.max_output_size", defaults.Execution.MaxOutputSize)
	v.SetDefault("execution.shell", defaults.Execution.Shell)
	v.SetDefault("ui.show_hidden", defaults.UI.ShowHidden)
	v.SetDefault("security.max_execution_time", defaults.Security.MaxExecutionTime)
	v.SetDefault("security.max_output_size", defaults.Security.MaxOutputSize)
	v.SetDefault("logging.level", defaults.Logging.Level)
}

// Conversion functions between models and contracts
func convertExecutionConfig(config models.ExecutionConfig) contracts.ExecutionConfig {
	return contracts.ExecutionConfig{
		Timeout:       config.Timeout,
		MaxOutputSize: config.MaxOutputSize,
		Shell:         config.Shell,
		WorkingDir:    config.WorkingDir,
	}
}

func convertFromExecutionConfig(config contracts.ExecutionConfig) models.ExecutionConfig {
	return models.ExecutionConfig{
		Timeout:       config.Timeout,
		MaxOutputSize: config.MaxOutputSize,
		Shell:         config.Shell,
		WorkingDir:    config.WorkingDir,
	}
}

func convertUIConfig(config models.UIConfig) contracts.UIConfig {
	return contracts.UIConfig{
		ShowHidden:       config.ShowHidden,
		RefreshOnFocus:   config.RefreshOnFocus,
		ConfirmOnExecute: config.ConfirmOnExecute,
		Theme:            convertThemeConfig(config.Theme),
		Layout:           convertLayoutConfig(config.Layout),
	}
}

func convertFromUIConfig(config contracts.UIConfig) models.UIConfig {
	return models.UIConfig{
		ShowHidden:       config.ShowHidden,
		RefreshOnFocus:   config.RefreshOnFocus,
		ConfirmOnExecute: config.ConfirmOnExecute,
		Theme:            convertFromThemeConfig(config.Theme),
		Layout:           convertFromLayoutConfig(config.Layout),
	}
}

func convertThemeConfig(config models.ThemeConfig) contracts.ThemeConfig {
	return contracts.ThemeConfig{
		Primary:    config.Primary,
		Secondary:  config.Secondary,
		Background: config.Background,
		Foreground: config.Foreground,
		Border:     config.Border,
		Focused:    config.Focused,
		Selected:   config.Selected,
		Error:      config.Error,
		Success:    config.Success,
	}
}

func convertFromThemeConfig(config contracts.ThemeConfig) models.ThemeConfig {
	return models.ThemeConfig{
		Primary:    config.Primary,
		Secondary:  config.Secondary,
		Background: config.Background,
		Foreground: config.Foreground,
		Border:     config.Border,
		Focused:    config.Focused,
		Selected:   config.Selected,
		Error:      config.Error,
		Success:    config.Success,
	}
}

func convertLayoutConfig(config models.LayoutConfig) contracts.LayoutConfig {
	return contracts.LayoutConfig{
		MinTerminalWidth:  config.MinTerminalWidth,
		MinTerminalHeight: config.MinTerminalHeight,
		SidebarRatio:      config.SidebarRatio,
		MaxSidebarWidth:   config.MaxSidebarWidth,
		MinSidebarWidth:   config.MinSidebarWidth,
	}
}

func convertFromLayoutConfig(config contracts.LayoutConfig) models.LayoutConfig {
	return models.LayoutConfig{
		MinTerminalWidth:  config.MinTerminalWidth,
		MinTerminalHeight: config.MinTerminalHeight,
		SidebarRatio:      config.SidebarRatio,
		MaxSidebarWidth:   config.MaxSidebarWidth,
		MinSidebarWidth:   config.MinSidebarWidth,
	}
}

func convertSecurityConfig(config models.SecurityConfig) contracts.SecurityPolicy {
	return contracts.SecurityPolicy{
		AllowedDirectories: config.AllowedDirectories,
		AllowedExtensions:  config.AllowedExtensions,
		MaxExecutionTime:   config.MaxExecutionTime,
		MaxOutputSize:      config.MaxOutputSize,
		RestrictedCommands: config.RestrictedCommands,
	}
}

func convertFromSecurityConfig(config contracts.SecurityPolicy) models.SecurityConfig {
	return models.SecurityConfig{
		AllowedDirectories: config.AllowedDirectories,
		AllowedExtensions:  config.AllowedExtensions,
		MaxExecutionTime:   config.MaxExecutionTime,
		MaxOutputSize:      config.MaxOutputSize,
		RestrictedCommands: config.RestrictedCommands,
	}
}

func convertLoggingConfig(config models.LoggingConfig) contracts.LoggingConfig {
	return contracts.LoggingConfig{
		Level:      config.Level,
		File:       config.File,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
	}
}

func convertFromLoggingConfig(config contracts.LoggingConfig) models.LoggingConfig {
	return models.LoggingConfig{
		Level:      config.Level,
		File:       config.File,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
	}
}
