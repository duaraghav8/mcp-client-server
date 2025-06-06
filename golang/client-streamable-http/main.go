package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"log"
	"os"
)

func main() {
	// Replace this with your MCP server URL.
	serverURL := "http://127.0.0.1:8080/mcp"

	// Create the streamable HTTP MCP client using the SDK's helper.
	mcpClient, err := client.NewStreamableHttpClient(serverURL)
	if err != nil {
		log.Fatalf("Failed to create streamable HTTP client: %v", err)
	}

	fmt.Println("Initializing client...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP client using streamable http",
		Version: "0.0.1",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	serverInfo, err := mcpClient.Initialize(context.Background(), initRequest)
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	fmt.Println("Server Info:")
	fmt.Println(serverInfo.ServerInfo.Name)
	fmt.Println(serverInfo.Capabilities.Tools)

	fmt.Println("Pinging server...")
	if mcpClient.Ping(context.Background()) != nil {
		log.Fatalf("Failed to ping server: %v", err)
	}
	fmt.Println("Ping successful!")

	ltr := mcp.ListToolsRequest{}
	capabilities, err := mcpClient.ListTools(context.Background(), ltr)
	if err != nil {
		log.Fatalf("Failed to list tools: %v", err)
	}
	fmt.Println("Server Tool Capabilities:")

	for _, tool := range capabilities.Tools {
		fmt.Printf("Tool Name: %s\n", tool.GetName())
		fmt.Printf("Tool Description: %s\n", tool.Description)
		fmt.Printf("Tool Input schema: %v\n", tool.InputSchema)

		j, err := tool.MarshalJSON()
		if err != nil {
			log.Fatalf("Failed to marshal tool to JSON: %v", err)
		}
		fmt.Printf("Tool Raw Input schema: %v\n", string(j))

		fmt.Printf("Annotations: %v\n", tool.Annotations)
		fmt.Println("==============================")
	}

	fmt.Println("Calling tool...")
	callToolReq := mcp.CallToolRequest{}
	callToolReq.Params.Name = "calculator/subtract"
	callToolReq.Params.Arguments = map[string]any{
		"a": 25,
		"b": 10,
	}
	result, err := mcpClient.CallTool(context.Background(), callToolReq)
	if err != nil {
		log.Fatalf("Failed to call tool: %v", err)
	}
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		log.Fatalf("Failed to convert content to TextContent: %v", err)
	}
	fmt.Println("Tool Result:", textContent.Text)

	fmt.Println("Calling tool to get image...")
	callToolReq = mcp.CallToolRequest{}
	callToolReq.Params.Name = "calculator/return_image"

	result, err = mcpClient.CallTool(context.Background(), callToolReq)
	if err != nil {
		log.Fatalf("Failed to call tool: %v", err)
	}
	imageContent, ok := mcp.AsImageContent(result.Content[0])
	if !ok {
		log.Fatalf("Failed to convert content to TextContent: %v", err)
	}
	fmt.Println("Tool Result:", imageContent.Data)

	decoded, err := base64.StdEncoding.DecodeString(imageContent.Data)
	if err != nil {
		log.Fatalf("Failed to decode base64 image data: %v", err)
	}
	err = os.WriteFile("output_image.png", decoded, 0644)
	if err != nil {
		log.Fatalf("Failed to write image to disk: %v", err)
	}

	err = mcpClient.Close()
	if err != nil {
		return
	}
}
