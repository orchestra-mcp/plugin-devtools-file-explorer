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

// FileInfoSchema returns the JSON Schema for the file_info tool.
func FileInfoSchema() *structpb.Struct {
	s, _ := structpb.NewStruct(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the file or directory to inspect",
			},
		},
		"required": []any{"path"},
	})
	return s
}

// FileInfo returns a tool handler that returns metadata about a file or directory.
func FileInfo() func(context.Context, *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
	return func(ctx context.Context, req *pluginv1.ToolRequest) (*pluginv1.ToolResponse, error) {
		if err := helpers.ValidateRequired(req.Arguments, "path"); err != nil {
			return helpers.ErrorResult("validation_error", err.Error()), nil
		}

		path := helpers.GetString(req.Arguments, "path")

		info, err := os.Stat(path)
		if err != nil {
			return helpers.ErrorResult("stat_error", err.Error()), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Path:     %s\n", path))
		sb.WriteString(fmt.Sprintf("Size:     %s\n", formatSize(info.Size())))
		sb.WriteString(fmt.Sprintf("IsDir:    %v\n", info.IsDir()))
		sb.WriteString(fmt.Sprintf("Mode:     %s\n", info.Mode().String()))
		sb.WriteString(fmt.Sprintf("Modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05 MST")))

		if !info.IsDir() {
			mime := mimeFromExtension(filepath.Ext(path))
			sb.WriteString(fmt.Sprintf("MIME:     %s\n", mime))
		}

		return helpers.TextResult(sb.String()), nil
	}
}

// mimeFromExtension returns a best-guess MIME type based on file extension.
func mimeFromExtension(ext string) string {
	ext = strings.ToLower(ext)
	mimeMap := map[string]string{
		// Text / source code
		".go":    "text/x-go",
		".rs":    "text/x-rust",
		".py":    "text/x-python",
		".js":    "text/javascript",
		".ts":    "text/typescript",
		".tsx":   "text/typescript",
		".jsx":   "text/javascript",
		".html":  "text/html",
		".htm":   "text/html",
		".css":   "text/css",
		".scss":  "text/x-scss",
		".json":  "application/json",
		".yaml":  "text/yaml",
		".yml":   "text/yaml",
		".toml":  "text/x-toml",
		".xml":   "application/xml",
		".svg":   "image/svg+xml",
		".md":    "text/markdown",
		".txt":   "text/plain",
		".sh":    "application/x-sh",
		".bash":  "application/x-sh",
		".zsh":   "application/x-sh",
		".fish":  "text/x-fish",
		".sql":   "application/sql",
		".proto": "text/x-protobuf",
		".c":     "text/x-c",
		".cpp":   "text/x-c++",
		".h":     "text/x-c",
		".java":  "text/x-java",
		".kt":    "text/x-kotlin",
		".swift": "text/x-swift",
		".rb":    "text/x-ruby",
		".php":   "text/x-php",
		".cs":    "text/x-csharp",
		".lua":   "text/x-lua",
		".r":     "text/x-r",
		".scala": "text/x-scala",
		// Images
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".webp": "image/webp",
		".ico":  "image/x-icon",
		// Archives
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".bz2":  "application/x-bzip2",
		".xz":   "application/x-xz",
		// Binaries / other
		".pdf":  "application/pdf",
		".wasm": "application/wasm",
		".exe":  "application/octet-stream",
		".so":   "application/octet-stream",
		".dylib": "application/octet-stream",
	}

	if mime, ok := mimeMap[ext]; ok {
		return mime
	}
	if ext == "" {
		return "application/octet-stream"
	}
	return "application/octet-stream"
}
