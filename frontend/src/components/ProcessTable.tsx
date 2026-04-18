import React, { useState } from 'react';
import { ChevronDown, ChevronUp, RefreshCw, Search, Trash2 } from 'lucide-react';

interface ProcessInfo {
  pid: string;
  user: string;
  cpu: string;
  mem: string;
  command: string;
}

interface ProcessTableProps {
  processes: ProcessInfo[];
  onAction?: (pid: string, signal: string, value?: number) => Promise<void>;
  onSearch?: (query: string) => void;
  isSearching?: boolean;
}

const ProcessTable: React.FC<ProcessTableProps> = ({ processes, onAction, onSearch, isSearching }) => {
  const [sortField, setSortField] = useState<keyof ProcessInfo>('cpu');
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc');
  const [loadingPid, setLoadingPid] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const handleSort = (field: keyof ProcessInfo) => {
    if (sortField === field) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDir('desc');
    }
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setSearchQuery(val);
    if (onSearch) {
      onSearch(val);
    }
  };

  const handleActionClick = async (pid: string, signal: string, value?: number) => {
    if (!onAction) return;
    setLoadingPid(`${pid}-${signal}`);
    try {
      await onAction(pid, signal, value);
    } finally {
      setLoadingPid(null);
    }
  };

  const sortedProcesses = [...processes].sort((a, b) => {
    let valA = a[sortField];
    let valB = b[sortField];

    // Numeric sort for CPU, MEM, and PID
    if (['cpu', 'mem', 'pid'].includes(sortField)) {
      const numA = parseFloat(valA);
      const numB = parseFloat(valB);
      return sortDir === 'asc' ? numA - numB : numB - numA;
    }

    // String sort for User and Command
    if (sortDir === 'asc') {
      return valA.localeCompare(valB);
    } else {
      return valB.localeCompare(valA);
    }
  });

  const SortIcon = ({ field }: { field: keyof ProcessInfo }) => {
    if (sortField !== field) return null;
    return sortDir === 'asc' ? <ChevronUp size={14} /> : <ChevronDown size={14} />;
  };

  return (
    <div className="process-table-container">
      <div className="flex items-center justify-between mb-4">
        <div className="relative flex-1 max-w-md">
          <div className="absolute inset-y-0 left-3 flex items-center pointer-events-none text-slate-500">
            <Search size={16} />
          </div>
          <input
            type="text"
            className="w-full bg-slate-800 border border-slate-700 rounded-lg py-2 pl-10 pr-4 text-sm focus:outline-none focus:border-blue-500 transition-colors"
            placeholder="Search processes by name, PID, or user..."
            value={searchQuery}
            onChange={handleSearchChange}
          />
        </div>
        {isSearching && (
          <div className="text-xs text-blue-400 animate-pulse flex items-center gap-2">
            <RefreshCw size={12} className="spin" />
            Searching for "{searchQuery}"...
          </div>
        )}
      </div>

      <div className="overflow-x-auto rounded-xl border border-slate-800 bg-slate-900/50">
        <table className="data-table">
          <thead>
            <tr>
              <th onClick={() => handleSort('pid')} className="cursor-pointer">
                <div className="flex items-center gap-1">PID <SortIcon field="pid" /></div>
              </th>
              <th onClick={() => handleSort('user')} className="cursor-pointer">
                <div className="flex items-center gap-1">User <SortIcon field="user" /></div>
              </th>
              <th onClick={() => handleSort('cpu')} className="cursor-pointer">
                <div className="flex items-center gap-1 text-blue-400">CPU% <SortIcon field="cpu" /></div>
              </th>
              <th onClick={() => handleSort('mem')} className="cursor-pointer">
                <div className="flex items-center gap-1 text-emerald-400">MEM% <SortIcon field="mem" /></div>
              </th>
              <th onClick={() => handleSort('command')} className="cursor-pointer">
                <div className="flex items-center gap-1">Command <SortIcon field="command" /></div>
              </th>
              <th className="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {sortedProcesses.map((proc, idx) => (
              <tr key={`${proc.pid}-${idx}`} className="hover:bg-slate-800/30 transition-colors">
                <td className="font-mono text-slate-500">{proc.pid}</td>
                <td>{proc.user}</td>
                <td>
                  <span className={`px-2 py-0.5 rounded-md text-[10px] font-bold ${parseFloat(proc.cpu) > 10 ? 'bg-blue-500/10 text-blue-400 border border-blue-500/20' : 'text-slate-400'}`}>
                    {proc.cpu}%
                  </span>
                </td>
                <td>
                  <span className={`px-2 py-0.5 rounded-md text-[10px] font-bold ${parseFloat(proc.mem) > 10 ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'text-slate-400'}`}>
                    {proc.mem}%
                  </span>
                </td>
                <td className="max-w-[200px] truncate font-mono text-xs text-slate-300" title={proc.command}>
                  {proc.command}
                </td>
                <td className="text-right">
                  <div className="flex items-center justify-end gap-2">
                    <select 
                      className="bg-slate-800 border border-slate-700 text-[10px] rounded px-1 py-0.5 focus:outline-none focus:border-blue-500"
                      onChange={(e) => handleActionClick(proc.pid, 'NICE', parseInt(e.target.value))}
                      defaultValue="0"
                      disabled={loadingPid === `${proc.pid}-NICE`}
                    >
                      <option value="-10">High (-10)</option>
                      <option value="0">Normal (0)</option>
                      <option value="10">Low (10)</option>
                      <option value="19">Lowest (19)</option>
                    </select>
                    <button 
                      onClick={() => handleActionClick(proc.pid, 'TERM')}
                      disabled={loadingPid === `${proc.pid}-TERM`}
                      className="p-1 hover:bg-orange-500/20 text-orange-400 rounded transition-colors"
                      title="SIGTERM (Graceful)"
                    >
                      <RefreshCw size={14} className={loadingPid === `${proc.pid}-TERM` ? 'spin' : ''} />
                    </button>
                    <button 
                      onClick={() => handleActionClick(proc.pid, 'KILL')}
                      disabled={loadingPid === `${proc.pid}-KILL`}
                      className="p-1 hover:bg-red-500/20 text-red-400 rounded transition-colors"
                      title="SIGKILL (Forced)"
                    >
                      <Trash2 size={14} className={loadingPid === `${proc.pid}-KILL` ? 'spin' : ''} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {sortedProcesses.length === 0 && (
              <tr>
                <td colSpan={6} className="text-center py-12 text-slate-500 italic">
                  {searchQuery ? `No processes found matching "${searchQuery}"` : 'No process data available.'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default ProcessTable;
