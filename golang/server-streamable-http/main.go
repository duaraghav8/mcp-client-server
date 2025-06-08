package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"math/rand"
	"time"
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

	audioTool := mcp.NewTool(
		"return_audio",
		mcp.WithDescription("returns random audio"),
	)
	mcpServer.AddTool(audioTool, handleAudioToolCall)

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
