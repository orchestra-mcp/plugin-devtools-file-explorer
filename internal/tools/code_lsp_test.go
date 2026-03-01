package tools_test

import (
	"context"
	"testing"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/storage"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/tools"
	"google.golang.org/protobuf/types/known/structpb"
)

// ---------- Helpers ----------

// makeArgs builds a *structpb.Struct from a plain map for use in ToolRequest.
func makeArgs(t *testing.T, m map[string]any) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m)
	if err != nil {
		t.Fatalf("structpb.NewStruct: %v", err)
	}
	return s
}

// toolRespWithText builds a PluginResponse whose ToolCall carries a text result,
// mirroring what helpers.TextResult produces ({"text": "..."} inside Result).
func toolRespWithText(text string) *pluginv1.PluginResponse {
	result, _ := structpb.NewStruct(map[string]any{"text": text})
	return &pluginv1.PluginResponse{
		Response: &pluginv1.PluginResponse_ToolCall{
			ToolCall: &pluginv1.ToolResponse{
				Success: true,
				Result:  result,
			},
		},
	}
}

// respText extracts the "text" field from a ToolResponse Result struct,
// which is how helpers.TextResult stores the string value.
func respText(resp *pluginv1.ToolResponse) string {
	if resp == nil || resp.Result == nil {
		return ""
	}
	v, ok := resp.Result.Fields["text"]
	if !ok {
		return ""
	}
	sv, ok := v.Kind.(*structpb.Value_StringValue)
	if !ok {
		return ""
	}
	return sv.StringValue
}

// ---------- Mock: single-call client ----------

// mockLSPClient records the last tool call and returns a configurable response.
type mockLSPClient struct {
	calledTool   string
	calledArgs   map[string]any
	responseText string
	err          error
}

func (m *mockLSPClient) Send(_ context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	if tc := req.GetToolCall(); tc != nil {
		m.calledTool = tc.GetToolName()
		if tc.GetArguments() != nil {
			m.calledArgs = tc.GetArguments().AsMap()
		}
	}
	return toolRespWithText(m.responseText), nil
}

// ---------- Mock: multi-call client ----------

// multiCallMock records every tool call and returns responses from a queue.
type multiCallMock struct {
	calledTools []string
	responses   []string
	idx         int
}

func (m *multiCallMock) Send(_ context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error) {
	if tc := req.GetToolCall(); tc != nil {
		m.calledTools = append(m.calledTools, tc.GetToolName())
	}
	text := ""
	if m.idx < len(m.responses) {
		text = m.responses[m.idx]
		m.idx++
	}
	return toolRespWithText(text), nil
}

// ---------- code_symbols ----------

func TestCodeSymbols_CallsLspOpenDocument(t *testing.T) {
	mock := &mockLSPClient{responseText: "func Foo() {}\nstruct Bar {}"}
	ds := storage.NewDataStorage(mock)
	handler := tools.CodeSymbols(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_symbols",
		Arguments: makeArgs(t, map[string]any{
			"path":    "/src/main.go",
			"content": "package main\nfunc Foo() {}",
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if respText(resp) == "" {
		t.Error("expected non-empty response text")
	}
	if mock.calledTool != "lsp_open_document" {
		t.Errorf("expected lsp_open_document, got %q", mock.calledTool)
	}
	if got := mock.calledArgs["path"]; got != "/src/main.go" {
		t.Errorf("path arg: got %v, want /src/main.go", got)
	}
}

func TestCodeSymbols_MissingPath(t *testing.T) {
	ds := storage.NewDataStorage(&mockLSPClient{})
	handler := tools.CodeSymbols(ds)

	req := &pluginv1.ToolRequest{
		ToolName:  "code_symbols",
		Arguments: makeArgs(t, map[string]any{"content": "package main"}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSuccess() {
		t.Error("expected error result when path is missing")
	}
	if resp.GetErrorCode() == "" {
		t.Error("expected non-empty error code")
	}
}

// ---------- code_goto_definition ----------

func TestCodeGotoDefinition_CallsLspGotoDefinition(t *testing.T) {
	mock := &mockLSPClient{responseText: "/src/defs.go:10:5"}
	ds := storage.NewDataStorage(mock)
	handler := tools.CodeGotoDefinition(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_goto_definition",
		Arguments: makeArgs(t, map[string]any{
			"path": "/src/main.go",
			"line": float64(5),
			"col":  float64(12),
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if respText(resp) == "" {
		t.Error("expected non-empty response text")
	}
	if mock.calledTool != "lsp_goto_definition" {
		t.Errorf("expected lsp_goto_definition, got %q", mock.calledTool)
	}
	if got := mock.calledArgs["path"]; got != "/src/main.go" {
		t.Errorf("path arg: got %v, want /src/main.go", got)
	}
}

func TestCodeGotoDefinition_MissingLine(t *testing.T) {
	ds := storage.NewDataStorage(&mockLSPClient{})
	handler := tools.CodeGotoDefinition(ds)

	req := &pluginv1.ToolRequest{
		ToolName:  "code_goto_definition",
		Arguments: makeArgs(t, map[string]any{"path": "/src/main.go", "col": float64(0)}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSuccess() {
		t.Error("expected error result when line is missing")
	}
}

// ---------- code_workspace_symbols ----------

func TestCodeWorkspaceSymbols_CallsLspWorkspaceSymbols(t *testing.T) {
	mock := &mockLSPClient{responseText: "Foo: function\nBar: struct"}
	ds := storage.NewDataStorage(mock)
	handler := tools.CodeWorkspaceSymbols(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_workspace_symbols",
		Arguments: makeArgs(t, map[string]any{
			"query": "Foo",
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if respText(resp) == "" {
		t.Error("expected non-empty response text")
	}
	if mock.calledTool != "lsp_workspace_symbols" {
		t.Errorf("expected lsp_workspace_symbols, got %q", mock.calledTool)
	}
	if got := mock.calledArgs["query"]; got != "Foo" {
		t.Errorf("query arg: got %v, want Foo", got)
	}
}

func TestCodeWorkspaceSymbols_MissingQuery(t *testing.T) {
	ds := storage.NewDataStorage(&mockLSPClient{})
	handler := tools.CodeWorkspaceSymbols(ds)

	req := &pluginv1.ToolRequest{
		ToolName:  "code_workspace_symbols",
		Arguments: makeArgs(t, map[string]any{}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSuccess() {
		t.Error("expected error result when query is missing")
	}
}

// ---------- code_diagnostics ----------

func TestCodeDiagnostics_CallsLspDiagnostics(t *testing.T) {
	mock := &mockLSPClient{responseText: "no errors"}
	ds := storage.NewDataStorage(mock)
	handler := tools.CodeDiagnostics(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_diagnostics",
		Arguments: makeArgs(t, map[string]any{
			"path": "/src/main.go",
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if got := respText(resp); got != "no errors" {
		t.Errorf("expected 'no errors', got %q", got)
	}
	if mock.calledTool != "lsp_diagnostics" {
		t.Errorf("expected lsp_diagnostics, got %q", mock.calledTool)
	}
	if got := mock.calledArgs["path"]; got != "/src/main.go" {
		t.Errorf("path arg: got %v, want /src/main.go", got)
	}
}

func TestCodeDiagnostics_MissingPath(t *testing.T) {
	ds := storage.NewDataStorage(&mockLSPClient{})
	handler := tools.CodeDiagnostics(ds)

	req := &pluginv1.ToolRequest{
		ToolName:  "code_diagnostics",
		Arguments: makeArgs(t, map[string]any{}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSuccess() {
		t.Error("expected error result when path is missing")
	}
}

// ---------- code_actions ----------

func TestCodeActions_DelegatesToLspDiagnostics(t *testing.T) {
	mock := &mockLSPClient{responseText: "fix import"}
	ds := storage.NewDataStorage(mock)
	handler := tools.CodeActions(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_actions",
		Arguments: makeArgs(t, map[string]any{
			"path": "/src/main.go",
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if respText(resp) == "" {
		t.Error("expected non-empty response text")
	}
	// code_actions reuses lsp_diagnostics under the hood
	if mock.calledTool != "lsp_diagnostics" {
		t.Errorf("expected lsp_diagnostics, got %q", mock.calledTool)
	}
}

// ---------- code_namespace ----------

func TestCodeNamespace_OpensThenQueriesWorkspaceSymbols(t *testing.T) {
	multiMock := &multiCallMock{
		responses: []string{"symbols opened", "namespace result"},
	}
	ds := storage.NewDataStorage(multiMock)
	handler := tools.CodeNamespace(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_namespace",
		Arguments: makeArgs(t, map[string]any{
			"path":    "/src/utils.go",
			"content": "package utils",
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if respText(resp) == "" {
		t.Error("expected non-empty response text")
	}
	// Must call lsp_open_document first, then lsp_workspace_symbols
	if len(multiMock.calledTools) < 2 {
		t.Fatalf("expected 2 tool calls, got %d: %v", len(multiMock.calledTools), multiMock.calledTools)
	}
	if multiMock.calledTools[0] != "lsp_open_document" {
		t.Errorf("first call: expected lsp_open_document, got %q", multiMock.calledTools[0])
	}
	if multiMock.calledTools[1] != "lsp_workspace_symbols" {
		t.Errorf("second call: expected lsp_workspace_symbols, got %q", multiMock.calledTools[1])
	}
}

// ---------- code_imports ----------

func TestCodeImports_CallsLspOpenDocument(t *testing.T) {
	mock := &mockLSPClient{responseText: "import \"fmt\""}
	ds := storage.NewDataStorage(mock)
	handler := tools.CodeImports(ds)

	req := &pluginv1.ToolRequest{
		ToolName: "code_imports",
		Arguments: makeArgs(t, map[string]any{
			"path":    "/src/main.go",
			"content": "package main\nimport \"fmt\"",
		}),
	}

	resp, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.GetSuccess() {
		t.Errorf("expected success, got error: %s", resp.GetErrorMessage())
	}
	if respText(resp) == "" {
		t.Error("expected non-empty response text")
	}
	if mock.calledTool != "lsp_open_document" {
		t.Errorf("expected lsp_open_document, got %q", mock.calledTool)
	}
}
