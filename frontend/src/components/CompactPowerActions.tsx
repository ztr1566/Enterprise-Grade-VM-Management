import React, { useState, useRef, useEffect } from 'react';
import { Power, RefreshCw, Moon } from 'lucide-react';

interface CompactPowerActionsProps {
  vmId: string;
  onPowerAction: (vmId: string, action: string) => void;
}

const CompactPowerActions: React.FC<CompactPowerActionsProps> = ({ vmId, onPowerAction }) => {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    if (open) document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, [open]);

  return (
    <div className="relative" ref={ref}>
      <button
        onClick={() => setOpen(!open)}
        className="w-10 h-10 bg-slate-800 hover:bg-slate-700 border border-slate-700 hover:border-slate-600 text-slate-400 hover:text-white rounded-xl transition-all flex items-center justify-center"
        title="Power Options"
      >
        <Power className="w-4 h-4" />
      </button>
      {open && (
        <div
          className="compact-power-dropdown"
          style={{
            position: 'absolute',
            bottom: '100%',
            right: 0,
            marginBottom: '0.5rem',
            width: '11rem',
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '12px',
            boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5)',
            zIndex: 200,
            overflow: 'hidden',
          }}
        >
          <div className="px-3 py-2 border-b border-slate-800 bg-slate-800/50">
            <span className="text-[10px] font-bold text-slate-500 uppercase">Power Options</span>
          </div>
          <button
            onClick={() => { onPowerAction(vmId, 'reboot'); setOpen(false); }}
            className="w-full text-left px-3 py-2 text-sm text-blue-400 hover:bg-blue-500/10 transition-colors flex items-center gap-2"
          >
            <RefreshCw size={14} />
            Reboot
          </button>
          <button
            onClick={() => { onPowerAction(vmId, 'sleep'); setOpen(false); }}
            className="w-full text-left px-3 py-2 text-sm text-amber-400 hover:bg-amber-500/10 transition-colors flex items-center gap-2"
          >
            <Moon size={14} />
            Sleep
          </button>
          <button
            onClick={() => { onPowerAction(vmId, 'shutdown'); setOpen(false); }}
            className="w-full text-left px-3 py-2 text-sm text-red-400 hover:bg-red-500/10 transition-colors flex items-center gap-2"
          >
            <Power size={14} />
            Shutdown
          </button>
        </div>
      )}
    </div>
  );
};

export default CompactPowerActions;
