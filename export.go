package devtoolsfileexplorer

import (
	"context"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/storage"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

// Sender is satisfied by the InProcessRouter.
type Sender interface {
	Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error)
}

// Register adds all file-explorer tools to the builder.
func Register(builder *plugin.PluginBuilder, sender Sender) {
	ds := storage.NewDataStorage(sender)
	tp := internal.NewToolsPlugin(ds)
	tp.RegisterTools(builder)
}
