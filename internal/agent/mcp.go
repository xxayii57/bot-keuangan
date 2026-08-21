package agent

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// MCPServer represents a connected MCP server process.
type MCPServer struct {
	Name     string
	Command  string
	Args     []string
	Tools    []MCPTool
	cmd      *exec.Cmd
	isClosed bool
}

// MCPTool is a tool advertised by an MCP server.
type MCPTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
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

// Connect starts an MCP server process and begins reading its stdio JSON-RPC output.
func (m *MCPServerManager) Connect(name, command string, args []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if args == nil {
		args = []string{}
	}

	cmd := exec.Command(command, args...)
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	srv := &MCPServer{
		Name:    name,
		Command: command,
		Args:    args,
		cmd:     cmd,
	}

	m.servers[name] = srv

	if err := cmd.Start(); err != nil {
		fmt.Printf("[mcp] failed to start server '%s': %v\n", name, err)
		srv.isClosed = true
		return
	}

	// Read stdout for JSON-RPC responses (tools/list results etc.)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			// TODO: parse JSON-RPC responses, extract tool listings
			_ = line
		}
		srv.isClosed = true
	}()

	// Drain stderr
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			// discard for now
		}
	}()

	go func() {
		if err := cmd.Wait(); err != nil {
			fmt.Printf("[mcp] server '%s' exited: %v\n", name, err)
		}
		srv.isClosed = true
	}()

	fmt.Printf("[mcp] connected to '%s' (%s %v)\n", name, command, args)
}
