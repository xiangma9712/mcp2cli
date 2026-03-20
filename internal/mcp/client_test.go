package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestClientInitializeAndListTools(t *testing.T) {
	var reqCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch req.Method {
		case "initialize":
			reqCount.Add(1)
			w.Header().Set("Mcp-Session-Id", "test-session-123")
			w.Header().Set("Content-Type", "application/json")
			result, _ := json.Marshal(InitializeResult{
				ProtocolVersion: "2025-03-26",
				ServerInfo:      AppInfo{Name: "test-server", Version: "1.0.0"},
			})
			resp := Response{JSONRPC: "2.0", ID: req.ID, Result: (*RawResult)(&result)}
			json.NewEncoder(w).Encode(resp)

		case "notifications/initialized":
			reqCount.Add(1)
			w.WriteHeader(http.StatusAccepted)

		case "tools/list":
			reqCount.Add(1)
			tools := ToolsListResult{
				Tools: []Tool{
					{
						Name:        "get_weather",
						Description: "Get weather for a location",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"location": map[string]any{
									"type":        "string",
									"description": "City name",
								},
							},
							"required": []any{"location"},
						},
					},
				},
			}
			result, _ := json.Marshal(tools)
			resp := Response{JSONRPC: "2.0", ID: req.ID, Result: (*RawResult)(&result)}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		case "tools/call":
			reqCount.Add(1)
			callResult := ToolCallResult{
				Content: []ContentItem{
					{Type: "text", Text: "Sunny, 72F"},
				},
			}
			result, _ := json.Marshal(callResult)
			resp := Response{JSONRPC: "2.0", ID: req.ID, Result: (*RawResult)(&result)}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)

		default:
			http.Error(w, "unknown method: "+req.Method, http.StatusBadRequest)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client := NewClient(server.URL)

	// Initialize
	initResult, err := client.Initialize(ctx, "test-client", "0.1.0")
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if initResult.ServerInfo.Name != "test-server" {
		t.Errorf("expected server name test-server, got %s", initResult.ServerInfo.Name)
	}
	if client.sessionID != "test-session-123" {
		t.Errorf("expected session ID test-session-123, got %s", client.sessionID)
	}

	// ListTools
	tools, err := client.ListTools(ctx)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "get_weather" {
		t.Errorf("expected tool name get_weather, got %s", tools[0].Name)
	}

	// CallTool
	result, err := client.CallTool(ctx, "get_weather", map[string]any{"location": "NYC"})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if len(result.Content) != 1 || result.Content[0].Text != "Sunny, 72F" {
		t.Errorf("unexpected result: %+v", result)
	}

	// Verify all expected requests were made (initialize + initialized + tools/list + tools/call)
	if got := reqCount.Load(); got != 4 {
		t.Errorf("expected 4 requests, got %d", got)
	}
}

func TestClientSSEResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		json.NewDecoder(r.Body).Decode(&req)

		if req.Method == "initialize" {
			result, _ := json.Marshal(InitializeResult{
				ProtocolVersion: "2025-03-26",
				ServerInfo:      AppInfo{Name: "sse-server", Version: "1.0.0"},
			})
			// Respond with SSE format
			w.Header().Set("Content-Type", "text/event-stream")
			resp, _ := json.Marshal(Response{JSONRPC: "2.0", ID: req.ID, Result: (*RawResult)(&result)})
			w.Write([]byte("data: " + string(resp) + "\n\n"))
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.Initialize(context.Background(), "test", "0.1.0")
	if err != nil {
		t.Fatalf("Initialize with SSE: %v", err)
	}
	if result.ServerInfo.Name != "sse-server" {
		t.Errorf("expected sse-server, got %s", result.ServerInfo.Name)
	}
}
