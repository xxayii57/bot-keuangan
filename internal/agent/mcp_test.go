package agent

import (
	"encoding/json"
	"testing"
)

func TestMCPJSONRPCMarshal(t *testing.T) {
	id := 1
	req := mcpRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  mcpMethodInitialize,
		Params: mcpInitializeParams{
			ProtocolVersion: "2024-11-05",
			Capabilities:    mcpClientCaps{Tools: &struct{}{}},
			ClientInfo:      mcpClientInfo{Name: "intimclaw", Version: "0.1.0"},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	// Verify it's valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("marshaled JSON is invalid: %v", err)
	}

	if parsed["jsonrpc"] != "2.0" {
		t.Errorf("jsonrpc = %v, want 2.0", parsed["jsonrpc"])
	}
	if parsed["method"] != "initialize" {
		t.Errorf("method = %v, want initialize", parsed["method"])
	}
}

func TestMCPJSONRPCResponse(t *testing.T) {
	respJSON := `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"test","version":"1.0"}}}`

	var resp mcpResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.ID == nil || *resp.ID != 1 {
		t.Errorf("id = %v, want 1", resp.ID)
	}
	if resp.Error != nil {
		t.Errorf("unexpected error: %v", resp.Error)
	}

	var initResult mcpInitializeResult
	if err := json.Unmarshal(resp.Result, &initResult); err != nil {
		t.Fatalf("failed to parse init result: %v", err)
	}
	if initResult.ServerInfo.Name != "test" {
		t.Errorf("server name = %v, want test", initResult.ServerInfo.Name)
	}
}

func TestMCPToolsListResponse(t *testing.T) {
	respJSON := `{"jsonrpc":"2.0","id":2,"result":{"tools":[{"name":"echo","description":"Echo back input"},{"name":"add","description":"Add two numbers","inputSchema":{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}}}}]}}`

	var resp mcpResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	var listResult mcpToolsListResult
	if err := json.Unmarshal(resp.Result, &listResult); err != nil {
		t.Fatalf("failed to parse tools/list result: %v", err)
	}

	if len(listResult.Tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(listResult.Tools))
	}
	if listResult.Tools[0].Name != "echo" {
		t.Errorf("tool[0].name = %v, want echo", listResult.Tools[0].Name)
	}
	if listResult.Tools[1].Name != "add" {
		t.Errorf("tool[1].name = %v, want add", listResult.Tools[1].Name)
	}
}

func TestMCPToolsCallResult(t *testing.T) {
	respJSON := `{"jsonrpc":"2.0","id":3,"result":{"content":[{"type":"text","text":"hello world"}],"isError":false}}`

	var resp mcpResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	var callResult mcpToolsCallResult
	if err := json.Unmarshal(resp.Result, &callResult); err != nil {
		t.Fatalf("failed to parse tools/call result: %v", err)
	}

	if callResult.IsError {
		t.Error("expected no error")
	}
	if len(callResult.Content) != 1 {
		t.Fatalf("got %d content items, want 1", len(callResult.Content))
	}
	if callResult.Content[0].Text != "hello world" {
		t.Errorf("text = %v, want hello world", callResult.Content[0].Text)
	}
}

func TestMCPToolsCallError(t *testing.T) {
	respJSON := `{"jsonrpc":"2.0","id":4,"result":{"content":[{"type":"text","text":"tool not found"}],"isError":true}}`

	var resp mcpResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	var callResult mcpToolsCallResult
	if err := json.Unmarshal(resp.Result, &callResult); err != nil {
		t.Fatalf("failed to parse tools/call result: %v", err)
	}

	if !callResult.IsError {
		t.Error("expected error flag to be true")
	}
}

func TestMCPNotifyNoID(t *testing.T) {
	// Notifications (like notifications/initialized) have no id.
	req := mcpRequest{
		JSONRPC: "2.0",
		Method:  mcpMethodInitialized,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal notification: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Notifications should not have "id" field.
	if _, exists := parsed["id"]; exists {
		t.Error("notification should not have id field")
	}
}

func TestMCPMalformedJSON(t *testing.T) {
	var resp mcpResponse
	err := json.Unmarshal([]byte(`not json at all`), &resp)
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestMCPErrorCode(t *testing.T) {
	respJSON := `{"jsonrpc":"2.0","id":5,"error":{"code":-32601,"message":"Method not found"}}`

	var resp mcpResponse
	if err := json.Unmarshal([]byte(respJSON), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}

	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != -32601 {
		t.Errorf("error code = %d, want -32601", resp.Error.Code)
	}
	if resp.Error.Message != "Method not found" {
		t.Errorf("error message = %v, want 'Method not found'", resp.Error.Message)
	}
}
