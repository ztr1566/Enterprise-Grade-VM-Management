import React, { useEffect, useRef, useState } from 'react';
import { Terminal, X, Loader2, RefreshCw } from 'lucide-react';
import { getBackoffDelay } from '../utils/wsBackoff';
import ConnectionOverlay from './ConnectionOverlay';

interface Props {
  vmId: string;
  service: string;
  token: string;
  onClose: () => void;
}

const LogViewer: React.FC<Props> = ({ vmId, service, token, onClose }) => {
  const [logs, setLogs] = useState<string[]>([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [reconnectAttempt, setReconnectAttempt] = useState(0);
  const [isReconnecting, setIsReconnecting] = useState(false);

  const scrollRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let reconnectTimeout: any;

    const connect = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const host = window.location.hostname + ':8080';
      const url = `${protocol}//${host}/api/vms/${vmId}/logs/${service}?token=${token}`;

      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        setError(null);
        setReconnectAttempt(0);
        setIsReconnecting(false);
      };

      ws.onmessage = (event) => {
        setLogs(prev => [...prev, event.data].slice(-500));
      };

      ws.onerror = () => {
        // Error will be followed by onclose
      };

      ws.onclose = () => {
        setConnected(false);
        setIsReconnecting(true);
        const delay = getBackoffDelay(reconnectAttempt);
        reconnectTimeout = setTimeout(() => {
          setReconnectAttempt(prev => prev + 1);
        }, delay);
      };
    };

    connect();

    return () => {
      if (wsRef.current) wsRef.current.close();
      clearTimeout(reconnectTimeout);
    };
  }, [vmId, service, token, reconnectAttempt]);

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight;
    }
  }, [logs]);

  return (
    <div className="flex flex-col h-[400px] bg-slate-950 border border-slate-800 rounded-xl overflow-hidden shadow-2xl">
      {/* Header */}
      <div className="px-4 py-3 bg-slate-900 border-b border-slate-800 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-500/10 rounded-lg flex items-center justify-center">
            <Terminal className="w-4 h-4 text-blue-400" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-bold text-white tracking-tight">{service}.service</span>
              {connected ? (
                <span className="flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
                  <span className="text-[10px] text-emerald-500 font-bold uppercase tracking-widest">Live</span>
                </span>
              ) : error ? (
                <span className="text-[10px] text-red-500 font-bold uppercase tracking-widest">Error</span>
              ) : (
                <span className="text-[10px] text-slate-500 font-bold uppercase tracking-widest">Connecting...</span>
              )}
            </div>
            <p className="text-[10px] font-mono text-slate-500">journalctl -u {service} -f -n 100</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button 
            onClick={() => { setLogs([]); }}
            className="p-1.5 text-slate-500 hover:text-white hover:bg-slate-800 rounded-md transition-all"
            title="Clear Logs"
          >
            <RefreshCw className="w-4 h-4" />
          </button>
          <button 
            onClick={onClose} 
            className="p-1.5 text-slate-500 hover:text-white hover:bg-slate-800 rounded-md transition-all"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>
      
      {/* Log Body */}
      <div 
        ref={scrollRef}
        className="flex-1 p-4 font-mono text-[11px] leading-relaxed text-slate-300 overflow-y-auto bg-[#020617] scrollbar-thin scrollbar-thumb-slate-800"
      >
        {error ? (
          <div className="flex flex-col items-center justify-center h-full text-red-400 gap-2">
            <X className="w-8 h-8 opacity-20" />
            <p className="font-sans font-medium">{error}</p>
          </div>
        ) : logs.length === 0 && !connected ? (
          <div className="flex items-center justify-center h-full gap-3 text-slate-500 italic">
            <Loader2 className="w-4 h-4 animate-spin" />
            Establishing live log stream...
          </div>
        ) : logs.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-2 text-slate-600 italic">
            <Terminal className="w-8 h-8 opacity-10" />
            Waiting for new log entries...
          </div>
        ) : (
          <div className="space-y-0.5">
            {logs.map((log, i) => (
              <div key={i} className="group flex gap-3 hover:bg-slate-900/50 -mx-4 px-4 py-0.5 transition-colors">
                <span className="text-slate-700 select-none w-8 text-right shrink-0">{i + 1}</span>
                <span className="break-all whitespace-pre-wrap">{log}</span>
              </div>
            ))}
          </div>
        )}
      </div>

      <ConnectionOverlay 
        isConnecting={isReconnecting} 
        attempt={reconnectAttempt} 
      />
    </div>
  );
};

export default LogViewer;
