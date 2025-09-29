package models

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Directory represents a directory node in the hierarchical script organization
type Directory struct {
	Path        string       `json:"path"`
	Name        string       `json:"name"`
	Parent      *Directory   `json:"-"` // Exclude from JSON to prevent cycles
	Children    []*Directory `json:"children,omitempty"`
	Scripts     []*Script    `json:"scripts,omitempty"`
	IsRoot      bool         `json:"is_root"`
	IsExpanded  bool         `json:"is_expanded"`
	ScriptCount int          `json:"script_count"`
	LastScan    time.Time    `json:"last_scan"`
}

// NewDirectory creates a new Directory instance
func NewDirectory(path string) *Directory {
	cleanPath := filepath.Clean(path)
	name := filepath.Base(cleanPath)

	return &Directory{
		Path:        cleanPath,
		Name:        name,
		Children:    make([]*Directory, 0),
		Scripts:     make([]*Script, 0),
		IsRoot:      false,
		IsExpanded:  false,
		ScriptCount: 0,
		LastScan:    time.Now(),
	}
}

// NewRootDirectory creates a root directory
func NewRootDirectory(path string) *Directory {
	dir := NewDirectory(path)
	dir.IsRoot = true
	dir.IsExpanded = true // Root directories start expanded
	return dir
}

// Validate checks if the directory is valid
func (d *Directory) Validate() error {
	if d.Path == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	if !filepath.IsAbs(d.Path) {
		return fmt.Errorf("directory path must be absolute: %s", d.Path)
	}

	// Path traversal check
	if !filepath.IsLocal(d.Path) {
		return fmt.Errorf("directory path contains path traversal: %s", d.Path)
	}

	// Validate parent-child relationships
	for _, child := range d.Children {
		if child.Parent != d {
			return fmt.Errorf("child directory parent mismatch: %s", child.Path)
		}

		// Check that child path is actually under parent
		if !strings.HasPrefix(child.Path, d.Path+string(filepath.Separator)) {
			return fmt.Errorf("child directory not under parent: %s not under %s", child.Path, d.Path)
		}
	}

	return nil
}

// AddChild adds a child directory
func (d *Directory) AddChild(child *Directory) error {
	if child == nil {
		return fmt.Errorf("child directory cannot be nil")
	}

	// Check for duplicates
	for _, existing := range d.Children {
		if existing.Path == child.Path {
			return fmt.Errorf("child directory already exists: %s", child.Path)
		}
	}

	// Validate that child is actually under this directory
	expectedParent := filepath.Dir(child.Path)
	if expectedParent != d.Path {
		return fmt.Errorf("child directory %s is not under parent %s", child.Path, d.Path)
	}

	child.Parent = d
	d.Children = append(d.Children, child)

	// Keep children sorted by name
	sort.Slice(d.Children, func(i, j int) bool {
		return d.Children[i].Name < d.Children[j].Name
	})

	d.updateScriptCount()
	return nil
}

// RemoveChild removes a child directory
func (d *Directory) RemoveChild(path string) bool {
	for i, child := range d.Children {
		if child.Path == path {
			child.Parent = nil
			d.Children = append(d.Children[:i], d.Children[i+1:]...)
			d.updateScriptCount()
			return true
		}
	}
	return false
}

// AddScript adds a script to this directory
func (d *Directory) AddScript(script *Script) error {
	if script == nil {
		return fmt.Errorf("script cannot be nil")
	}

	// Validate that script is in this directory
	scriptDir := filepath.Dir(script.Path)
	if scriptDir != d.Path {
		return fmt.Errorf("script %s is not in directory %s", script.Path, d.Path)
	}

	// Check for duplicates
	for _, existing := range d.Scripts {
		if existing.Path == script.Path {
			return fmt.Errorf("script already exists: %s", script.Path)
		}
	}

	d.Scripts = append(d.Scripts, script)

	// Keep scripts sorted by name
	sort.Slice(d.Scripts, func(i, j int) bool {
		return d.Scripts[i].Name < d.Scripts[j].Name
	})

	d.updateScriptCount()
	return nil
}

// RemoveScript removes a script from this directory
func (d *Directory) RemoveScript(path string) bool {
	for i, script := range d.Scripts {
		if script.Path == path {
			d.Scripts = append(d.Scripts[:i], d.Scripts[i+1:]...)
			d.updateScriptCount()
			return true
		}
	}
	return false
}

// FindScript finds a script by path in this directory
func (d *Directory) FindScript(path string) *Script {
	for _, script := range d.Scripts {
		if script.Path == path {
			return script
		}
	}
	return nil
}

// FindChild finds a child directory by path
func (d *Directory) FindChild(path string) *Directory {
	for _, child := range d.Children {
		if child.Path == path {
			return child
		}
	}
	return nil
}

// GetRoot returns the root directory of this tree
func (d *Directory) GetRoot() *Directory {
	current := d
	for current.Parent != nil {
		current = current.Parent
	}
	return current
}

// GetDepth returns the depth of this directory from root
func (d *Directory) GetDepth() int {
	depth := 0
	current := d
	for current.Parent != nil {
		depth++
		current = current.Parent
	}
	return depth
}

// GetPath returns the breadcrumb path from root to this directory
func (d *Directory) GetBreadcrumbs() []string {
	path := make([]string, 0)
	current := d

	// Build path from current to root
	for current != nil {
		path = append([]string{current.Name}, path...)
		current = current.Parent
	}

	return path
}

// Expand sets the directory as expanded
func (d *Directory) Expand() {
	d.IsExpanded = true
}

// Collapse sets the directory as collapsed
func (d *Directory) Collapse() {
	d.IsExpanded = false
}

// Toggle toggles the expansion state
func (d *Directory) Toggle() {
	d.IsExpanded = !d.IsExpanded
}

// updateScriptCount recursively updates script counts
func (d *Directory) updateScriptCount() {
	count := len(d.Scripts)

	for _, child := range d.Children {
		count += child.ScriptCount
	}

	d.ScriptCount = count

	// Update parent counts recursively
	if d.Parent != nil {
		d.Parent.updateScriptCount()
	}
}

// GetAllScripts returns all scripts in this directory and subdirectories
func (d *Directory) GetAllScripts() []*Script {
	scripts := make([]*Script, 0)

	// Add scripts from this directory
	scripts = append(scripts, d.Scripts...)

	// Add scripts from child directories
	for _, child := range d.Children {
		scripts = append(scripts, child.GetAllScripts()...)
	}

	return scripts
}

// FilterScripts returns scripts matching the given filter function
func (d *Directory) FilterScripts(filter func(*Script) bool) []*Script {
	var filtered []*Script

	// Filter scripts from this directory
	for _, script := range d.Scripts {
		if filter(script) {
			filtered = append(filtered, script)
		}
	}

	// Filter scripts from child directories
	for _, child := range d.Children {
		filtered = append(filtered, child.FilterScripts(filter)...)
	}

	return filtered
}

// Walk traverses the directory tree and calls the function for each directory
func (d *Directory) Walk(fn func(*Directory) error) error {
	// Process current directory
	if err := fn(d); err != nil {
		return err
	}

	// Process children
	for _, child := range d.Children {
		if err := child.Walk(fn); err != nil {
			return err
		}
	}

	return nil
}

// Clone creates a deep copy of the directory (without parent reference)
func (d *Directory) Clone() *Directory {
	clone := &Directory{
		Path:        d.Path,
		Name:        d.Name,
		Parent:      nil, // Don't clone parent to avoid cycles
		IsRoot:      d.IsRoot,
		IsExpanded:  d.IsExpanded,
		ScriptCount: d.ScriptCount,
		LastScan:    d.LastScan,
	}

	// Clone children
	if d.Children != nil {
		clone.Children = make([]*Directory, len(d.Children))
		for i, child := range d.Children {
			clone.Children[i] = child.Clone()
			clone.Children[i].Parent = clone
		}
	}

	// Clone scripts
	if d.Scripts != nil {
		clone.Scripts = make([]*Script, len(d.Scripts))
		for i, script := range d.Scripts {
			clone.Scripts[i] = script.Clone()
		}
	}

	return clone
}