package parser

// ScriptMetadata holds extracted metadata from a script file
type ScriptMetadata struct {
	// Description is the extracted description from comments or docstrings
	Description string `json:"description"`

	// FullContent is the complete script content or a preview
	FullContent string `json:"full_content"`

	// LineCount is the total number of lines in the script
	LineCount int `json:"line_count"`

	// PreviewLines is the number of lines included in the preview
	PreviewLines int `json:"preview_lines"`

	// IsTruncated indicates whether the content has been truncated
	IsTruncated bool `json:"is_truncated"`

	// Interpreter is the shebang interpreter (e.g., "/bin/bash", "python3")
	Interpreter string `json:"interpreter,omitempty"`

	// Tags are auto-extracted tags from the script (optional)
	Tags []string `json:"tags,omitempty"`
}

// ParseConfig holds configuration for script parsing
type ParseConfig struct {
	// MaxPreviewLines is the maximum number of lines to include in preview
	MaxPreviewLines int

	// FullScriptThreshold is the line count below which we show the full script
	FullScriptThreshold int

	// DescriptionMaxChars is the maximum length for description
	DescriptionMaxChars int
}

// DefaultParseConfig returns the default parsing configuration
func DefaultParseConfig() ParseConfig {
	return ParseConfig{
		MaxPreviewLines:     50,
		FullScriptThreshold: 30,
		DescriptionMaxChars: 300,
	}
}

// NewScriptMetadata creates a new empty ScriptMetadata
func NewScriptMetadata() *ScriptMetadata {
	return &ScriptMetadata{
		Tags: make([]string, 0),
	}
}
