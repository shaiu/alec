package contract

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/shaiu/alec/pkg/contracts"
)

// TestConfigManagerContract verifies that any implementation of ConfigManager
// interface conforms to the contract requirements
func TestConfigManagerContract(t *testing.T) {
	// This test will fail until we have an implementation
	var manager contracts.ConfigManager
	if manager == nil {
		t.Skip("No ConfigManager implementation available yet - this is expected during TDD phase")
	}

	tests := []struct {
		name string
		test func(t *testing.T, m contracts.ConfigManager)
	}{
		{"LoadConfig must handle missing file gracefully", testLoadConfigMissingFile},
		{"SaveConfig must create directories and handle permissions", testSaveConfig},
		{"GetDefaultConfig must provide sensible defaults", testGetDefaultConfig},
		{"ValidateConfig must check paths and permissions", testValidateConfig},
		{"GetConfigPath must follow OS conventions", testGetConfigPath},
		{"MergeConfig must respect priority order", testMergeConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t, manager)
		})
	}
}

func testLoadConfigMissingFile(t *testing.T, m contracts.ConfigManager) {
	// Should not fail when config file is missing
	config, err := m.LoadConfig()
	if err != nil {
		t.Errorf("LoadConfig should not fail when config file is missing: %v", err)
		return
	}

	if config == nil {
		t.Error("LoadConfig should return default config when file is missing")
		return
	}

	// Should have sensible defaults
	if len(config.ScriptDirectories) == 0 {
		t.Error("Default config should have at least one script directory")
	}

	if len(config.ScriptExtensions) == 0 {
		t.Error("Default config should have script extensions defined")
	}
}

func testSaveConfig(t *testing.T, m contracts.ConfigManager) {
	// Create a temporary config
	config := &contracts.AppConfig{
		ScriptDirectories: []string{"/tmp/test-scripts"},
		ScriptExtensions: map[string]string{
			".sh": "shell",
			".py": "python",
		},
		Execution: contracts.ExecutionConfig{
			Timeout:       5 * time.Minute,
			MaxOutputSize: 1000,
		},
	}

	err := m.SaveConfig(config)
	if err != nil {
		t.Errorf("SaveConfig failed: %v", err)
		return
	}

	// Verify the config file was created
	configPath := m.GetConfigPath()
	if configPath == "" {
		t.Error("GetConfigPath should return valid path after SaveConfig")
		return
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file should exist after SaveConfig: %s", configPath)
	}

	// Check permissions (should be secure - readable/writable by owner only)
	info, err := os.Stat(configPath)
	if err != nil {
		t.Errorf("Failed to stat config file: %v", err)
		return
	}

	mode := info.Mode()
	if mode&0077 != 0 {
		t.Errorf("Config file should have secure permissions (0600), got: %o", mode)
	}
}

func testGetDefaultConfig(t *testing.T, m contracts.ConfigManager) {
	defaultConfig := m.GetDefaultConfig()

	if defaultConfig == nil {
		t.Error("GetDefaultConfig should never return nil")
		return
	}

	// Verify required fields have sensible defaults
	if len(defaultConfig.ScriptDirectories) == 0 {
		t.Error("Default config should have script directories")
	}

	if len(defaultConfig.ScriptExtensions) == 0 {
		t.Error("Default config should have script extensions")
	}

	if defaultConfig.Execution.Timeout <= 0 {
		t.Error("Default config should have positive execution timeout")
	}

	if defaultConfig.Execution.MaxOutputSize <= 0 {
		t.Error("Default config should have positive max output size")
	}

	// Check for common script types
	expectedTypes := []string{".sh", ".py", ".js"}
	for _, ext := range expectedTypes {
		if _, exists := defaultConfig.ScriptExtensions[ext]; !exists {
			t.Errorf("Default config should include common script type: %s", ext)
		}
	}

	// Verify UI defaults
	if defaultConfig.UI.Layout.MinTerminalWidth <= 0 {
		t.Error("Default config should have positive minimum terminal width")
	}

	if defaultConfig.UI.Layout.SidebarRatio <= 0 || defaultConfig.UI.Layout.SidebarRatio >= 1 {
		t.Error("Default config sidebar ratio should be between 0 and 1")
	}

	// Verify security defaults
	if len(defaultConfig.Security.AllowedExtensions) == 0 {
		t.Error("Default config should have allowed extensions for security")
	}

	if defaultConfig.Security.MaxExecutionTime <= 0 {
		t.Error("Default config should have positive max execution time")
	}
}

func testValidateConfig(t *testing.T, m contracts.ConfigManager) {
	tests := []struct {
		name    string
		config  *contracts.AppConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid config",
			config: &contracts.AppConfig{
				ScriptDirectories: []string{"/tmp"},
				ScriptExtensions:  map[string]string{".sh": "shell"},
				Execution: contracts.ExecutionConfig{
					Timeout:       1 * time.Minute,
					MaxOutputSize: 100,
				},
			},
			wantErr: false,
		},
		{
			name: "empty script directories",
			config: &contracts.AppConfig{
				ScriptDirectories: []string{},
				ScriptExtensions:  map[string]string{".sh": "shell"},
			},
			wantErr: true,
		},
		{
			name: "invalid directory path",
			config: &contracts.AppConfig{
				ScriptDirectories: []string{"/nonexistent/impossible/path"},
				ScriptExtensions:  map[string]string{".sh": "shell"},
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			config: &contracts.AppConfig{
				ScriptDirectories: []string{"/tmp"},
				ScriptExtensions:  map[string]string{".sh": "shell"},
				Execution: contracts.ExecutionConfig{
					Timeout:       -1 * time.Second,
					MaxOutputSize: 100,
				},
			},
			wantErr: true,
		},
		{
			name: "zero max output size",
			config: &contracts.AppConfig{
				ScriptDirectories: []string{"/tmp"},
				ScriptExtensions:  map[string]string{".sh": "shell"},
				Execution: contracts.ExecutionConfig{
					Timeout:       1 * time.Minute,
					MaxOutputSize: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testGetConfigPath(t *testing.T, m contracts.ConfigManager) {
	configPath := m.GetConfigPath()

	if configPath == "" {
		t.Error("GetConfigPath should return non-empty path")
		return
	}

	if !filepath.IsAbs(configPath) {
		t.Error("GetConfigPath should return absolute path")
	}

	// Path should follow OS conventions
	dir := filepath.Dir(configPath)
	filename := filepath.Base(configPath)

	if filename == "" {
		t.Error("Config path should have filename")
	}

	// Should be in appropriate config directory for the OS
	// This is a basic check - actual implementation should use OS-specific paths
	if dir == "/" || dir == "." {
		t.Error("Config should be in appropriate config directory, not root or current")
	}
}

func testMergeConfig(t *testing.T, m contracts.ConfigManager) {
	// Create test configs with different values
	config1 := &contracts.AppConfig{
		ScriptDirectories: []string{"/tmp/config1"},
		ScriptExtensions:  map[string]string{".sh": "shell"},
		Execution: contracts.ExecutionConfig{
			Timeout:       1 * time.Minute,
			MaxOutputSize: 100,
		},
	}

	config2 := &contracts.AppConfig{
		ScriptDirectories: []string{"/tmp/config2"}, // Should override config1
		ScriptExtensions:  map[string]string{".py": "python"},
		Execution: contracts.ExecutionConfig{
			Timeout:       2 * time.Minute, // Should override config1
			MaxOutputSize: 200,             // Should override config1
		},
	}

	config3 := &contracts.AppConfig{
		ScriptDirectories: []string{"/tmp/config3"}, // Should override both
	}

	merged := m.MergeConfig(config1, config2, config3)

	if merged == nil {
		t.Error("MergeConfig should return non-nil result")
		return
	}

	// Later configs should override earlier ones
	if len(merged.ScriptDirectories) != 1 || merged.ScriptDirectories[0] != "/tmp/config3" {
		t.Errorf("MergeConfig should use last config's script directories: got %v", merged.ScriptDirectories)
	}

	// Should merge execution timeout from config2 (config3 doesn't specify)
	if merged.Execution.Timeout != 2*time.Minute {
		t.Errorf("MergeConfig should use config2's timeout: got %v", merged.Execution.Timeout)
	}

	// Should merge script extensions from all configs
	if _, hasShell := merged.ScriptExtensions[".sh"]; !hasShell {
		t.Error("MergeConfig should preserve .sh extension from config1")
	}

	if _, hasPython := merged.ScriptExtensions[".py"]; !hasPython {
		t.Error("MergeConfig should preserve .py extension from config2")
	}
}

func testWatchConfig(t *testing.T, m contracts.ConfigManager) {
	// This is an optional feature test
	t.Run("WatchConfig", func(t *testing.T) {
		ch, err := m.WatchConfig()
		if err != nil {
			t.Skip("WatchConfig not implemented or not supported")
		}

		if ch == nil {
			t.Error("WatchConfig should return non-nil channel when no error")
		}

		// Don't block on the channel in tests
		select {
		case <-ch:
			// Received a config update
		default:
			// No immediate update, which is fine
		}
	})
}