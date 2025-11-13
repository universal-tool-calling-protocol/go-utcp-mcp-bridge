![MCP vs. UTCP](https://github.com/universal-tool-calling-protocol/.github/raw/main/assets/banner.png)

**UTCP MCP Bridge** connects the [Universal Tool Calling Protocol (UTCP)](https://github.com/universal-tool-calling-protocol/go-utcp) to an [MCP (Mark3Labs Communication Protocol) server](https://github.com/mark3labs/mcp-go), allowing UTCP tools to be exposed and called via MCP-compatible clients.

This bridge enables seamless integration between UTCP tools and any MCP-based ecosystem, providing standard tool invocation, search, streaming, and provider registration functionalities.

---

## Features

- **Call UTCP tools via MCP** (`utcp_call_tool`)
- **Stream UTCP tool output** (`utcp_call_tool_stream`)
- **Search available UTCP tools** (`utcp_search_tools`)
- **Register new UTCP providers dynamically** (`utcp_register_provider`)
- Compatible with all UTCP provider types: CLI, GraphQL, gRPC, HTTP, SSE, TCP, UDP, WebSocket, WebRTC, Text, Streamable HTTP, and MCP.

---

## Installation

```bash
git clone https://github.com/your-org/utcp-mcp-bridge.git
cd utcp-mcp-bridge
go mod tidy
go build -o utcp-mcp-bridge main.go
sudo mv utcp-mcp-bridge /usr/local/bin/
```
