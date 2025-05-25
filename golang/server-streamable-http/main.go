package main

import (
	"context"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	mcpServer := server.NewMCPServer(
		"math-tools",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	echoTool := mcp.NewTool(
		"echo",
		mcp.WithDescription("echoes back your message"),
		mcp.WithString(
			"message",
			mcp.Description("Your message"),
			mcp.Required(),
		),
	)
	mcpServer.AddTool(echoTool, handleEchoToolCall)

	httpServer := server.NewStreamableHTTPServer(mcpServer)
	fmt.Printf("Listening on port :9000/mcp\n")
	if err := httpServer.Start(":9000"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		return
	}
}

func handleEchoToolCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fmt.Println(request.GetString("message", ""))
	fmt.Println(request)
	arguments := request.GetArguments()
	message, ok := arguments["message"].(string)
	fmt.Println(message, ok)
	if !ok {
		return nil, fmt.Errorf("invalid message argument")
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Echo: %s", message),
			},
		},
	}, nil
}
