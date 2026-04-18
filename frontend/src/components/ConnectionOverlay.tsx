import React from 'react';
import { RefreshCcw, AlertCircle, WifiOff } from 'lucide-react';

interface Props {
  isConnecting: boolean;
  attempt: number;
  error?: string | null;
}

/**
 * ConnectionOverlay provides a non-intrusive floating status indicator
 * for active WebSocket reconnection attempts.
 */
const ConnectionOverlay: React.FC<Props> = ({ isConnecting, attempt, error }) => {
  if (!isConnecting && !error) return null;

  return (
    <div className="fixed bottom-6 right-6 z-[200] animate-in fade-in slide-in-from-bottom-4 duration-300">
      <div className={`
        flex items-center gap-4 px-5 py-4 rounded-2xl border shadow-2xl backdrop-blur-md
        ${error 
          ? 'bg-red-500/10 border-red-500/20 text-red-400' 
          : 'bg-slate-900/90 border-slate-800 text-slate-300'}
      `}>
        <div className={`
          w-10 h-10 rounded-full flex items-center justify-center
          ${error ? 'bg-red-500/20' : 'bg-blue-500/10'}
        `}>
          {error ? (
            <AlertCircle className="w-5 h-5 text-red-500" />
          ) : isConnecting ? (
            <RefreshCcw className="w-5 h-5 text-blue-400 animate-spin" />
          ) : (
            <WifiOff className="w-5 h-5 text-slate-500" />
          )}
        </div>
        
        <div>
          <h4 className="text-sm font-bold text-white tracking-tight">
            {error ? 'Connection Lost' : 'Reconnecting...'}
          </h4>
          <p className="text-[11px] font-medium opacity-70">
            {error 
              ? 'Attempting to restore link to host' 
              : `Retry attempt #${attempt}`}
          </p>
        </div>

        {isConnecting && !error && (
          <div className="flex gap-1 ml-2">
            {[1, 2, 3].map(i => (
              <div 
                key={i} 
                className="w-1 h-1 rounded-full bg-blue-400 animate-pulse" 
                style={{ animationDelay: `${i * 150}ms` }}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ConnectionOverlay;
