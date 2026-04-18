import React, { useState, useRef, useEffect } from 'react';
import { Power, RefreshCw, Moon, ChevronDown } from 'lucide-react';

interface FullPowerActionsProps {
  onPowerAction: (action: string) => void;
}

const FullPowerActions: React.FC<FullPowerActionsProps> = ({ onPowerAction }) => {
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
        className="flex items-center gap-2 px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 hover:border-slate-600 text-slate-300 hover:text-white rounded-xl transition-all"
      >
        <Power className="w-4 h-4" />
        <span className="text-sm font-semibold">Power Actions</span>
        <ChevronDown className={`w-4 h-4 transition-transform ${open ? 'rotate-180' : ''}`} />
      </button>
      {open && (
        <div
          style={{
            position: 'absolute',
            top: '100%',
            right: 0,
            marginTop: '0.5rem',
            width: '12rem',
            background: 'var(--bg-surface)',
            border: '1px solid var(--border-subtle)',
            borderRadius: '12px',
            boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5)',
            zIndex: 200,
            overflow: 'hidden',
          }}
        >
          <button
            onClick={() => { onPowerAction('reboot'); setOpen(false); }}
            className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-blue-400 hover:bg-blue-400/10 transition-colors"
          >
            <RefreshCw size={16} />
            Reboot VM
          </button>
          <button
            onClick={() => { onPowerAction('shutdown'); setOpen(false); }}
            className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-red-400 hover:bg-red-400/10 transition-colors"
          >
            <Power size={16} />
            Shutdown VM
          </button>
          <button
            onClick={() => { onPowerAction('sleep'); setOpen(false); }}
            className="w-full flex items-center gap-3 px-4 py-2.5 text-sm text-amber-400 hover:bg-amber-400/10 transition-colors"
          >
            <Moon size={16} />
            Sleep Mode
          </button>
        </div>
      )}
    </div>
  );
};

export default FullPowerActions;
