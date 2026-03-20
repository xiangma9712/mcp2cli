package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

const protocolVersion = "2025-03-26"

const defaultTimeout = 30 * time.Second

// Client communicates with a remote MCP server using Streamable HTTP transport.
//
// The MCP Streamable HTTP transport uses a single HTTP endpoint for all communication.
// The client sends JSON-RPC 2.0 requests via POST and receives responses as either
// application/json or text/event-stream (SSE). A session ID returned by the server
// on initialize is included in subsequent requests via the Mcp-Session-Id header.
//
// Typical flow: Initialize → tools/list → tools/call.
type Client struct {
	endpoint  string
	sessionID string
	nextID    atomic.Int64
	http      *http.Client
}

// NewClient creates a new MCP client for the given endpoint URL.
// The URL must use http or https scheme.
func NewClient(endpoint string) (*Client, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s (only http and https are allowed)", u.Scheme)
	}
	return &Client{
		endpoint: endpoint,
		http:     &http.Client{Timeout: defaultTimeout},
	}, nil
}

// checkResponse validates that a JSON-RPC response is present and has no error.
func checkResponse(resp *Response, method string) error {
	if resp == nil {
		return fmt.Errorf("%s: no response from server", method)
	}
	if resp.Error != nil {
		return fmt.Errorf("%s error: %s", method, resp.Error.Message)
	}
	if resp.Result == nil {
		return fmt.Errorf("%s: empty result from server", method)
	}
	return nil
}

// SetHTTPClient replaces the default HTTP client.
func (c *Client) SetHTTPClient(hc *http.Client) {
	c.http = hc
}

func (c *Client) newID() int {
	return int(c.nextID.Add(1))
}

func (c *Client) post(ctx context.Context, req *Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if c.sessionID != "" {
		httpReq.Header.Set("Mcp-Session-Id", c.sessionID)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Limit response body to 10MB to prevent DoS from malicious servers.
	const maxResponseSize = 10 * 1024 * 1024
	limitedBody := io.LimitReader(resp.Body, maxResponseSize)

	if resp.StatusCode == http.StatusAccepted {
		return nil, nil // notification acknowledged
	}

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(limitedBody)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(errBody))
	}

	// Capture session ID from initialize response
	if sid := resp.Header.Get("Mcp-Session-Id"); sid != "" {
		c.sessionID = sid
	}

	ct := resp.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "text/event-stream") {
		return c.readSSEResponse(limitedBody, req.ID)
	}

	var rpcResp Response
	if err := json.NewDecoder(limitedBody).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &rpcResp, nil
}

func (c *Client) readSSEResponse(r io.Reader, requestID int) (*Response, error) {
	scanner := bufio.NewScanner(r)
	var lastResponse *Response
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var resp Response
		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			log.Printf("debug: malformed SSE data (skipping): %v", err)
			continue
		}
		if resp.ID == requestID {
			lastResponse = &resp
		}
	}
	if lastResponse == nil {
		return nil, fmt.Errorf("no response received for request %d", requestID)
	}
	return lastResponse, scanner.Err()
}

// Initialize performs the MCP initialize handshake.
func (c *Client) Initialize(ctx context.Context, clientName, clientVersion string) (*InitializeResult, error) {
	req := &Request{
		JSONRPC: "2.0",
		ID:      c.newID(),
		Method:  "initialize",
		Params: InitializeParams{
			ProtocolVersion: protocolVersion,
			Capabilities:    ClientCaps{},
			ClientInfo:      AppInfo{Name: clientName, Version: clientVersion},
		},
	}

	resp, err := c.post(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("initialize: %w", err)
	}
	if err := checkResponse(resp, "initialize"); err != nil {
		return nil, err
	}

	var result InitializeResult
	if err := json.Unmarshal(*resp.Result, &result); err != nil {
		return nil, fmt.Errorf("decode initialize result: %w", err)
	}

	// Send initialized notification
	notif := &Request{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}
	if _, err := c.post(ctx, notif); err != nil {
		return nil, fmt.Errorf("send initialized notification: %w", err)
	}

	return &result, nil
}

// ListTools retrieves all available tools from the server.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	var allTools []Tool
	var cursor *string

	for {
		params := map[string]any{}
		if cursor != nil {
			params["cursor"] = *cursor
		}

		req := &Request{
			JSONRPC: "2.0",
			ID:      c.newID(),
			Method:  "tools/list",
		}
		if len(params) > 0 {
			req.Params = params
		}

		resp, err := c.post(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("tools/list: %w", err)
		}
		if err := checkResponse(resp, "tools/list"); err != nil {
			return nil, err
		}

		var result ToolsListResult
		if err := json.Unmarshal(*resp.Result, &result); err != nil {
			return nil, fmt.Errorf("decode tools/list result: %w", err)
		}

		allTools = append(allTools, result.Tools...)

		if result.NextCursor == nil || *result.NextCursor == "" {
			break
		}
		cursor = result.NextCursor
	}

	return allTools, nil
}

// CallTool invokes a tool on the server.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]any) (*ToolCallResult, error) {
	req := &Request{
		JSONRPC: "2.0",
		ID:      c.newID(),
		Method:  "tools/call",
		Params: ToolCallParams{
			Name:      name,
			Arguments: arguments,
		},
	}

	resp, err := c.post(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("tools/call: %w", err)
	}
	if err := checkResponse(resp, "tools/call"); err != nil {
		return nil, err
	}

	var result ToolCallResult
	if err := json.Unmarshal(*resp.Result, &result); err != nil {
		return nil, fmt.Errorf("decode tools/call result: %w", err)
	}

	return &result, nil
}
