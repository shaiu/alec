package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// TokenType represents different types of tokens in a script
type TokenType int

const (
	TokenShebang TokenType = iota
	TokenComment
	TokenDocstring
	TokenDescriptionMarker
	TokenCode
	TokenEOF
)

// Token represents a single lexical token
type Token struct {
	Type  TokenType
	Value string
	Line  int
}

// ScriptLexer is the interface that all script-type-specific lexers must implement
type ScriptLexer interface {
	// Parse parses the script file and extracts metadata
	Parse(reader io.Reader, config ParseConfig) (*ScriptMetadata, error)

	// ExtractDescription extracts just the description from the script
	ExtractDescription(reader io.Reader) (string, error)
}

// ParseScript is the main entry point for parsing a script file
func ParseScript(scriptPath, scriptType string, config ParseConfig) (*ScriptMetadata, error) {
	file, err := os.Open(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open script: %w", err)
	}
	defer file.Close()

	var lexer ScriptLexer
	switch scriptType {
	case "shell":
		lexer = NewShellLexer()
	case "python":
		lexer = NewPythonLexer()
	default:
		// For unsupported types, return basic metadata
		return parseGenericScript(file, config)
	}

	return lexer.Parse(file, config)
}

// parseGenericScript handles scripts with unknown types
func parseGenericScript(reader io.Reader, config ParseConfig) (*ScriptMetadata, error) {
	metadata := NewScriptMetadata()
	scanner := bufio.NewScanner(reader)

	var lines []string
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		line := scanner.Text()
		lines = append(lines, line)

		// Stop reading after we have enough for preview
		if lineCount > config.MaxPreviewLines {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading script: %w", err)
	}

	// Check if we need to continue reading to get accurate line count
	if lineCount > config.MaxPreviewLines {
		for scanner.Scan() {
			lineCount++
		}
	}

	metadata.LineCount = lineCount
	metadata.PreviewLines = len(lines)
	metadata.IsTruncated = lineCount > len(lines)
	metadata.FullContent = strings.Join(lines, "\n")

	return metadata, nil
}

// truncateDescription truncates a description to the maximum allowed length
func truncateDescription(desc string, maxChars int) string {
	if len(desc) <= maxChars {
		return desc
	}

	// Find the last space before maxChars-3 to avoid cutting words
	truncateAt := maxChars - 3
	if idx := strings.LastIndex(desc[:truncateAt], " "); idx > 0 {
		truncateAt = idx
	}

	return desc[:truncateAt] + "..."
}

// readLines reads lines from a reader with a limit
func readLines(reader io.Reader, maxLines int) ([]string, int, error) {
	scanner := bufio.NewScanner(reader)
	var lines []string
	lineCount := 0

	for scanner.Scan() {
		lineCount++
		if lineCount <= maxLines {
			lines = append(lines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	return lines, lineCount, nil
}
