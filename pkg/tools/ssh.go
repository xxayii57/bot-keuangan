package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/xxayii57/bot-keuangan/pkg/config"
)

const sshToolName = "ssh_exec"

// SSHTool runs commands on whitelisted remote hosts over SSH. Password auth
// uses sshpass when available; hosts must be configured in tools.ssh.
type SSHTool struct {
	hosts    map[string]sshHostEntry
	defaultUser string
	port     int
	timeout  time.Duration
}

type sshHostEntry struct {
	user     string
	password string
	port     int
}

// NewSSHTool builds the tool from config values (already resolved by caller).
func NewSSHTool(hosts map[string]sshHostEntry, defaultUser string, port, timeoutSeconds int) *SSHTool {
	if port <= 0 {
		port = 22
	}
	if timeoutSeconds <= 0 {
		timeoutSeconds = 60
	}
	return &SSHTool{
		hosts:       hosts,
		defaultUser: defaultUser,
		port:        port,
		timeout:     time.Duration(timeoutSeconds) * time.Second,
	}
}

func (t *SSHTool) Name() string { return sshToolName }

func (t *SSHTool) Description() string {
	return "Run a shell command on a whitelisted remote server via SSH. " +
		"Only hosts listed in tools.ssh.allowed_hosts/credentials are reachable."
}

func (t *SSHTool) Parameters() map[string]any {
	hosts := make([]string, 0, len(t.hosts))
	for h := range t.hosts {
		hosts = append(hosts, h)
	}
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"host": map[string]any{
				"type":        "string",
				"description": "Hostname or IP of the whitelisted remote server. Known hosts: " + strings.Join(hosts, ", "),
			},
			"command": map[string]any{
				"type":        "string",
				"description": "Shell command to execute on the remote host",
			},
			"user": map[string]any{
				"type":        "string",
				"description": "SSH user; defaults to the per-host or global default",
			},
			"timeout_seconds": map[string]any{
				"type":        "integer",
				"description": "Optional timeout in seconds",
			},
		},
		"required": []string{"host", "command"},
	}
}

func (t *SSHTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	host, _ := args["host"].(string)
	command, _ := args["command"].(string)
	user, _ := args["user"].(string)

	host = strings.TrimSpace(host)
	command = strings.TrimSpace(command)
	if host == "" || command == "" {
		return ErrorResult("host and command are required")
	}

	entry, ok := t.hosts[host]
	if !ok {
		return ErrorResult(fmt.Sprintf("host %q is not in the ssh allowlist; known hosts: %s", host, strings.Join(mapKeys(t.hosts), ", ")))
	}
	if entry.user != "" && user == "" {
		user = entry.user
	}
	if user == "" {
		user = t.defaultUser
	}
	if user == "" {
		user = "root"
	}

	port := entry.port
	if port <= 0 {
		port = t.port
	}

	timeout := t.timeout
	if ts, ok := args["timeout_seconds"].(float64); ok && ts > 0 {
		timeout = time.Duration(ts) * time.Second
	}

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var cmd *exec.Cmd
	password := entry.password
	switch {
	case password != "":
		sshpassPath, err := exec.LookPath("sshpass")
		if err != nil {
			return ErrorResult("password auth requires sshpass installed on this machine")
		}
		cmd = exec.CommandContext(runCtx, sshpassPath, "-p", password,
			"ssh", "-o", "StrictHostKeyChecking=no",
			"-o", fmt.Sprintf("Port=%d", port),
			"-o", "ConnectTimeout=15",
			fmt.Sprintf("%s@%s", user, host),
			command)
	default:
		cmd = exec.CommandContext(runCtx, "ssh",
			"-o", "BatchMode=yes",
			"-o", "StrictHostKeyChecking=accept-new",
			"-o", fmt.Sprintf("Port=%d", port),
			"-o", "ConnectTimeout=15",
			fmt.Sprintf("%s@%s", user, host),
			command)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return ErrorResult(fmt.Sprintf("failed to start ssh: %v", err))
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	var waitErr error
	select {
	case waitErr = <-done:
	case <-runCtx.Done():
		_ = cmd.Process.Kill()
		<-done
		if ctx.Err() == nil {
			return ErrorResult(fmt.Sprintf("ssh command timed out after %s", timeout))
		}
		return ErrorResult("ssh command canceled")
	}

	out := strings.TrimSpace(stdout.String())
	errText := strings.TrimSpace(stderr.String())
	switch {
	case waitErr == nil && out == "" && errText == "":
		return NewToolResult("[No Output - command executed successfully]")
	case waitErr == nil:
		if errText != "" {
			return NewToolResult(out + "\n[STDERR]\n" + errText)
		}
		return NewToolResult(out)
	default:
		msg := fmt.Sprintf("ssh failed: %v", waitErr)
		if errText != "" {
			msg += "\n[STDERR]\n" + errText
		}
		if out != "" {
			msg += "\n[STDOUT]\n" + out
		}
		return ErrorResult(msg)
	}
}

func mapKeys(m map[string]sshHostEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// NewSSHToolFromConfig builds an SSHTool from tools.ssh configuration.
// Returns nil when no hosts are configured (tool stays unavailable).
func NewSSHToolFromConfig(cfg *config.Config) *SSHTool {
	if cfg == nil {
		return nil
	}
	sshCfg := cfg.Tools.SSH
	if len(sshCfg.AllowedHosts) == 0 && len(sshCfg.Credentials) == 0 {
		return nil
	}

	hosts := make(map[string]sshHostEntry)
	for _, cred := range sshCfg.Credentials {
		if cred.Host == "" {
			continue
		}
		hosts[cred.Host] = sshHostEntry{
			user:     cred.User,
			password: cred.Password.String(),
			port:     cred.Port,
		}
	}
	for _, h := range sshCfg.AllowedHosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if _, exists := hosts[h]; !exists {
			hosts[h] = sshHostEntry{}
		}
	}

	return NewSSHTool(hosts, sshCfg.DefaultUser, sshCfg.Port, sshCfg.TimeoutSeconds)
}
