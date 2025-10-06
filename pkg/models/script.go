package models

import (
	"crypto/md5"
	"fmt"
	"path/filepath"
	"time"

	"github.com/your-org/alec/pkg/parser"
)

// ScriptStatus represents the current state of a script
type ScriptStatus string

const (
	StatusDiscovered ScriptStatus = "discovered"
	StatusValidated  ScriptStatus = "validated"
	StatusReady      ScriptStatus = "ready"
	StatusExecuting  ScriptStatus = "executing"
	StatusCompleted  ScriptStatus = "completed"
	StatusFailed     ScriptStatus = "failed"
	StatusRefreshing ScriptStatus = "refreshing"
	StatusRemoved    ScriptStatus = "removed"
)

// Script represents an executable script file with associated metadata
type Script struct {
	ID             string               `json:"id"`
	Name           string               `json:"name"`
	Path           string               `json:"path"`
	Type           string               `json:"type"`
	Size           int64                `json:"size"`
	ModifiedTime   time.Time            `json:"modified_time"`
	Permissions    string               `json:"permissions"`
	IsExecutable   bool                 `json:"is_executable"`
	Description    string               `json:"description,omitempty"`
	Tags           []string             `json:"tags,omitempty"`
	Status         ScriptStatus         `json:"status"`
	LastAccessed   *time.Time           `json:"last_accessed,omitempty"`
	ExecutionCount int                  `json:"execution_count"`
	Metadata       *parser.ScriptMetadata `json:"metadata,omitempty"`
}

// NewScript creates a new Script instance with generated ID
func NewScript(path string) *Script {
	name := filepath.Base(path)
	// Remove extension from name
	ext := filepath.Ext(name)
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}

	return &Script{
		ID:             generateScriptID(path),
		Name:           name,
		Path:           filepath.Clean(path),
		Status:         StatusDiscovered,
		ExecutionCount: 0,
	}
}

// generateScriptID creates a unique ID based on path and modification time
func generateScriptID(path string) string {
	hash := md5.Sum([]byte(path))
	return fmt.Sprintf("script_%x", hash[:8])
}

// Validate checks if the script meets all validation requirements
func (s *Script) Validate() error {
	if s.Path == "" {
		return fmt.Errorf("script path cannot be empty")
	}

	if !filepath.IsAbs(s.Path) {
		return fmt.Errorf("script path must be absolute: %s", s.Path)
	}

	// Path traversal check using Go 1.21+ feature
	if !filepath.IsLocal(s.Path) {
		return fmt.Errorf("script path contains path traversal: %s", s.Path)
	}

	if s.Type == "" {
		return fmt.Errorf("script type must be specified")
	}

	return nil
}

// UpdateStatus transitions the script to a new status
func (s *Script) UpdateStatus(newStatus ScriptStatus) error {
	// Validate status transitions
	switch s.Status {
	case StatusDiscovered:
		if newStatus != StatusValidated && newStatus != StatusRemoved {
			return fmt.Errorf("invalid status transition from %s to %s", s.Status, newStatus)
		}
	case StatusValidated:
		if newStatus != StatusReady && newStatus != StatusRemoved {
			return fmt.Errorf("invalid status transition from %s to %s", s.Status, newStatus)
		}
	case StatusReady:
		if newStatus != StatusExecuting && newStatus != StatusRefreshing && newStatus != StatusRemoved {
			return fmt.Errorf("invalid status transition from %s to %s", s.Status, newStatus)
		}
	case StatusExecuting:
		if newStatus != StatusCompleted && newStatus != StatusFailed {
			return fmt.Errorf("invalid status transition from %s to %s", s.Status, newStatus)
		}
	case StatusCompleted, StatusFailed:
		if newStatus != StatusReady && newStatus != StatusRemoved {
			return fmt.Errorf("invalid status transition from %s to %s", s.Status, newStatus)
		}
	case StatusRefreshing:
		if newStatus != StatusReady && newStatus != StatusRemoved {
			return fmt.Errorf("invalid status transition from %s to %s", s.Status, newStatus)
		}
	}

	s.Status = newStatus
	return nil
}

// MarkAccessed updates the last accessed time
func (s *Script) MarkAccessed() {
	now := time.Now()
	s.LastAccessed = &now
}

// IncrementExecutionCount increments the execution counter
func (s *Script) IncrementExecutionCount() {
	s.ExecutionCount++
}

// GetTypeFromExtension determines script type from file extension
func GetTypeFromExtension(path string) string {
	ext := filepath.Ext(path)

	typeMap := map[string]string{
		".sh":   "shell",
		".bash": "shell",
		".zsh":  "shell",
		".py":   "python",
		".js":   "node",
		".ts":   "node",
		".rb":   "ruby",
		".pl":   "perl",
		".php":  "php",
		".go":   "go",
		".rs":   "rust",
	}

	return typeMap[ext]
}

// IsSupported checks if a file extension is supported
func IsSupported(path string) bool {
	return GetTypeFromExtension(path) != ""
}

// SetDescription extracts description from script comments
func (s *Script) SetDescription(description string) {
	// Limit description length
	if len(description) > 200 {
		description = description[:200] + "..."
	}
	s.Description = description
}

// AddTag adds a tag to the script
func (s *Script) AddTag(tag string) {
	// Check if tag already exists
	for _, existingTag := range s.Tags {
		if existingTag == tag {
			return
		}
	}
	s.Tags = append(s.Tags, tag)
}

// RemoveTag removes a tag from the script
func (s *Script) RemoveTag(tag string) {
	for i, existingTag := range s.Tags {
		if existingTag == tag {
			s.Tags = append(s.Tags[:i], s.Tags[i+1:]...)
			return
		}
	}
}

// HasTag checks if the script has a specific tag
func (s *Script) HasTag(tag string) bool {
	for _, existingTag := range s.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the script
func (s *Script) Clone() *Script {
	clone := *s

	// Deep copy slices
	if s.Tags != nil {
		clone.Tags = make([]string, len(s.Tags))
		copy(clone.Tags, s.Tags)
	}

	// Deep copy pointer fields
	if s.LastAccessed != nil {
		lastAccessed := *s.LastAccessed
		clone.LastAccessed = &lastAccessed
	}

	return &clone
}