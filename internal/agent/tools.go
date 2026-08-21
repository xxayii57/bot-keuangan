package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// ToolInfo describes a registered tool.
type ToolInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// toolFunc is the signature every tool implements.
type toolFunc func(args map[string]interface{}) (string, error)

// ToolRegistry is a name→handler map with thread-safe access.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]toolFunc
	meta  map[string]string // name → description
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]toolFunc),
		meta:  make(map[string]string),
	}
}

func (r *ToolRegistry) Register(name, description string, fn toolFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[name] = fn
	r.meta[name] = description
}

func (r *ToolRegistry) Execute(name string, args map[string]interface{}) (string, error) {
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

	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	var out []byte
	var err error
	if payload != "" {
		cmd := fmt.Sprintf(`curl -s -X POST -H "Content-Type: application/json" -d '%s' "%s"`, payload, url)
		c := exec.Command("sh", "-c", cmd)
		out, err = c.CombinedOutput()
	} else {
		cmd := fmt.Sprintf(`curl -s "%s"`, url)
		c := exec.Command("sh", "-c", cmd)
		out, err = c.CombinedOutput()
	}
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
	cmd := fmt.Sprintf(`curl -s -X POST "https://api.telegram.org/bot%s/sendDocument" -F "chat_id=%s" -F "document=@%s"`, token, chatID, filePath)
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
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
	a.mcp.mu.RLock()
	srv, ok := a.mcp.servers[server]
	a.mcp.mu.RUnlock()
	if !ok {
		return fmt.Sprintf("MCP server '%s' not connected.", server), nil
	}
	if srv.isClosed {
		return fmt.Sprintf("MCP server '%s' is closed.", server), nil
	}
	return fmt.Sprintf("MCP call %s/%s: not yet implemented (stdio transport pending).", server, tool), nil
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
