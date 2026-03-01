package tools

import (
	"context"
	"fmt"
	"os"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// MoveFileSchema returns the JSON Schema for the move_file tool.
func MoveFileSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"source": map[string]any{
				"type":        "string",
				"description": "Source file or directory path",
			},
			"destination": map[string]any{
				"type":        "string",
				"description": "Destination file or directory path",
			},
		},
		"required": []any{"source", "destination"},
	})
	return s
}

// MoveFile returns a tool handler that moves (renames) a file or directory.
func MoveFile() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "source", "destination"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		source := helpers.GetString(req.Arguments, "source")
		destination := helpers.GetString(req.Arguments, "destination")

		if err := os.Rename(source, destination); err != nil {
			return helpers.ErrorResult("move_error", err.Error()), nil
		}

		return helpers.TextResult(fmt.Sprintf("Moved %s → %s", source, destination)), nil
	}
}
