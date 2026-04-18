import React, { useEffect, useRef, useState } from 'react';
import { Terminal } from 'xterm';
import { FitAddon } from '@xterm/addon-fit';
import 'xterm/css/xterm.css';
import { getBackoffDelay } from '../utils/wsBackoff';
import ConnectionOverlay from './ConnectionOverlay';

interface TerminalViewProps {
  vmId: string;
  vmName?: string;
  token: string;
  loginAs?: string;
  onClose: () => void;
}


const TerminalView: React.FC<TerminalViewProps> = ({ vmId, vmName, token, loginAs, onClose }) => {
  // wrapperRef is the CONSTRAINED container — position:absolute, inset:0, overflow:hidden
  // xterm mounts into this, measures it, and will NEVER cause it to grow.
  const wrapperRef = useRef<HTMLDivElement>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const [handshakeDone, setHandshakeDone] = useState(false);
  const [reconnectAttempt, setReconnectAttempt] = useState(0);
  const [isReconnecting, setIsReconnecting] = useState(false);

  useEffect(() => {
    if (!wrapperRef.current) return;
    let aborted = false;
    let reconnectTimeout: any;

    // ── 1. Initialise xterm ──────────────────────────────────────────────────
    const term = new Terminal({
      cursorBlink: false, // activated after handshake
      fontSize: 14,
      fontFamily: '"JetBrains Mono", "Cascadia Code", "Fira Code", monospace',
      scrollback: 5000,
      theme: {
        background: '#0a0e1a',
        foreground: '#e2e8f0',
        cursor: '#3b82f6',
        selectionBackground: '#3b82f640',
        black: '#1e293b',   brightBlack: '#334155',
        blue: '#3b82f6',    brightBlue: '#60a5fa',
        cyan: '#06b6d4',    brightCyan: '#22d3ee',
        green: '#10b981',   brightGreen: '#34d399',
        red: '#ef4444',     brightRed: '#f87171',
        white: '#e2e8f0',   brightWhite: '#f8fafc',
        yellow: '#f59e0b',  brightYellow: '#fbbf24',
        magenta: '#8b5cf6', brightMagenta: '#a78bfa',
      },
    });

    const fitAddon = new FitAddon();
    fitAddonRef.current = fitAddon;
    term.loadAddon(fitAddon);
    // Mount xterm into the CONSTRAINED wrapper (overflow:hidden, absolute positioned)
    term.open(wrapperRef.current);

    // ── 2. Fit BEFORE opening WebSocket (so initial cols/rows are accurate) ──
    fitAddon.fit();
    const initialCols = term.cols;
    const initialRows = term.rows;

    // ── 3. Open WebSocket ────────────────────────────────────────────────────
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    let url = `${protocol}//${host}/api/vms/${vmId}/terminal?token=${token}`;
    if (loginAs) {
      url += `&login_as=${loginAs}`;
    }
    const ws = new WebSocket(url);
    ws.binaryType = 'arraybuffer';
    wsRef.current = ws;

    ws.onopen = () => {
      if (aborted) { ws.close(); return; }
      setHandshakeDone(true);
      setIsReconnecting(false);
      setReconnectAttempt(0);
      term.options.cursorBlink = true;
      term.focus();

      // First message = handshake: tell the backend exactly how big to create the PTY
      ws.send(JSON.stringify({ type: 'resize', cols: initialCols, rows: initialRows }));

      term.onData((data) => {
        if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'input', data }));
      });

      term.onResize(({ cols, rows }) => {
        if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ type: 'resize', cols, rows }));
      });
    };

    ws.onmessage = (event) => {
      if (!handshakeDone) {
        setHandshakeDone(true);
        term.options.cursorBlink = true;
        term.focus();
      }
      if (event.data instanceof ArrayBuffer) {
        term.write(new Uint8Array(event.data));
      } else {
        term.write(event.data as string);
      }
    };

    ws.onerror = () => {};
    ws.onclose = () => { 
      if (aborted) return;
      setHandshakeDone(false);
      setIsReconnecting(true);
      term.writeln('\r\n\x1b[33m[INFO] Connection lost. Attempting to reconnect...\x1b[0m');
      
      const delay = getBackoffDelay(reconnectAttempt);
      reconnectTimeout = setTimeout(() => {
        setReconnectAttempt(prev => prev + 1);
      }, delay);
    };

    // ── 4. Debounced fit via ResizeObserver ──────────────────────────────────
    // ResizeObserver watches the WRAPPER element directly (more precise than window resize).
    // Debounced 100ms to prevent rapid-fire calls during drag resizing.
    let debounceTimer: ReturnType<typeof setTimeout> | null = null;
    const observer = new ResizeObserver(() => {
      if (debounceTimer) clearTimeout(debounceTimer);
      debounceTimer = setTimeout(() => {
        // fitAddon.fit() FIRST — recalculates cols/rows from container dimensions
        // then term.onResize fires automatically, which sends the JSON resize message
        fitAddon.fit();
      }, 100);
    });
    observer.observe(wrapperRef.current);

    return () => {
      aborted = true;
      if (debounceTimer) clearTimeout(debounceTimer);
      clearTimeout(reconnectTimeout);
      observer.disconnect();
      ws.close();
      term.dispose();
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [vmId, token, reconnectAttempt]);

  return (
    // ── Outer overlay ────────────────────────────────────────────────────────
    <div className="fixed inset-0 z-50 flex items-center justify-center p-6 bg-slate-950/90 backdrop-blur-sm">

      {/* ── Modal shell — flex column, max-size, does NOT grow beyond viewport ── */}
      <div
        className="w-full max-w-5xl bg-[#0a0e1a] border border-slate-700 rounded-2xl shadow-2xl shadow-blue-500/10 overflow-hidden flex flex-col"
        style={{ height: 'min(80vh, 680px)' }}
      >
        {/* Title bar — fixed height */}
        <div
          className="flex-none flex items-center justify-between px-5 py-3 bg-slate-900 border-b border-slate-700 select-none"
          style={{ height: '44px' }}
        >
          <div className="flex items-center gap-3">
            <div className="flex gap-1.5">
              <button
                onClick={() => { wsRef.current?.close(); onClose(); }}
                className="w-3 h-3 rounded-full bg-red-500 hover:bg-red-400 transition-colors"
                aria-label="Close"
              />
              <div className="w-3 h-3 rounded-full bg-yellow-500/60" />
              <button
                className="w-3 h-3 rounded-full bg-green-500/60 hover:bg-green-400 transition-colors"
                onClick={() => fitAddonRef.current?.fit()}
                title="Re-fit terminal"
              />
            </div>
            <span className="text-slate-400 text-xs font-mono tracking-wide">
              {vmName ? `${vmName} — ` : ''}ssh@{vmId.slice(0, 8)}
            </span>
          </div>
          <div className="flex items-center gap-3">
            {!handshakeDone && (
              <span className="text-slate-500 text-xs font-mono animate-pulse">connecting…</span>
            )}
            <span className="text-slate-600 text-xs font-mono">bash</span>
          </div>
        </div>

        {/*
          ── Terminal area — flex:1 fills remaining modal height ─────────────
          The outer div (flex:1, position:relative) constrains the space.
          The inner div (position:absolute, inset:0) is what xterm mounts into.
          overflow:hidden on the inner prevents xterm from escaping its bounds.
          This breaks the feedback loop: xterm measures inner (fixed size),
          renders into it, and can never cause the outer to grow.
        */}
        <div className="flex-1 relative" style={{ minHeight: 0 }}>
          <div
            ref={wrapperRef}
            style={{
              position: 'absolute',
              inset: 0,
              padding: '10px',
              overflow: 'hidden',     /* CRITICAL: xterm cannot expand this */
              backgroundColor: '#0a0e1a',
            }}
          />
        </div>
      </div>
      
      <ConnectionOverlay 
        isConnecting={isReconnecting} 
        attempt={reconnectAttempt} 
      />
    </div>
  );
};

export default TerminalView;
