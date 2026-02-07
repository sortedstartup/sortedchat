package sortedagents

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPToolAdapter adapts an MCP tool to the sortedagents.Tool interface
type MCPToolAdapter struct {
	session *mcp.ClientSession
	tool    mcp.Tool
	schema  *JSONSchema
}

// NewMCPToolAdapter creates a new adapter for an MCP tool
func NewMCPToolAdapter(session *mcp.ClientSession, tool *mcp.Tool) *MCPToolAdapter {
	adapter := &MCPToolAdapter{
		session: session,
		tool:    *tool,
	}
	adapter.schema = adapter.convertSchema(tool.InputSchema)
	return adapter
}

func (t *MCPToolAdapter) Name() string {
	return t.tool.Name
}

func (t *MCPToolAdapter) Description() string {
	return t.tool.Description
}

func (t *MCPToolAdapter) Parameters() *JSONSchema {
	return t.schema
}

func (t *MCPToolAdapter) Execute(ctx context.Context, args map[string]any) (any, error) {
	// Use CallTool with CallToolParams
	params := &mcp.CallToolParams{
		Name:      t.tool.Name,
		Arguments: args,
	}

	result, err := t.session.CallTool(ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// convertSchema converts the generic MCP input schema to our JSONSchema struct
func (t *MCPToolAdapter) convertSchema(inputSchema interface{}) *JSONSchema {
	// MCP schemas are typically generic maps or structs.
	// We use JSON round-tripping to convert them to our internal JSONSchema definition.
	bytes, err := json.Marshal(inputSchema)
	if err != nil {
		// Fallback for failed marshalling
		return &JSONSchema{Type: "object"}
	}

	var schema JSONSchema
	if err := json.Unmarshal(bytes, &schema); err != nil {
		// Fallback for failed unmarshalling
		return &JSONSchema{Type: "object"}
	}

	return &schema
}

// ConnectToMCPStdioServer connects to an MCP server running via stdio
// command is the executable (e.g. "npx", "python"), args are the arguments
func ConnectToMCPStdioServer(ctx context.Context, command string, args ...string) (*mcp.ClientSession, func(), error) {
	transport := &mcp.CommandTransport{
		Command: exec.Command(command, args...),
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "sortedagents-client",
		Version: "1.0.0",
	}, nil)

	session, err := mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MCP server: %w", err)
	}

	// cleanup function
	cleanup := func() {
		session.Close()
	}

	return session, cleanup, nil
}

// ConnectToMCPHTTPServer connects to a remote MCP server using Streamable HTTP transport
func ConnectToMCPHTTPServer(ctx context.Context, url string) (*mcp.ClientSession, func(), error) {
	transport := &mcp.StreamableClientTransport{
		Endpoint: url,
	}

	mcpClient := mcp.NewClient(&mcp.Implementation{
		Name:    "sortedagents-client",
		Version: "1.0.0",
	}, nil)

	session, err := mcpClient.Connect(ctx, transport, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to MCP HTTP server: %w", err)
	}

	cleanup := func() {
		session.Close()
	}

	return session, cleanup, nil
}

// LoadMCPTools connects to an MCP server, fetches tools, and returns them as []Tool
func LoadMCPTools(ctx context.Context, command string, args ...string) ([]Tool, func(), error) {
	session, cleanup, err := ConnectToMCPStdioServer(ctx, command, args...)
	if err != nil {
		return nil, nil, err
	}

	listResp, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("failed to list tools: %w", err)
	}

	var tools []Tool
	for _, tool := range listResp.Tools {
		tools = append(tools, NewMCPToolAdapter(session, tool))
	}

	return tools, cleanup, nil
}
