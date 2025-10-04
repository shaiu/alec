// Contract: Script Discovery Interface
// This file defines the contract for script discovery and management
// Implementation must pass all associated tests

package contracts

import (
	"context"
	"time"
)

// ScriptInfo represents metadata about a discovered script
type ScriptInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Type         string    `json:"type"`
	Size         int64     `json:"size"`
	ModifiedTime time.Time `json:"modified_time"`
	IsExecutable bool      `json:"is_executable"`
	Description  string    `json:"description,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}

// DirectoryInfo represents a directory in the script hierarchy
type DirectoryInfo struct {
	Path        string          `json:"path"`
	Name        string          `json:"name"`
	Children    []DirectoryInfo `json:"children,omitempty"`
	Scripts     []ScriptInfo    `json:"scripts,omitempty"`
	ScriptCount int             `json:"script_count"`
	LastScan    time.Time       `json:"last_scan"`
}

// ScriptDiscovery interface defines the contract for script discovery operations
type ScriptDiscovery interface {
	// ScanDirectories scans configured directories for executable scripts
	// Returns directory tree with discovered scripts
	// Must handle security validation and path restrictions
	ScanDirectories(ctx context.Context, directories []string) ([]DirectoryInfo, error)

	// ValidateScript checks if a script is valid and executable
	// Returns error if script fails validation
	ValidateScript(scriptPath string) (*ScriptInfo, error)

	// RefreshScript updates metadata for a single script
	// Returns updated script info or error if script no longer exists
	RefreshScript(scriptPath string) (*ScriptInfo, error)

	// FilterScripts filters scripts based on query string
	// Supports name matching, type filtering, and tag filtering
	FilterScripts(scripts []ScriptInfo, query string) []ScriptInfo

	// WatchDirectory monitors directory for changes (future enhancement)
	// Returns channel of change events
	WatchDirectory(ctx context.Context, dirPath string) (<-chan DirectoryChange, error)
}

// DirectoryChange represents a change event in a watched directory
type DirectoryChange struct {
	Type      ChangeType `json:"type"`
	Path      string     `json:"path"`
	Timestamp time.Time  `json:"timestamp"`
}

// ChangeType represents the type of directory change
type ChangeType string

const (
	ChangeTypeCreate ChangeType = "create"
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeDelete ChangeType = "delete"
	ChangeTypeRename ChangeType = "rename"
)

// Contract Requirements:
// 1. ScanDirectories MUST return consistent results for unchanged directories
// 2. ValidateScript MUST prevent path traversal attacks
// 3. All paths MUST be absolute and cleaned
// 4. Script types MUST be determined by file extension
// 5. IsExecutable MUST reflect actual file permissions
// 6. Scanning MUST be interruptible via context cancellation
// 7. Large directories MUST be handled efficiently (pagination/streaming)
// 8. Error messages MUST be user-friendly and actionable