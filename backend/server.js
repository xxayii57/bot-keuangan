import express from 'express';
import cors from 'cors';
import { WebSocketServer } from 'ws';
import http from 'http';
import dotenv from 'dotenv';
import path from 'path';
import fs from 'fs';

// Import Route Handlers / Helpers
import { getLocalToken, authGuard } from './security/authGuard.js';
import { getWorkspaceDir, setWorkspaceDir } from './security/pathGuard.js';
import { 
  getMemoryStore, 
  saveOAuthTokens, 
  clearOAuthTokens, 
  getActiveSettings, 
  updateEnvFile 
} from './auth/tokenStore.js';
import { 
  registerPendingToolCall, 
  getPendingRequest, 
  updateRequestStatus, 
  getAllRequests,
  removeRequest 
} from './security/approvalPolicy.js';

import { listDirFiles, readFileContent, writeFileContent, deleteFile } from './tools/fileTools.js';
import { runTerminalCommand } from './tools/terminalTools.js';
import { 
  getSavedVPSProfiles, 
  saveVPSProfile, 
  deleteVPSProfile, 
  connectVPS, 
  runVPSCommand, 
  disconnectVPS, 
  getVPSStatus 
} from './tools/vpsTools.js';

import { generateContentGeminiApiKey } from './providers/geminiApiKey.js';
import { generateContentGeminiOAuth } from './providers/geminiOAuth.js';
import { generateContentOpenAI } from './providers/openai.js';
import { generateContentCustomOpenAI } from './providers/customOpenAI.js';
import { checkCLIProviderStatus, triggerCLILogin } from './providers/cliProvider.js';
import { checkBridgeStatus } from './tools/shizukuBridge.js';

dotenv.config();

const app = express();
const server = http.createServer(app);
const port = process.env.PORT || process.env.BACKEND_PORT || 3001;
const host = process.env.HOST || '127.0.0.1';

// Setup Middlewares
app.use(cors({ origin: '*' }));
app.use(express.json());

// Serve static frontend files
const distPath = '/data/data/com.termux/files/home/frontend/dist';
app.use(express.static(distPath));

// Token check on all routes except bypassable ones
app.use(authGuard);

// Global list of active WebSocket connections
const aliveSockets = new Set();

// --- Express REST Routing API ---

// 1. Health
app.get('/api/health', (req, res) => {
  res.json({
    status: 'ok',
    os: process.platform,
    cwd: getWorkspaceDir(),
    timestamp: new Date().toISOString()
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'ok', detail: 'IntimClaw backend' });
});

// 2. Chat / Agent Execution
app.post('/api/chat', async (req, res) => {
  const { messages, model, runAgentTools } = req.body;

  if (!messages || !Array.isArray(messages)) {
    return res.status(400).json({ error: 'Messages array is required.' });
  }

  const settings = getActiveSettings();
  const provider = settings.activeProvider;
  const activeModel = model || settings.customModel || 'gemini-3.5-flash';

  try {
    let aiResponseText = '';

    // Route message to active model provider
    if (provider === 'gemini_api_key') {
      aiResponseText = await generateContentGeminiApiKey(messages, activeModel);
    } else if (provider === 'gemini_oauth') {
      aiResponseText = await generateContentGeminiOAuth(messages, activeModel);
    } else if (provider === 'openai') {
      aiResponseText = await generateContentOpenAI(messages, activeModel);
    } else if (provider === 'custom_openai') {
      aiResponseText = await generateContentCustomOpenAI(messages, activeModel);
    } else if (provider === 'cli_provider') {
      throw new Error('Official CLI Provider is only for CLI environment authentication, check statuses below.');
    } else {
      throw new Error('No active AI Provider configured.');
    }

    // Agent capability parsing: Check if the AI output hints at any tool actions
    // Simple regex or parse structure for tools:
    // Support a clean declarative tool JSON block in text response: ```tool_call {...}```
    let pendingTool = null;
    const toolMatch = aiResponseText.match(/```json\s*(\{[\s\S]*?"tool"[\s\S]*?\})\s*```/);
    if (toolMatch && runAgentTools) {
      try {
        const parsedTool = JSON.parse(toolMatch[1]);
        if (parsedTool.tool) {
          // Register as pending tool approval
          pendingTool = registerPendingToolCall(parsedTool.tool, parsedTool.arguments || {});
        }
      } catch (err) {
        // Failed tool block parsing, treat as normal text
      }
    }

    res.json({
      response: aiResponseText,
      toolCall: pendingTool
    });

  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 3. Authentications & Providers
app.get('/api/auth/status', async (req, res) => {
  const settings = getActiveSettings();
  const cliStatus = await checkCLIProviderStatus();
  const bridgeStatus = checkBridgeStatus();

  res.json({
    activeProvider: settings.activeProvider,
    oauthConnected: settings.oauthConnected,
    cliStatus,
    bridgeStatus,
    credentialStatus: {
      geminiApiKey: settings.geminiApiKey ? 'configured' : 'empty',
      openaiApiKey: settings.openaiApiKey ? 'configured' : 'empty',
      customApiKey: settings.customApiKey ? 'configured' : 'empty',
      customBaseUrl: settings.customBaseUrl || ''
    }
  });
});

// Start Google OAuth Callback Initiator
app.post('/api/auth/gemini/start', (req, res) => {
  const { clientId, clientSecret, redirectUri } = req.body;
  if (!clientId || !clientSecret) {
    return res.status(400).json({ error: 'Client ID & Client Secret are required for OAuth.' });
  }

  // Update .env with client keys temporarily
  updateEnvFile({
    GOOGLE_OAUTH_CLIENT_ID: clientId,
    GOOGLE_OAUTH_CLIENT_SECRET: clientSecret,
    GOOGLE_OAUTH_REDIRECT_URI: redirectUri || 'http://127.0.0.1:3001/api/auth/gemini/callback'
  });

  // Construct target Google authorization URI
  const scopes = 'https://www.googleapis.com/auth/generative-language';
  const state = getLocalToken();
  const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?` +
    `client_id=${encodeURIComponent(clientId)}&` +
    `redirect_uri=${encodeURIComponent(redirectUri)}&` +
    `response_type=code&` +
    `scope=${encodeURIComponent(scopes)}&` +
    `access_type=offline&` +
    `prompt=consent&` +
    `state=${state}`;

  res.json({ url: authUrl });
});

// Callback handler of Google Auth
app.get('/api/auth/gemini/callback', async (req, res) => {
  const { code, state, error } = req.query;

  if (error) {
    return res.send(`<h2>Authentication Failed</h2><p>${error}</p>`);
  }

  // Verify State Token for preventing CSRF attacks
  if (state !== getLocalToken()) {
    return res.status(403).send('<h2>Forbidden</h2><p>State verification failed.</p>');
  }

  const clientId = process.env.GOOGLE_OAUTH_CLIENT_ID;
  const clientSecret = process.env.GOOGLE_OAUTH_CLIENT_SECRET;
  const redirectUri = process.env.GOOGLE_OAUTH_REDIRECT_URI;

  try {
    const tokenUrl = 'https://oauth2.googleapis.com/token';
    const params = new URLSearchParams({
      code: code,
      client_id: clientId,
      client_secret: clientSecret,
      redirect_uri: redirectUri,
      grant_type: 'authorization_code'
    });

    const response = await fetch(tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: params.toString()
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(`Token exchange failed: ${errText}`);
    }

    const tokenData = await response.json();
    saveOAuthTokens(tokenData.access_token, tokenData.refresh_token, tokenData.expires_in);

    // Save success response with instruction to return to browser UI
    res.send(`
      <div style="font-family: sans-serif; text-align: center; padding: 40px; background: #0f172a; color: #f8fafc; height: 100vh; box-sizing: border-box;">
        <h2 style="color: #10b981;">✓ Google OAuth Connected Successfully!</h2>
        <p>Your local Node.js agent backend has securely received the access tokens.</p>
        <p>You can close this tab and return to your AI agent dashboard!</p>
        <button style="background: #10b981; color: white; border: none; padding: 10px 20px; border-radius: 6px; cursor: pointer; margin-top: 20px;" onclick="window.close()">Close Windows</button>
      </div>
    `);
  } catch (err) {
    res.status(500).send(`<h2>Auth Code Exchange Error</h2><p>${err.message}</p>`);
  }
});

app.post('/api/auth/gemini/logout', (req, res) => {
  clearOAuthTokens();
  res.json({ success: true, message: 'Google OAuth token deleted' });
});

// 4. Configuration Settings
app.post('/api/settings/save', (req, res) => {
  const { activeProvider, geminiApiKey, openaiApiKey, customApiKey, customBaseUrl, customModel } = req.body;
  
  const updates = {};
  if (geminiApiKey !== undefined) updates.GEMINI_API_KEY = geminiApiKey;
  if (openaiApiKey !== undefined) updates.OPENAI_API_KEY = openaiApiKey;
  if (customApiKey !== undefined) updates.CUSTOM_API_KEY = customApiKey;
  if (customBaseUrl !== undefined) updates.CUSTOM_BASE_URL = customBaseUrl;
  if (customModel !== undefined) updates.CUSTOM_MODEL = customModel;

  const store = getMemoryStore();
  if (activeProvider) {
    store.activeProvider = activeProvider;
    updates.DEFAULT_PROVIDER = activeProvider;
  }
  if (customModel) {
    store.customModel = customModel;
  }

  const result = updateEnvFile(updates);
  res.json({ success: result, settings: getActiveSettings() });
});

app.get('/api/settings/load', (req, res) => {
  res.json(getActiveSettings());
});

app.post('/api/settings/test-provider', async (req, res) => {
  const { provider, model } = req.body;
  const testMessage = [{ role: 'user', content: 'Ping! Keep response to 1 word only.' }];
  
  try {
    let text = '';
    if (provider === 'gemini_api_key') {
      text = await generateContentGeminiApiKey(testMessage, model || 'gemini-3.5-flash');
    } else if (provider === 'gemini_oauth') {
      text = await generateContentGeminiOAuth(testMessage, model || 'gemini-3.5-flash');
    } else if (provider === 'openai') {
      text = await generateContentOpenAI(testMessage, model || 'gpt-4o-mini');
    } else if (provider === 'custom_openai') {
      text = await generateContentCustomOpenAI(testMessage);
    } else {
      throw new Error(`Testing not configured for: ${provider}`);
    }
    res.json({ success: true, reply: text });
  } catch (error) {
    res.json({ success: false, error: error.message });
  }
});

// 5. Native/Local Tools Router with Approval Handlers
app.get('/api/tools/approvals', (req, res) => {
  res.json(getAllRequests());
});

app.post('/api/tools/list-files', (req, res) => {
  const { path: dirPath } = req.body;
  try {
    const files = listDirFiles(dirPath);
    res.json({ files });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

app.post('/api/tools/read-file', (req, res) => {
  const { path: filePath } = req.body;
  try {
    const content = readFileContent(filePath);
    res.json({ content });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

// Write file requires explicit manual approval
app.post('/api/tools/write-file', (req, res) => {
  const { path: filePath, content, approvedId } = req.body;

  try {
    if (approvedId) {
      const item = getPendingRequest(approvedId);
      if (!item || item.status !== 'approved') {
        return res.status(403).json({ error: 'Tool was not approved' });
      }

      writeFileContent(filePath, content);
      updateRequestStatus(approvedId, 'completed', { success: true });
      return res.json({ success: true, message: 'File written successfully' });
    }

    // Register a pending request
    const pending = registerPendingToolCall('write_file', { path: filePath, content });
    res.json({ pendingApproval: pending });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

// Delete file requires approval
app.post('/api/tools/delete-file', (req, res) => {
  const { path: filePath, approvedId } = req.body;

  try {
    if (approvedId) {
      const item = getPendingRequest(approvedId);
      if (!item || item.status !== 'approved') {
        return res.status(403).json({ error: 'Deletion not approved.' });
      }

      deleteFile(filePath);
      updateRequestStatus(approvedId, 'completed', { success: true });
      return res.json({ success: true, message: 'Deleted file successfully' });
    }

    const pending = registerPendingToolCall('delete_file', { path: filePath });
    res.json({ pendingApproval: pending });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

// Action Approve/Reject tool calls
app.post('/api/tools/approve', (req, res) => {
  const { id, approve } = req.body;
  const status = approve ? 'approved' : 'rejected';
  
  const updated = updateRequestStatus(id, status);
  if (!updated) {
    return res.status(404).json({ error: 'Approval ID not found' });
  }

  res.json({ success: true, request: updated });
});

// Cancel or dismiss a completed/rejected request
app.post('/api/tools/dismiss', (req, res) => {
  const { id } = req.body;
  removeRequest(id);
  res.json({ success: true });
});

app.post('/api/tools/workspace', (req, res) => {
  const { path: nPath } = req.body;
  try {
    const resolved = setWorkspaceDir(nPath);
    res.json({ path: resolved });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

// Upload file directly to downloads/ folder
app.post('/api/tools/upload', (req, res) => {
  const { name, base64Data } = req.body;

  if (!name || !base64Data) {
    return res.status(400).json({ error: 'Filename and base64Data are required.' });
  }

  try {
    const buffer = Buffer.from(base64Data, 'base64');
    const destDir = path.join(getWorkspaceDir(), 'downloads');
    
    // Ensure downloads dir exists
    if (!fs.existsSync(destDir)) {
      fs.mkdirSync(destDir, { recursive: true });
    }

    const destPath = path.join(destDir, name);
    fs.writeFileSync(destPath, buffer);

    res.json({ success: true, path: destPath, message: `File saved successfully to downloads/${name}` });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

// Run command (requires approval)
app.post('/api/tools/run-command', (req, res) => {
  const { command, approvedId } = req.body;

  try {
    if (approvedId) {
      const item = getPendingRequest(approvedId);
      if (!item || item.status !== 'approved') {
        return res.status(403).json({ error: 'Command not approved by policy.' });
      }

      // Execute command and steam socket updates
      let outputAcc = '';
      broadcastWebSocket({ type: 'terminal_start', command });

      const proc = runTerminalCommand(
        command,
        (data, type) => {
          outputAcc += data;
          broadcastWebSocket({ type: 'terminal_chunk', chunk: data, streamType: type });
        },
        (code) => {
          broadcastWebSocket({ type: 'terminal_end', code, fullOutput: outputAcc });
          updateRequestStatus(approvedId, 'completed', { code, output: outputAcc });
        }
      );

      return res.json({ success: true, running: true });
    }

    const pending = registerPendingToolCall('run_terminal_command', { command });
    res.json({ pendingApproval: pending });

  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

// 6. VPS Operations
app.get('/api/vps/profiles', (req, res) => {
  res.json(getSavedVPSProfiles());
});

app.post('/api/vps/profile', (req, res) => {
  const profile = saveVPSProfile(req.body);
  res.json(profile);
});

app.delete('/api/vps/profile/:id', (req, res) => {
  deleteVPSProfile(req.params.id);
  res.json({ success: true });
});

app.post('/api/vps/connect', (req, res) => {
  const { id } = req.body;
  try {
    connectVPS(
      id,
      (data, type) => {
        broadcastWebSocket({ type: 'vps_chunk', profileId: id, chunk: data, streamType: type });
      },
      (err, success) => {
        broadcastWebSocket({ type: 'vps_connection_status', profileId: id, success, error: err ? err.message : null });
      }
    );
    res.json({ success: true, status: 'connecting' });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

app.post('/api/vps/run', (req, res) => {
  const { id: profileId, command, approvedId } = req.body;

  try {
    if (approvedId) {
      const item = getPendingRequest(approvedId);
      if (!item || item.status !== 'approved') {
        return res.status(403).json({ error: 'VPS Command not approved.' });
      }

      runVPSCommand(
        profileId,
        command,
        (data, type) => {
          broadcastWebSocket({ type: 'vps_chunk', profileId, chunk: data, streamType: type });
        },
        (code) => {
          broadcastWebSocket({ type: 'vps_end', profileId, code });
          updateRequestStatus(approvedId, 'completed', { code });
        }
      );
      return res.json({ success: true });
    }

    const pending = registerPendingToolCall('run_vps_command', { id: profileId, command });
    res.json({ pendingApproval: pending });
  } catch (e) {
    res.status(500).json({ error: e.message });
  }
});

app.post('/api/vps/disconnect', (req, res) => {
  const { id } = req.body;
  disconnectVPS(id);
  res.json({ success: true });
});

// --- WebSocket Broker Setup ---
const wss = new WebSocketServer({ noServer: true });

server.on('upgrade', (request, socket, head) => {
  // Simple token query extraction on WS Upgrade
  const url = new URL(request.url, `http://${request.headers.host}`);
  const token = url.searchParams.get('token');

  if (token !== getLocalToken()) {
    socket.write('HTTP/1.1 401 Unauthorized\r\n\r\n');
    socket.destroy();
    return;
  }

  wss.handleUpgrade(request, socket, head, (ws) => {
    wss.emit('connection', ws, request);
  });
});

wss.on('connection', (ws) => {
  aliveSockets.add(ws);
  
  ws.send(JSON.stringify({ 
    type: 'welcome', 
    message: 'Secure websocket stream broker connected.',
    workspace: getWorkspaceDir() 
  }));

  ws.on('close', () => {
    aliveSockets.delete(ws);
  });
});

function broadcastWebSocket(payload) {
  const data = JSON.stringify(payload);
  for (const client of aliveSockets) {
    if (client.readyState === 1) { // OPEN
      client.send(data);
    }
  }
}

// Catch-all route to serve React frontend SPA routes
app.get('*', (req, res, next) => {
  if (req.path.startsWith('/api/')) {
    return next();
  }
  res.sendFile(path.join(distPath, 'index.html'));
});

// Start Server
server.listen(port, host, () => {
  console.log(`Backend local agent listening on http://${host}:${port}`);
  console.log(`LOCAL_ACCESS_TOKEN configuration: ${getLocalToken()}`);
});
