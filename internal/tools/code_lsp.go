package tools

import (
	"context"
	"fmt"
	"path/filepath"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/storage"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// lspText extracts the text field from a ToolResponse Result struct.
// The convention is that TextResult stores {"text": "..."} in the Result field.
func lspText(resp *pluginv1.ToolResponse) string {
	if resp == nil || resp.Result == nil {
		return ""
	}
	if v, ok := resp.Result.Fields["text"]; ok {
		if sv, ok := v.Kind.(*structpb.Value_StringValue); ok {
			return sv.StringValue
		}
	}
	return ""
}

// hasFields returns an error if any of the given field names are absent from s.
// Unlike ValidateRequired (which uses GetString and rejects numeric 0-values),
// this checks key presence directly — allowing integer fields with value 0.
func hasFields(s *structpb.Struct, fields ...string) error {
	if s == nil {
		return fmt.Errorf("arguments are required: %s", joinFields(fields))
	}
	var missing []string
	for _, f := range fields {
		if _, ok := s.Fields[f]; !ok {
			missing = append(missing, f)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", joinFields(missing))
	}
	return nil
}

func joinFields(fields []string) string {
	result := ""
	for i, f := range fields {
		if i > 0 {
			result += ", "
		}
		result += f
	}
	return result
}

// ---------- code_symbols ----------

// CodeSymbolsSchema returns the JSON schema for the code_symbols tool.
func CodeSymbolsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":    map[string]any{"type": "string", "description": "File path"},
			"content": map[string]any{"type": "string", "description": "Source code content"},
		},
		"required": []any{"path", "content"},
	})
	return s
}

// CodeSymbols returns a handler that lists symbols in a file via the LSP.
func CodeSymbols(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path", "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		content := helpers.GetString(req.Arguments, "content")

		resp, err := ds.CallTool(ctx, "lsp_open_document", map[string]any{
			"path":    path,
			"content": content,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_goto_definition ----------

// CodeGotoDefinitionSchema returns the JSON schema for the code_goto_definition tool.
func CodeGotoDefinitionSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "File path"},
			"line": map[string]any{"type": "integer", "description": "Line number (0-indexed)"},
			"col":  map[string]any{"type": "integer", "description": "Column number (0-indexed)"},
		},
		"required": []any{"path", "line", "col"},
	})
	return s
}

// CodeGotoDefinition returns a handler that finds a symbol's definition location.
func CodeGotoDefinition(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := hasFields(req.Arguments, "path", "line", "col"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		line := helpers.GetInt(req.Arguments, "line")
		col := helpers.GetInt(req.Arguments, "col")

		resp, err := ds.CallTool(ctx, "lsp_goto_definition", map[string]any{
			"path": path,
			"line": line,
			"col":  col,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_find_references ----------

// CodeFindReferencesSchema returns the JSON schema for the code_find_references tool.
func CodeFindReferencesSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "File path"},
			"line": map[string]any{"type": "integer", "description": "Line number (0-indexed)"},
			"col":  map[string]any{"type": "integer", "description": "Column number (0-indexed)"},
		},
		"required": []any{"path", "line", "col"},
	})
	return s
}

// CodeFindReferences returns a handler that finds all references to a symbol.
func CodeFindReferences(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := hasFields(req.Arguments, "path", "line", "col"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		line := helpers.GetInt(req.Arguments, "line")
		col := helpers.GetInt(req.Arguments, "col")

		resp, err := ds.CallTool(ctx, "lsp_find_references", map[string]any{
			"path": path,
			"line": line,
			"col":  col,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_hover ----------

// CodeHoverSchema returns the JSON schema for the code_hover tool.
func CodeHoverSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "File path"},
			"line": map[string]any{"type": "integer", "description": "Line number (0-indexed)"},
			"col":  map[string]any{"type": "integer", "description": "Column number (0-indexed)"},
		},
		"required": []any{"path", "line", "col"},
	})
	return s
}

// CodeHover returns a handler that provides hover information for a symbol.
func CodeHover(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := hasFields(req.Arguments, "path", "line", "col"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		line := helpers.GetInt(req.Arguments, "line")
		col := helpers.GetInt(req.Arguments, "col")

		resp, err := ds.CallTool(ctx, "lsp_hover", map[string]any{
			"path": path,
			"line": line,
			"col":  col,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_complete ----------

// CodeCompleteSchema returns the JSON schema for the code_complete tool.
func CodeCompleteSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":   map[string]any{"type": "string", "description": "File path"},
			"line":   map[string]any{"type": "integer", "description": "Line number (0-indexed)"},
			"col":    map[string]any{"type": "integer", "description": "Column number (0-indexed)"},
			"prefix": map[string]any{"type": "string", "description": "Optional text prefix to filter completions"},
		},
		"required": []any{"path", "line", "col"},
	})
	return s
}

// CodeComplete returns a handler that provides completion candidates at a cursor position.
func CodeComplete(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := hasFields(req.Arguments, "path", "line", "col"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		line := helpers.GetInt(req.Arguments, "line")
		col := helpers.GetInt(req.Arguments, "col")
		prefix := helpers.GetString(req.Arguments, "prefix")

		args := map[string]any{
			"path": path,
			"line": line,
			"col":  col,
		}
		if prefix != "" {
			args["prefix"] = prefix
		}

		resp, err := ds.CallTool(ctx, "lsp_complete", args)
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_diagnostics ----------

// CodeDiagnosticsSchema returns the JSON schema for the code_diagnostics tool.
func CodeDiagnosticsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "File path to diagnose"},
		},
		"required": []any{"path"},
	})
	return s
}

// CodeDiagnostics returns a handler that reports parse errors and diagnostics for a file.
func CodeDiagnostics(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")

		resp, err := ds.CallTool(ctx, "lsp_diagnostics", map[string]any{
			"path": path,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_actions ----------

// CodeActionsSchema returns the JSON schema for the code_actions tool.
func CodeActionsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "File path to get code actions for"},
		},
		"required": []any{"path"},
	})
	return s
}

// CodeActions returns a handler that suggests code actions for a file based on diagnostics.
func CodeActions(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")

		resp, err := ds.CallTool(ctx, "lsp_diagnostics", map[string]any{
			"path": path,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_workspace_symbols ----------

// CodeWorkspaceSymbolsSchema returns the JSON schema for the code_workspace_symbols tool.
func CodeWorkspaceSymbolsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{"type": "string", "description": "Symbol query to search across all open documents"},
		},
		"required": []any{"query"},
	})
	return s
}

// CodeWorkspaceSymbols returns a handler that searches for symbols across all open documents.
func CodeWorkspaceSymbols(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "query"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		query := helpers.GetString(req.Arguments, "query")

		resp, err := ds.CallTool(ctx, "lsp_workspace_symbols", map[string]any{
			"query": query,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(lspText(resp)), nil
	}
}

// ---------- code_namespace ----------

// CodeNamespaceSchema returns the JSON schema for the code_namespace tool.
func CodeNamespaceSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":    map[string]any{"type": "string", "description": "File path"},
			"content": map[string]any{"type": "string", "description": "Source code content"},
		},
		"required": []any{"path", "content"},
	})
	return s
}

// CodeNamespace returns a handler that retrieves the namespace/module structure for a file.
func CodeNamespace(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path", "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		content := helpers.GetString(req.Arguments, "content")

		// Open the document first so it is registered in the LSP.
		_, err := ds.CallTool(ctx, "lsp_open_document", map[string]any{
			"path":    path,
			"content": content,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}

		// Use the file basename as the namespace query.
		query := filepath.Base(path)
		resp, err := ds.CallTool(ctx, "lsp_workspace_symbols", map[string]any{
			"query": query,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(fmt.Sprintf("Namespace for %s:\n%s", path, lspText(resp))), nil
	}
}

// ---------- code_imports ----------

// CodeImportsSchema returns the JSON schema for the code_imports tool.
func CodeImportsSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":    map[string]any{"type": "string", "description": "File path"},
			"content": map[string]any{"type": "string", "description": "Source code content"},
		},
		"required": []any{"path", "content"},
	})
	return s
}

// CodeImports returns a handler that retrieves import declarations for a file.
func CodeImports(ds *storage.DataStorage) func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path", "content"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}
		path := helpers.GetString(req.Arguments, "path")
		content := helpers.GetString(req.Arguments, "content")

		resp, err := ds.CallTool(ctx, "lsp_open_document", map[string]any{
			"path":    path,
			"content": content,
		})
		if err != nil {
			return helpers.ErrorResult("lsp_error", err.Error()), nil
		}
		return helpers.TextResult(fmt.Sprintf("Imports for %s:\n%s", path, lspText(resp))), nil
	}
}
