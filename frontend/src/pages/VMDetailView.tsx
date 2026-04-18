import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { ArrowLeft, Cpu, HardDrive, RefreshCw, MemoryStick, Trash2 } from 'lucide-react';
import api from '../services/api';
import NetworkChart from '../components/NetworkChart';
import ProcessTable from '../components/ProcessTable';
import ServiceTable from '../components/ServiceTable';
import SecurityPanel from '../components/SecurityPanel';
import FullPowerActions from '../components/FullPowerActions';

interface VM {
  id: string;
  name: string;
  host: string;
  management_username: string;
  status: string;
  auth_type: string;
  provisioning_status: string;
}

interface VMStats {
  cpu: number;
  ram: number;
  disk: number;
  network?: {
    rx_bytes_sec: number;
    tx_bytes_sec: number;
    active_connections: number;
  };
  processes?: Array<{
    pid: string;
    user: string;
    cpu: string;
    mem: string;
    command: string;
  }>;
  timestamp: string;
}

const VMDetailView: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [vm, setVm] = useState<VM | null>(null);
  const [stats, setStats] = useState<VMStats | null>(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [loadingStats, setLoadingStats] = useState(false);
  const [statsError, setStatsError] = useState<string | null>(null);
  const [networkHistory, setNetworkHistory] = useState<any[]>([]);
  const [processQuery, setProcessQuery] = useState('');
  
  const [services, setServices] = useState<any[]>([]);
  const [loadingServices, setLoadingServices] = useState(false);
  const [servicesError, setServicesError] = useState<string | null>(null);



  const [security, setSecurity] = useState<any>(null);
  const [loadingSecurity, setLoadingSecurity] = useState(false);

  useEffect(() => {
    const fetchVM = async () => {
      try {
        const response = await api.get(`/vms/${id}`);
        setVm(response.data);
      } catch (err) {
        console.error('Failed to fetch VM', err);
      }
    };
    fetchVM();
  }, [id]);

  const fetchStats = async () => {
    if (!id) return;
    setLoadingStats(true);
    try {
      const response = await api.get(`/vms/${id}/stats?q=${encodeURIComponent(processQuery)}`);
      setStats(response.data);
      setStatsError(null);
    } catch (err: any) {
      console.error('Failed to fetch VM stats', err);
      if (err.response && err.response.data && err.response.data.message) {
        setStatsError(err.response.data.message);
      } else {
        setStatsError(err.message || 'Failed to fetch stats');
      }
    } finally {
      setLoadingStats(false);
    }
  };

  const fetchServices = async () => {
    if (!id) return;
    setLoadingServices(true);
    try {
      const response = await api.get(`/vms/${id}/services`);
      setServices(response.data);
      setServicesError(null);
    } catch (err: any) {
      console.error('Failed to fetch services', err);
      if (err.response && err.response.data && err.response.data.message) {
        setServicesError(err.response.data.message);
      } else {
        setServicesError(err.message || 'Failed to fetch services');
      }
    } finally {
      setLoadingServices(false);
    }
  };

  useEffect(() => {
    if (stats?.network) {
      setNetworkHistory(prev => {
        const newPoint = {
          time: new Date(stats.timestamp).toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' }),
          rx: stats.network!.rx_bytes_sec,
          tx: stats.network!.tx_bytes_sec
        };
        const updated = [...prev, newPoint].slice(-30); // Keep last 30 samples (2.5 mins)
        return updated;
      });
    }
  }, [stats]);

  const fetchSecurity = async () => {
    if (!id) return;
    setLoadingSecurity(true);
    try {
      const response = await api.get(`/vms/${id}/security`);
      setSecurity(response.data);
    } catch (err) {
      console.error('Failed to fetch security status', err);
    } finally {
      setLoadingSecurity(false);
    }
  };

  const handleAddFirewallRule = async (port: string, protocol: string) => {
    try {
      await api.post(`/vms/${id}/firewall/rules`, { action: 'add', port, protocol });
      fetchSecurity();
    } catch (err) {
      console.error('Failed to add firewall rule', err);
      alert('Failed to add firewall rule.');
    }
  };

  const handleRemoveFirewallRule = async (port: string, protocol: string) => {
    if (window.confirm(`Remove firewall rule for port ${port}/${protocol}?`)) {
      try {
        await api.post(`/vms/${id}/firewall/rules`, { action: 'remove', port, protocol });
        fetchSecurity();
      } catch (err) {
        console.error('Failed to remove firewall rule', err);
        alert('Failed to remove firewall rule.');
      }
    }
  };

  const handleSetSELinuxMode = async (mode: string) => {
    if (window.confirm(`Set SELinux to ${mode} mode?`)) {
      try {
        await api.post(`/vms/${id}/selinux`, { mode });
        fetchSecurity();
      } catch (err) {
        console.error('Failed to set SELinux mode', err);
        alert('Failed to set SELinux mode.');
      }
    }
  };

  const handleInstallSecurityComponent = async (component: string) => {
    try {
      await api.post(`/vms/${id}/security/install`, { component });
      fetchSecurity();
    } catch (err) {
      console.error('Failed to install security component', err);
      alert('Failed to install security component. Check logs.');
    }
  };

  useEffect(() => {
    let interval: any;
    if ((activeTab === 'metrics' || activeTab === 'network' || activeTab === 'processes') && id) {
      fetchStats();
      interval = setInterval(fetchStats, 2000); // 2s polling for high precision (Phase 6.5)
    } else if (activeTab === 'services' && id) {
      fetchServices();
      interval = setInterval(fetchServices, 5000); // 5s polling for services
    } else if (activeTab === 'security' && id) {
      fetchSecurity();
      interval = setInterval(fetchSecurity, 10000); // 10s polling for security
    }
    return () => clearInterval(interval);
  }, [activeTab, id, processQuery]);

  const handleProcessAction = async (pid: string, signal: string, value?: number) => {
    try {
      await api.post(`/vms/${id}/processes/${pid}/signal`, { signal, value });
      fetchStats();
    } catch (err) {
      console.error('Failed to send signal', err);
      alert('Failed to send signal. Check logs.');
    }
  };

  const handleServiceAction = async (serviceName: string, action: string) => {
    try {
      await api.post(`/vms/${id}/services/${serviceName}/action`, { action });
      fetchServices();
    } catch (err) {
      console.error('Failed to perform service action', err);
      alert('Failed to perform service action. Check logs.');
    }
  };

  const handlePowerAction = async (action: string) => {
    if (window.confirm(`Are you sure you want to ${action} this VM?`)) {
      try {
        await api.post(`/vms/${id}/power`, { action });
      } catch (err) {
        console.error('Power action failed', err);
        alert('Power action failed. Check console.');
      }
    }
  };

  const handleProcessSearch = (query: string) => {
    setProcessQuery(query);
  };

  const tabs = ['overview', 'metrics', 'access', 'services', 'security', 'network', 'processes'];

  return (
    <div className="page-enter">
      <header className="vmdetail-header">
        <Link to="/vms" className="vmdetail-back">
          <ArrowLeft size={18} />
        </Link>
        <div className="flex-1">
          <h1 className="vmdetail-title">{vm?.name ?? 'Loading...'}</h1>
          <p className="vmdetail-host">{vm?.host ?? ''}</p>
        </div>
        
        <div className="flex items-center gap-3">
          {vm && (
            <span className={`status-badge status-badge--${vm.status}`}>
              <span className="status-dot" />
              {vm.status}
            </span>
          )}

          <FullPowerActions onPowerAction={handlePowerAction} />
        </div>
      </header>

      <div className="card">
        <div className="tab-bar">
          {tabs.map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`tab-item ${activeTab === tab ? 'tab-item--active' : ''}`}
            >
              {tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </div>

        <div className="card-body">
          {activeTab === 'overview' && vm && (
            <div className="vmdetail-overview">
              <div className="vmdetail-field">
                <span className="vmdetail-field-label">ID</span>
                <span className="vmdetail-field-value font-mono">{vm.id}</span>
              </div>
              <div className="vmdetail-field">
                <span className="vmdetail-field-label">Host</span>
                <span className="vmdetail-field-value">{vm.host}</span>
              </div>
              <div className="vmdetail-field">
                <span className="vmdetail-field-label">Management User</span>
                <span className="vmdetail-field-value">{vm.management_username}</span>
              </div>
              <div className="vmdetail-field">
                <span className="vmdetail-field-label">Auth Type</span>
                <span className="vmdetail-field-value">{vm.auth_type}</span>
              </div>
              <div className="vmdetail-field">
                <span className="vmdetail-field-label">Provisioning</span>
                <span className={`prov-badge prov-badge--${vm.provisioning_status}`}>
                  {vm.provisioning_status}
                </span>
              </div>
            </div>
          )}

          {activeTab === 'metrics' && (
            <div className="metrics-panel">
              <div className="metrics-toolbar">
                <button onClick={fetchStats} disabled={loadingStats} className="btn-secondary">
                  <RefreshCw size={14} className={loadingStats ? 'spin' : ''} />
                  Refresh
                </button>
              </div>
              
              {statsError && (
                <div className="dashboard-alert dashboard-alert--error mb-4">
                  <span><strong>Error fetching metrics:</strong> {statsError}</span>
                </div>
              )}

              {stats ? (
                <div className="metrics-grid">
                  <div className="metric-card metric-card--blue">
                    <Cpu size={24} />
                    <div>
                      <p className="metric-label">CPU Usage</p>
                      <p className="metric-value">{stats.cpu.toFixed(1)}%</p>
                    </div>
                    <div className="metric-bar">
                      <div className="metric-bar-fill metric-bar-fill--blue" style={{ width: `${stats.cpu}%` }} />
                    </div>
                  </div>
                  <div className="metric-card metric-card--violet">
                    <MemoryStick size={24} />
                    <div>
                      <p className="metric-label">RAM Usage</p>
                      <p className="metric-value">{stats.ram.toFixed(1)}%</p>
                    </div>
                    <div className="metric-bar">
                      <div className="metric-bar-fill metric-bar-fill--violet" style={{ width: `${stats.ram}%` }} />
                    </div>
                  </div>
                  <div className="metric-card metric-card--amber">
                    <HardDrive size={24} />
                    <div>
                      <p className="metric-label">Disk Usage</p>
                      <p className="metric-value">{stats.disk.toFixed(1)}%</p>
                    </div>
                    <div className="metric-bar">
                      <div className="metric-bar-fill metric-bar-fill--amber" style={{ width: `${stats.disk}%` }} />
                    </div>
                  </div>
                </div>
              ) : (
                !statsError && (
                  <p className="text-muted">
                    {loadingStats ? 'Fetching metrics via SSH...' : 'Click Refresh to fetch live metrics.'}
                  </p>
                )
              )}
            </div>
          )}

          {activeTab === 'access' && (
            <AccessManager vmId={id!} />
          )}

          {activeTab === 'services' && (
            <div className="services-panel">
              <div className="metrics-toolbar mb-4">
                <button onClick={fetchServices} disabled={loadingServices} className="btn-secondary">
                  <RefreshCw size={14} className={loadingServices ? 'spin' : ''} />
                  Refresh
                </button>
              </div>
              {servicesError && (
                <div className="dashboard-alert dashboard-alert--error mb-4">
                  <span><strong>Error fetching services:</strong> {servicesError}</span>
                </div>
              )}
              <ServiceTable services={services} onAction={handleServiceAction} />
            </div>
          )}
          {activeTab === 'security' && (
            <SecurityPanel
              security={security}
              loading={loadingSecurity}
              onRefresh={fetchSecurity}
              onAddRule={handleAddFirewallRule}
              onRemoveRule={handleRemoveFirewallRule}
              onSetSELinuxMode={handleSetSELinuxMode}
              onInstallComponent={handleInstallSecurityComponent}
            />
          )}
          {activeTab === 'network' && (
            <div className="network-panel">
              <div className="metrics-toolbar mb-4">
                <button onClick={fetchStats} disabled={loadingStats} className="btn-secondary">
                  <RefreshCw size={14} className={loadingStats ? 'spin' : ''} />
                  Refresh
                </button>
              </div>

              {statsError && (
                <div className="dashboard-alert dashboard-alert--error mb-4">
                  <span><strong>Error fetching metrics:</strong> {statsError}</span>
                </div>
              )}

              {networkHistory.length > 0 ? (
                <NetworkChart 
                  data={networkHistory} 
                  activeConnections={stats?.network?.active_connections || 0} 
                />
              ) : (
                <div className="h-[300px] flex items-center justify-center border border-dashed border-slate-800 rounded-xl">
                  <p className="text-sm text-slate-500 italic">
                    {loadingStats ? 'Initializing network probe...' : 'No network data available. Click refresh to start.'}
                  </p>
                </div>
              )}
            </div>
          )}
          {activeTab === 'processes' && (
            <div className="process-explorer">
              <div className="metrics-toolbar mb-4">
                <button onClick={fetchStats} disabled={loadingStats} className="btn-secondary">
                  <RefreshCw size={14} className={loadingStats ? 'spin' : ''} />
                  Refresh
                </button>
              </div>

              {statsError && (
                <div className="dashboard-alert dashboard-alert--error mb-4">
                  <span><strong>Error fetching metrics:</strong> {statsError}</span>
                </div>
              )}

              <ProcessTable 
                processes={stats?.processes || []} 
                onAction={handleProcessAction} 
                onSearch={handleProcessSearch}
                isSearching={!!processQuery}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

interface AccessManagerProps {
  vmId: string;
}

const AccessManager: React.FC<AccessManagerProps> = ({ vmId }) => {
  const [users, setUsers] = useState<string[]>([]);
  const [newUser, setNewUser] = useState('');
  const [loading, setLoading] = useState(false);

  const fetchUsers = async () => {
    try {
      const response = await api.get(`/vms/${vmId}/users`);
      setUsers(response.data || []);
    } catch (err) {
      console.error('Failed to fetch VM users', err);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, [vmId]);

  const handleAddUser = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newUser.trim()) return;
    setLoading(true);
    try {
      await api.post(`/vms/${vmId}/users`, { username: newUser.trim() });
      setNewUser('');
      fetchUsers();
    } catch (err) {
      console.error('Failed to add user', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRevokeUser = async (username: string) => {
    if (!window.confirm(`Are you sure you want to revoke OS access for user "${username}"?`)) {
      return;
    }
    
    try {
      await api.delete(`/vms/${vmId}/access/${username}`);
      fetchUsers();
    } catch (err) {
      console.error('Failed to revoke user access', err);
      alert('Failed to revoke user access. Check logs for details.');
    }
  };

  return (
    <div className="access-manager">
      <h3 className="text-lg font-bold mb-4">OS Access Control</h3>
      <p className="text-sm text-slate-500 mb-6">
        Manage OS-level users that are authorized to access this VM. 
        Adding a user will automatically trigger a provisioning task to configure sudoers.
      </p>

      <form onSubmit={handleAddUser} className="flex gap-2 mb-8">
        <input
          type="text"
          value={newUser}
          onChange={(e) => setNewUser(e.target.value)}
          placeholder="Enter OS username (e.g. ztr)"
          className="flex-1 bg-slate-800 border border-slate-700 rounded-lg px-4 py-2 text-sm focus:outline-none focus:border-blue-500"
        />
        <button
          type="submit"
          disabled={loading || !newUser.trim()}
          className="btn-secondary bg-blue-600/10 text-blue-400 border-blue-500/20 hover:bg-blue-600/20 disabled:opacity-50"
        >
          {loading ? 'Adding...' : 'Add User'}
        </button>
      </form>

      <div className="space-y-2">
        <h4 className="text-xs font-bold uppercase tracking-wider text-slate-600 mb-2">Authorized Users</h4>
        {users.length > 0 ? (
          users.map((user) => (
            <div key={user} className="flex items-center justify-between p-3 bg-slate-800/50 rounded-xl border border-slate-800">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 rounded-lg bg-slate-800 flex items-center justify-center">
                  <span className="text-xs font-bold text-slate-400">{user.charAt(0).toUpperCase()}</span>
                </div>
                <span className="text-sm font-medium">{user}</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="text-[10px] px-2 py-0.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 rounded-md">
                  Active
                </span>
                <button
                  onClick={() => handleRevokeUser(user)}
                  className="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-400/10 rounded-lg transition-colors"
                  title="Revoke Access"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))
        ) : (
          <p className="text-sm text-slate-600 italic">No additional users configured.</p>
        )}
      </div>
    </div>
  );
};

export default VMDetailView;
