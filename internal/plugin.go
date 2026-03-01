package internal

import (
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/storage"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/tools"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// ToolsPlugin registers all file-explorer tools with the plugin builder.
type ToolsPlugin struct {
	ds *storage.DataStorage
}

// NewToolsPlugin creates a ToolsPlugin with the given DataStorage for cross-plugin LSP calls.
func NewToolsPlugin(ds *storage.DataStorage) *ToolsPlugin {
	return &ToolsPlugin{ds: ds}
}

// RegisterTools registers all 17 file-explorer tools.
func (tp *ToolsPlugin) RegisterTools(builder *plugin.PluginBuilder) {
	// File system tools
	builder.RegisterTool("list_directory", "List directory contents with optional hidden-file and recursive support", tools.ListDirectorySchema(), tools.ListDirectory())
	builder.RegisterTool("read_file", "Read file contents with optional line offset and limit", tools.ReadFileSchema(), tools.ReadFile())
	builder.RegisterTool("write_file", "Write content to a file, optionally creating parent directories", tools.WriteFileSchema(), tools.WriteFile())
	builder.RegisterTool("move_file", "Move or rename a file or directory", tools.MoveFileSchema(), tools.MoveFile())
	builder.RegisterTool("delete_file", "Delete a file or directory, optionally recursively", tools.DeleteFileSchema(), tools.DeleteFile())
	builder.RegisterTool("file_info", "Return metadata (size, permissions, MIME type) for a file or directory", tools.FileInfoSchema(), tools.FileInfo())
	builder.RegisterTool("file_search", "Search for files matching a glob pattern within a directory", tools.FileSearchSchema(), tools.FileSearch())

	// IDE/LSP tools
	builder.RegisterTool("code_symbols", "List symbols (functions, structs, classes) in a file by opening it in the LSP document store", tools.CodeSymbolsSchema(), tools.CodeSymbols(tp.ds))
	builder.RegisterTool("code_goto_definition", "Return the definition location for a symbol at a given file position", tools.CodeGotoDefinitionSchema(), tools.CodeGotoDefinition(tp.ds))
	builder.RegisterTool("code_find_references", "Find all references to the symbol at a given position across open documents", tools.CodeFindReferencesSchema(), tools.CodeFindReferences(tp.ds))
	builder.RegisterTool("code_hover", "Return hover information (doc comment, type signature) for the symbol at a position", tools.CodeHoverSchema(), tools.CodeHover(tp.ds))
	builder.RegisterTool("code_complete", "Get completion candidates at a cursor position", tools.CodeCompleteSchema(), tools.CodeComplete(tp.ds))
	builder.RegisterTool("code_diagnostics", "Get parse errors and diagnostics for a file", tools.CodeDiagnosticsSchema(), tools.CodeDiagnostics(tp.ds))
	builder.RegisterTool("code_actions", "Get suggested code actions for a file (based on diagnostics)", tools.CodeActionsSchema(), tools.CodeActions(tp.ds))
	builder.RegisterTool("code_workspace_symbols", "Search for symbols across all open documents", tools.CodeWorkspaceSymbolsSchema(), tools.CodeWorkspaceSymbols(tp.ds))
	builder.RegisterTool("code_namespace", "Get the namespace/module structure for a file", tools.CodeNamespaceSchema(), tools.CodeNamespace(tp.ds))
	builder.RegisterTool("code_imports", "Get import declarations for a file", tools.CodeImportsSchema(), tools.CodeImports(tp.ds))
}
