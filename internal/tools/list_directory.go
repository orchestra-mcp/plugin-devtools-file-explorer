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

// ListDirectorySchema returns the JSON Schema for the list_directory tool.
func ListDirectorySchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Directory path to list",
			},
			"show_hidden": map[string]any{
				"type":        "boolean",
				"description": "Show hidden files and directories (default: false)",
			},
			"recursive": map[string]any{
				"type":        "boolean",
				"description": "Recursively list subdirectories up to 3 levels deep (default: false)",
			},
		},
		"required": []any{"path"},
	})
	return s
}

// ListDirectory returns a tool handler that lists directory contents.
func ListDirectory() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		path := helpers.GetString(req.Arguments, "path")
		showHidden := helpers.GetBool(req.Arguments, "show_hidden")
		recursive := helpers.GetBool(req.Arguments, "recursive")

		var sb strings.Builder

		if recursive {
			err := walkDir(&sb, path, "", 0, 3, showHidden)
			if err != nil {
				return helpers.ErrorResult("read_error", err.Error()), nil
			}
		} else {
			entries, err := os.ReadDir(path)
			if err != nil {
				return helpers.ErrorResult("read_error", err.Error()), nil
			}
			for _, entry := range entries {
				if !showHidden && strings.HasPrefix(entry.Name(), ".") {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					continue
				}
				formatEntry(&sb, entry.Name(), info, "", entry.IsDir())
			}
		}

		return helpers.TextResult(sb.String()), nil
	}
}

// walkDir recursively walks directories up to maxDepth levels.
func walkDir(sb *strings.Builder, dirPath, prefix string, depth, maxDepth int, showHidden bool) error {
	if depth >= maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	// Filter hidden if needed
	filtered := entries[:0:len(entries)]
	filtered = filtered[:0]
	for _, e := range entries {
		if !showHidden && strings.HasPrefix(e.Name(), ".") {
			continue
		}
		filtered = append(filtered, e)
	}

	for i, entry := range filtered {
		isLast := i == len(filtered)-1
		connector := "├── "
		childPrefix := prefix + "│   "
		if isLast {
			connector = "└── "
			childPrefix = prefix + "    "
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		line := fmt.Sprintf("%s%s", prefix+connector, entry.Name())
		if entry.IsDir() {
			line += "/"
		} else {
			line += fmt.Sprintf("  (%s, %s)", formatSize(info.Size()), info.ModTime().Format("2006-01-02 15:04"))
		}
		sb.WriteString(line + "\n")

		if entry.IsDir() && depth+1 < maxDepth {
			subPath := filepath.Join(dirPath, entry.Name())
			_ = walkDir(sb, subPath, childPrefix, depth+1, maxDepth, showHidden)
		}
	}
	return nil
}

// formatEntry writes a single flat directory entry line.
func formatEntry(sb *strings.Builder, name string, info os.FileInfo, prefix string, isDir bool) {
	if isDir {
		sb.WriteString(fmt.Sprintf("%s%s/  (dir, %s)\n", prefix, name, info.ModTime().Format("2006-01-02 15:04")))
	} else {
		sb.WriteString(fmt.Sprintf("%s%s  (%s, %s)\n", prefix, name, formatSize(info.Size()), info.ModTime().Format("2006-01-02 15:04")))
	}
}

// formatSize returns a human-readable file size.
func formatSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
