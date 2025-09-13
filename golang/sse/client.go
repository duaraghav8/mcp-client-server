package main

import (
	"context"
	"fmt"
	"log"
	"time"

	mcpc "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// adjust if your server mounted the endpoint elsewhere (default is /mcp)
const baseURL = "http://localhost:9000/sse"

func main() {
	// Optional: attach headers (e.g., bearer token) via mcpc.WithHeaders(...)
	cli, err := mcpc.NewSSEMCPClient(
		baseURL,
		// mcpc.WithHeaders(map[string]string{"Authorization": "Bearer <token>"}),
	)
	if err != nil {
		log.Fatalf("new SSE client: %v", err)
	}
	defer func() {
		if cerr := cli.Close(); cerr != nil {
			log.Printf("close error: %v", cerr)
		}
	}()

	// Receive async notifications (server->client) over SSE
	cli.OnNotification(func(n mcp.JSONRPCNotification) {
		log.Printf("notification received: method=%s data=%v", n.Method, n.Params)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1) Connect transport (SSE GET + POST wiring)
	if err := cli.Start(ctx); err != nil {
		log.Fatalf("start: %v", err)
	}

	// 2) Initialize (protocol negotiation + capabilities)
	// ProtocolVersion string is date-based per spec. 2024-11-05 is widely used/accepted.
	initReq := mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcp.ClientCapabilities{}, // fill if you advertise extras
			ClientInfo:      mcp.Implementation{Name: "mcp-go-sse-client", Version: "0.1.0"},
		},
	}
	initRes, err := cli.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("initialize: %v", err)
	}
	fmt.Printf("Server: %s %s (protocol %s)\n",
		initRes.ServerInfo.Name, initRes.ServerInfo.Version, initRes.ProtocolVersion)

	// 3) List tools
	toolsRes, err := cli.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("list tools: %v", err)
	}
	fmt.Println("Tools:")
	for _, t := range toolsRes.Tools {
		fmt.Printf("  - %s: %s\n", t.Name, t.Description)
	}

	// 4) Optionally call a tool if the server exposes one named "ping" or similar
	//    Change "ping" and arguments to something your server actually implements.
	tryCall(ctx, cli, "ping", map[string]any{"message": "hello from mcp-go SSE client"})

	time.Sleep(20 * time.Second)
}

func tryCall(ctx context.Context, cli *mcpc.Client, tool string, args map[string]any) {
	res, err := cli.CallTool(ctx, mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      tool,
			Arguments: args,
		},
	})
	if err != nil {
		log.Printf("call tool %q error: %v", tool, err)
		return
	}
	fmt.Printf("Tool %q result:\n", tool)
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			fmt.Printf("  text: %s\n", tc.Text)
		} else {
			fmt.Printf("  content: %#v\n", c)
		}
	}
}
