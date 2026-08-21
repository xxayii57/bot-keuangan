import React, { useState } from 'react';
import { useAgent } from '../stores/agentStore';
import { Shield, Key, FolderOpen, RefreshCw, Cpu } from 'lucide-react';
import { api } from '../lib/api';

export const Header: React.FC = () => {
  const { 
    activeProvider, 
    oauthConnected, 
    localToken, 
    updateLocalToken, 
    currentPath, 
    setCurrentPath,
    loadFileList
  } = useAgent();

  const [showTokenInput, setShowTokenInput] = useState(false);
  const [tokenInput, setTokenInput] = useState(localToken);
  const [workspaceInput, setWorkspaceInput] = useState('');
  const [showWorkModal, setShowWorkModal] = useState(false);

  const handleSaveToken = () => {
    updateLocalToken(tokenInput);
    setShowTokenInput(false);
    alert('Local access token successfully updated inside browser. Reconnecting streams...');
  };

  const handleUpdateWorkspace = async () => {
    try {
      const res = await api.setWorkspace(workspaceInput || currentPath);
      setCurrentPath(res.path);
      loadFileList(res.path);
      setShowWorkModal(false);
      alert(`Workspace successfully mapped to: ${res.path}`);
    } catch (e: any) {
      alert(`Error updating workspace: ${e.message}`);
    }
  };

  return (
    <header className="bg-[#F3EDF7] border-b border-[#E6E0E9] px-4 py-2.5 flex flex-wrap items-center justify-between gap-4 sticky top-0 z-40 select-none text-[#1D1B20]">
      <div className="flex items-center gap-3">
        <div className="h-10 w-10 rounded-xl overflow-hidden shadow-sm flex items-center justify-center border border-[#E6E0E9] bg-white shrink-0">
          <img src="/logo.png" alt="Codex Agent Logo" className="h-full w-full object-cover" />
        </div>
        <div>
          <h1 className="text-[16px] font-semibold leading-tight flex items-center gap-2 text-[#1D1B20]">
            INTIMCLAW AI AGENT
            <span className="text-[10px] bg-[#E8DEF8] text-[#1D192B] px-2 py-0.5 rounded font-mono font-bold border border-[#D0BCFF]">v1.2</span>
          </h1>
          <p className="text-[11px] text-[#49454F] font-medium uppercase tracking-wider">Safe Local Terminal & Bridge Assistant</p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        {/* Active Provider Status Indicator */}
        <div className="flex items-center gap-2 bg-[#E8DEF8]/50 px-3 py-1.5 rounded-full border border-[#D0BCFF]">
          <Shield className="h-4 w-4 text-[#6750A4]" />
          <span className="text-[11px] font-bold text-[#49454F] uppercase tracking-tight">Provider:</span>
          <span className="text-[11px] font-mono font-bold text-[#21005D] uppercase bg-[#EADDFF] px-2.5 py-0.5 rounded-full border border-[#D0BCFF]">
            {activeProvider.replace('_', ' ')}
          </span>
          {activeProvider === 'gemini_oauth' && (
            <span className={`h-2.5 w-2.5 rounded-full ${oauthConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`} />
          )}
        </div>

        {/* Workspace directory */}
        <button 
          onClick={() => {
            setWorkspaceInput(currentPath);
            setShowWorkModal(true);
          }}
          className="flex items-center gap-2 bg-white hover:bg-[#F3EDF7] text-[#1D1B20] hover:text-[#6750A4] px-3 py-1.5 rounded-xl border border-[#CAC4D0] text-xs transition-colors shadow-sm font-semibold"
        >
          <FolderOpen className="h-4 w-4 text-amber-600" />
          <span className="font-mono truncate max-w-[200px]">{currentPath || 'Loading workspace...'}</span>
        </button>

        {/* Local Access Token Key toggle */}
        <div className="relative">
          <button 
            onClick={() => setShowTokenInput(!showTokenInput)}
            className="flex items-center gap-2 bg-white hover:bg-[#F3EDF7] text-[#1D1B20] hover:text-[#6750A4] px-3 py-1.5 rounded-xl border border-[#CAC4D0] text-xs transition-colors shadow-sm font-semibold"
          >
            <Key className="h-4 w-4 text-indigo-500" />
            <span>Local Token</span>
          </button>
          
          {showTokenInput && (
            <div className="absolute right-0 mt-2 bg-white border border-[#CAC4D0] rounded-2xl p-4 shadow-xl w-72 z-50 text-[#1D1B20]">
              <h4 className="text-xs font-bold text-[#1D192B] mb-2 flex items-center gap-1.5">
                <Shield className="h-4 w-4 text-[#6750A4]" /> Secure Connection Passcode
              </h4>
              <p className="text-[11px] text-[#49454F] mb-3 leading-relaxed font-sans">
                Backend local Anda mengamankan endpoint API dengan LOCAL_ACCESS_TOKEN. Masukkan token Anda untuk tersambung:
              </p>
              <input 
                type="password" 
                value={tokenInput} 
                onChange={(e) => setTokenInput(e.target.value)}
                placeholder="LOCAL_ACCESS_TOKEN" 
                className="w-full bg-[#F3EDF7] text-[#1D1B20] px-3 py-2 text-xs rounded-xl border border-[#CAC4D0] font-mono mb-3 focus:border-[#6750A4] outline-none"
              />
              <div className="flex justify-end gap-2">
                <button 
                  onClick={() => setShowTokenInput(false)}
                  className="px-2.5 py-1 text-xs font-semibold text-[#49454F] hover:text-[#1D1B20] rounded"
                >
                  Cancel
                </button>
                <button 
                  onClick={handleSaveToken}
                  className="bg-[#6750A4] hover:bg-[#523B8B] text-white px-3 py-1.5 text-xs rounded-full font-bold shadow-sm"
                >
                  Apply Token
                </button>
              </div>
            </div>
          )}
        </div>
        {/* About Dialog Button */}
        <button 
          onClick={() => alert("Tentang IntimClaw\n\nIntimClaw adalah aplikasi AI Agent lokal yang tangguh untuk mengelola VPS, SSH tunnel, dan bot automation secara aman.\n\nSitus Resmi: http://intim.my.id\nVersi: 1.2.0")}
          className="flex items-center gap-2 bg-white hover:bg-[#F3EDF7] text-[#1D1B20] hover:text-[#6750A4] px-3 py-1.5 rounded-xl border border-[#CAC4D0] text-xs transition-colors shadow-sm font-semibold"
        >
          <span>Tentang</span>
        </button>
      </div>

      {/* Manual Workspace Path Change Modal */}
      {showWorkModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 z-50 animate-fade-in">
          <div className="bg-white border border-[#CAC4D0] rounded-2xl max-w-md w-full p-5 shadow-2xl text-[#1D1B20]">
            <h3 className="text-sm font-bold text-[#1D192B] flex items-center gap-2 mb-2">
              <FolderOpen className="h-4 w-4 text-amber-600" /> Set Working Directory
            </h3>
            <p className="text-xs text-[#49454F] mb-4 leading-relaxed font-sans">
              Tentukan direktori kerja utama local agent Anda. File explorer, terminal command, dan file writes akan terkunci di folder ini demi keselamatan.
            </p>
            <input 
              type="text" 
              value={workspaceInput} 
              onChange={(e) => setWorkspaceInput(e.target.value)}
              placeholder="e.g. /data/data/com.termux/files/home" 
              className="w-full bg-[#F3EDF7] text-[#1D1B20] px-3 py-2 text-xs rounded-xl border border-[#CAC4D0] font-mono mb-4 focus:border-[#6750A4] outline-none"
            />
            <div className="flex justify-end gap-2">
              <button 
                onClick={() => setShowWorkModal(false)}
                className="px-3 py-1.5 text-xs text-[#49454F] hover:text-[#1D1B20] font-semibold"
              >
                Cancel
              </button>
              <button 
                onClick={handleUpdateWorkspace}
                className="bg-[#6750A4] hover:bg-[#523B8B] text-white px-4 py-1.5 text-xs rounded-full font-bold shadow-md"
              >
                Save Work Dir
              </button>
            </div>
          </div>
        </div>
      )}
    </header>
  );
};
