package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

const smallFileThreshold = 100 * 1024 // 100 KB

// ReadFileSchema returns the JSON Schema for the read_file tool.
func ReadFileSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to read",
			},
			"offset": map[string]any{
				"type":        "integer",
				"description": "Line number to start reading from (1-based, optional)",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of lines to return (optional)",
			},
		},
		"required": []any{"path"},
	})
	return s
}

// ReadFile returns a tool handler that reads file contents.
func ReadFile() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		path := helpers.GetString(req.Arguments, "path")
		offset := helpers.GetInt(req.Arguments, "offset")
		limit := helpers.GetInt(req.Arguments, "limit")

		useLineMode := offset > 0 || limit > 0

		info, err := os.Stat(path)
		if err != nil {
			return helpers.ErrorResult("stat_error", err.Error()), nil
		}
		if info.IsDir() {
			return helpers.ErrorResult("is_directory", fmt.Sprintf("%s is a directory", path)), nil
		}

		// Small file without line constraints: read all at once
		if !useLineMode && info.Size() < smallFileThreshold {
			data, err := os.ReadFile(path)
			if err != nil {
				return helpers.ErrorResult("read_error", err.Error()), nil
			}
			return helpers.TextResult(string(data)), nil
		}

		// Line-based reading (also used for large files)
		f, err := os.Open(path)
		if err != nil {
			return helpers.ErrorResult("open_error", err.Error()), nil
		}
		defer f.Close()

		var sb strings.Builder
		scanner := bufio.NewScanner(f)

		// Increase default buffer for long lines
		buf := make([]byte, 64*1024)
		scanner.Buffer(buf, 1024*1024)

		lineNum := 0
		linesWritten := 0

		// Normalise: offset 0 means "no offset" (start at 1)
		startLine := offset
		if startLine < 1 {
			startLine = 1
		}

		for scanner.Scan() {
			lineNum++
			if lineNum < startLine {
				continue
			}
			if limit > 0 && linesWritten >= limit {
				break
			}
			if useLineMode {
				sb.WriteString(fmt.Sprintf("%6d\t%s\n", lineNum, scanner.Text()))
			} else {
				sb.WriteString(scanner.Text() + "\n")
			}
			linesWritten++
		}

		if err := scanner.Err(); err != nil {
			return helpers.ErrorResult("scan_error", err.Error()), nil
		}

		return helpers.TextResult(sb.String()), nil
	}
}
