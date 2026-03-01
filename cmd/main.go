package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	pluginv1 "github.com/orchestra-mcp/gen-go/orchestra/plugin/v1"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal"
	"github.com/orchestra-mcp/plugin-devtools-file-explorer/internal/storage"
	"github.com/orchestra-mcp/sdk-go/plugin"
)

func main() {
	builder := plugin.New("devtools.file-explorer").
		Version("0.1.0").
		Description("File explorer with read/write and code intelligence").
		Author("Orchestra").
		Binary("devtools-file-explorer")

	// Create a lazy client adapter that forwards calls to the orchestrator once
	// the QUIC connection is established during p.Run(). Tool registration
	// happens before Run(), so we wire the plugin reference in after Build().
	adapter := &clientAdapter{}
	ds := storage.NewDataStorage(adapter)

	tp := internal.NewToolsPlugin(ds)
	tp.RegisterTools(builder)

	p := builder.BuildWithTools()
	// Wire the plugin into the adapter so it can retrieve the orchestrator
	// client after Run connects.
	adapter.p = p

	p.ParseFlags()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	if err := p.Run(ctx); err != nil {
		log.Fatalf("devtools.file-explorer: %v", err)
	}
}

// clientAdapter implements storage.StorageClient by forwarding to the plugin's
// orchestrator client. This allows tool handlers to reach the engine-rag plugin
// via the QUIC connection that is established during Run.
type clientAdapter struct {
	p *plugin.Plugin
}

func (a *clientAdapter) Send(ctx context.Context, req *pluginv1.PluginRequest) (*pluginv1.PluginResponse, error) {
	client := a.p.OrchestratorClient()
	if client == nil {
		return nil, fmt.Errorf("orchestrator client not connected")
	}
	return client.Send(ctx, req)
}
