package storage

import (
	"context"
	"fmt"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/sdk-go/helpers"
	"google.golang.org/protobuf/types/known/structpb"
)

// StorageClient sends requests to the orchestrator for storage operations.
type StorageClient interface {
	Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error)
}

// DataStorage wraps the storage client for tool handlers.
type DataStorage struct {
	client StorageClient
}

// NewDataStorage creates a new DataStorage with the given client.
func NewDataStorage(client StorageClient) *DataStorage {
	return &DataStorage{client: client}
}

// CallTool sends a ToolRequest to the orchestrator, routing to the target plugin.
func (ds *DataStorage) CallTool(ctx context.Context, toolName string, args map[string]any) (*pluginv1.ToolResponse, error) {
	argsStruct, err := structpb.NewStruct(args)
	if err != nil {
		return nil, fmt.Errorf("build args for %s: %w", toolName, err)
	}
	req := &pluginv1.PluginRequest{
		RequestId: helpers.NewUUID(),
		Request: &pluginv1.PluginRequest_ToolCall{
			ToolCall: &pluginv1.ToolRequest{
				ToolName:  toolName,
				Arguments: argsStruct,
			},
		},
	}
	resp, err := ds.client.Send(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("call tool %s: %w", toolName, err)
	}
	tc := resp.GetToolCall()
	if tc == nil {
		return nil, fmt.Errorf("unexpected response type for %s", toolName)
	}
	return tc, nil
}
