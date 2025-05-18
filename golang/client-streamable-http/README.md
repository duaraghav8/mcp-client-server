This project implements a MCP client that uses streamable HTTP transport in Golang.

It connects to the server (assuming mcp server is running), performs initialization, gets server capabilities and calls tool(s).

As of this writing, Anthropic doesn't provide an official Golang sdk for MCP. So we use mcp-go.