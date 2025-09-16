package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var (
		mu       sync.RWMutex
		sessions = make(map[string]struct{}) // sessionID -> present
	)

	hooks := &server.Hooks{}
	hooks.AddOnRegisterSession(func(ctx context.Context, sess server.ClientSession) {
		mu.Lock()
		defer mu.Unlock()
		sessions[sess.SessionID()] = struct{}{}
		log.Printf("[sessions] + %s (now %d)", sess.SessionID(), len(sessions))
	})

	hooks.AddOnUnregisterSession(func(ctx context.Context, sess server.ClientSession) {
		mu.Lock()
		defer mu.Unlock()
		delete(sessions, sess.SessionID())
		log.Printf("[sessions] - %s (now %d)", sess.SessionID(), len(sessions))
	})

	// 1) Core MCP server (name/version are arbitrary)
	s := server.NewMCPServer(
		"mcp-go-sse-demo",
		"0.1.0",
		server.WithLogging(),               // enable logging notifications
		server.WithToolCapabilities(false), // advertise tools
		server.WithHooks(hooks),
	)

	// 2) Register simple tools
	registerTools(s)

	// 3) HTTP+SSE transport on :9000
	//    - BaseURL is REQUIRED so the server can tell clients where to POST messages.
	//    - Endpoints default to /mcp/sse and /mcp/message. We set a static base path /mcp.
	sse := server.NewSSEServer(
		s,
		//server.WithBaseURL("http://localhost:9000"),
		//server.WithStaticBasePath("/mcp"),          // base path prefix
		//server.WithSSEEndpoint("/mcp/sse"),         // GET (event stream)
		//server.WithMessageEndpoint("/mcp/message"), // POST (JSON-RPC)
	)

	// 4) Graceful shutdown on SIGINT/SIGTERM
	shutdown := make(chan struct{})
	go func() {
		if err := sse.Start("localhost:9000"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("SSE server failed: %v", err)
		}
		close(shutdown)
	}()

	log.Println("SSE MCP server listening on http://localhost:9000 (SSE: /mcp/sse, POST: /mcp/message)")

	// 5) Send notifications periodically
	go func() {
		time.Sleep(10 * time.Second) // wait a moment for client to connect
		for i := 1; i <= 5; i++ {
			msg := map[string]any{
				"info":  "server notification",
				"count": i,
				"time":  time.Now().Format(time.RFC3339),
			}
			log.Printf("Sending notification #%d to all clients", i)
			// method name is arbitrary; client will see it in n.Method
			s.SendNotificationToAllClients("server/ping", msg)
			time.Sleep(2 * time.Second)

			var target string
			mu.RLock()
			for sid := range sessions {
				target = sid
				break
			}
			mu.RUnlock()

			fmt.Println("Sending private notification to ", target)

			err := s.SendNotificationToSpecificClient(
				target,
				"server/ping",
				map[string]any{
					"info": "private message to this MF",
					"time": time.Now().Format(time.RFC3339),
				},
			)
			if err != nil {
				fmt.Printf("Failed to send private notification: %v", err)
			}

			time.Sleep(2 * time.Second)
		}
	}()

	// Wait for signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	fmt.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sse.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	<-shutdown
	log.Println("Server stopped.")
}

func registerTools(s *server.MCPServer) {
	// Tool: ping — echoes a message
	pingTool := mcp.NewTool(
		"ping",
		mcp.WithDescription("Echo a message back (for connectivity checks)."),
		mcp.WithString("message",
			mcp.Description("Any text to echo back."),
			mcp.Required(),
		),
	)

	s.AddTool(pingTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		msg := req.GetString("message", "")
		if msg == "" {
			return mcp.NewToolResultError("missing required arg: message"), nil
		}
		return mcp.NewToolResultText("pong: " + msg), nil
	})

	// Tool: add — sums two numbers and returns a structured result
	addTool := mcp.NewTool(
		"add",
		mcp.WithDescription("Return a+b."),
		mcp.WithNumber("a", mcp.Description("First addend."), mcp.Required()),
		mcp.WithNumber("b", mcp.Description("Second addend."), mcp.Required()),
	)
	s.AddTool(addTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		a := req.GetFloat("a", 0)
		b := req.GetFloat("b", 0)
		return mcp.NewToolResultStructured(map[string]any{
			"sum": a + b,
		}, "sum computed"), nil
	})
}
