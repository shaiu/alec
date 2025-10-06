package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ShellLexer parses shell scripts (bash, sh, zsh)
type ShellLexer struct{}

// NewShellLexer creates a new shell script lexer
func NewShellLexer() *ShellLexer {
	return &ShellLexer{}
}

// Parse parses a shell script and extracts metadata
func (l *ShellLexer) Parse(reader io.Reader, config ParseConfig) (*ScriptMetadata, error) {
	metadata := NewScriptMetadata()
	scanner := bufio.NewScanner(reader)

	var allLines []string
	lineNum := 0
	inHeaderComments := true
	var descriptionParts []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		allLines = append(allLines, line)

		// Extract shebang from first line
		if lineNum == 1 {
			if strings.HasPrefix(line, "#!") {
				metadata.Interpreter = strings.TrimSpace(strings.TrimPrefix(line, "#!"))
				continue
			}
			// If no shebang on first line, still process it as potential comment
		}

		// Extract header comments for description
		if inHeaderComments {
			trimmed := strings.TrimSpace(line)

			// Check for custom description markers
			if desc := l.extractMarkedDescription(trimmed); desc != "" {
				descriptionParts = append(descriptionParts, desc)
				continue
			}

			// Regular comment line
			if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!") {
				comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
				if comment != "" {
					descriptionParts = append(descriptionParts, comment)
				}
				continue
			}

			// Empty line - continue in header zone
			if trimmed == "" {
				continue
			}

			// Non-comment, non-empty line - exit header zone
			inHeaderComments = false
		}

		// Stop reading full content if we've exceeded preview limit
		if lineNum > config.MaxPreviewLines && len(allLines) >= config.MaxPreviewLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning shell script: %w", err)
	}

	// If we stopped reading early, count remaining lines
	totalLineCount := lineNum
	if lineNum > config.MaxPreviewLines {
		for scanner.Scan() {
			totalLineCount++
		}
	}

	// Build description from collected parts
	if len(descriptionParts) > 0 {
		fullDesc := strings.Join(descriptionParts, " ")
		metadata.Description = truncateDescription(fullDesc, config.DescriptionMaxChars)
	}

	// Determine if we should include full content or truncated preview
	metadata.LineCount = totalLineCount
	if totalLineCount <= config.FullScriptThreshold {
		// Script is short enough - include everything
		metadata.FullContent = strings.Join(allLines, "\n")
		metadata.PreviewLines = totalLineCount
		metadata.IsTruncated = false
	} else {
		// Include preview only
		previewLineCount := config.MaxPreviewLines
		if len(allLines) < previewLineCount {
			previewLineCount = len(allLines)
		}
		metadata.FullContent = strings.Join(allLines[:previewLineCount], "\n")
		metadata.PreviewLines = previewLineCount
		metadata.IsTruncated = true
	}

	return metadata, nil
}

// ExtractDescription extracts just the description from a shell script
func (l *ShellLexer) ExtractDescription(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	var descriptionParts []string
	lineNum := 0
	inHeaderComments := true

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip shebang
		if lineNum == 1 && strings.HasPrefix(trimmed, "#!") {
			continue
		}

		if inHeaderComments {
			// Check for custom description markers
			if desc := l.extractMarkedDescription(trimmed); desc != "" {
				descriptionParts = append(descriptionParts, desc)
				continue
			}

			// Regular comment line
			if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!") {
				comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
				if comment != "" {
					descriptionParts = append(descriptionParts, comment)
				}
				continue
			}

			// Empty line - continue
			if trimmed == "" {
				continue
			}

			// Non-comment line - stop
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading shell script: %w", err)
	}

	return strings.Join(descriptionParts, " "), nil
}

// extractMarkedDescription checks for custom description markers and extracts the content
func (l *ShellLexer) extractMarkedDescription(line string) string {
	markers := []string{
		"# Description:",
		"# @desc",
		"# @description",
		"# Summary:",
		"# @summary",
		"#Description:",
		"#@desc",
	}

	for _, marker := range markers {
		if strings.HasPrefix(line, marker) {
			return strings.TrimSpace(strings.TrimPrefix(line, marker))
		}
	}

	return ""
}
