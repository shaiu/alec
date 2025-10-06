package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// PythonLexer parses Python scripts
type PythonLexer struct{}

// NewPythonLexer creates a new Python script lexer
func NewPythonLexer() *PythonLexer {
	return &PythonLexer{}
}

// Parse parses a Python script and extracts metadata
func (l *PythonLexer) Parse(reader io.Reader, config ParseConfig) (*ScriptMetadata, error) {
	metadata := NewScriptMetadata()
	scanner := bufio.NewScanner(reader)

	var allLines []string
	lineNum := 0
	inDocstring := false
	docstringDelimiter := ""
	var docstringLines []string
	var commentLines []string
	foundModuleDocstring := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		allLines = append(allLines, line)
		trimmed := strings.TrimSpace(line)

		// Extract shebang from first line
		if lineNum == 1 && strings.HasPrefix(trimmed, "#!") {
			metadata.Interpreter = strings.TrimSpace(strings.TrimPrefix(trimmed, "#!"))
			continue
		}

		// Look for module-level docstring (within first few lines, after shebang/comments)
		if !foundModuleDocstring && lineNum <= 20 {
			// Check for start of docstring
			if !inDocstring {
				// Check for triple-quoted string
				if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
					if strings.HasPrefix(trimmed, `"""`) {
						docstringDelimiter = `"""`
					} else {
						docstringDelimiter = `'''`
					}

					inDocstring = true
					// Extract content from same line if it's a one-liner
					content := strings.TrimPrefix(trimmed, docstringDelimiter)
					if strings.HasSuffix(content, docstringDelimiter) {
						// One-line docstring
						content = strings.TrimSuffix(content, docstringDelimiter)
						content = strings.TrimSpace(content)
						if content != "" {
							docstringLines = append(docstringLines, content)
						}
						inDocstring = false
						foundModuleDocstring = true
					} else {
						// Multi-line docstring started
						content = strings.TrimSpace(content)
						if content != "" {
							docstringLines = append(docstringLines, content)
						}
					}
					continue
				}

				// Check for custom description markers in comments
				if desc := l.extractMarkedDescription(trimmed); desc != "" {
					commentLines = append(commentLines, desc)
					continue
				}

				// Regular comment line (after shebang)
				if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!") {
					comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
					if comment != "" {
						commentLines = append(commentLines, comment)
					}
					continue
				}

				// Empty line - continue looking
				if trimmed == "" {
					continue
				}

				// Non-docstring code found before docstring - stop looking
				if !strings.HasPrefix(trimmed, `"""`) && !strings.HasPrefix(trimmed, `'''`) && trimmed != "" {
					foundModuleDocstring = true
				}
			} else {
				// We're inside a docstring - look for end delimiter
				if strings.Contains(trimmed, docstringDelimiter) {
					// End of docstring
					content := strings.Split(trimmed, docstringDelimiter)[0]
					content = strings.TrimSpace(content)
					if content != "" {
						docstringLines = append(docstringLines, content)
					}
					inDocstring = false
					foundModuleDocstring = true
				} else {
					// Continue collecting docstring lines
					content := strings.TrimSpace(trimmed)
					if content != "" {
						docstringLines = append(docstringLines, content)
					}
				}
				continue
			}
		}

		// Stop reading full content if we've exceeded preview limit
		if lineNum > config.MaxPreviewLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning Python script: %w", err)
	}

	// If we stopped reading early, count remaining lines
	totalLineCount := lineNum
	if lineNum > config.MaxPreviewLines {
		for scanner.Scan() {
			totalLineCount++
		}
	}

	// Build description: prioritize docstring, fallback to comments
	if len(docstringLines) > 0 {
		fullDesc := strings.Join(docstringLines, " ")
		metadata.Description = truncateDescription(fullDesc, config.DescriptionMaxChars)
	} else if len(commentLines) > 0 {
		fullDesc := strings.Join(commentLines, " ")
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

// ExtractDescription extracts just the description from a Python script
func (l *PythonLexer) ExtractDescription(reader io.Reader) (string, error) {
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	inDocstring := false
	docstringDelimiter := ""
	var docstringLines []string
	var commentLines []string

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Skip shebang
		if lineNum == 1 && strings.HasPrefix(trimmed, "#!") {
			continue
		}

		// Only look in first ~20 lines for module docstring
		if lineNum > 20 {
			break
		}

		if !inDocstring {
			// Check for triple-quoted string
			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				if strings.HasPrefix(trimmed, `"""`) {
					docstringDelimiter = `"""`
				} else {
					docstringDelimiter = `'''`
				}

				inDocstring = true
				content := strings.TrimPrefix(trimmed, docstringDelimiter)
				if strings.HasSuffix(content, docstringDelimiter) {
					// One-line docstring
					content = strings.TrimSuffix(content, docstringDelimiter)
					docstringLines = append(docstringLines, strings.TrimSpace(content))
					break
				} else {
					docstringLines = append(docstringLines, strings.TrimSpace(content))
				}
				continue
			}

			// Check for custom markers
			if desc := l.extractMarkedDescription(trimmed); desc != "" {
				commentLines = append(commentLines, desc)
				continue
			}

			// Regular comments
			if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!") {
				comment := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
				if comment != "" {
					commentLines = append(commentLines, comment)
				}
				continue
			}

			// Empty line
			if trimmed == "" {
				continue
			}

			// Code without docstring - stop
			break
		} else {
			// Inside docstring
			if strings.Contains(trimmed, docstringDelimiter) {
				content := strings.Split(trimmed, docstringDelimiter)[0]
				if content != "" {
					docstringLines = append(docstringLines, strings.TrimSpace(content))
				}
				break
			} else {
				if trimmed != "" {
					docstringLines = append(docstringLines, strings.TrimSpace(trimmed))
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading Python script: %w", err)
	}

	// Prioritize docstring over comments
	if len(docstringLines) > 0 {
		return strings.Join(docstringLines, " "), nil
	}
	return strings.Join(commentLines, " "), nil
}

// extractMarkedDescription checks for custom description markers in Python comments
func (l *PythonLexer) extractMarkedDescription(line string) string {
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
