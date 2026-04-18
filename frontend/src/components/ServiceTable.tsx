import React, { useState } from 'react';
import { Play, Square, RefreshCw, CheckCircle, XCircle, Search, ToggleLeft, ToggleRight } from 'lucide-react';

interface Service {
  name: string;
  load_state: string;
  active_state: string;
  sub_state: string;
  description: string;
}

interface ServiceTableProps {
  services: Service[];
  onAction: (serviceName: string, action: string) => void;
}

const ServiceTable: React.FC<ServiceTableProps> = ({ services, onAction }) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [showSystemServices, setShowSystemServices] = useState(false);

  const isSystemService = (name: string) => {
    return (
      name.startsWith('systemd-') ||
      name.startsWith('dbus') ||
      name.startsWith('snap.') ||
      name.startsWith('getty') ||
      name.startsWith('polkit') ||
      name.startsWith('rsyslog')
    );
  };

  const filteredServices = services.filter((svc) => {
    const term = searchQuery.toLowerCase();
    const matchesSearch =
      svc.name.toLowerCase().includes(term) ||
      svc.description.toLowerCase().includes(term) ||
      svc.active_state.toLowerCase().includes(term);

    if (!matchesSearch) return false;
    if (!showSystemServices && isSystemService(svc.name)) return false;
    return true;
  });

  const handleAction = (serviceName: string, action: string) => {
    if (window.confirm(`Are you sure you want to ${action} the service "${serviceName}"?`)) {
      onAction(serviceName, action);
    }
  };

  return (
    <div className="process-table-container border border-slate-800 rounded-xl overflow-hidden flex flex-col bg-slate-900/50">
      <div className="table-toolbar flex items-center justify-between p-4 bg-slate-800/50 border-b border-slate-800">
        <div className="search-box flex items-center bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 focus-within:border-blue-500 w-64">
          <Search size={16} className="text-slate-500 mr-2" />
          <input
            type="text"
            placeholder="Search services..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="bg-transparent border-none outline-none text-sm w-full"
          />
        </div>
        <button
          onClick={() => setShowSystemServices(!showSystemServices)}
          className="flex items-center gap-2 text-sm text-slate-400 hover:text-slate-200 transition-colors"
        >
          {showSystemServices ? <ToggleRight size={20} className="text-blue-400" /> : <ToggleLeft size={20} />}
          Show System Services
        </button>
      </div>

      <div className="table-wrapper overflow-auto max-h-[400px]">
        <table className="enterprise-table w-full text-left border-collapse">
          <thead className="sticky top-0 bg-slate-900 z-10 shadow-md">
            <tr>
              <th className="py-3 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 w-[25%]">Service</th>
              <th className="py-3 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 w-[40%]">Description</th>
              <th className="py-3 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 w-[15%]">Status</th>
              <th className="py-3 px-4 text-xs font-bold uppercase tracking-wider text-slate-500 w-[20%] text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filteredServices.length > 0 ? (
              filteredServices.map((svc) => (
                <tr key={svc.name} className="hover:bg-slate-800/30 transition-colors">
                  <td className="py-3 px-4 font-medium text-sm w-[25%] truncate max-w-0" title={svc.name}>
                    {svc.name}
                  </td>
                  <td className="py-3 px-4 text-slate-400 text-xs w-[40%] truncate max-w-0" title={svc.description}>
                    {svc.description}
                  </td>
                  <td className="py-3 px-4 w-[15%]">
                    <span
                      className={`inline-flex items-center text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full border ${
                        svc.active_state === 'active'
                          ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_8px_rgba(16,185,129,0.2)]'
                          : svc.active_state === 'failed'
                          ? 'bg-red-500/10 text-red-400 border-red-500/30 shadow-[0_0_8px_rgba(239,68,68,0.3)] animate-pulse'
                          : 'bg-slate-500/10 text-slate-400 border-slate-500/30'
                      }`}
                    >
                      {svc.active_state}
                      {svc.sub_state !== 'running' && svc.sub_state !== 'dead' && ` (${svc.sub_state})`}
                    </span>
                  </td>
                  <td className="py-3 px-4 w-[20%] text-right">
                    <div className="flex items-center justify-end gap-1">
                      {svc.active_state !== 'active' ? (
                        <button
                          onClick={() => handleAction(svc.name, 'start')}
                          className="p-1.5 text-emerald-400 hover:bg-emerald-400/20 rounded-lg transition-colors"
                          title="Start Service"
                        >
                          <Play size={14} />
                        </button>
                      ) : (
                        <button
                          onClick={() => handleAction(svc.name, 'stop')}
                          className="p-1.5 text-slate-400 hover:bg-red-400/20 hover:text-red-400 rounded-lg transition-colors"
                          title="Stop Service"
                        >
                          <Square size={14} />
                        </button>
                      )}
                      <button
                        onClick={() => handleAction(svc.name, 'restart')}
                        className="p-1.5 text-orange-400 hover:bg-orange-400/20 rounded-lg transition-colors"
                        title="Restart Service"
                      >
                        <RefreshCw size={14} />
                      </button>
                      <button
                        onClick={() => handleAction(svc.name, 'enable')}
                        className="p-1.5 text-blue-400 hover:bg-blue-400/20 rounded-lg transition-colors"
                        title="Enable on Boot"
                      >
                        <CheckCircle size={14} />
                      </button>
                      <button
                        onClick={() => handleAction(svc.name, 'disable')}
                        className="p-1.5 text-slate-400 hover:bg-slate-400/20 hover:text-slate-200 rounded-lg transition-colors"
                        title="Disable on Boot"
                      >
                        <XCircle size={14} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            ) : (
              <tr>
                <td colSpan={4} className="py-8 text-center text-slate-500 text-sm italic">
                  {searchQuery ? `No services found matching "${searchQuery}"` : 'No services available'}
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default ServiceTable;
