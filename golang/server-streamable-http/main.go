package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type CustomMCPServer struct {
	*server.MCPServer
}

func (s *CustomMCPServer) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	fmt.Printf("********** Adding tool: %s\n", tool.Name)
	s.MCPServer.AddTool(tool, handler)
}

type Person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type InputSchema struct {
	Sender   Person `json:"sender"`
	Receiver Person `json:"receiver"`
}

type OutputSchema struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func main() {
	mcpServer := server.NewMCPServer(
		"math-tools",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	customServer := &CustomMCPServer{mcpServer}

	echoTool := mcp.NewTool(
		"echo",
		mcp.WithDescription("echoes back your message"),
		mcp.WithString(
			"message",
			mcp.Description("Your message"),
			mcp.Required(),
		),
	)
	customServer.AddTool(echoTool, handleEchoToolCall)

	audioTool := mcp.NewTool(
		"return_audio",
		mcp.WithDescription("returns random audio"),
	)
	customServer.AddTool(audioTool, handleAudioToolCall)

	structuredContentTool := mcp.NewTool(
		"structured_content",
		mcp.WithDescription("returns structured content"),
	)
	customServer.AddTool(structuredContentTool, handleStructuredContentCall)

	myResource := mcp.NewResource(
		"file://sample.txt",
		"sample_text_file",
		mcp.WithResourceDescription("Sample text file resource"),
		mcp.WithMIMEType("text/plain"),
	)
	customServer.AddResource(myResource, resourceHandler)

	withSchemaTool := mcp.NewTool(
		"tool_with_output_schema",
		mcp.WithDescription("has schema for output"),
		mcp.WithInputSchema[InputSchema](),
		mcp.WithOutputSchema[OutputSchema](),
	)
	customServer.AddTool(withSchemaTool, handleWithSchemaCall)

	egPrompt := mcp.NewPrompt(
		"echo",
		mcp.WithArgument(
			"message",
			mcp.RequiredArgument(),
			mcp.ArgumentDescription("Message to echo"),
		),
	)

	customServer.AddPrompt(egPrompt, egPromptHandler)

	httpServer := server.NewStreamableHTTPServer(customServer.MCPServer)
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

	embedded := mcp.EmbeddedResource{
		Type: "resource",
		Resource: mcp.TextResourceContents{
			URI:      "resource://embedded-tool",
			MIMEType: "text/plain",
			Text:     "This is embedded resource content!",
		},
		Meta: nil,
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Echo: %s", message),
			},
			mcp.NewResourceLink(
				"file:///example/resource.txt",
				"sample text",
				"An example text resource",
				"text/plain",
			),
			embedded,
		},
	}, nil
}

func handleAudioToolCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	const (
		sampleRate = 8000
		duration   = 1 // seconds
		numSamples = sampleRate * duration
	)

	// WAV header for 16-bit PCM mono
	var buf bytes.Buffer
	// RIFF header
	buf.WriteString("RIFF")
	binary.Write(&buf, binary.LittleEndian, uint32(36+numSamples*2))
	buf.WriteString("WAVE")
	// fmt chunk
	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(16)) // Subchunk1Size
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // AudioFormat PCM
	binary.Write(&buf, binary.LittleEndian, uint16(1))  // NumChannels
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, uint32(sampleRate*2)) // ByteRate
	binary.Write(&buf, binary.LittleEndian, uint16(2))            // BlockAlign
	binary.Write(&buf, binary.LittleEndian, uint16(16))           // BitsPerSample
	// data chunk
	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(numSamples*2))

	// Write random samples
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numSamples; i++ {
		sample := int16(rand.Intn(65536) - 32768)
		binary.Write(&buf, binary.LittleEndian, sample)
	}

	data := buf.Bytes()
	encoded := base64.StdEncoding.EncodeToString(data)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.NewAudioContent(encoded, "audio/wav"),
		},
	}, nil
}

func handleStructuredContentCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	structuredData := map[string]interface{}{
		"title":       "Sample Structured Content",
		"description": "This is an example of structured content returned by a tool.",
		"items": []map[string]interface{}{
			{
				"id":    1,
				"name":  "Item One",
				"value": 100,
			},
			{
				"id":    2,
				"name":  "Item Two",
				"value": 200,
			},
		},
	}

	return mcp.NewToolResultStructured(map[string]any{"data": structuredData}, "Returned structured data"), nil
}

func resourceHandler(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "file://sample.txt",
			MIMEType: "text/plain",
			Text:     "Hello, this is a sample text file content.",
		},
	}, nil
}

func handleWithSchemaCall(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output := OutputSchema{
		Title:       "Sample Title",
		Description: "This is a sample description.",
	}
	return mcp.NewToolResultStructured(output, "Data returned in structured format"), nil
}

func egPromptHandler(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	message, ok := request.Params.Arguments["message"]
	if !ok {
		message = "No message provided"
	}

	messages := []mcp.PromptMessage{
		{
			Role:    mcp.RoleAssistant,
			Content: mcp.NewTextContent("Your message: " + message),
		},
	}
	return &mcp.GetPromptResult{
		Result:      mcp.Result{},
		Description: "Yeh le tera prompt result",
		Messages:    messages,
	}, nil
}
