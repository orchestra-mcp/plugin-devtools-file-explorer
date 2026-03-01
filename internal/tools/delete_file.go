package tools

import (
	"context"
	"fmt"
	"os"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// DeleteFileSchema returns the JSON Schema for the delete_file tool.
func DeleteFileSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file or directory to delete",
			},
			"recursive": map[string]any{
				"type":        "boolean",
				"description": "Remove directory and all its contents recursively (default: false)",
			},
		},
		"required": []any{"path"},
	})
	return s
}

// DeleteFile returns a tool handler that deletes a file or directory.
func DeleteFile() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		path := helpers.GetString(req.Arguments, "path")
		recursive := helpers.GetBool(req.Arguments, "recursive")

		var err error
		if recursive {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil {
			return helpers.ErrorResult("delete_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Deleted %s", path)), nil
	}
}
