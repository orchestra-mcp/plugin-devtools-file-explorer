package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// WriteFileSchema returns the JSON Schema for the write_file tool.
func WriteFileSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file to write",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "Content to write to the file",
			},
			"create_dirs": map[string]any{
				"type":        "boolean",
				"description": "Create parent directories if they do not exist (default: false)",
			},
		},
		"required": []any{"path", "content"},
	})
	return s
}

// WriteFile returns a tool handler that writes content to a file.
func WriteFile() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path", "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		path := helpers.GetString(req.Arguments, "path")
		content := helpers.GetString(req.Arguments, "content")
		createDirs := helpers.GetBool(req.Arguments, "create_dirs")

		if createDirs {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return helpers.ErrorResult("mkdir_error", err.Error()), nil
			}
		}

		data := []byte(content)
		if err := os.WriteFile(path, data, 0644); err != nil {
			return helpers.ErrorResult("write_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Written %d bytes to %s", len(data), path)), nil
	}
}
