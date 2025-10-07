package services

import (
	"github.com/shaiu/alec/pkg/contracts"
	"github.com/shaiu/alec/pkg/models"
)

// ServiceRegistry provides centralized access to all services
type ServiceRegistry struct {
	ConfigManager     contracts.ConfigManager
	ScriptDiscovery   contracts.ScriptDiscovery
	ScriptExecutor    contracts.ScriptExecutor
	SecurityValidator *SecurityValidator
}

// NewServiceRegistry creates a new service registry with all services initialized
func NewServiceRegistry() (*ServiceRegistry, error) {
	// Initialize config manager first
	configManager := NewConfigManagerService()

	// Load configuration
	config, err := configManager.LoadConfig()
	if err != nil {
		return nil, err
	}

	// Create security validator
	securityValidator := NewSecurityValidator(
		config.ScriptDirectories,
		config.Security.AllowedExtensions,
	)

	// Initialize script discovery service
	scriptDiscovery := NewScriptDiscoveryService(
		config.ScriptDirectories,
		config.ScriptExtensions,
	)

	// Initialize script executor service
	executionConfig := &models.ExecutionConfig{
		Timeout:       config.Execution.Timeout,
		MaxOutputSize: config.Execution.MaxOutputSize,
		Shell:         config.Execution.Shell,
		WorkingDir:    config.Execution.WorkingDir,
	}
	scriptExecutor := NewScriptExecutorService(securityValidator, executionConfig)

	return &ServiceRegistry{
		ConfigManager:     configManager,
		ScriptDiscovery:   scriptDiscovery,
		ScriptExecutor:    scriptExecutor,
		SecurityValidator: securityValidator,
	}, nil
}

// GetScriptDiscovery returns the script discovery service
func (sr *ServiceRegistry) GetScriptDiscovery() contracts.ScriptDiscovery {
	return sr.ScriptDiscovery
}

// GetScriptExecutor returns the script executor service
func (sr *ServiceRegistry) GetScriptExecutor() contracts.ScriptExecutor {
	return sr.ScriptExecutor
}

// GetConfigManager returns the configuration manager service
func (sr *ServiceRegistry) GetConfigManager() contracts.ConfigManager {
	return sr.ConfigManager
}

// Reload reloads all services with updated configuration
func (sr *ServiceRegistry) Reload() error {
	// Reload configuration
	config, err := sr.ConfigManager.LoadConfig()
	if err != nil {
		return err
	}

	// Update security validator
	sr.SecurityValidator = NewSecurityValidator(
		config.ScriptDirectories,
		config.Security.AllowedExtensions,
	)

	// Recreate script discovery with new config
	sr.ScriptDiscovery = NewScriptDiscoveryService(
		config.ScriptDirectories,
		config.ScriptExtensions,
	)

	// Recreate script executor with new config
	executionConfig := &models.ExecutionConfig{
		Timeout:       config.Execution.Timeout,
		MaxOutputSize: config.Execution.MaxOutputSize,
		Shell:         config.Execution.Shell,
		WorkingDir:    config.Execution.WorkingDir,
	}
	sr.ScriptExecutor = NewScriptExecutorService(sr.SecurityValidator, executionConfig)

	return nil
}