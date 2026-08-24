#!/bin/bash
# deploy.sh — One-shot deployment for intimclaw on STB
# Run this from STB when connected to home WiFi:
#   bash deploy.sh
#
# Prerequisites: STB at 192.168.1.81, root access, internet connection

set -e
STB_USER="root"
STB_HOST="192.168.1.81"
INSTALL_DIR="$HOME/intimclaw-new"
CONFIG_DIR="$HOME/.intimclaw"

echo "================================================"
echo "  IntimClaw STB Deployment"
echo "================================================"

# 1. Stop old processes
echo "[1/6] Stopping old processes..."
pkill -f "intimclaw.*gateway" 2>/dev/null || true
pkill -f "intimclaw-launcher" 2>/dev/null || true
sleep 2

# 2. Install cloudflared
echo "[2/6] Installing cloudflared..."
if ! command -v cloudflared &>/dev/null; then
    ARCH=$(uname -m)
    case "$ARCH" in
        aarch64|arm64) CF_ARCH="arm64" ;;
        x86_64|amd64) CF_ARCH="amd64" ;;
        armv7l|armhf) CF_ARCH="arm" ;;
        *) echo "Unsupported arch: $ARCH"; exit 1 ;;
    esac
    curl -sL "https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-${CF_ARCH}" -o /usr/local/bin/cloudflared
    chmod +x /usr/local/bin/cloudflared
    echo "  cloudflared installed ($CF_ARCH)"
else
    echo "  cloudflared already installed"
fi

# 3. Install microclaw MCP bridge
echo "[3/6] Setting up microclaw MCP bridge..."
if [ -d "$HOME/microclaw" ]; then
    cat > /tmp/microclaw-mcp.py << 'MCPEOF'
#!/usr/bin/env python3
"""MicroClaw MCP Server — exposes microclaw tools as MCP for intimclaw."""
import json, sys, os, subprocess, importlib, traceback

TOOLS_DIR = os.path.expanduser("~/microclaw/tools")
sys.path.insert(0, os.path.expanduser("~/microclaw"))
from tools import TOOL_REGISTRY, load_tools

# Load all microclaw tools
load_tools()

def handle_request(req):
    method = req.get("method", "")
    if method == "initialize":
        tools = []
        for name, func in TOOL_REGISTRY.items():
            desc = (func.__doc__ or name).strip().split("\n")[0]
            schema = {"type": "object", "properties": {}, "required": []}
            import inspect
            sig = inspect.signature(func)
            for pname, param in sig.parameters.items():
                schema["properties"][pname] = {"type": "string", "description": pname}
                if param.default is inspect.Parameter.empty:
                    schema["required"].append(pname)
            tools.append({"name": name, "description": desc, "inputSchema": schema})
        return {"jsonrpc": "2.0", "id": req.get("id"), "result": {
            "protocolVersion": "2024-11-05",
            "capabilities": {"tools": {"listChanged": False}},
            "serverInfo": {"name": "microclaw", "version": "1.0.0"}
        }, "tools": tools}
    elif method == "tools/list":
        tools = []
        for name, func in TOOL_REGISTRY.items():
            desc = (func.__doc__ or name).strip().split("\n")[0]
            import inspect
            sig = inspect.signature(func)
            props = {}
            req = []
            for pname, param in sig.parameters.items():
                props[pname] = {"type": "string", "description": pname}
                if param.default is inspect.Parameter.empty:
                    req.append(pname)
            tools.append({"name": name, "description": desc,
                         "inputSchema": {"type": "object", "properties": props, "required": req}})
        return {"jsonrpc": "2.0", "id": req.get("id"), "result": {"tools": tools}}
    elif method == "tools/call":
        tool_name = req.get("params", {}).get("name", "")
        arguments = req.get("params", {}).get("arguments", {})
        if tool_name not in TOOL_REGISTRY:
            return {"jsonrpc": "2.0", "id": req.get("id"), "error": {"code": -32601, "message": f"Tool {tool_name} not found"}}
        try:
            result = TOOL_REGISTRY[tool_name](**arguments)
            return {"jsonrpc": "2.0", "id": req.get("id"), "result": {
                "content": [{"type": "text", "text": str(result)}], "isError": False}}
        except Exception as e:
            return {"jsonrpc": "2.0", "id": req.get("id"), "result": {
                "content": [{"type": "text", "text": f"Error: {e}\n{traceback.format_exc()}"}], "isError": True}}
    elif method == "ping":
        return {"jsonrpc": "2.0", "id": req.get("id"), "result": {}}
    return {"jsonrpc": "2.0", "id": req.get("id"), "error": {"code": -32601, "message": "Method not found"}}

if __name__ == "__main__":
    for line in sys.stdin:
        try:
            req = json.loads(line.strip())
            resp = handle_request(req)
            sys.stdout.write(json.dumps(resp) + "\n")
            sys.stdout.flush()
        except: pass
MCPEOF
    cp /tmp/microclaw-mcp.py ~/microclaw/mcp_server.py
    echo "  MCP bridge ready at ~/microclaw/mcp_server.py"
else
    echo "  WARNING: microclaw not found at ~/microclaw, skipping MCP bridge"
fi

# 4. Create systemd services
echo "[4/6] Creating systemd services..."
cat > /etc/systemd/system/intimclaw-gateway.service << SVCEOF
[Unit]
Description=IntimClaw Gateway (Telegram + Agent)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Environment=HOME=/root
WorkingDirectory=/root
ExecStart=/root/intimclaw-new/bin/intimclaw gateway
Restart=always
RestartSec=10
StandardOutput=append:/var/log/intimclaw-gateway.log
StandardError=append:/var/log/intimclaw-gateway.log

[Install]
WantedBy=multi-user.target
SVCEOF

cat > /etc/systemd/system/intimclaw-webui.service << SVCEOF
[Unit]
Description=IntimClaw WebUI Launcher
After=intimclaw-gateway.service
Wants=intimclaw-gateway.service

[Service]
Type=simple
User=root
Environment=HOME=/root
WorkingDirectory=/root
ExecStart=/root/intimclaw-new/bin/intimclaw-launcher -public --port 18080
Restart=always
RestartSec=10
StandardOutput=append:/var/log/intimclaw-webui.log
StandardError=append:/var/log/intimclaw-webui.log

[Install]
WantedBy=multi-user.target
SVCEOF

cat > /etc/systemd/system/intimclaw-tunnel.service << SVCEOF
[Unit]
Description=Cloudflare Tunnel for IntimClaw WebUI
After=intimclaw-webui.service
Wants=intimclaw-webui.service

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/cloudflared tunnel --url http://localhost:18080 --no-autoupdate
Restart=always
RestartSec=30
StandardOutput=append:/var/log/intimclaw-tunnel.log
StandardError=append:/var/log/intimclaw-tunnel.log

[Install]
WantedBy=multi-user.target
SVCEOF

cat > /etc/systemd/system/intimclaw-mcp.service << SVCEOF
[Unit]
Description=MicroClaw MCP Bridge (74 tools for IntimClaw)
After=network-online.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/microclaw
ExecStart=/usr/bin/python3 /root/microclaw/mcp_server.py
Restart=always
RestartSec=10
StandardOutput=append:/var/log/intimclaw-mcp.log
StandardError=append:/var/log/intimclaw-mcp.log

[Install]
WantedBy=multi-user.target
SVCEOF

systemctl daemon-reload
echo "  Services created"

# 5. Enable all services
echo "[5/6] Enabling services..."
systemctl enable intimclaw-gateway intimclaw-webui intimclaw-tunnel intimclaw-mcp
echo "  Services enabled (will auto-start on boot)"

# 6. Start services
echo "[6/6] Starting services..."
systemctl start intimclaw-gateway
sleep 3
systemctl start intimclaw-webui
sleep 3
systemctl start intimclaw-mcp
sleep 3
systemctl start intimclaw-tunnel
sleep 5

# Show status
echo ""
echo "================================================"
echo "  STATUS"
echo "================================================"
for svc in intimclaw-gateway intimclaw-webui intimclaw-tunnel intimclaw-mcp; do
    STATUS=$(systemctl is-active $svc 2>/dev/null || echo "inactive")
    printf "  %-30s %s\n" "$svc" "$STATUS"
done

# Show tunnel URL
echo ""
echo "  Tunnel URL:"
TUNNEL_LOG=$(journalctl -u intimclaw-tunnel --no-pager -n 20 2>/dev/null || cat /var/log/intimclaw-tunnel.log 2>/dev/null)
TUNNEL_URL=$(echo "$TUNNEL_LOG" | grep -oP "https://[a-z0-9-]+\.trycloudflare\.com" | tail -1)
if [ -n "$TUNNEL_URL" ]; then
    echo "  $TUNNEL_URL"
else
    echo "  (checking... run: journalctl -u intimclaw-tunnel -f)"
fi

echo ""
echo "  Telegram: @intimclawbot"
echo "  WebUI: $TUNNEL_URL (or http://192.168.1.81:18080 on LAN)"
echo "  Logs: journalctl -u intimclaw-gateway -f"
echo ""
