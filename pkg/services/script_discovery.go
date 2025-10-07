package services

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/shaiu/alec/pkg/contracts"
	"github.com/shaiu/alec/pkg/models"
	"github.com/shaiu/alec/pkg/parser"
)

// ScriptDiscoveryService implements the ScriptDiscovery contract
type ScriptDiscoveryService struct {
	allowedDirs       []string
	supportedTypes    map[string]string
	securityValidator *SecurityValidator
}

// NewScriptDiscoveryService creates a new script discovery service
func NewScriptDiscoveryService(allowedDirs []string, supportedTypes map[string]string) *ScriptDiscoveryService {
	return &ScriptDiscoveryService{
		allowedDirs:       allowedDirs,
		supportedTypes:    supportedTypes,
		securityValidator: NewSecurityValidator(allowedDirs, getSupportedExtensions(supportedTypes)),
	}
}

// ScanDirectories scans configured directories for executable scripts
func (s *ScriptDiscoveryService) ScanDirectories(ctx context.Context, directories []string) ([]contracts.DirectoryInfo, error) {
	var results []contracts.DirectoryInfo

	for _, dir := range directories {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		dirInfo, err := s.scanSingleDirectory(ctx, dir)
		if err != nil {
			// Log error but continue with other directories
			continue
		}

		if dirInfo != nil {
			results = append(results, *dirInfo)
		}
	}

	return results, nil
}

// scanSingleDirectory scans a single directory and builds the tree structure
func (s *ScriptDiscoveryService) scanSingleDirectory(ctx context.Context, rootPath string) (*contracts.DirectoryInfo, error) {

	// Validate and clean the path
	cleanPath := filepath.Clean(rootPath)

	// Don't validate directory paths using ValidateScriptPath (which is for files)
	// Just do basic path traversal protection
	if !filepath.IsAbs(cleanPath) && !filepath.IsLocal(cleanPath) {
		return nil, fmt.Errorf("path traversal detected: %s", rootPath)
	}

	// Check if directory exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access directory %s: %w", cleanPath, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", cleanPath)
	}

	// Build directory tree
	dirInfo := &contracts.DirectoryInfo{
		Path:        cleanPath,
		Name:        filepath.Base(cleanPath),
		Children:    make([]contracts.DirectoryInfo, 0),
		Scripts:     make([]contracts.ScriptInfo, 0),
		ScriptCount: 0,
		LastScan:    time.Now(),
	}

	err = s.walkDirectory(ctx, cleanPath, dirInfo)
	if err != nil {
		return nil, err
	}

	// No post-processing needed - directory structure built during walk

	return dirInfo, nil
}

// walkDirectory recursively walks a directory and populates the structure
func (s *ScriptDiscoveryService) walkDirectory(ctx context.Context, rootPath string, dirInfo *contracts.DirectoryInfo) error {
	return filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Skip files we can't access
			return nil
		}

		// Skip the root directory itself
		if path == rootPath {
			return nil
		}

		if d.IsDir() {
			// For directories, just do basic path validation
			if !filepath.IsAbs(path) && !filepath.IsLocal(path) {
				return nil // Skip paths with traversal attempts
			}
			// Create subdirectory info
			subDirInfo := contracts.DirectoryInfo{
				Path:        path,
				Name:        filepath.Base(path),
				Children:    make([]contracts.DirectoryInfo, 0),
				Scripts:     make([]contracts.ScriptInfo, 0),
				ScriptCount: 0,
				LastScan:    time.Now(),
			}

			// Find the parent directory in our structure
			parentPath := filepath.Dir(path)
			if parentPath == rootPath {
				// Direct child of root
				dirInfo.Children = append(dirInfo.Children, subDirInfo)
			} else {
				// Find and update parent in tree (simplified for now)
				s.addToParent(dirInfo, path, &subDirInfo)
			}
		} else {
			// For files, validate path security
			if err := s.securityValidator.ValidateScriptPath(path); err != nil {
				return nil // Skip invalid script paths
			}

			// Check if it's a supported script
			if s.isSupported(path) {
				scriptInfo, err := s.createScriptInfo(path)
				if err != nil {
					return nil // Skip files we can't process
				}

				// Collect all scripts in a flat list, will be organized for display later
				dirInfo.Scripts = append(dirInfo.Scripts, *scriptInfo)
				dirInfo.ScriptCount++
			}
		}

		return nil
	})
}

// addToParent adds a subdirectory to its parent in the tree
func (s *ScriptDiscoveryService) addToParent(root *contracts.DirectoryInfo, childPath string, child *contracts.DirectoryInfo) {
	// Simplified implementation - in a full implementation, this would properly traverse the tree
	parentPath := filepath.Dir(childPath)
	if strings.HasPrefix(parentPath, root.Path) {
		root.Children = append(root.Children, *child)
	}
}

// addScriptToParent adds a script to its parent directory in the tree
func (s *ScriptDiscoveryService) addScriptToParent(root *contracts.DirectoryInfo, scriptPath string, script *contracts.ScriptInfo) {
	// Simplified implementation
	parentPath := filepath.Dir(scriptPath)
	if strings.HasPrefix(parentPath, root.Path) {
		root.Scripts = append(root.Scripts, *script)
		root.ScriptCount++
	}
}

// addScriptToCorrectParent adds a script to its proper parent directory
func (s *ScriptDiscoveryService) addScriptToCorrectParent(root *contracts.DirectoryInfo, scriptPath string, script *contracts.ScriptInfo) {
	scriptDir := filepath.Dir(scriptPath)

	// If script is in root directory, add to root
	if scriptDir == root.Path {
		root.Scripts = append(root.Scripts, *script)
		root.ScriptCount++
		return
	}

	// Find the correct subdirectory to add the script to
	s.addScriptToDirectory(root, scriptDir, script)
}

// addScriptToDirectory recursively finds the correct directory and adds the script
func (s *ScriptDiscoveryService) addScriptToDirectory(dir *contracts.DirectoryInfo, targetPath string, script *contracts.ScriptInfo) {
	// Check if this is the target directory
	if dir.Path == targetPath {
		dir.Scripts = append(dir.Scripts, *script)
		dir.ScriptCount++
		return
	}

	// Search in children
	for i := range dir.Children {
		if strings.HasPrefix(targetPath, dir.Children[i].Path) {
			s.addScriptToDirectory(&dir.Children[i], targetPath, script)
			return
		}
	}

	// If we couldn't find the subdirectory, create it
	if strings.HasPrefix(targetPath, dir.Path) {
		// Create missing directory structure
		relPath, _ := filepath.Rel(dir.Path, targetPath)
		parts := strings.Split(relPath, string(filepath.Separator))

		currentDir := dir
		currentPath := dir.Path

		for _, part := range parts {
			if part == "" {
				continue
			}

			currentPath = filepath.Join(currentPath, part)

			// Check if this subdirectory already exists
			found := false
			for i := range currentDir.Children {
				if currentDir.Children[i].Path == currentPath {
					currentDir = &currentDir.Children[i]
					found = true
					break
				}
			}

			// Create new subdirectory if it doesn't exist
			if !found {
				newDir := contracts.DirectoryInfo{
					Path:        currentPath,
					Name:        part,
					Children:    make([]contracts.DirectoryInfo, 0),
					Scripts:     make([]contracts.ScriptInfo, 0),
					ScriptCount: 0,
					LastScan:    dir.LastScan,
				}
				currentDir.Children = append(currentDir.Children, newDir)
				currentDir = &currentDir.Children[len(currentDir.Children)-1]
			}
		}

		// Add script to the final directory
		currentDir.Scripts = append(currentDir.Scripts, *script)
		currentDir.ScriptCount++
	}
}

// ValidateScript checks if a script is valid and executable
func (s *ScriptDiscoveryService) ValidateScript(scriptPath string) (*contracts.ScriptInfo, error) {
	// Security validation
	if err := s.securityValidator.ValidateScriptPath(scriptPath); err != nil {
		return nil, err
	}

	// Check if file exists
	info, err := os.Stat(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("script not found: %s", scriptPath)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("path is a directory, not a script: %s", scriptPath)
	}

	// Check if supported type
	if !s.isSupported(scriptPath) {
		return nil, fmt.Errorf("unsupported script type: %s", filepath.Ext(scriptPath))
	}

	return s.createScriptInfo(scriptPath)
}

// RefreshScript updates metadata for a single script
func (s *ScriptDiscoveryService) RefreshScript(scriptPath string) (*contracts.ScriptInfo, error) {
	return s.ValidateScript(scriptPath) // Same logic for now
}

// FilterScripts filters scripts based on query string
func (s *ScriptDiscoveryService) FilterScripts(scripts []contracts.ScriptInfo, query string) []contracts.ScriptInfo {
	if query == "" {
		return scripts
	}

	query = strings.ToLower(query)
	var filtered []contracts.ScriptInfo

	for _, script := range scripts {
		// Check name match
		if strings.Contains(strings.ToLower(script.Name), query) {
			filtered = append(filtered, script)
			continue
		}

		// Check type match
		if strings.Contains(strings.ToLower(script.Type), query) {
			filtered = append(filtered, script)
			continue
		}

		// Check tags match
		for _, tag := range script.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				filtered = append(filtered, script)
				break
			}
		}
	}

	return filtered
}

// WatchDirectory monitors directory for changes (placeholder implementation)
func (s *ScriptDiscoveryService) WatchDirectory(ctx context.Context, dirPath string) (<-chan contracts.DirectoryChange, error) {
	// This would use fsnotify in a real implementation
	ch := make(chan contracts.DirectoryChange)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

// createScriptInfo creates a ScriptInfo from a file path
func (s *ScriptDiscoveryService) createScriptInfo(path string) (*contracts.ScriptInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Check if executable
	isExecutable := info.Mode()&0111 != 0

	// Get script type from extension
	scriptType := models.GetTypeFromExtension(path)

	// Parse script metadata using the parser
	config := parser.DefaultParseConfig()
	metadata, err := parser.ParseScript(path, scriptType, config)
	// If parsing fails, continue without metadata (graceful degradation)
	if err != nil {
		metadata = nil
	}

	// Extract description from metadata if available
	description := ""
	if metadata != nil && metadata.Description != "" {
		description = metadata.Description
	}

	// Convert parser.ScriptMetadata to contracts.ScriptMetadata
	var contractMetadata *contracts.ScriptMetadata
	if metadata != nil {
		contractMetadata = &contracts.ScriptMetadata{
			Description:  metadata.Description,
			FullContent:  metadata.FullContent,
			LineCount:    metadata.LineCount,
			PreviewLines: metadata.PreviewLines,
			IsTruncated:  metadata.IsTruncated,
			Interpreter:  metadata.Interpreter,
			Tags:         metadata.Tags,
		}
	}

	// Create script info
	scriptInfo := &contracts.ScriptInfo{
		ID:           generateScriptID(path, info.ModTime()),
		Name:         getScriptName(path),
		Path:         path,
		Type:         scriptType,
		Size:         info.Size(),
		ModifiedTime: info.ModTime(),
		IsExecutable: isExecutable,
		Description:  description,
		Tags:         make([]string, 0),
		Metadata:     contractMetadata,
	}

	return scriptInfo, nil
}

// isSupported checks if a file extension is supported
func (s *ScriptDiscoveryService) isSupported(path string) bool {
	ext := filepath.Ext(path)
	_, supported := s.supportedTypes[ext]
	return supported
}

// Helper functions
func generateScriptID(path string, modTime time.Time) string {
	return fmt.Sprintf("script_%x", []byte(path+modTime.String())[:8])
}

func getScriptName(path string) string {
	name := filepath.Base(path)
	ext := filepath.Ext(name)
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}
	return name
}

func getSupportedExtensions(types map[string]string) []string {
	var exts []string
	for ext := range types {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

// organizeScriptsIntoDirectories reorganizes scripts from flat list into directory structure
func (s *ScriptDiscoveryService) organizeScriptsIntoDirectories(root *contracts.DirectoryInfo) {
	// Get all scripts and clear the root scripts list
	allScripts := root.Scripts
	root.Scripts = make([]contracts.ScriptInfo, 0)
	root.ScriptCount = 0

	// Group scripts by directory
	scriptsByDir := make(map[string][]contracts.ScriptInfo)

	for _, script := range allScripts {
		scriptDir := filepath.Dir(script.Path)

		// If script is in root directory, keep it in root
		if scriptDir == root.Path {
			root.Scripts = append(root.Scripts, script)
			root.ScriptCount++
		} else {
			// Group by subdirectory
			scriptsByDir[scriptDir] = append(scriptsByDir[scriptDir], script)
		}
	}

	// Create directory structure and add scripts
	for dirPath, scripts := range scriptsByDir {
		s.createDirectoryStructureAndAddScripts(root, dirPath, scripts)
	}
}

// createDirectoryStructureAndAddScripts creates directory structure and adds scripts
func (s *ScriptDiscoveryService) createDirectoryStructureAndAddScripts(root *contracts.DirectoryInfo, dirPath string, scripts []contracts.ScriptInfo) {
	// Get relative path from root
	relPath, err := filepath.Rel(root.Path, dirPath)
	if err != nil {
		return
	}

	// Split path into parts
	parts := strings.Split(relPath, string(filepath.Separator))
	currentDir := root
	currentPath := root.Path

	// Create directory structure
	for _, part := range parts {
		if part == "." || part == "" {
			continue
		}

		currentPath = filepath.Join(currentPath, part)

		// Find or create this directory
		found := false
		for i := range currentDir.Children {
			if currentDir.Children[i].Path == currentPath {
				currentDir = &currentDir.Children[i]
				found = true
				break
			}
		}

		if !found {
			// Create new directory
			newDir := contracts.DirectoryInfo{
				Path:        currentPath,
				Name:        part,
				Children:    make([]contracts.DirectoryInfo, 0),
				Scripts:     make([]contracts.ScriptInfo, 0),
				ScriptCount: 0,
				LastScan:    root.LastScan,
			}
			currentDir.Children = append(currentDir.Children, newDir)
			currentDir = &currentDir.Children[len(currentDir.Children)-1]
		}
	}

	// Add scripts to the final directory
	currentDir.Scripts = append(currentDir.Scripts, scripts...)
	currentDir.ScriptCount += len(scripts)
}

// addScriptToTree adds a script to the correct location in the directory tree
func (s *ScriptDiscoveryService) addScriptToTree(root *contracts.DirectoryInfo, scriptPath string, script *contracts.ScriptInfo) {
	scriptDir := filepath.Dir(scriptPath)

	// If script is in root directory, add to root
	if scriptDir == root.Path {
		root.Scripts = append(root.Scripts, *script)
		root.ScriptCount++
		return
	}

	// Find or create the directory structure and add script there
	s.findOrCreateDirectoryAndAddScript(root, scriptDir, script)
}

// findOrCreateDirectoryAndAddScript finds or creates directory structure and adds script
func (s *ScriptDiscoveryService) findOrCreateDirectoryAndAddScript(root *contracts.DirectoryInfo, targetDir string, script *contracts.ScriptInfo) {
	// Get relative path from root to target directory
	relPath, err := filepath.Rel(root.Path, targetDir)
	if err != nil || relPath == "." {
		// If we can't get relative path, add to root as fallback
		root.Scripts = append(root.Scripts, *script)
		root.ScriptCount++
		return
	}

	// Split path into parts
	parts := strings.Split(relPath, string(filepath.Separator))
	currentDir := root
	currentPath := root.Path

	// Create/find directory structure
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		currentPath = filepath.Join(currentPath, part)

		// Find this directory in children
		found := false
		for i := range currentDir.Children {
			if currentDir.Children[i].Path == currentPath {
				currentDir = &currentDir.Children[i]
				found = true
				break
			}
		}

		// Create directory if not found
		if !found {
			newDir := contracts.DirectoryInfo{
				Path:        currentPath,
				Name:        part,
				Children:    make([]contracts.DirectoryInfo, 0),
				Scripts:     make([]contracts.ScriptInfo, 0),
				ScriptCount: 0,
				LastScan:    root.LastScan,
			}
			currentDir.Children = append(currentDir.Children, newDir)
			currentDir = &currentDir.Children[len(currentDir.Children)-1]
		}
	}

	// Add script to the final directory
	currentDir.Scripts = append(currentDir.Scripts, *script)
	currentDir.ScriptCount++
}