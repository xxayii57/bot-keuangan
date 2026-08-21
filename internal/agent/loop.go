// intimclaw.ic — Built by xxayii — IntimClaw Agent Engine v0.1.0
package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/xxayii/intimclaw/internal/config"
)

type Agent struct {
	cfg     *config.Config
	tools   *ToolRegistry
	history []Message
	mcp     *MCPServerManager
	skills  *SkillsLoader
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AgentEvent struct {
	Type    string `json:"type"`    // "thought", "action", "observation", "message", "error"
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Role             string `json:"role"`
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []AnthropicMessage `json:"messages"`
}

type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func New(cfg *config.Config) *Agent {
	a := &Agent{
		cfg:    cfg,
		tools:  NewToolRegistry(),
		mcp:    NewMCPServerManager(),
		skills: NewSkillsLoader(cfg.Skills.Directories),
	}
	a.registerBuiltinTools()
	a.skills.Load()

	if cfg.MCP.Enabled {
		for _, s := range cfg.MCP.Servers {
			fmt.Printf("[intimclaw] Connecting to MCP server: %s (%s %v)...\n", s.Name, s.Command, s.Args)
			a.mcp.Connect(s.Name, s.Command, s.Args)
		}
	}
	return a
}

func NewFromConfig(cfg *config.Config) *Agent {
	return New(cfg)
}

func (a *Agent) registerBuiltinTools() {
	// Core tools
	a.tools.Register("exec", "Execute a shell command", a.toolExec)
	a.tools.Register("http_request", "Make HTTP request", a.toolHTTPRequest)
	a.tools.Register("file_read", "Read file contents", a.toolFileRead)
	a.tools.Register("file_write", "Write content to a file", a.toolFileWrite)
	a.tools.Register("file_edit", "Edit a file by replacing text", a.toolFileEdit)
	a.tools.Register("list_dir", "List directory contents", a.toolListDir)
	a.tools.Register("web_fetch", "Fetch a URL and return content", a.toolWebFetch)
	a.tools.Register("memory_save", "Save a memory", a.toolMemorySave)
	a.tools.Register("memory_search", "Search memories", a.toolMemorySearch)
	// Remote tools
	a.tools.Register("ssh_command", "Execute command on remote host via SSH", a.toolSSHCommand)
	a.tools.Register("telegram_api", "Call Telegram Bot API directly", a.toolTelegramAPI)
	a.tools.Register("send_file", "Send file via Telegram", a.toolSendFile)
	// Data tools
	a.tools.Register("manage_wallets", "Manage crypto wallets (add/list/delete)", a.toolManageWallets)
	a.tools.Register("manage_airdrop_tasks", "Manage airdrop task tracking", a.toolManageAirdropTasks)
	a.tools.Register("cost_ledger", "Track API/service costs", a.toolCostLedger)
	a.tools.Register("alerts", "Create/list/dismiss alerts", a.toolAlerts)
	// System tools
	a.tools.Register("backup", "Backup/restore config files", a.toolBackup)
	a.tools.Register("skill_engine", "List/search/load skills", a.toolSkillEngine)
	a.tools.Register("cron", "Manage scheduled tasks", a.toolCron)
	a.tools.Register("system_info", "Get system status/processes/network", a.toolSystemInfo)
	a.tools.Register("git", "Git operations (status/log/diff)", a.toolGit)
	a.tools.Register("process", "List/kill/monitor processes", a.toolProcess)
	a.tools.Register("docker", "Docker operations (ps/logs)", a.toolDocker)
	// MCP tools
	a.tools.Register("mcp_connect", "Connect to MCP server", a.toolMCPConnect)
	a.tools.Register("mcp_call", "Call MCP tool", a.toolMCPCall)
	a.tools.Register("mcp_list", "List MCP tools", a.toolMCPList)
	// Skills tools
	a.tools.Register("skill_list", "List available skills", a.toolSkillList)
	a.tools.Register("skill_load", "Load skill content", a.toolSkillLoad)
}

func (a *Agent) SetModel(model string) {
	if model == "" {
		return
	}
	a.cfg.Agent.Model = model
	for _, p := range a.cfg.Providers {
		for _, m := range p.Models {
			if m == model {
				a.cfg.Agent.Provider = p.Name
				return
			}
		}
	}
}

func (a *Agent) callLLM(currentModel string, messages []Message, systemPrompt string, provider config.ProviderConfig) (string, error) {
	fallbackModels := []string{
		"jr/f/deepseek-v4-flash-free",
		"jr/f/mimo-v2.5-free",
		"jr/f/nemotron-3-ultra-free",
	}

	var resp *http.Response
	var err error
	var client = &http.Client{Timeout: 60 * time.Second}
	var req *http.Request

	for try := 0; try < len(fallbackModels)+1; try++ {
		var reqJSON []byte
		var url string
		isAnthropic := provider.Type == "anthropic" || strings.Contains(strings.ToLower(provider.Name), "anthropic")

		if isAnthropic {
			var antMessages []AnthropicMessage
			for _, m := range messages {
				if m.Role == "system" {
					continue
				}
				antMessages = append(antMessages, AnthropicMessage{
					Role:    m.Role,
					Content: m.Content,
				})
			}
			url = provider.BaseURL
			if url == "" {
				url = "https://api.anthropic.com"
			}
			url = strings.TrimSuffix(url, "/") + "/v1/messages"

			antReq := AnthropicRequest{
				Model:     currentModel,
				MaxTokens: 4000,
				System:    systemPrompt,
				Messages:  antMessages,
			}
			reqJSON, _ = json.Marshal(antReq)
		} else {
			url = provider.BaseURL
			if url == "" {
				if provider.Type == "groq" || strings.Contains(strings.ToLower(provider.Name), "groq") {
					url = "https://api.groq.com/openai"
				} else if provider.Type == "ollama" || strings.Contains(strings.ToLower(provider.Name), "ollama") {
					url = "http://localhost:11434"
				}
			}
			url = strings.TrimSuffix(url, "/") + "/v1" // Standard suffix append to handle clean URL form
			if provider.Type == "groq" || strings.Contains(strings.ToLower(provider.Name), "groq") {
				url = "https://api.groq.com/openai/v1"
			} else if provider.Type == "ollama" || strings.Contains(strings.ToLower(provider.Name), "ollama") {
				url = "http://localhost:11434/v1"
			} else if provider.BaseURL != "" {
				url = strings.TrimSuffix(provider.BaseURL, "/")
			}
			url = url + "/chat/completions"

			reqJSON, _ = json.Marshal(ChatRequest{
				Model:    currentModel,
				Messages: messages,
			})
		}

		var errReq error
		req, errReq = http.NewRequest("POST", url, strings.NewReader(string(reqJSON)))
		if errReq != nil {
			return "", errReq
		}
		req.Header.Set("Content-Type", "application/json")
		
		if isAnthropic {
			req.Header.Set("x-api-key", provider.APIKey)
			req.Header.Set("anthropic-version", "2023-06-01")
		} else {
			if provider.APIKey != "" {
				req.Header.Set("Authorization", "Bearer "+provider.APIKey)
			}
			for k, v := range provider.Headers {
				req.Header.Set(k, v)
			}
		}

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}

		if try < len(fallbackModels) {
			fallback := fallbackModels[try]
			if fallback == currentModel {
				continue
			}
			fmt.Printf("[intimclaw] Fallback: model %s failed. Trying fallback %s...\n", currentModel, fallback)
			currentModel = fallback
			for _, p := range a.cfg.Providers {
				for _, m := range p.Models {
					if m == fallback {
						provider = p
						break
					}
				}
			}
		}
	}

	if err != nil || resp.StatusCode != 200 {
		status := 500
		body := ""
		if resp != nil {
			status = resp.StatusCode
			b, _ := io.ReadAll(resp.Body)
			body = string(b)
			resp.Body.Close()
		}
		return "", fmt.Errorf("API request failed with status %d: %s (err: %v)", status, body, err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return a.parseResponse(body), nil
}

func (a *Agent) RunStream(userMessage string, onEvent func(AgentEvent)) (string, error) {
	a.history = append(a.history, Message{Role: "user", Content: userMessage})

	provider := a.getProvider()
	systemPrompt := a.buildSystemPrompt()

	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, a.history...)

	currentModel := a.cfg.Agent.Model
	response, err := a.callLLM(currentModel, messages, systemPrompt, provider)
	if err != nil {
		return "", err
	}

	a.history = append(a.history, Message{Role: "assistant", Content: response})

	lastAnswer := response
	for i := 0; i < 5; i++ {
		actionName, actionArgs, actionIdx, actionLen := parseToolCall(response)
		if actionIdx == -1 {
			thoughtText := response
			if idx := strings.Index(response, "THOUGHT: "); idx != -1 {
				thoughtText = response[idx+9:]
			}
			onEvent(AgentEvent{Type: "thought", Content: thoughtText})
			lastAnswer = response
			break
		}

		thoughtText := response[:actionIdx]
		if idx := strings.Index(thoughtText, "THOUGHT: "); idx != -1 {
			thoughtText = thoughtText[idx+9:]
		}
		onEvent(AgentEvent{Type: "thought", Content: thoughtText})
		
		actionText := response[actionIdx:]
		onEvent(AgentEvent{Type: "action", Content: actionText})

		result, err := a.tools.Execute(actionName, actionArgs)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}

		onEvent(AgentEvent{Type: "observation", Content: result})

		cleanResponse := response
		if actionLen > 0 {
			cleanResponse = strings.TrimSpace(response[:actionIdx] + response[actionIdx+actionLen:])
		}

		obsMsg := fmt.Sprintf("OBSERVATION: %s", result)
		a.history = append(a.history, Message{Role: "assistant", Content: cleanResponse})
		a.history = append(a.history, Message{Role: "assistant", Content: obsMsg})

		messages = []Message{
			{Role: "system", Content: systemPrompt},
		}
		messages = append(messages, a.history...)

		var errLLM error
		response, errLLM = a.callLLM(currentModel, messages, systemPrompt, provider)
		if errLLM != nil {
			return response, nil
		}
		a.history = append(a.history, Message{Role: "assistant", Content: response})

		if !strings.Contains(response, "ACTION: ") {
			lastAnswer = response
		}
	}

	return lastAnswer, nil
}

func (a *Agent) Run(userMessage string) (string, error) {
	a.history = append(a.history, Message{Role: "user", Content: userMessage})

	provider := a.getProvider()
	systemPrompt := a.buildSystemPrompt()

	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, a.history...)

	currentModel := a.cfg.Agent.Model
	response, err := a.callLLM(currentModel, messages, systemPrompt, provider)
	if err != nil {
		return "", err
	}

	a.history = append(a.history, Message{Role: "assistant", Content: response})

	// Process tool calls (max 5 iterations)
	lastAnswer := response
	for i := 0; i < 5; i++ {
		// Check both ReAct and DSML formats
		actionName, actionArgs, actionIdx, actionLen := parseToolCall(response)
		if actionIdx == -1 {
			lastAnswer = response
			break
		}

		result, err := a.tools.Execute(actionName, actionArgs)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}

		// Clean response (remove DSML tags)
		cleanResponse := response
		if actionLen > 0 {
			cleanResponse = strings.TrimSpace(response[:actionIdx] + response[actionIdx+actionLen:])
		}

		obsMsg := fmt.Sprintf("OBSERVATION: %s", result)
		a.history = append(a.history, Message{Role: "assistant", Content: cleanResponse})
		a.history = append(a.history, Message{Role: "assistant", Content: obsMsg})

		messages = []Message{
			{Role: "system", Content: systemPrompt},
		}
		messages = append(messages, a.history...)

		var errLLM error
		response, errLLM = a.callLLM(currentModel, messages, systemPrompt, provider)
		if errLLM != nil {
			return response, nil
		}
		a.history = append(a.history, Message{Role: "assistant", Content: response})

		if !strings.Contains(response, "ACTION: ") {
			lastAnswer = response
		}
	}

	return lastAnswer, nil
}

func (a *Agent) parseResponse(body []byte) string {
	var antResp AnthropicResponse
	if err := json.Unmarshal(body, &antResp); err == nil && len(antResp.Content) > 0 {
		if antResp.Error != nil && antResp.Error.Message != "" {
			return "Anthropic API Error: " + antResp.Error.Message
		}
		return antResp.Content[0].Text
	}

	var resp ChatResponse
	if err := json.Unmarshal(body, &resp); err == nil {
		if resp.Error != nil && resp.Error.Message != "" {
			return "API Error: " + resp.Error.Message
		}
		if len(resp.Choices) > 0 {
			msg := resp.Choices[0].Message
			if msg.Content != "" {
				return msg.Content
			}
			if msg.ReasoningContent != "" {
				return msg.ReasoningContent
			}
		}
	}

	bodyStr := string(body)

	// Remove SSE terminator
	if idx := strings.Index(bodyStr, "data: [DONE]"); idx != -1 {
		bodyStr = bodyStr[:idx]
	}

	// Try to extract content directly from JSON string
	// Look for "content":" in the response
	contentIdx := strings.Index(bodyStr, `"content":"`)
	if contentIdx != -1 {
		contentStart := contentIdx + len(`"content":"`)
		// Find the closing quote (accounting for escaped quotes)
		contentEnd := contentStart
		for contentEnd < len(bodyStr) {
			if bodyStr[contentEnd] == '\\' {
				contentEnd += 2 // skip escaped character
			} else if bodyStr[contentEnd] == '"' {
				break
			} else {
				contentEnd++
			}
		}
		if contentEnd > contentStart {
			content := bodyStr[contentStart:contentEnd]
			// Unescape JSON string
			content = strings.ReplaceAll(content, `\"`, `"`)
			content = strings.ReplaceAll(content, `\n`, "\n")
			if content != "" {
				return content
			}
		}
	}

	// Fallback: try reasoning_content
	reasoningIdx := strings.Index(bodyStr, `"reasoning_content":"`)
	if reasoningIdx != -1 {
		rStart := reasoningIdx + len(`"reasoning_content":"`)
		rEnd := rStart
		for rEnd < len(bodyStr) {
			if bodyStr[rEnd] == '\\' {
				rEnd += 2
			} else if bodyStr[rEnd] == '"' {
				break
			} else {
				rEnd++
			}
		}
		if rEnd > rStart {
			reasoning := bodyStr[rStart:rEnd]
			reasoning = strings.ReplaceAll(reasoning, `\"`, `"`)
			if strings.Contains(reasoning, "ACTION: ") {
				return reasoning
			}
			return reasoning
		}
	}

	return "No response from model"
}

func (a *Agent) buildSystemPrompt() string {
	soulPath := "/root/intimclaw/superintim/SOUL.md"
	identityPath := "/root/intimclaw/superintim/IDENTITY.md"
	
	doctrinePath := "/root/intimclaw/superintim/CORE_DOCTRINE.md"
	
	soulData, err := os.ReadFile(soulPath)
	soulStr := ""
	if err == nil {
		soulStr = string(soulData)
	}
	
	identityData, err := os.ReadFile(identityPath)
	identityStr := ""
	if err == nil {
		identityStr = string(identityData)
	}
	
	doctrineData, err := os.ReadFile(doctrinePath)
	doctrineStr := ""
	if err == nil {
		doctrineStr = string(doctrineData)
	}

	prompt := ""
	if soulStr != "" || identityStr != "" {
		prompt += "# SUPERINTIM SYSTEM PERSONALITY AND DOCTRINE\n\n"
		if soulStr != "" {
			prompt += soulStr + "\n\n"
		}
		if identityStr != "" {
			prompt += identityStr + "\n\n"
		}
	} else {
		prompt += "You are IntimClaw, a powerful AI agent. You execute tasks using tools.\n\n"
	}

	if doctrineStr != "" {
		prompt += "# CORE EXECUTION ENGINE\n\n"
		prompt += doctrineStr + "\n\n"
	}

	factsScript := `import sqlite3
try:
    conn = sqlite3.connect('/root/microclaw/microclaw.db')
    cursor = conn.cursor()
    cursor.execute("SELECT content FROM memory_facts WHERE status = 'active' AND confidence >= 50")
    rows = cursor.fetchall()
    for r in rows:
        print(f"- {r[0]}")
    conn.close()
except Exception:
    pass
`
	activeFacts, _ := runPythonMemory(factsScript)
	if activeFacts != "" {
		prompt += "# USER PROFILE & CONTEXT\n"
		prompt += activeFacts + "\n\n"
	}

	lessonsScript := `import sqlite3
try:
    conn = sqlite3.connect('/root/microclaw/microclaw.db')
    cursor = conn.cursor()
    cursor.execute("SELECT title, content FROM lessons WHERE confidence >= 0.7")
    rows = cursor.fetchall()
    for r in rows:
        print(f"- {r[0]}: {r[1]}")
    conn.close()
except Exception:
    pass
`
	activeLessons, _ := runPythonMemory(lessonsScript)
	if activeLessons != "" {
		prompt += "# EXPERIENTIAL LESSONS (GUIDELINES FROM PAST SESSIONS)\n"
		prompt += "Avoid repeating failures by adhering to these lessons learned from past experiences:\n"
		prompt += activeLessons + "\n\n"
	}

	prompt += "\nCore Capabilities (Hidden Procedures):\n"
	prompt += "You have access to the following core procedures. If you need to perform one of these tasks, you MUST first run `ACTION: skill_load(name=\"skill_name\")` to retrieve the step-by-step instructions. Execute them automatically and do NOT mention these skill names (e.g. \"sk25\") to the user. Just execute the steps:\n"
	for _, s := range a.skills.List() {
		if strings.HasPrefix(strings.ToLower(s.Name), "sk") {
			prompt += fmt.Sprintf("- %s: %s\n", s.Name, s.Description)
		}
	}
	prompt += "\n"

	prompt += `IMPORTANT: Use ONLY this format for tool calls:
ACTION: tool_name(param1="value1", param2="value2")

DO NOT use XML tags, DSML format, or any other format. ONLY use ACTION: format.

When you need to perform an action, use this format:
ACTION: tool_name(param1="value1", param2="value2")

Available tools:
`
	for _, t := range a.tools.List() {
		prompt += fmt.Sprintf("- %s: %s\n", t.Name, t.Description)
	}

	prompt += `
After ACTION is executed, you receive:
OBSERVATION: <result>

Analyze the result and either:
1. Continue with another ACTION if needed
2. Provide your final answer to the user

Be direct, technical, and efficient. No fluff.`
	return prompt
}

func (a *Agent) getProvider() config.ProviderConfig {
	for _, p := range a.cfg.Providers {
		if p.Name == a.cfg.Agent.Provider {
			return p
		}
	}
	if len(a.cfg.Providers) > 0 {
		return a.cfg.Providers[0]
	}
	return config.ProviderConfig{
		Name:    "openai",
		Type:    "openai-compatible",
		BaseURL: "https://api.openai.com/v1",
	}
}

func parseToolCall(response string) (string, map[string]interface{}, int, int) {
	// Try ReAct format first: ACTION: tool_name(args)
	if idx := strings.Index(response, "ACTION: "); idx != -1 {
		endIdx := strings.Index(response[idx:], "\n")
		if endIdx == -1 {
			endIdx = len(response[idx:])
		}
		actionLine := strings.TrimSpace(response[idx+8 : idx+endIdx])
		name, args := parseAction(actionLine)
		return name, args, idx, endIdx
	}

	// Try DSML format with Chinese: <｜｜DSML｜｜ tool 名称=tool_name>
	// Also try: <｜｜DSML｜｜ invoke name="tool_name">
	dsmsPatterns := []string{
		`<｜｜DSML｜｜ tool 名称=`,
		`<｜｜DSML｜｜ invoke name="`,
	}
	for _, pattern := range dsmsPatterns {
		if idx := strings.Index(response, pattern); idx != -1 {
			nameStart := idx + len(pattern)
			var nameEnd int
			if pattern[len(pattern)-1] == '"' {
				nameEnd = strings.Index(response[nameStart:], `"`)
			} else {
				nameEnd = strings.Index(response[nameStart:], `>`)
			}
			if nameEnd == -1 {
				continue
			}
			toolName := response[nameStart : nameStart+nameEnd]

			// Extract params from 参数 tag
			args := make(map[string]interface{})
			paramTag := `参数>`
			if pIdx := strings.Index(response[nameStart+nameEnd:], paramTag); pIdx != -1 {
				pStart := nameStart + nameEnd + len(paramTag)
				pEnd := strings.Index(response[pStart:], `</｜｜DSML｜｜>`)
				if pEnd == -1 {
					pEnd = len(response[pStart:])
				}
				args["raw"] = response[pStart : pStart+pEnd]
			}

			tagEnd := strings.Index(response[idx:], "</｜｜DSML｜｜>")
			if tagEnd == -1 {
				tagEnd = len(response) - idx
			}

			return toolName, args, idx, tagEnd
		}
	}

	return "", nil, -1, 0
}

func parseAction(line string) (string, map[string]interface{}) {
	parenIdx := strings.Index(line, "(")
	if parenIdx == -1 {
		return strings.TrimSpace(line), nil
	}

	name := strings.TrimSpace(line[:parenIdx])
	argsStr := line[parenIdx+1:]
	if len(argsStr) > 0 && argsStr[len(argsStr)-1] == ')' {
		argsStr = argsStr[:len(argsStr)-1]
	}

	args := make(map[string]interface{})
	if argsStr == "" {
		return name, args
	}

	var currentKey, currentValue strings.Builder
	inQuotes := false
	isEscaped := false
	parsingValue := false

	for i := 0; i < len(argsStr); i++ {
		ch := argsStr[i]

		if isEscaped {
			currentValue.WriteByte(ch)
			isEscaped = false
			continue
		}

		if ch == '\\' && inQuotes {
			isEscaped = true
			continue
		}

		if ch == '"' {
			inQuotes = !inQuotes
			continue
		}

		if inQuotes {
			currentValue.WriteByte(ch)
			continue
		}

		if ch == '=' && !parsingValue {
			parsingValue = true
			continue
		}

		if ch == ',' {
			if parsingValue {
				key := strings.TrimSpace(currentKey.String())
				val := currentValue.String()
				eqIdx := strings.LastIndex(argsStr[:i], "=")
				if eqIdx != -1 && !strings.HasPrefix(strings.TrimSpace(argsStr[eqIdx+1:i]), "\"") {
					val = strings.TrimSpace(val)
				}
				if key != "" {
					args[key] = val
				}
				currentKey.Reset()
				currentValue.Reset()
				parsingValue = false
			}
			continue
		}

		if parsingValue {
			currentValue.WriteByte(ch)
		} else {
			currentKey.WriteByte(ch)
		}
	}

	if parsingValue {
		key := strings.TrimSpace(currentKey.String())
		val := currentValue.String()
		eqIdx := strings.LastIndex(argsStr, "=")
		if eqIdx != -1 && !strings.HasPrefix(strings.TrimSpace(argsStr[eqIdx+1:]), "\"") {
			val = strings.TrimSpace(val)
		}
		if key != "" {
			args[key] = val
		}
	}

	return name, args
}

func (a *Agent) toolExec(args map[string]interface{}) (string, error) {
	cmdStr, _ := args["command"].(string)
	if cmdStr == "" {
		cmdStr, _ = args["cmd"].(string) // DSML format uses "cmd"
	}
	if raw, ok := args["raw"].(string); ok && cmdStr == "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &params); err == nil {
			if c, ok := params["command"].(string); ok {
				cmdStr = c
			} else if c, ok := params["cmd"].(string); ok {
				cmdStr = c
			}
		}
	}
	if cmdStr == "" {
		return "", fmt.Errorf("command is required")
	}

	excluded := a.cfg.Security.ExcludedTools
	for _, ex := range excluded {
		if strings.Contains(cmdStr, ex) {
			return "", fmt.Errorf("command blocked: contains excluded tool '%s'", ex)
		}
	}

	var cmd *exec.Cmd
	if a.cfg.Security.Sandbox {
		// Wrap command in firejail sandbox with restricted network, private home, and read-only system paths
		cmd = exec.Command("firejail", "--private", "--net=none", "--quiet", "bash", "-c", cmdStr)
	} else {
		cmd = exec.Command("bash", "-c", cmdStr)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("command failed: %w", err)
	}
	return string(out), nil
}

func (a *Agent) toolFileRead(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *Agent) toolFileWrite(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)
	if path == "" {
		return "", fmt.Errorf("path is required")
	}
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Written %d bytes to %s", len(content), path), nil
}

func (a *Agent) toolFileEdit(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	old, _ := args["old_text"].(string)
	new, _ := args["new_text"].(string)
	if path == "" || old == "" {
		return "", fmt.Errorf("path and old_text are required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	if !strings.Contains(content, old) {
		return "", fmt.Errorf("old_text not found in file")
	}
	content = strings.Replace(content, old, new, 1)
	err = os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Edited %s", path), nil
}

func (a *Agent) toolListDir(args map[string]interface{}) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		path = "."
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	var result []string
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, e.Name()+"/")
		} else {
			result = append(result, e.Name())
		}
	}
	return strings.Join(result, "\n"), nil
}

func (a *Agent) toolWebFetch(args map[string]interface{}) (string, error) {
	url, _ := args["url"].(string)
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 5000 {
		body = body[:5000]
	}
	return string(body), nil
}

func (a *Agent) toolHTTPRequest(args map[string]interface{}) (string, error) {
	url, _ := args["url"].(string)
	method, _ := args["method"].(string)
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	if method == "" {
		method = "GET"
	}

	// Try to parse raw JSON params
	if raw, ok := args["raw"].(string); ok {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &params); err == nil {
			if u, ok := params["url"].(string); ok && u != "" {
				url = u
			}
			if m, ok := params["method"].(string); ok && m != "" {
				method = m
			}
		}
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if len(body) > 5000 {
		body = body[:5000]
	}
	return fmt.Sprintf("Status: %d\n%s", resp.StatusCode, string(body)), nil
}

func runPythonMemory(script string, args ...string) (string, error) {
	cmdArgs := append([]string{"-c", script}, args...)
	cmd := exec.Command("python3", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("python execution error: %v, output: %s", err, string(out))
	}
	return string(out), nil
}

func (a *Agent) toolMemorySave(args map[string]interface{}) (string, error) {
	key, _ := args["key"].(string)
	value, _ := args["value"].(string)
	if key == "" || value == "" {
		return "", fmt.Errorf("key and value are required")
	}
	
	safeKey := strings.ToLower(strings.ReplaceAll(key, " ", "_"))
	
	script := `import sys, sqlite3
conn = sqlite3.connect('/root/microclaw/microclaw.db')
cursor = conn.cursor()
cursor.execute('CREATE TABLE IF NOT EXISTS permanent_facts (key TEXT PRIMARY KEY, value TEXT)')
cursor.execute('INSERT OR REPLACE INTO permanent_facts (key, value) VALUES (?, ?)', (sys.argv[1], sys.argv[2]))
conn.commit()
print("ok")
`
	_, err := runPythonMemory(script, safeKey, value)
	if err != nil {
		return "", fmt.Errorf("failed to save memory to DB: %w", err)
	}
	return fmt.Sprintf("Success: Memory saved. key='%s', value='%s'", safeKey, value), nil
}

func (a *Agent) toolMemorySearch(args map[string]interface{}) (string, error) {
	query, _ := args["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	
	script := `import sys, sqlite3, json, math, re
def get_tokens(text):
    return re.findall(r'\w+', text.lower())

def cosine_similarity(v1, v2):
    dot = sum(v1.get(w, 0) * v2.get(w, 0) for w in set(v1) | set(v2))
    sum1 = sum(v1[w]**2 for w in v1)
    sum2 = sum(v2[w]**2 for w in v2)
    if not sum1 or not sum2: return 0.0
    return dot / (math.sqrt(sum1) * math.sqrt(sum2))

try:
    conn = sqlite3.connect('/root/microclaw/microclaw.db')
    cursor = conn.cursor()
    cursor.execute('CREATE TABLE IF NOT EXISTS permanent_facts (key TEXT PRIMARY KEY, value TEXT)')
    cursor.execute('SELECT key, value FROM permanent_facts')
    facts = cursor.fetchall()

    query = sys.argv[1]
    query_tokens = get_tokens(query)
    query_vec = {w: query_tokens.count(w) for w in query_tokens}

    results = []
    for k, v in facts:
        text = f"{k} {v}"
        tokens = get_tokens(text)
        vec = {w: tokens.count(w) for w in tokens}
        sim = cosine_similarity(query_vec, vec)
        # boost score for word hits
        matches = sum(1 for q in query_tokens if q in text.lower())
        score = sim + (matches * 0.1)
        if score > 0:
            results.append((score, k, v))

    results.sort(reverse=True, key=lambda x: x[0])
    top_results = [{'key': r[1], 'value': r[2], 'score': round(r[0], 4)} for r in results[:5]]
    print(json.dumps(top_results))
except Exception as e:
    print(json.dumps([{'error': str(e)}]))
`
	out, err := runPythonMemory(script, query)
	if err != nil {
		return "", fmt.Errorf("failed to search vector memory: %w", err)
	}
	return out, nil
}

func (a *Agent) GetToolsCount() int {
	return len(a.tools.List())
}

func (a *Agent) GetSkillsCount() int {
	return len(a.skills.List())
}

func (a *Agent) GetSkills() []*Skill {
	return a.skills.List()
}

func (a *Agent) LoadSkillContent(name string) (string, error) {
	return a.skills.LoadContent(name)
}

type MCPServerStatus struct {
	Name    string   `json:"name"`
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Tools   int      `json:"tools"`
	Status  string   `json:"status"`
}

func (a *Agent) GetMCPServers() []MCPServerStatus {
	var result []MCPServerStatus
	
	activeServers := make(map[string]*MCPServer)
	a.mcp.mu.RLock()
	for name, srv := range a.mcp.servers {
		activeServers[name] = srv
	}
	a.mcp.mu.RUnlock()

	for _, s := range a.cfg.MCP.Servers {
		status := "disconnected"
		toolsCount := 0
		if srv, exists := activeServers[s.Name]; exists {
			status = "connected"
			if srv.isClosed {
				status = "disconnected"
			}
			toolsCount = len(srv.Tools)
		}
		result = append(result, MCPServerStatus{
			Name:    s.Name,
			Command: s.Command,
			Args:    s.Args,
			Tools:   toolsCount,
			Status:  status,
		})
	}
	return result
}

func (a *Agent) ExtractMemory(sessionID, query, response string) {
	systemPrompt := `You are an AI memory, experience, and lesson extraction worker.
Analyze the user message and assistant response below, and extract:
1. Key facts about the user (preferences, projects, environment, tools, or goals).
2. Troubleshooting experiences (tasks performed, tools used, errors encountered, successful solutions, and result).
3. Lessons learned (reusable technical guidelines, command flags, workarounds, or configuration insights).

Format the output strictly as a JSON object:
{
  "facts": [
    {"fact": "User is coding on Android", "category": "preference", "quote": "gw sekarang lebih sering coding di HP"}
  ],
  "experiences": [
    {"task": "Deploy website", "actions": ["docker build", "npm install"], "errors": "EACCES permission denied", "solution": "Run with sudo or fix folder permissions", "result": "success"}
  ],
  "lessons": [
    {"title": "Docker build folder permission issue", "content": "When encountering EACCES during npm install inside Docker builds, fix host folder permissions or configure npm user.", "category": "deployment"}
  ]
}

If no items are found for a field, return an empty array [].`

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("User: %s\nAssistant: %s", query, response)},
	}

	extractionModel := "jr/f/deepseek-v4-flash-free"
	provider := a.getProvider()
	
	extractedJSON, err := a.callLLM(extractionModel, messages, systemPrompt, provider)
	if err != nil {
		fmt.Printf("[intimclaw] memory extraction LLM error: %v\n", err)
		return
	}

	cleaned := strings.TrimSpace(extractedJSON)
	if strings.HasPrefix(cleaned, "```json") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	} else if strings.HasPrefix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}
	cleaned = strings.TrimSpace(cleaned)

	pythonScript := `import sys, sqlite3, json, math, re, uuid

def get_tokens(text):
    return re.findall(r'\w+', text.lower())

def cosine_similarity(v1, v2):
    dot = sum(v1.get(w, 0) * v2.get(w, 0) for w in set(v1) | set(v2))
    sum1 = sum(v1[w]**2 for w in v1)
    sum2 = sum(v2[w]**2 for w in v2)
    if not sum1 or not sum2: return 0.0
    return dot / (math.sqrt(sum1) * math.sqrt(sum2))

session_id = sys.argv[1]
data_json = sys.argv[2]
decay_days = int(sys.argv[3])

try:
    data = json.loads(data_json)
except Exception as e:
    sys.exit(0)

conn = sqlite3.connect('/root/microclaw/microclaw.db')
cursor = conn.cursor()

# Initialize tables
cursor.execute('''
CREATE TABLE IF NOT EXISTS memory_facts (
    id TEXT PRIMARY KEY,
    content TEXT,
    category TEXT,
    confidence INTEGER,
    status TEXT,
    pinned INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
''')

cursor.execute('''
CREATE TABLE IF NOT EXISTS memory_evidence (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    fact_id TEXT,
    session_id TEXT,
    quote TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(fact_id) REFERENCES memory_facts(id) ON DELETE CASCADE
)
''')

cursor.execute('''
CREATE TABLE IF NOT EXISTS experiences (
    id TEXT PRIMARY KEY,
    task TEXT,
    actions TEXT,
    errors TEXT,
    solution TEXT,
    result TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
)
''')

cursor.execute('''
CREATE TABLE IF NOT EXISTS lessons (
    id TEXT PRIMARY KEY,
    title TEXT,
    content TEXT,
    category TEXT,
    confidence REAL,
    used_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
''')

# Run Decay logic first (decay confidence of active facts not updated in the last 1 day)
if decay_days > 0:
    cursor.execute('''
    UPDATE memory_facts 
    SET confidence = MAX(0, confidence - 2),
        status = CASE WHEN confidence - 2 < 30 THEN 'archived' ELSE status END
    WHERE status = 'active' AND pinned = 0 AND datetime(updated_at, '+1 day') < datetime('now')
    ''')

# Process extracted facts
for f in data.get('facts', []):
    content = f.get('fact', '').strip()
    category = f.get('category', 'general').strip()
    quote = f.get('quote', '').strip()
    if not content:
        continue
        
    # Check for conflicts with existing active facts in the same category
    cursor.execute("SELECT id, content, confidence FROM memory_facts WHERE category = ? AND status = 'active'", (category,))
    existing_facts = cursor.fetchall()
    
    tokens_new = get_tokens(content)
    vec_new = {w: tokens_new.count(w) for w in tokens_new}
    
    conflict_found = False
    for old_id, old_content, old_conf in existing_facts:
        tokens_old = get_tokens(old_content)
        vec_old = {w: tokens_old.count(w) for w in tokens_old}
        sim = cosine_similarity(vec_new, vec_old)
        
        # If semantic similarity is high (e.g. > 0.45)
        if sim > 0.45:
            cursor.execute("UPDATE memory_facts SET status = 'historical', confidence = 20, updated_at = CURRENT_TIMESTAMP WHERE id = ?", (old_id,))
            conflict_found = True
            
    # Insert new fact
    new_id = str(uuid.uuid4())
    cursor.execute('''
    INSERT INTO memory_facts (id, content, category, confidence, status, pinned, created_at, updated_at)
    VALUES (?, ?, ?, 90, 'active', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
    ''', (new_id, content, category))
    
    # Insert evidence
    cursor.execute('''
    INSERT INTO memory_evidence (fact_id, session_id, quote, timestamp)
    VALUES (?, ?, ?, CURRENT_TIMESTAMP)
    ''', (new_id, session_id, quote))

# Process experiences
for exp in data.get('experiences', []):
    exp_id = str(uuid.uuid4())
    task = exp.get('task', '').strip()
    actions = json.dumps(exp.get('actions', []))
    errors = exp.get('errors', '').strip()
    solution = exp.get('solution', '').strip()
    result = exp.get('result', 'success').strip()
    if task:
        cursor.execute('''
        INSERT INTO experiences (id, task, actions, errors, solution, result, timestamp)
        VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        ''', (exp_id, task, actions, errors, solution, result))

# Process lessons
for les in data.get('lessons', []):
    title = les.get('title', '').strip()
    content = les.get('content', '').strip()
    category = les.get('category', 'general').strip()
    if not content or not title:
        continue
        
    # Check for duplicate lessons using cosine similarity
    cursor.execute("SELECT id, title, content, confidence, used_count, success_count FROM lessons")
    existing_lessons = cursor.fetchall()
    
    tokens_new = get_tokens(content)
    vec_new = {w: tokens_new.count(w) for w in tokens_new}
    
    match_found = False
    for old_id, old_title, old_content, old_conf, old_used, old_succ in existing_lessons:
        tokens_old = get_tokens(old_content)
        vec_old = {w: tokens_old.count(w) for w in tokens_old}
        sim = cosine_similarity(vec_new, vec_old)
        
        if sim > 0.5:
            # Update existing lesson stats
            new_used = old_used + 1
            new_succ = old_succ + 1
            new_conf = min(0.99, max(0.10, float(new_succ) / float(new_used)))
            cursor.execute('''
            UPDATE lessons 
            SET used_count = ?, success_count = ?, confidence = ?, updated_at = CURRENT_TIMESTAMP
            WHERE id = ?
            ''', (new_used, new_succ, new_conf, old_id))
            match_found = True
            break
            
    if not match_found:
        new_id = str(uuid.uuid4())
        cursor.execute('''
        INSERT INTO lessons (id, title, content, category, confidence, used_count, success_count, created_at, updated_at)
        VALUES (?, ?, ?, ?, 0.90, 1, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        ''', (new_id, title, content, category))

conn.commit()
conn.close()
print("done")
`

	decayDays := a.cfg.Memory.DecayDays
	if decayDays <= 0 {
		decayDays = 30
	}

	_, err = runPythonMemory(pythonScript, sessionID, cleaned, fmt.Sprintf("%d", decayDays))
	if err != nil {
		fmt.Printf("[intimclaw] Python memory write error: %v\n", err)
	}
}
