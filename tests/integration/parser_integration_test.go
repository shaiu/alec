package integration

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/your-org/alec/pkg/parser"
)

// TestParseRealScripts tests parsing of actual script files
func TestParseRealScripts(t *testing.T) {
	fixturesDir := "../fixtures/scripts"

	tests := []struct {
		name           string
		filename       string
		scriptType     string
		wantDescPrefix string
		wantInterp     string
		minLines       int
	}{
		{
			name:           "backup shell script",
			filename:       "backup.sh",
			scriptType:     "shell",
			wantDescPrefix: "Performs automated backups",
			wantInterp:     "/bin/bash",
			minLines:       20,
		},
		{
			name:           "deploy python script",
			filename:       "deploy.py",
			scriptType:     "python",
			wantDescPrefix: "Automated deployment script",
			wantInterp:     "/usr/bin/env python3",
			minLines:       30,
		},
		{
			name:           "simple shell script",
			filename:       "simple.sh",
			scriptType:     "shell",
			wantDescPrefix: "Simple hello world",
			wantInterp:     "/bin/bash",
			minLines:       3,
		},
		{
			name:           "quick test python script",
			filename:       "quick_test.py",
			scriptType:     "python",
			wantDescPrefix: "Quick test runner",
			wantInterp:     "/usr/bin/env python3",
			minLines:       5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(fixturesDir, tt.filename)
			config := parser.DefaultParseConfig()

			metadata, err := parser.ParseScript(scriptPath, tt.scriptType, config)
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}

			// Check description starts with expected prefix
			if !strings.HasPrefix(metadata.Description, tt.wantDescPrefix) {
				t.Errorf("Description = %q, want prefix %q", metadata.Description, tt.wantDescPrefix)
			}

			// Check interpreter
			if metadata.Interpreter != tt.wantInterp {
				t.Errorf("Interpreter = %q, want %q", metadata.Interpreter, tt.wantInterp)
			}

			// Check minimum line count
			if metadata.LineCount < tt.minLines {
				t.Errorf("LineCount = %d, want at least %d", metadata.LineCount, tt.minLines)
			}

			// Check that preview content is not empty
			if metadata.FullContent == "" {
				t.Error("FullContent should not be empty")
			}

			// Check that preview lines is set correctly
			if metadata.PreviewLines == 0 {
				t.Error("PreviewLines should be > 0")
			}
		})
	}
}

// TestParseScript_ContentPreview tests that script content is properly previewed
func TestParseScript_ContentPreview(t *testing.T) {
	fixturesDir := "../fixtures/scripts"

	tests := []struct {
		name         string
		filename     string
		scriptType   string
		expectFull   bool // Whether we expect the full script or truncated
	}{
		{
			name:       "simple short script - full content",
			filename:   "simple.sh",
			scriptType: "shell",
			expectFull: true,
		},
		{
			name:       "longer script - may be truncated",
			filename:   "backup.sh",
			scriptType: "shell",
			expectFull: true, // Still short enough for full content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptPath := filepath.Join(fixturesDir, tt.filename)
			config := parser.DefaultParseConfig()

			metadata, err := parser.ParseScript(scriptPath, tt.scriptType, config)
			if err != nil {
				t.Fatalf("ParseScript() error = %v", err)
			}

			if tt.expectFull && metadata.IsTruncated {
				t.Error("Expected full script but got truncated content")
			}

			// Verify content contains expected elements
			if !strings.Contains(metadata.FullContent, "#!") && metadata.Interpreter != "" {
				t.Error("Expected content to contain shebang")
			}
		})
	}
}

// TestParseScript_ErrorHandling tests error cases
func TestParseScript_ErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		scriptPath string
		scriptType string
		wantErr    bool
	}{
		{
			name:       "nonexistent file",
			scriptPath: "/tmp/nonexistent_file_12345.sh",
			scriptType: "shell",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := parser.DefaultParseConfig()
			_, err := parser.ParseScript(tt.scriptPath, tt.scriptType, config)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCustomParseConfig tests custom parsing configuration
func TestCustomParseConfig(t *testing.T) {
	fixturesDir := "../fixtures/scripts"
	scriptPath := filepath.Join(fixturesDir, "backup.sh")

	config := parser.ParseConfig{
		MaxPreviewLines:     10,
		FullScriptThreshold: 5,
		DescriptionMaxChars: 50,
	}

	metadata, err := parser.ParseScript(scriptPath, "shell", config)
	if err != nil {
		t.Fatalf("ParseScript() error = %v", err)
	}

	// With these settings, the backup.sh script should be truncated
	if !metadata.IsTruncated {
		t.Error("Expected script to be truncated with custom config")
	}

	if metadata.PreviewLines > config.MaxPreviewLines {
		t.Errorf("PreviewLines = %d, want <= %d", metadata.PreviewLines, config.MaxPreviewLines)
	}

	if len(metadata.Description) > config.DescriptionMaxChars {
		t.Errorf("Description length = %d, want <= %d", len(metadata.Description), config.DescriptionMaxChars)
	}
}
