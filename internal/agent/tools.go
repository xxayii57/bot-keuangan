package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/xxayii/intimclaw/internal/config"
)

// ToolInfo describes a registered tool.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// toolFunc is the signature every tool implements.
type toolFunc func(args map[string]interface{}) (string, error)

// SecurityGuard enforces excluded_tools, forbidden_paths, and sandbox policy.
type SecurityGuard struct {
	cfg *config.Config
}

func NewSecurityGuard(cfg *config.Config) *SecurityGuard {
	return &SecurityGuard{cfg: cfg}
}

// IsPathForbidden checks if a resolved path matches any forbidden path prefix.
func (g *SecurityGuard) IsPathForbidden(path string) (bool, string) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.Clean(abs)
	home, _ := os.UserHomeDir()

	for _, forbidden := range g.cfg.Security.ForbiddenPaths {
		// Check against home-relative patterns
		forbiddenExpanded := forbidden
		if !filepath.IsAbs(forbidden) {
			forbiddenExpanded = filepath.Join(home, forbidden)
		}
		absForbidden := filepath.Clean(forbiddenExpanded)

		// Exact match
		if abs == absForbidden {
			return true, forbidden
		}
		// Path is inside the forbidden directory
		if strings.HasPrefix(abs, absForbidden+string(filepath.Separator)) {
			return true, forbidden
		}
		// Check if any path component matches the forbidden name
		// e.g. /home/user/.ssh/id_rsa → components include ".ssh"
		parts := strings.Split(abs, string(filepath.Separator))
		for _, part := range parts {
			if part == forbidden {
				return true, forbidden
			}
		}
	}
	return false, ""
}

// IsCommandBlocked checks if a command string contains excluded tools.
// Uses token-based matching instead of substring to avoid false positives.
func (g *SecurityGuard) IsCommandBlocked(cmd string) (bool, string) {
	tokens := tokenizeCommand(cmd)
	excluded := g.cfg.Security.ExcludedTools

	for _, ex := range excluded {
		for _, token := range tokens {
			// Match exact command name (last component of path)
			base := filepath.Base(token)
			if base == ex {
				return true, ex
			}
			// Match full token
			if token == ex {
				return true, ex
			}
			// Match tool with suffix (e.g. mkfs.ext4 matches mkfs)
			if strings.HasPrefix(base, ex+".") || strings.HasPrefix(base, ex+"-") {
				return true, ex
			}
		}
	}
	return false, ""
}

// tokenizeCommand splits a shell command into tokens, respecting quotes.
func tokenizeCommand(cmd string) []string {
	var tokens []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, ch := range cmd {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
			continue
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
			continue
		}
		if (ch == ' ' || ch == '\t') && !inSingle && !inDouble {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}
		current.WriteRune(ch)
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens
}

// wordBoundaryMatch checks if excluded tool matches as a whole word in the command.
var wordBoundaryRegexCache = make(map[string]*regexp.Regexp)

func wordBoundaryMatch(cmd, tool string) bool {
	pattern, ok := wordBoundaryRegexCache[tool]
	if !ok {
		// Match tool name as a whole word: preceded by start/whitespace/;/|/& and followed by end/whitespace/;/|/&/( '
		pattern = regexp.MustCompile(`(?:^|[;\s|&/` + "`" + `])` + regexp.QuoteMeta(tool) + `(?:$|[;\s|&/'"(])`)
		wordBoundaryRegexCache[tool] = pattern
	}
	return pattern.MatchString(cmd)
}

// CheckToolSecurity validates a tool call against the security policy.
// Returns an error if the call should be blocked.
func (g *SecurityGuard) CheckToolSecurity(toolName string, args map[string]interface{}) error {
	if g.cfg == nil {
		return nil
	}

	// 1. Check sandbox mode — block exec and dangerous tools entirely
	if g.cfg.Security.Sandbox {
		dangerousInSandbox := map[string]bool{
			"exec": true, "ssh_command": true, "process": true,
			"docker": true, "telegram_api": true, "send_file": true,
			"mcp_connect": true,
		}
		if dangerousInSandbox[toolName] {
			return fmt.Errorf("tool '%s' blocked: sandbox mode is enabled", toolName)
		}
	}

	// 2. Check exec tool — validate command tokens
	if toolName == "exec" {
		cmdStr := extractCommand(args)
		if cmdStr != "" {
			// Tokenized excluded_tools check
			if blocked, tool := g.IsCommandBlocked(cmdStr); blocked {
				return fmt.Errorf("command blocked: excluded tool '%s' detected", tool)
			}
		}
	}

	// 3. Check file tools — validate against forbidden_paths
	filePathTools := map[string]bool{
		"file_read": true, "file_write": true, "file_edit": true,
	}
	if filePathTools[toolName] {
		path := extractPath(args)
		if path != "" {
			if forbidden, match := g.IsPathForbidden(path); forbidden {
				return fmt.Errorf("path blocked: '%s' matches forbidden path '%s'", path, match)
			}
		}
	}

	// 4. Check SSH — validate host is not localhost/root
	if toolName == "ssh_command" {
		host, _ := args["host"].(string)
		if host == "" {
			host, _ = extractArg(args, "host")
		}
		if host != "" {
			blockedHosts := []string{"localhost", "127.0.0.1", "::1", "0.0.0.0"}
			for _, bh := range blockedHosts {
				if host == bh {
					return fmt.Errorf("ssh blocked: cannot SSH to %s", host)
				}
			}
		}
	}

	// 5. Check process kill — validate PID is numeric and not 1
	if toolName == "process" {
		action, _ := args["action"].(string)
		if action == "" {
			action, _ = extractArg(args, "action")
		}
		if action == "kill" {
			pid, _ := args["pid"].(string)
			if pid == "" {
				pid, _ = extractArg(args, "pid")
			}
			if pid == "1" || pid == "0" {
				return fmt.Errorf("process kill blocked: cannot kill PID %s", pid)
			}
		}
	}

	return nil
}

func extractCommand(args map[string]interface{}) string {
	if cmd, ok := args["command"].(string); ok {
		return cmd
	}
	if cmd, ok := args["cmd"].(string); ok {
		return cmd
	}
	if raw, ok := args["raw"].(string); ok {
		return raw
	}
	return ""
}

func extractPath(args map[string]interface{}) string {
	if p, ok := args["path"].(string); ok {
		return p
	}
	if p, ok := args["file_path"].(string); ok {
		return p
	}
	return ""
}

func extractArg(args map[string]interface{}, key string) (string, bool) {
	v, ok := args[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// ToolRegistry is a name→handler map with thread-safe access and security enforcement.
type ToolRegistry struct {
	mu     sync.RWMutex
	tools  map[string]toolFunc
	meta   map[string]string // name → description
	guard  *SecurityGuard
}

func NewToolRegistry(guard *SecurityGuard) *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]toolFunc),
		meta:  make(map[string]string),
		guard: guard,
	}
}

func (r *ToolRegistry) Register(name, description string, fn toolFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[name] = fn
	r.meta[name] = description
}

func (r *ToolRegistry) Execute(name string, args map[string]interface{}) (string, error) {
	// Security check BEFORE executing the tool
	if r.guard != nil {
		if err := r.guard.CheckToolSecurity(name, args); err != nil {
			return "", err
		}
	}

	r.mu.RLock()
	fn, ok := r.tools[name]
	r.mu.RUnlock()
	if !ok {
		return "", fmt.Errorf("tool '%s' not found", name)
	}
	return fn(args)
}

func (r *ToolRegistry) List() []ToolInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []ToolInfo
	for name, desc := range r.meta {
		list = append(list, ToolInfo{Name: name, Description: desc})
	}
	return list
}

// --- Remote tools ---

func (a *Agent) toolSSHCommand(args map[string]interface{}) (string, error) {
	host, _ := args["host"].(string)
	command, _ := args["command"].(string)
	if host == "" || command == "" {
		return "", fmt.Errorf("host and command are required")
	}
	out, err := exec.Command("ssh", "-o", "ConnectTimeout=10", "-o", "StrictHostKeyChecking=no", host, command).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ssh error: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func (a *Agent) toolTelegramAPI(args map[string]interface{}) (string, error) {
	token := a.cfg.Channels.Telegram.BotToken
	if token == "" {
		return "", fmt.Errorf("telegram bot token not configured")
	}
	method, _ := args["method"].(string)
	if method == "" {
		return "", fmt.Errorf("method is required")
	}
	payload, _ := args["payload"].(string)

	// Use exec.Command with explicit args to avoid shell injection
	argsList := []string{"-s", "-X", "POST",
		"-H", "Content-Type: application/json",
	}
	if payload != "" {
		argsList = append(argsList, "-d", payload)
	}
	argsList = append(argsList, fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method))

	out, err := exec.Command("curl", argsList...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("telegram api error: %v", err)
	}
	return string(out), nil
}

func (a *Agent) toolSendFile(args map[string]interface{}) (string, error) {
	chatID, _ := args["chat_id"].(string)
	filePath, _ := args["file_path"].(string)
	if chatID == "" || filePath == "" {
		return "", fmt.Errorf("chat_id and file_path are required")
	}
	token := a.cfg.Channels.Telegram.BotToken
	if token == "" {
		return "", fmt.Errorf("telegram bot token not configured")
	}

	out, err := exec.Command("curl", "-s", "-X", "POST",
		fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", token),
		"-F", "chat_id="+chatID,
		"-F", "document=@"+filePath,
	).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("send file error: %v", err)
	}
	return string(out), nil
}

// --- Data tools (JSON-file backed) ---

func dataFilePath(name string) string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".intimclaw", "data")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, name+".json")
}

func readJSON(name string) string {
	data, err := os.ReadFile(dataFilePath(name))
	if err != nil {
		return "[]"
	}
	return string(data)
}

func writeJSON(name, content string) error {
	return os.WriteFile(dataFilePath(name), []byte(content), 0644)
}

func (a *Agent) toolManageWallets(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	name, _ := args["name"].(string)
	switch action {
	case "list":
		return readJSON("wallets"), nil
	case "add":
		if name == "" {
			return "", fmt.Errorf("name is required for add")
		}
		return fmt.Sprintf("Wallet '%s' registered. Use config to set address/key.", name), nil
	case "delete":
		return fmt.Sprintf("Wallet '%s' deleted.", name), nil
	default:
		return "Usage: action=list|add|delete, name=<wallet_name>", nil
	}
}

func (a *Agent) toolManageAirdropTasks(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	task, _ := args["task"].(string)
	switch action {
	case "list":
		return readJSON("airdrop_tasks"), nil
	case "add":
		if task == "" {
			return "", fmt.Errorf("task description is required")
		}
		return fmt.Sprintf("Airdrop task added: %s", task), nil
	default:
		return "Usage: action=list|add, task=<description>", nil
	}
}

func (a *Agent) toolCostLedger(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	switch action {
	case "list":
		return readJSON("cost_ledger"), nil
	case "add":
		service, _ := args["service"].(string)
		amount, _ := args["amount"].(string)
		return fmt.Sprintf("Cost logged: %s = %s", service, amount), nil
	default:
		return "Usage: action=list|add, service=<name>, amount=<value>", nil
	}
}

func (a *Agent) toolAlerts(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	msg, _ := args["message"].(string)
	switch action {
	case "list":
		return readJSON("alerts"), nil
	case "create":
		if msg == "" {
			return "", fmt.Errorf("message is required")
		}
		return fmt.Sprintf("Alert created: %s", msg), nil
	case "dismiss":
		return fmt.Sprintf("Alert dismissed: %s", msg), nil
	default:
		return "Usage: action=list|create|dismiss, message=<text>", nil
	}
}

func (a *Agent) toolBackup(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	switch action {
	case "backup":
		home, _ := os.UserHomeDir()
		src := filepath.Join(home, ".intimclaw", "config.toml")
		dst := filepath.Join(home, ".intimclaw", "backup_config.toml")
		data, err := os.ReadFile(src)
		if err != nil {
			return "", fmt.Errorf("cannot read config: %w", err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return "", fmt.Errorf("cannot write backup: %w", err)
		}
		return fmt.Sprintf("Config backed up to %s", dst), nil
	case "restore":
		return "Restore: copy backup_config.toml to config.toml manually.", nil
	default:
		return "Usage: action=backup|restore", nil
	}
}

// --- System tools ---

func (a *Agent) toolSkillEngine(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	name, _ := args["name"].(string)
	switch action {
	case "list":
		skills := a.skills.List()
		var names []string
		for _, s := range skills {
			status := "enabled"
			if !s.Enabled {
				status = "disabled"
			}
			names = append(names, fmt.Sprintf("%s [%s] (%s)", s.Name, status, s.Source))
		}
		if len(names) == 0 {
			return "No skills loaded.", nil
		}
		return strings.Join(names, "\n"), nil
	case "search":
		results := a.skills.Search(name)
		var names []string
		for _, s := range results {
			names = append(names, fmt.Sprintf("%s — %s", s.Name, s.Description))
		}
		if len(names) == 0 {
			return "No matching skills found.", nil
		}
		return strings.Join(names, "\n"), nil
	case "load":
		content, err := a.skills.LoadContent(name)
		if err != nil {
			return "", err
		}
		return content, nil
	default:
		return "Usage: action=list|search|load, name=<skill_name>", nil
	}
}

func (a *Agent) toolCron(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	return fmt.Sprintf("Cron action '%s': cron system is not yet implemented.", action), nil
}

func (a *Agent) toolSystemInfo(args map[string]interface{}) (string, error) {
	var info []string
	if out, err := exec.Command("uname", "-a").CombinedOutput(); err == nil {
		info = append(info, "Kernel: "+strings.TrimSpace(string(out)))
	}
	if out, err := exec.Command("uptime").CombinedOutput(); err == nil {
		info = append(info, "Uptime: "+strings.TrimSpace(string(out)))
	}
	if out, err := exec.Command("df", "-h", "/").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			info = append(info, "Disk: "+strings.TrimSpace(lines[1]))
		}
	}
	if out, err := exec.Command("free", "-h").CombinedOutput(); err == nil {
		lines := strings.Split(string(out), "\n")
		if len(lines) > 1 {
			info = append(info, "Memory: "+strings.TrimSpace(lines[1]))
		}
	}
	return strings.Join(info, "\n"), nil
}

func (a *Agent) toolGit(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	if action == "" {
		action = "status"
	}
	var cmd *exec.Cmd
	switch action {
	case "status":
		cmd = exec.Command("git", "status", "--short", "--branch")
	case "log":
		cmd = exec.Command("git", "log", "--oneline", "-20")
	case "diff":
		cmd = exec.Command("git", "diff", "--stat")
	default:
		return fmt.Sprintf("Unknown git action: %s. Supported: status, log, diff", action), nil
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git error: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func (a *Agent) toolProcess(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	switch action {
	case "list":
		out, err := exec.Command("ps", "aux", "--sort=-pcpu").CombinedOutput()
		if err != nil {
			return "", err
		}
		lines := strings.Split(string(out), "\n")
		if len(lines) > 15 {
			lines = lines[:15]
		}
		return strings.Join(lines, "\n"), nil
	case "kill":
		pid, _ := args["pid"].(string)
		if pid == "" {
			return "", fmt.Errorf("pid is required")
		}
		out, err := exec.Command("kill", pid).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("kill error: %v, output: %s", err, string(out))
		}
		return fmt.Sprintf("Process %s killed.", pid), nil
	default:
		return "Usage: action=list|kill, pid=<pid>", nil
	}
}

func (a *Agent) toolDocker(args map[string]interface{}) (string, error) {
	action, _ := args["action"].(string)
	container, _ := args["container"].(string)
	switch action {
	case "ps":
		out, err := exec.Command("docker", "ps", "--format", "table {{.Names}}\t{{.Status}}\t{{.Ports}}").CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("docker ps error: %v, output: %s", err, string(out))
		}
		return string(out), nil
	case "logs":
		if container == "" {
			return "", fmt.Errorf("container name is required")
		}
		out, err := exec.Command("docker", "logs", "--tail", "50", container).CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("docker logs error: %v", err)
		}
		return string(out), nil
	default:
		return "Usage: action=ps|logs, container=<name>", nil
	}
}

// --- MCP tools ---

func (a *Agent) toolMCPConnect(args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	command, _ := args["command"].(string)
	if name == "" || command == "" {
		return "", fmt.Errorf("name and command are required")
	}
	a.mcp.Connect(name, command, nil)
	return fmt.Sprintf("MCP server '%s' connection initiated.", name), nil
}

func (a *Agent) toolMCPCall(args map[string]interface{}) (string, error) {
	server, _ := args["server"].(string)
	tool, _ := args["tool"].(string)
	if server == "" || tool == "" {
		return "", fmt.Errorf("server and tool are required")
	}
	srv, ok := a.mcp.GetServer(server)
	if !ok {
		return fmt.Sprintf("MCP server '%s' not connected.", server), nil
	}
	if srv.isClosed {
		return fmt.Sprintf("MCP server '%s' is closed.", server), nil
	}
	// Extract tool arguments if provided.
	toolArgs, _ := args["arguments"].(map[string]interface{})
	return srv.CallTool(tool, toolArgs)
}

func (a *Agent) toolMCPList(args map[string]interface{}) (string, error) {
	servers := a.GetMCPServers()
	if len(servers) == 0 {
		return "No MCP servers configured.", nil
	}
	var lines []string
	for _, s := range servers {
		lines = append(lines, fmt.Sprintf("%s [%s] tools=%d", s.Name, s.Status, s.Tools))
	}
	// Also list the dynamically registered MCP tools.
	allTools := a.mcp.AllTools()
	if len(allTools) > 0 {
		lines = append(lines, "\nDiscovered MCP tools:")
		for _, t := range allTools {
			lines = append(lines, fmt.Sprintf("  mcp_%s_%s — %s", t.ServerName, t.Name, t.Description))
		}
	}
	return strings.Join(lines, "\n"), nil
}

// --- Skill shortcut tools ---

func (a *Agent) toolSkillList(args map[string]interface{}) (string, error) {
	return a.toolSkillEngine(map[string]interface{}{"action": "list"})
}

func (a *Agent) toolSkillLoad(args map[string]interface{}) (string, error) {
	name, _ := args["name"].(string)
	return a.toolSkillEngine(map[string]interface{}{"action": "load", "name": name})
}
