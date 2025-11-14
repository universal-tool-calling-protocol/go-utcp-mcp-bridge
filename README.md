![MCP vs. UTCP](https://github.com/universal-tool-calling-protocol/.github/raw/main/assets/banner.png)

This **utcp bridge** enables seamless integration between UTCP tools and any MCP-based ecosystem, providing standard tool invocation, search, streaming, and provider registration functionalities.
A lightweight Go-based bridge that exposes **UTCP tools**, **UTCP chains**, and **UTCP CodeMode** execution as **MCP tools** — enabling any MCP-compatible client (Claude Desktop, Claude CLI, LLM runtimes implementing MCP) to call UTCP tools seamlessly.

This bridge lets you:

- 🔌 Load UTCP providers dynamically from JSON  
- 🛠 Call UTCP tools via MCP  
- 🔍 Search the UTCP tool registry  
- 🔄 Stream UTCP tool results over MCP  
- ⛓️ Execute multi-step **UTCP Chains** via a single MCP call  
- 🧩 Run **Go CodeMode** snippets through UTCP (inline Go execution)  
- 🤝 Register new providers dynamically at runtime  

Designed with flexibility in mind, the bridge can power anything from local tool-automation setups to distributed LLM agent workflows.

---

## Features

### ✓ UTCP → MCP Tool Mapping

| MCP Tool Name             | Description |
|---------------------------|-------------|
| `utcp_call_tool`          | Call any UTCP tool with arguments |
| `utcp_search_tools`       | Fuzzy-search tools in UTCP registry |
| `utcp_call_tool_stream`   | Stream responses from UTCP tools |
| `utcp_register_provider`  | Register new UTCP provider at runtime |
| `utcp_run_chain`          | Execute UTCP tool chain (ChainStep[]) |
| `utcp_run_code`           | Execute Go CodeMode                   |

---


## Installation

```bash
git clone https://github.com/your-org/utcp-mcp-bridge.git
cd utcp-mcp-bridge
go mod tidy
go build -o utcp-mcp-bridge main.go
sudo mv utcp-mcp-bridge /usr/local/bin/
```
