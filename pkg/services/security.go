package services

import (
	"fmt"
	"path/filepath"
	"strings"
)

// SecurityValidator provides path validation and permission checks
type SecurityValidator struct {
	allowedDirs []string
	allowedExts []string
}

// NewSecurityValidator creates a new security validator
func NewSecurityValidator(allowedDirs, allowedExts []string) *SecurityValidator {
	return &SecurityValidator{
		allowedDirs: allowedDirs,
		allowedExts: allowedExts,
	}
}

// ValidateScriptPath validates a script path against security policies
func (sv *SecurityValidator) ValidateScriptPath(path string) error {
	// Clean and resolve the path
	cleanPath := filepath.Clean(path)

	// For absolute paths, use different validation logic
	if filepath.IsAbs(cleanPath) {
		// For absolute paths, just check they don't contain suspicious elements
		if strings.Contains(cleanPath, "..") || strings.Contains(cleanPath, "./") {
			return fmt.Errorf("path traversal detected: %s", path)
		}
	} else {
		// For relative paths, use Go 1.21+ traversal-resistant API
		if !filepath.IsLocal(cleanPath) {
			return fmt.Errorf("path traversal detected: %s", path)
		}
	}

	// Convert to absolute path for validation
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Validate against allowed directories
	if len(sv.allowedDirs) > 0 {
		allowed := false
		for _, dir := range sv.allowedDirs {
			absDir, err := filepath.Abs(dir)
			if err != nil {
				continue
			}

			// Check if path is under allowed directory
			if strings.HasPrefix(absPath, absDir) {
				// Ensure it's actually under the directory, not just a prefix match
				if absPath == absDir || strings.HasPrefix(absPath, absDir+string(filepath.Separator)) {
					allowed = true
					break
				}
			}
		}

		if !allowed {
			return fmt.Errorf("path not in allowed directories: %s", path)
		}
	}

	// Validate file extension if it's a file
	if !strings.HasSuffix(cleanPath, string(filepath.Separator)) {
		ext := filepath.Ext(cleanPath)
		if ext != "" && len(sv.allowedExts) > 0 {
			allowed := false
			for _, allowedExt := range sv.allowedExts {
				if ext == allowedExt {
					allowed = true
					break
				}
			}

			if !allowed {
				return fmt.Errorf("file extension not allowed: %s", ext)
			}
		}
	}

	return nil
}

// IsPathAllowed checks if a path is allowed without detailed error
func (sv *SecurityValidator) IsPathAllowed(path string) bool {
	return sv.ValidateScriptPath(path) == nil
}

// SanitizeArgs removes potentially dangerous arguments
func (sv *SecurityValidator) SanitizeArgs(args []string) []string {
	safe := make([]string, 0, len(args))

	for _, arg := range args {
		// Basic sanitization - remove shell metacharacters
		if !strings.ContainsAny(arg, ";|&`$(){}[]<>*?~") {
			safe = append(safe, arg)
		}
	}

	return safe
}