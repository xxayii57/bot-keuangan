package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// MCP JSON-RPC 2.0 types.

type mcpRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      *int          `json:"id,omitempty"` // nil for notifications
	Method  string        `json:"method"`
	Params  interface{}   `json:"params,omitempty"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      *int            `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCP protocol method names.
const (
	mcpMethodInitialize      = "initialize"
	mcpMethodInitialized     = "notifications/initialized"
	mcpMethodToolsList       = "tools/list"
	mcpMethodToolsCall       = "tools/call"
)

// Initialize params/response.
type mcpInitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    mcpClientCaps  `json:"capabilities"`
	ClientInfo      mcpClientInfo  `json:"clientInfo"`
}

type mcpClientCaps struct {
	Tools *struct{} `json:"tools,omitempty"`
}

type mcpClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type mcpInitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    mcpServerCaps  `json:"capabilities"`
	ServerInfo      mcpServerInfo  `json:"serverInfo"`
}

type mcpServerCaps struct {
	Tools *mcpToolsCap `json:"tools,omitempty"`
}

type mcpToolsCap struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type mcpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// tools/list response.
type mcpToolsListResult struct {
	Tools []mcpToolDef `json:"tools"`
}

type mcpToolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema,omitempty"`
}

// tools/call params/response.
type mcpToolsCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type mcpToolsCallResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// MCPServer represents a connected MCP server process.
type MCPServer struct {
	Name     string
	Command  string
	Args     []string
	Tools    []MCPTool
	cmd      *exec.Cmd
	isClosed bool

	mu       sync.Mutex
	nextID   int
	stdin    io.WriteCloser
	pending  map[int]chan *mcpResponse // id → response channel
}

// MCPTool is a tool advertised by an MCP server.
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ServerName  string `json:"server_name"` // which server owns this tool
}

// MCPServerManager manages multiple MCP server connections.
type MCPServerManager struct {
	mu      sync.RWMutex
	servers map[string]*MCPServer
}

func NewMCPServerManager() *MCPServerManager {
	return &MCPServerManager{
		servers: make(map[string]*MCPServer),
	}
}

// Connect starts an MCP server process, performs the initialize handshake,
// discovers tools via tools/list, and begins reading responses.
func (m *MCPServerManager) Connect(name, command string, args []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if args == nil {
		args = []string{}
	}

	cmd := exec.Command(command, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("[mcp] failed to get stdout pipe for '%s': %v\n", name, err)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf("[mcp] failed to get stderr pipe for '%s': %v\n", name, err)
		return
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Printf("[mcp] failed to get stdin pipe for '%s': %v\n", name, err)
		return
	}

	srv := &MCPServer{
		Name:    name,
		Command: command,
		Args:    args,
		cmd:     cmd,
		nextID:  1,
		stdin:   stdin,
		pending: make(map[int]chan *mcpResponse),
	}

	m.servers[name] = srv

	if err := cmd.Start(); err != nil {
		fmt.Printf("[mcp] failed to start server '%s': %v\n", name, err)
		srv.isClosed = true
		return
	}

	// Start response reader goroutine.
	go srv.readLoop(stdout)

	// Drain stderr for logging.
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				fmt.Printf("[mcp:%s] stderr: %s\n", name, line)
			}
		}
	}()

	// Wait for process exit in background.
	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("[mcp] server '%s' exited: %v\n", name, err)
		}
		srv.isClosed = true
	}()

	fmt.Printf("[mcp] connected to '%s' (%s %v)\n", name, command, args)

	// Perform initialize handshake.
	if err := srv.initialize(); err != nil {
		fmt.Printf("[mcp] initialize failed for '%s': %v\n", name, err)
		srv.isClosed = true
		return
	}

	// Discover tools.
	if err := srv.discoverTools(); err != nil {
		fmt.Printf("[mcp] tools/list failed for '%s': %v\n", name, err)
	}
}

// initialize performs the MCP initialize handshake.
func (s *MCPServer) initialize() error {
	params := mcpInitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    mcpClientCaps{Tools: &struct{}{}},
		ClientInfo:      mcpClientInfo{Name: "intimclaw", Version: "0.1.0"},
	}

	result, err := s.rpcCall(mcpMethodInitialize, params)
	if err != nil {
		return fmt.Errorf("initialize call failed: %w", err)
	}

	var initResult mcpInitializeResult
	if err := json.Unmarshal(result, &initResult); err != nil {
		return fmt.Errorf("failed to parse initialize result: %w", err)
	}

	// Send initialized notification (no id → no response expected).
	s.rpcNotify(mcpMethodInitialized, nil)

	return nil
}

// discoverTools fetches the list of tools from the server.
func (s *MCPServer) discoverTools() error {
	result, err := s.rpcCall(mcpMethodToolsList, nil)
	if err != nil {
		return fmt.Errorf("tools/list call failed: %w", err)
	}

	var listResult mcpToolsListResult
	if err := json.Unmarshal(result, &listResult); err != nil {
		return fmt.Errorf("failed to parse tools/list result: %w", err)
	}

	s.Tools = nil
	for _, t := range listResult.Tools {
		s.Tools = append(s.Tools, MCPTool{
			Name:        t.Name,
			Description: t.Description,
			ServerName:  s.Name,
		})
	}

	fmt.Printf("[mcp:%s] discovered %d tools\n", s.Name, len(s.Tools))
	return nil
}

// CallTool invokes a tool on this MCP server via tools/call.
func (s *MCPServer) CallTool(toolName string, arguments map[string]interface{}) (string, error) {
	if s.isClosed {
		return "", fmt.Errorf("MCP server '%s' is closed", s.Name)
	}

	params := mcpToolsCallParams{
		Name:      toolName,
		Arguments: arguments,
	}

	result, err := s.rpcCall(mcpMethodToolsCall, params)
	if err != nil {
		return "", fmt.Errorf("tools/call failed: %w", err)
	}

	var callResult mcpToolsCallResult
	if err := json.Unmarshal(result, &callResult); err != nil {
		return "", fmt.Errorf("failed to parse tools/call result: %w", err)
	}

	if callResult.IsError {
		var texts []string
		for _, c := range callResult.Content {
			if c.Type == "text" && c.Text != "" {
				texts = append(texts, c.Text)
			}
		}
		return "", fmt.Errorf("MCP tool error: %s", strings.Join(texts, "; "))
	}

	// Extract text content.
	var texts []string
	for _, c := range callResult.Content {
		if c.Type == "text" && c.Text != "" {
			texts = append(texts, c.Text)
		}
	}
	return strings.Join(texts, "\n"), nil
}

// rpcCall sends a JSON-RPC request and waits for the matching response.
func (s *MCPServer) rpcCall(method string, params interface{}) (json.RawMessage, error) {
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.mu.Unlock()

	req := mcpRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	data = append(data, '\n')

	ch := make(chan *mcpResponse, 1)
	s.mu.Lock()
	s.pending[id] = ch
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.pending, id)
		s.mu.Unlock()
	}()

	if _, err := s.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("write to stdin: %w", err)
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// rpcNotify sends a JSON-RPC notification (no id, no response expected).
func (s *MCPServer) rpcNotify(method string, params interface{}) {
	req := mcpRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return
	}
	data = append(data, '\n')
	s.stdin.Write(data)
}

// readLoop reads newline-delimited JSON-RPC responses from stdout.
func (s *MCPServer) readLoop(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	// Increase buffer size for large tool results.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var resp mcpResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			fmt.Printf("[mcp:%s] malformed response: %v\n", s.Name, err)
			continue
		}

		if resp.ID == nil {
			// Server-initiated notification — ignore for now.
			continue
		}

		s.mu.Lock()
		ch, ok := s.pending[*resp.ID]
		s.mu.Unlock()

		if ok {
			ch <- &resp
		} else {
			fmt.Printf("[mcp:%s] unexpected response for id %d\n", s.Name, *resp.ID)
		}
	}
}

// AllTools returns all tools from all connected servers.
func (m *MCPServerManager) AllTools() []MCPTool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []MCPTool
	for _, srv := range m.servers {
		if srv.isClosed {
			continue
		}
		all = append(all, srv.Tools...)
	}
	return all
}

// GetServer returns a specific server by name.
func (m *MCPServerManager) GetServer(name string) (*MCPServer, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	srv, ok := m.servers[name]
	return srv, ok
}
