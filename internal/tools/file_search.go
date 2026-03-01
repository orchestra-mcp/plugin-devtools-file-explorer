package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

const defaultMaxResults = 50

// FileSearchSchema returns the JSON Schema for the file_search tool.
func FileSearchSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"directory": map[string]any{
				"type":        "string",
				"description": "Root directory to search within",
			},
			"pattern": map[string]any{
				"type":        "string",
				"description": "Glob pattern to match. If it contains '/', matched against the full relative path; otherwise matched against the file/directory basename.",
			},
			"max_results": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results to return (default: 50)",
			},
		},
		"required": []any{"directory", "pattern"},
	})
	return s
}

// FileSearch returns a tool handler that searches for files matching a glob pattern.
func FileSearch() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "directory", "pattern"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		directory := helpers.GetString(req.Arguments, "directory")
		pattern := helpers.GetString(req.Arguments, "pattern")
		maxResults := helpers.GetInt(req.Arguments, "max_results")
		if maxResults <= 0 {
			maxResults = defaultMaxResults
		}

		pathPattern := strings.Contains(pattern, string(os.PathSeparator)) || strings.Contains(pattern, "/")

		var matches []string

		err := filepath.WalkDir(directory, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				// Skip unreadable entries without aborting the walk.
				return nil
			}
			if len(matches) >= maxResults {
				return filepath.SkipAll
			}

			rel, relErr := filepath.Rel(directory, path)
			if relErr != nil {
				rel = path
			}

			var matched bool
			if pathPattern {
				matched, _ = filepath.Match(pattern, rel)
			} else {
				matched, _ = filepath.Match(pattern, d.Name())
			}

			if matched {
				matches = append(matches, path)
			}
			return nil
		})

		if err != nil {
			return helpers.ErrorResult("walk_error", err.Error()), nil
		}

		if len(matches) == 0 {
			return helpers.TextResult(fmt.Sprintf("No files found matching %q in %s", pattern, directory)), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d match(es) for %q in %s:\n\n", len(matches), pattern, directory))
		for _, m := range matches {
			sb.WriteString(m + "\n")
		}

		return helpers.TextResult(sb.String()), nil
	}
}
