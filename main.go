package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	mark3lab "github.com/mark3labs/mcp-go/mcp"

	"github.com/mark3labs/mcp-go/server"
	"github.com/universal-tool-calling-protocol/go-utcp"
	"github.com/universal-tool-calling-protocol/go-utcp/src/plugins/chain"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/base"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/cli"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/graphql"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/grpc"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/http"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/mcp"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/sse"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/streamable"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/tcp"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/text"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/udp"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/webrtc"
	"github.com/universal-tool-calling-protocol/go-utcp/src/providers/websocket"
)

// UnmarshalProvider converts JSON into the appropriate providers.Provider implementation.
func UnmarshalProvider(data []byte) (base.Provider, error) {
	var peek struct {
		ProviderType string `json:"provider_type"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider_type: %w", err)
	}

	switch peek.ProviderType {
	case "cli":
		var p cli.CliProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal CLIProvider: %w", err)
		}
		return &p, nil
	case "graphql":
		var p graphql.GraphQLProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal GraphQLProvider: %w", err)
		}
		return &p, nil
	case "grpc":
		var p grpc.GRPCProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal GRPCProvider: %w", err)
		}
		return &p, nil
	case "http":
		var p http.HttpProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal HttpProvider: %w", err)
		}
		return &p, nil
	case "mcp":
		var p mcp.MCPProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal MCPProvider: %w", err)
		}
		return &p, nil
	case "sse":
		var p sse.SSEProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal SSEProvider: %w", err)
		}
		return &p, nil
	case "streamable":
		var p streamable.StreamableHttpProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal StreamableHttpProvider: %w", err)
		}
		return &p, nil
	case "tcp":
		var p tcp.TCPProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal TCPProvider: %w", err)
		}
		return &p, nil
	case "text":
		var p text.TextProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal TextProvider: %w", err)
		}
		return &p, nil
	case "udp":
		var p udp.UDPProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal UDPProvider: %w", err)
		}
		return &p, nil
	case "webrtc":
		var p webrtc.WebRTCProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal WebRTCProvider: %w", err)
		}
		return &p, nil
	case "websocket":
		var p websocket.WebSocketProvider
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal WebSocketProvider: %w", err)
		}
		return &p, nil
	default:
		return nil, fmt.Errorf("unsupported provider_type: %s", peek.ProviderType)
	}
}

// UTCPMCPBridge connects a UTCP client to an MCP server
type UTCPMCPBridge struct {
	utcpClient utcp.UtcpClientInterface
	utcpChain  chain.UtcpChainClient
	mcpServer  *server.MCPServer
}

func NewUTCPMCPBridge(utcpClient utcp.UtcpClientInterface, utcpChain chain.UtcpChainClient) (*UTCPMCPBridge, error) {
	bridge := &UTCPMCPBridge{
		utcpClient: utcpClient,
		utcpChain:  utcpChain,
	}

	// Create MCP server
	mcpServer := server.NewMCPServer("utcp-bridge", "1.0.0", server.WithToolCapabilities(true))
	bridge.mcpServer = mcpServer

	// Register tools
	if err := bridge.registerToolHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register tool handlers: %w", err)
	}

	return bridge, nil
}

func (b *UTCPMCPBridge) registerToolHandlers() error {
	// CallTool
	b.mcpServer.AddTool(mark3lab.Tool{
		Name:        "utcp_call_tool",
		Description: "Call a UTCP tool by name with arguments",
		InputSchema: mark3lab.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"tool_name": map[string]interface{}{"type": "string"},
				"arguments": map[string]interface{}{"type": "object"},
			},
			Required: []string{"tool_name"},
		},
	}, b.handleCallTool)

	// SearchTools
	b.mcpServer.AddTool(mark3lab.Tool{
		Name:        "utcp_search_tools",
		Description: "Search for available UTCP tools",
		InputSchema: mark3lab.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"query": map[string]interface{}{"type": "string"},
				"limit": map[string]interface{}{"type": "integer", "default": 10},
			},
			Required: []string{"query"},
		},
	}, b.handleSearchTools)

	// CallToolStream
	b.mcpServer.AddTool(mark3lab.Tool{
		Name:        "utcp_call_tool_stream",
		Description: "Call a UTCP tool with streaming response",
		InputSchema: mark3lab.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"tool_name": map[string]interface{}{"type": "string"},
				"arguments": map[string]interface{}{"type": "object"},
			},
			Required: []string{"tool_name"},
		},
	}, b.handleCallToolStream)

	// RegisterToolProvider
	b.mcpServer.AddTool(mark3lab.Tool{
		Name:        "utcp_register_provider",
		Description: "Register a new UTCP tool provider",
		InputSchema: mark3lab.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"provider_config": map[string]interface{}{"type": "object"},
			},
			Required: []string{"provider_config"},
		},
	}, b.handleRegisterProvider)
	registerUTCPRunChain(b)

	return nil
}

func registerUTCPRunChain(b *UTCPMCPBridge) {
	b.mcpServer.AddTool(mark3lab.Tool{
		Name:        "utcp_run_chain",
		Description: "Run a UTCP tool chain using UtcpChainClient",
		InputSchema: mark3lab.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"steps": map[string]interface{}{
					"type":  "array",
					"items": map[string]interface{}{"type": "object"},
				},
				"timeout": map[string]interface{}{
					"type":    "integer",
					"default": 30000,
				},
			},
			Required: []string{"steps"},
		},
	}, b.handleRunChain)
}

func (b *UTCPMCPBridge) handleRunChain(
	ctx context.Context,
	request mark3lab.CallToolRequest,
) (*mark3lab.CallToolResult, error) {

	// ----- Extract arguments -----
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mark3lab.NewToolResultError("invalid arguments"), nil
	}

	rawSteps, ok := args["steps"].([]any)
	if !ok {
		return mark3lab.NewToolResultError("steps must be an array"), nil
	}

	// ----- Handle timeout -----
	timeout := 30 * time.Second
	if t, exists := args["timeout"]; exists {
		if tf, ok := t.(float64); ok {
			timeout = time.Duration(int(tf)) * time.Millisecond
		}
	}

	// ----- Map raw MCP steps -> UTCP ChainStep -----
	steps := make([]chain.ChainStep, 0, len(rawSteps))

	for _, raw := range rawSteps {
		m, _ := raw.(map[string]any)

		step := chain.ChainStep{
			ID:          castString(m["id"]),
			ToolName:    castString(m["tool_name"]),
			Inputs:      castMap(m["inputs"]),
			UsePrevious: castBool(m["use_previous"]),
			Stream:      castBool(m["stream"]),
		}

		// Basic validation
		if step.ToolName == "" {
			return mark3lab.NewToolResultError("each step requires tool_name"), nil
		}

		steps = append(steps, step)
	}

	// ----- Execute Chain -----
	result, err := b.utcpChain.CallToolChain(ctx, steps, timeout)
	if err != nil {
		return mark3lab.NewToolResultError(fmt.Sprintf("chain failed: %v", err)), nil
	}

	// ----- Return JSON result -----
	resJSON, _ := json.Marshal(result)
	return mark3lab.NewToolResultText(string(resJSON)), nil
}

// Handlers using server.CallToolRequest / server.CallToolResult
func (b *UTCPMCPBridge) handleCallTool(ctx context.Context, request mark3lab.CallToolRequest) (*mark3lab.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mark3lab.NewToolResultError("invalid arguments"), nil
	}

	toolName, ok := args["tool_name"].(string)
	if !ok {
		return mark3lab.NewToolResultError("tool_name must be string"), nil
	}

	var toolArgs map[string]any
	if a, exists := args["arguments"]; exists {
		toolArgs, _ = a.(map[string]any)
	} else {
		toolArgs = map[string]any{}
	}

	result, err := b.utcpClient.CallTool(ctx, toolName, toolArgs)
	if err != nil {
		return mark3lab.NewToolResultError(fmt.Sprintf("failed to call tool: %v", err)), nil
	}

	resJSON, _ := json.Marshal(result)
	return mark3lab.NewToolResultText(string(resJSON)), nil
}

func (b *UTCPMCPBridge) handleSearchTools(ctx context.Context, request mark3lab.CallToolRequest) (*mark3lab.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mark3lab.NewToolResultError("invalid arguments"), nil
	}
	query, _ := args["query"].(string)
	limit := 10
	if l, exists := args["limit"]; exists {
		if lf, ok := l.(float64); ok {
			limit = int(lf)
		}
	}

	tools, err := b.utcpClient.SearchTools(query, limit)
	if err != nil {
		return mark3lab.NewToolResultError(fmt.Sprintf("failed to search tools: %v", err)), nil
	}

	resJSON, _ := json.Marshal(tools)
	return mark3lab.NewToolResultText(string(resJSON)), nil
}

func (b *UTCPMCPBridge) handleCallToolStream(ctx context.Context, request mark3lab.CallToolRequest) (*mark3lab.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mark3lab.NewToolResultError("invalid arguments"), nil
	}

	toolName, _ := args["tool_name"].(string)
	var toolArgs map[string]any
	if a, exists := args["arguments"]; exists {
		toolArgs, _ = a.(map[string]any)
	} else {
		toolArgs = map[string]any{}
	}

	stream, err := b.utcpClient.CallToolStream(ctx, toolName, toolArgs)
	if err != nil {
		return mark3lab.NewToolResultError(fmt.Sprintf("stream failed: %v", err)), nil
	}

	var chunks []string
	for {
		chunk, err := stream.Next()
		if err != nil {
			return mark3lab.NewToolResultError(fmt.Sprintf("stream error: %v", err)), nil
		}
		if chunk == nil {
			break
		}
		cJSON, _ := json.Marshal(chunk)
		chunks = append(chunks, string(cJSON))
	}

	resJSON, _ := json.Marshal(map[string]any{"chunks": chunks})
	return mark3lab.NewToolResultText(string(resJSON)), nil
}

func (b *UTCPMCPBridge) handleRegisterProvider(ctx context.Context, request mark3lab.CallToolRequest) (*mark3lab.CallToolResult, error) {
	args, ok := request.Params.Arguments.(map[string]any)
	if !ok {
		return mark3lab.NewToolResultError("invalid arguments"), nil
	}

	config, ok := args["provider_config"].(map[string]any)
	if !ok {
		return mark3lab.NewToolResultError("provider_config must be object"), nil
	}

	cfgJSON, _ := json.Marshal(config)
	provider, err := UnmarshalProvider(cfgJSON)
	if err != nil {
		return mark3lab.NewToolResultError(fmt.Sprintf("failed to unmarshal provider: %v", err)), nil
	}

	tools, err := b.utcpClient.RegisterToolProvider(ctx, provider)
	if err != nil {
		return mark3lab.NewToolResultError(fmt.Sprintf("failed to register provider: %v", err)), nil
	}

	resJSON, _ := json.Marshal(map[string]any{
		"registered_tools": tools,
		"count":            len(tools),
	})
	return mark3lab.NewToolResultText(string(resJSON)), nil
}

func (b *UTCPMCPBridge) Start() error {
	return server.ServeStdio(b.mcpServer)
}

func main() {
	ctx := context.Background()

	config := utcp.NewClientConfig()
	if p := os.Getenv("UTCP_PROVIDERS_FILE"); p != "" {
		config.ProvidersFilePath = p
	}

	utcpClient, err := utcp.NewUTCPClient(ctx, config, nil, nil)
	if err != nil {
		log.Fatalf("Failed to create UTCP client: %v", err)
	}
	utcpChain := chain.UtcpChainClient{Client: utcpClient}
	bridge, err := NewUTCPMCPBridge(utcpClient, utcpChain)
	if err != nil {
		log.Fatalf("Failed to create MCP bridge: %v", err)
	}

	log.Println("Starting UTCP MCP Bridge...")
	if err := bridge.Start(); err != nil {
		log.Fatalf("Failed to start MCP bridge: %v", err)
	}
}

func castString(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func castBool(v any) bool {
	b, _ := v.(bool)
	return b
}

func castMap(v any) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}
