import React, { useEffect, useState } from 'react';
import api from '../services/api';
import { X, Activity, Server, Clock, HardDrive, Cpu, MemoryStick, AlertTriangle, List, Play, Square, RotateCw, Loader2, Power, RefreshCcw, FileText, ChevronDown } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import LogViewer from './LogViewer';
import {
  AreaChart, Area, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer
} from 'recharts';

interface VM {
  id: string;
  name: string;
  host: string;
}

interface VMStats {
  cpu: number;
  ram: number;
  disk: number;
  timestamp: string;
}

interface Props {
  vm: VM | null;
  onClose: () => void;
}

const VMDetails: React.FC<Props> = ({ vm, onClose }) => {
  const [history, setHistory] = useState<VMStats[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshRate, setRefreshRate] = useState<number>(5000);
  
  const { token } = useAuth();
  const [activeTab, setActiveTab] = useState<'metrics' | 'services' | 'logs'>('metrics');
  const [selectedServiceForLogs, setSelectedServiceForLogs] = useState<string | null>(null);
  const [services, setServices] = useState<{name: string, load: string, active: string, sub: string}[]>([]);
  const [servicesLoading, setServicesLoading] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const [powerAction, setPowerAction] = useState<'reboot' | 'shutdown' | null>(null);
  const [powerLoading, setPowerLoading] = useState(false);
  const [powerMessage, setPowerMessage] = useState<string | null>(null);

  useEffect(() => {
    if (!vm) return;

    let isSubscribed = true;

    const fetchStats = async () => {
      try {
        const response = await api.get(`/vms/${vm.id}/stats`);
        if (!isSubscribed) return;
        
        const newStats = response.data;
        setHistory(prev => {
          const updated = [...prev, newStats];
          return updated.slice(-20); // Maintain a buffer of only the last 20 data points
        });
        setError(null);
      } catch (err: any) {
        if (!isSubscribed) return;
        console.error('Failed to fetch VM stats', err);
        setError(err?.response?.data?.error || 'Failed to fetch metrics');
      } finally {
        if (isSubscribed) setLoading(false);
      }
    };

    // Initial fetch (only if we don't have history yet, to prevent jumping on refresh rate change)
    if (history.length === 0) {
      setLoading(true);
      fetchStats();
    }

    // Poll based on refreshRate
    const intervalId = setInterval(fetchStats, refreshRate);

    return () => {
      isSubscribed = false;
      clearInterval(intervalId);
    };
  }, [vm, refreshRate]);

  // Reset state when VM changes
  useEffect(() => {
    setHistory([]);
    setLoading(true);
    setError(null);
    setActiveTab('metrics');
    setActionError(null);
  }, [vm?.id]);

  const fetchServices = async () => {
    if (!vm) return;
    setServicesLoading(true);
    try {
      const response = await api.get(`/vms/${vm.id}/services`);
      setServices(response.data || []);
    } catch (err) {
      console.error('Failed to fetch services', err);
    } finally {
      setServicesLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'services') {
      fetchServices();
    }
  }, [activeTab, vm?.id]);

  const handleServiceAction = async (serviceName: string, action: string) => {
    setActionLoading(`${serviceName}-${action}`);
    setActionError(null);
    try {
      const res = await api.post(`/vms/${vm?.id}/services/${serviceName}/action`, { action });
      if (res.data.status === 'error') {
        setActionError(`Failed to ${action} ${serviceName}: ${res.data.error || res.data.output}`);
      } else {
        await fetchServices();
      }
    } catch (err: any) {
      console.error(`Failed to ${action} service ${serviceName}`, err);
      setActionError(`Failed to ${action} ${serviceName}: ` + (err?.response?.data?.error || err.message));
    } finally {
      setActionLoading(null);
    }
  };

  const handlePowerAction = async () => {
    if (!powerAction || !vm) return;
    setPowerLoading(true);
    try {
      await api.post(`/vms/${vm.id}/power`, { action: powerAction });
      setPowerMessage(`Command sent. Server is ${powerAction === 'reboot' ? 'restarting' : 'shutting down'}...`);
      
      // Auto-close after 3 seconds
      setTimeout(() => {
        onClose();
        setPowerAction(null);
        setPowerLoading(false);
        setPowerMessage(null);
      }, 3000);
    } catch (err: any) {
      console.error(`Power action ${powerAction} failed`, err);
      alert(`Power action failed: ${err?.response?.data?.error || err.message}`);
      setPowerAction(null);
      setPowerLoading(false);
    }
  };

  if (!vm) return null;

  const currentStats = history.length > 0 ? history[history.length - 1] : null;

  // Format timestamp for X-Axis
  const formatTime = (timeStr: string) => {
    const d = new Date(timeStr);
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}`;
  };

  const chartData = history.map(h => ({
    ...h,
    time: formatTime(h.timestamp)
  }));

  const MetricHeader = ({ title, value, icon, unit, colorClass }: { title: string, value: number, icon: React.ReactNode, unit: string, colorClass: string }) => (
    <div className="flex items-center justify-between mb-2">
      <div className="flex items-center gap-2 text-slate-400 font-medium">
        {icon}
        <span>{title}</span>
      </div>
      <div className={`text-xl font-bold ${colorClass} tracking-tight flex items-baseline gap-1`}>
        {value.toFixed(1)} <span className="text-sm opacity-70 font-normal">{unit}</span>
      </div>
    </div>
  );

  return (
    <div className="fixed inset-y-0 right-0 w-full max-w-xl bg-slate-900 border-l border-slate-800 shadow-2xl z-[60] flex flex-col transform transition-transform duration-300">
      <div className="p-6 border-b border-slate-800 flex items-center justify-between bg-slate-900/50 sticky top-0 z-10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center">
            <Activity className="text-blue-400 w-5 h-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">{vm.name}</h2>
            <p className="text-slate-400 text-sm font-mono">{vm.host}</p>
          </div>
        </div>
        <div className="flex items-center gap-4">
          <select 
            value={refreshRate}
            onChange={(e) => setRefreshRate(Number(e.target.value))}
            className="bg-slate-800 border border-slate-700 text-slate-300 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 block p-2"
          >
            <option value={1000}>1s</option>
            <option value={5000}>5s</option>
            <option value={10000}>10s</option>
            <option value={30000}>30s</option>
          </select>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors bg-slate-800 p-2 rounded-lg">
            <X className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Power Actions Bar */}
      <div className="px-6 py-3 bg-slate-800/30 border-b border-slate-800 flex items-center justify-between">
        <span className="text-xs font-bold text-slate-500 uppercase tracking-widest">Power Management</span>
        <div className="flex gap-2">
          <button 
            onClick={() => setPowerAction('reboot')}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-orange-500/10 hover:bg-orange-500/20 text-orange-400 border border-orange-500/20 rounded-lg text-xs font-semibold transition-all"
          >
            <RefreshCcw className="w-3.5 h-3.5" />
            Reboot
          </button>
          <button 
            onClick={() => setPowerAction('shutdown')}
            className="flex items-center gap-1.5 px-3 py-1.5 bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 rounded-lg text-xs font-semibold transition-all"
          >
            <Power className="w-3.5 h-3.5" />
            Shutdown
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex border-b border-slate-800 bg-slate-900 sticky z-10" style={{ top: '89px' }}>
        <button 
          onClick={() => setActiveTab('metrics')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 flex justify-center items-center gap-2 ${activeTab === 'metrics' ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-400 hover:text-slate-300'}`}
        >
          <Activity className="w-4 h-4" /> Metrics
        </button>
        <button 
          onClick={() => setActiveTab('services')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 flex justify-center items-center gap-2 ${activeTab === 'services' ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-400 hover:text-slate-300'}`}
        >
          <List className="w-4 h-4" /> Services
        </button>
        <button 
          onClick={() => setActiveTab('logs')}
          className={`flex-1 py-3 text-sm font-medium border-b-2 flex justify-center items-center gap-2 ${activeTab === 'logs' ? 'border-blue-500 text-blue-400' : 'border-transparent text-slate-400 hover:text-slate-300'}`}
        >
          <FileText className="w-4 h-4" /> Logs
        </button>
      </div>

      <div className="p-6 flex-1 overflow-y-auto">
        {activeTab === 'metrics' ? (
          <>
            <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <Server className="w-5 h-5 text-blue-400" />
              Real-time Metrics
            </h3>

            {loading && history.length === 0 ? (
              <div className="flex flex-col items-center justify-center py-20 text-slate-500">
                <Activity className="w-8 h-8 animate-pulse mb-4 text-blue-500/50" />
                <p>Connecting via SSH to gather metrics...</p>
              </div>
            ) : error && history.length === 0 ? (
              <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-5 text-red-400 text-sm">
                <span className="font-bold block mb-1">Metrics Error</span>
                {error}
              </div>
            ) : (
          <div className="space-y-6">
            {error && (
               <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-3 text-red-400 text-sm flex items-center gap-2">
                 <AlertTriangle className="w-4 h-4" />
                 <span>Connection lost. Retrying...</span>
               </div>
            )}
            
            {/* CPU Chart */}
            <div className="bg-slate-800/50 rounded-xl p-5 border border-slate-700/50">
              <MetricHeader title="CPU Usage" value={currentStats?.cpu || 0} unit="%" icon={<Cpu className="w-4 h-4 text-cyan-400" />} colorClass="text-cyan-400" />
              <div className="h-40 w-full mt-2">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorCpu" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#22d3ee" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#22d3ee" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={10} tickMargin={10} />
                    <YAxis stroke="#64748b" fontSize={10} domain={[0, 100]} />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '0.5rem' }}
                      itemStyle={{ color: '#22d3ee' }}
                    />
                    <Area type="monotone" dataKey="cpu" stroke="#22d3ee" strokeWidth={2} fillOpacity={1} fill="url(#colorCpu)" isAnimationActive={false} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* RAM Chart */}
            <div className="bg-slate-800/50 rounded-xl p-5 border border-slate-700/50">
              <MetricHeader title="Memory Usage" value={currentStats?.ram || 0} unit="%" icon={<MemoryStick className="w-4 h-4 text-purple-400" />} colorClass="text-purple-400" />
              <div className="h-40 w-full mt-2">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={chartData} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
                    <defs>
                      <linearGradient id="colorRam" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#a855f7" stopOpacity={0.3}/>
                        <stop offset="95%" stopColor="#a855f7" stopOpacity={0}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={10} tickMargin={10} />
                    <YAxis stroke="#64748b" fontSize={10} domain={[0, 100]} />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '0.5rem' }}
                      itemStyle={{ color: '#a855f7' }}
                    />
                    <Area type="monotone" dataKey="ram" stroke="#a855f7" strokeWidth={2} fillOpacity={1} fill="url(#colorRam)" isAnimationActive={false} />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Disk Chart */}
            <div className="bg-slate-800/50 rounded-xl p-5 border border-slate-700/50">
              <MetricHeader title="Disk Usage (/)" value={currentStats?.disk || 0} unit="%" icon={<HardDrive className="w-4 h-4 text-emerald-400" />} colorClass="text-emerald-400" />
              <div className="h-40 w-full mt-2">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={chartData} margin={{ top: 5, right: 0, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" vertical={false} />
                    <XAxis dataKey="time" stroke="#64748b" fontSize={10} tickMargin={10} />
                    <YAxis stroke="#64748b" fontSize={10} domain={[0, 100]} />
                    <Tooltip 
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '0.5rem' }}
                      itemStyle={{ color: '#34d399' }}
                    />
                    <Line type="stepAfter" dataKey="disk" stroke="#34d399" strokeWidth={2} dot={false} isAnimationActive={false} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
            
            <div className="mt-6 flex items-center justify-between text-xs text-slate-500">
              <div className="flex items-center gap-2">
                <Clock className="w-3.5 h-3.5" />
                <span>Last updated: {currentStats ? formatTime(currentStats.timestamp) : '...'}</span>
              </div>
              <div>Buffer: {history.length} / 20 points</div>
            </div>
          </div>
        )}
          </>
        ) : activeTab === 'services' ? (
          <>
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <List className="w-5 h-5 text-blue-400" />
                Core Services
              </h3>
              <button 
                onClick={fetchServices}
                disabled={servicesLoading}
                className="text-xs bg-slate-800 hover:bg-slate-700 text-slate-300 px-3 py-1.5 rounded-lg transition-colors flex items-center gap-2 disabled:opacity-50"
              >
                <RotateCw className={`w-3.5 h-3.5 ${servicesLoading ? 'animate-spin' : ''}`} />
                Refresh
              </button>
            </div>

            {servicesLoading && services.length === 0 ? (
              <div className="flex justify-center py-10">
                <Activity className="w-6 h-6 animate-pulse text-blue-500/50" />
              </div>
            ) : (
              <div className="space-y-3">
                {actionError && (
                  <div className="bg-red-500/10 border border-red-500/20 rounded-xl p-3 text-red-400 text-sm flex items-start gap-2 mb-4">
                    <AlertTriangle className="w-5 h-5 shrink-0 mt-0.5" />
                    <div className="flex-1 break-all whitespace-pre-wrap font-mono text-xs">{actionError}</div>
                    <button onClick={() => setActionError(null)} className="text-red-400/70 hover:text-red-400">
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                )}
                {services.map(svc => (
                  <div key={svc.name} className="bg-slate-800/50 border border-slate-700/50 rounded-xl p-4 flex items-center justify-between">
                    <div>
                      <h4 className="font-bold text-white mb-1">{svc.name}</h4>
                      <div className="flex gap-1.5 flex-wrap">
                        <span className={`px-2 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wider ${
                          (svc.active === 'active' || svc.active === 'running')
                            ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                            : (svc.active === 'failed' || svc.active === 'inactive')
                            ? 'bg-red-500/10 text-red-400 border border-red-500/20'
                            : 'bg-slate-700/50 text-slate-400 border border-slate-600/50'
                        }`}>
                          <span className={`inline-block w-1.5 h-1.5 rounded-full mr-1.5 ${
                            (svc.active === 'active' || svc.active === 'running') ? 'bg-emerald-400' : (svc.active === 'failed' || svc.active === 'inactive') ? 'bg-red-400' : 'bg-slate-500'
                          }`} />
                          {svc.active === 'failed' || svc.active === 'inactive' ? 'STOPPED' : svc.active}
                        </span>
                        {svc.sub && (
                          <span className="px-2 py-0.5 rounded-md text-[10px] font-mono text-slate-500 bg-slate-800 border border-slate-700">
                            {svc.sub}
                          </span>
                        )}
                      </div>
                    </div>
                    
                    <div className="flex gap-2">
                      <button 
                        onClick={() => { setSelectedServiceForLogs(svc.name); setActiveTab('logs'); }}
                        title="View Logs"
                        className="w-9 h-9 flex items-center justify-center bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-400 rounded-lg transition-colors"
                      >
                        <FileText className="w-4 h-4" />
                      </button>
                      <button 
                        onClick={() => handleServiceAction(svc.name, 'start')}
                        title="Start Service"
                        disabled={(svc.active === 'active' || svc.active === 'running') || actionLoading !== null}
                        className="w-9 h-9 flex items-center justify-center bg-slate-800 hover:bg-emerald-500/20 border border-slate-700 text-emerald-400 rounded-lg transition-colors disabled:opacity-30 disabled:hover:bg-slate-800"
                      >
                        {actionLoading === `${svc.name}-start` ? <Loader2 className="w-4 h-4 animate-spin" /> : <Play className="w-4 h-4 fill-current" />}
                      </button>
                      <button 
                        onClick={() => handleServiceAction(svc.name, 'restart')}
                        title="Restart Service"
                        disabled={actionLoading !== null}
                        className="w-9 h-9 flex items-center justify-center bg-slate-800 hover:bg-blue-500/20 border border-slate-700 text-blue-400 rounded-lg transition-colors disabled:opacity-30 disabled:hover:bg-slate-800"
                      >
                        {actionLoading === `${svc.name}-restart` ? <Loader2 className="w-4 h-4 animate-spin" /> : <RotateCw className="w-4 h-4" />}
                      </button>
                      <button 
                        onClick={() => handleServiceAction(svc.name, 'stop')}
                        title="Stop Service"
                        disabled={(svc.active !== 'active' && svc.active !== 'running') || actionLoading !== null}
                        className="w-9 h-9 flex items-center justify-center bg-slate-800 hover:bg-red-500/20 border border-slate-700 text-red-400 rounded-lg transition-colors disabled:opacity-30 disabled:hover:bg-slate-800"
                      >
                        {actionLoading === `${svc.name}-stop` ? <Loader2 className="w-4 h-4 animate-spin" /> : <Square className="w-4 h-4 fill-current" />}
                      </button>
                    </div>
                  </div>
                ))}
                
                {services.length === 0 && !servicesLoading && (
                  <p className="text-center text-slate-500 py-10 border-2 border-dashed border-slate-800 rounded-xl">
                    No core services detected.
                  </p>
                )}
              </div>
            )}
          </>
        ) : (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                <FileText className="w-5 h-5 text-blue-400" />
                Systemd Logs
              </h3>
              
              <div className="relative min-w-[200px]">
                <select
                  value={selectedServiceForLogs || ''}
                  onChange={(e) => setSelectedServiceForLogs(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 text-slate-200 text-xs rounded-lg px-3 py-2 appearance-none focus:ring-2 focus:ring-blue-500 outline-none"
                >
                  <option value="" disabled>Select a service...</option>
                  {services.map(s => (
                    <option key={s.name} value={s.name}>{s.name}</option>
                  ))}
                </select>
                <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-3 h-3 text-slate-500 pointer-events-none" />
              </div>
            </div>

            {selectedServiceForLogs ? (
              <LogViewer 
                vmId={vm.id}
                service={selectedServiceForLogs}
                token={token || ''}
                onClose={() => setSelectedServiceForLogs(null)}
              />
            ) : (
              <div className="flex flex-col items-center justify-center py-20 bg-slate-800/20 border-2 border-dashed border-slate-800 rounded-2xl text-slate-500">
                <FileText className="w-10 h-10 mb-4 opacity-20" />
                <p>Select a service above or from the Services tab to view live logs.</p>
              </div>
            )}
            
            <div className="bg-blue-500/5 border border-blue-500/10 rounded-xl p-4 flex items-start gap-3">
              <Activity className="w-5 h-5 text-blue-400 shrink-0 mt-0.5" />
              <div className="text-[11px] text-slate-400 leading-relaxed">
                <strong className="text-blue-300 block mb-1">About Log Observation</strong>
                Logs are streamed in real-time using <code className="text-blue-200">journalctl -f</code> over an agentless SSH pipe. 
                This allows you to diagnose start-up failures and runtime errors without needing a full terminal session.
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Power Confirmation Modal */}
      {powerAction && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl p-6 max-w-sm w-full shadow-2xl">
            {powerMessage ? (
              <div className="text-center py-4">
                <div className="w-12 h-12 bg-emerald-500/10 rounded-full flex items-center justify-center mx-auto mb-4">
                  <Activity className="w-6 h-6 text-emerald-500 animate-pulse" />
                </div>
                <p className="text-white font-medium">{powerMessage}</p>
              </div>
            ) : (
              <>
                <h3 className="text-xl font-bold text-white mb-2 capitalize">{powerAction} Server</h3>
                <p className="text-slate-400 mb-6">
                  Are you sure you want to <span className="text-white font-semibold">{powerAction}</span> this server? All active connections will be lost.
                </p>
                <div className="flex gap-3">
                  <button 
                    disabled={powerLoading}
                    onClick={() => setPowerAction(null)}
                    className="flex-1 px-4 py-2 bg-slate-800 hover:bg-slate-700 text-slate-300 rounded-xl transition-colors disabled:opacity-50"
                  >
                    Cancel
                  </button>
                  <button 
                    disabled={powerLoading}
                    onClick={handlePowerAction}
                    className={`flex-1 px-4 py-2 rounded-xl text-white font-bold transition-all flex items-center justify-center gap-2 ${
                      powerAction === 'shutdown' ? 'bg-red-600 hover:bg-red-500' : 'bg-orange-600 hover:bg-orange-500'
                    } disabled:opacity-50`}
                  >
                    {powerLoading ? <Loader2 className="w-4 h-4 animate-spin" /> : 'Confirm'}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default VMDetails;
