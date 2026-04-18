import React, { useState } from 'react';
import { Shield, ShieldAlert, ShieldCheck, Plus, Trash2, AlertTriangle, RefreshCw } from 'lucide-react';

interface FirewallRule {
  port: string;
  protocol: string;
  action: string;
  source?: string;
}

interface FirewallStatus {
  backend: string; // "firewalld" | "ufw" | "NOT_INSTALLED"
  active: boolean;
  rules: FirewallRule[];
}

interface SELinuxStatus {
  status: string; // "INSTALLED" | "NOT_INSTALLED"
  mode: string;
  policy: string;
}

interface SecurityData {
  firewall: FirewallStatus;
  selinux: SELinuxStatus;
}

interface SecurityPanelProps {
  security: SecurityData | null;
  loading: boolean;
  onRefresh: () => void;
  onAddRule: (port: string, protocol: string) => void;
  onRemoveRule: (port: string, protocol: string) => void;
  onSetSELinuxMode: (mode: string) => void;
  onInstallComponent: (component: string) => void;
}

const SecurityPanel: React.FC<SecurityPanelProps> = ({
  security,
  loading,
  onRefresh,
  onAddRule,
  onRemoveRule,
  onSetSELinuxMode,
  onInstallComponent,
}) => {
  const [newPort, setNewPort] = useState('');
  const [newProtocol, setNewProtocol] = useState('tcp');
  const [installing, setInstalling] = useState<string | null>(null);

  const handleAddRule = () => {
    if (!newPort || !/^\d{1,5}$/.test(newPort)) return;
    const portNum = parseInt(newPort, 10);
    if (portNum < 1 || portNum > 65535) return;
    onAddRule(newPort, newProtocol);
    setNewPort('');
  };

  const handleInstall = async (component: string) => {
    setInstalling(component);
    await onInstallComponent(component);
    setInstalling(null);
  };

  if (!security) {
    return (
      <div className="flex items-center justify-center py-16 text-slate-500">
        <RefreshCw size={18} className={loading ? 'animate-spin mr-2' : 'mr-2'} />
        {loading ? 'Loading security status...' : 'No security data available'}
      </div>
    );
  }

  const fw = security.firewall;
  const se = security.selinux;

  return (
    <div className="space-y-6">
      {/* SSH Warning Banner */}
      <div className="flex items-start gap-3 p-4 bg-amber-500/5 border border-amber-500/20 rounded-xl">
        <AlertTriangle className="w-5 h-5 text-amber-400 mt-0.5 shrink-0" />
        <div>
          <p className="text-amber-300 text-sm font-semibold">Caution: Live Firewall Modification</p>
          <p className="text-amber-400/70 text-xs mt-1">
            Modifying firewall rules can interrupt your current SSH session. Ensure port 22 (SSH) remains open before making changes.
          </p>
        </div>
      </div>

      {/* ── Firewall Control ── */}
      <div className="border border-slate-800 rounded-xl overflow-hidden bg-slate-900/50">
        <div className="flex items-center justify-between p-4 bg-slate-800/50 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <Shield className="w-5 h-5 text-blue-400" />
            <div>
              <h3 className="text-sm font-bold text-slate-200">Firewall Control</h3>
              <p className="text-[10px] text-slate-500 uppercase tracking-wider mt-0.5">
                Backend: {fw.backend === 'NOT_INSTALLED' ? 'Not detected' : fw.backend}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            {fw.backend !== 'NOT_INSTALLED' && (
              <span
                className={`inline-flex items-center text-[10px] uppercase tracking-wider px-2.5 py-1 rounded-full border ${
                  fw.active
                    ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_8px_rgba(16,185,129,0.2)]'
                    : 'bg-slate-500/10 text-slate-400 border-slate-500/30'
                }`}
              >
                {fw.active ? 'Active' : 'Inactive'}
              </span>
            )}
            <button onClick={onRefresh} className="p-1.5 text-slate-400 hover:text-white rounded-lg transition-colors" title="Refresh">
              <RefreshCw size={14} className={loading ? 'animate-spin' : ''} />
            </button>
          </div>
        </div>

        <div className="p-4">
          {fw.backend === 'NOT_INSTALLED' ? (
            <div className="py-6 flex flex-col items-center justify-center text-center space-y-4">
              <ShieldAlert className="w-12 h-12 text-slate-700" />
              <div>
                <p className="text-slate-300 font-semibold text-sm">Firewall is not installed</p>
                <p className="text-slate-500 text-xs mt-1 max-w-xs">
                  No supported firewall service (firewalld or ufw) was detected on this VM. Security management requires a firewall service.
                </p>
              </div>
              <button
                onClick={() => handleInstall('firewall')}
                disabled={installing === 'firewall'}
                className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-bold shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2"
              >
                {installing === 'firewall' ? <RefreshCw size={16} className="animate-spin" /> : <Shield size={16} />}
                Setup Firewall
              </button>
            </div>
          ) : !fw.active ? (
            <div className="py-8 flex flex-col items-center justify-center text-center space-y-4">
              <ShieldAlert className="w-12 h-12 text-amber-500/20" />
              <div>
                <p className="text-slate-300 font-semibold text-sm">Firewall is inactive</p>
                <p className="text-slate-500 text-xs mt-1 max-w-xs">
                  The {fw.backend} service is installed but not currently running. Protected ports may be exposed or blocked by default policies.
                </p>
              </div>
              <button
                onClick={() => handleInstall('firewall')}
                disabled={installing === 'firewall'}
                className="px-6 py-2.5 bg-amber-600 hover:bg-amber-500 text-white rounded-xl font-bold shadow-lg shadow-amber-500/20 transition-all flex items-center gap-2"
              >
                {installing === 'firewall' ? <RefreshCw size={16} className="animate-spin" /> : <ShieldCheck size={16} />}
                Enable Firewall Service
              </button>
            </div>
          ) : (
            <>
              {/* Rules Table */}
              <div className="overflow-auto max-h-[280px] mb-4">
                <table className="w-full text-left border-collapse">
                  <thead className="sticky top-0 bg-slate-900 z-10">
                    <tr>
                      <th className="py-2 px-3 text-xs font-bold uppercase tracking-wider text-slate-500 w-[30%]">Port</th>
                      <th className="py-2 px-3 text-xs font-bold uppercase tracking-wider text-slate-500 w-[25%]">Protocol</th>
                      <th className="py-2 px-3 text-xs font-bold uppercase tracking-wider text-slate-500 w-[25%]">Action</th>
                      <th className="py-2 px-3 text-xs font-bold uppercase tracking-wider text-slate-500 w-[20%] text-right">Remove</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800">
                    {fw.rules.length > 0 ? (
                      fw.rules.map((rule, i) => (
                        <tr key={`${rule.port}-${rule.protocol}-${i}`} className="hover:bg-slate-800/30 transition-colors">
                          <td className="py-2 px-3 text-sm font-mono text-slate-200">{rule.port}</td>
                          <td className="py-2 px-3 text-sm text-slate-400 uppercase">{rule.protocol}</td>
                          <td className="py-2 px-3">
                            <span className="text-[10px] uppercase tracking-wider px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                              {rule.action}
                            </span>
                          </td>
                          <td className="py-2 px-3 text-right">
                            <button
                              onClick={() => onRemoveRule(rule.port, rule.protocol)}
                              className="p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-400/10 rounded-lg transition-colors"
                              title="Remove rule"
                            >
                              <Trash2 size={14} />
                            </button>
                          </td>
                        </tr>
                      ))
                    ) : (
                      <tr>
                        <td colSpan={4} className="py-6 text-center text-slate-500 text-sm italic">
                          No firewall rules configured
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>

              {/* Add Rule Form */}
              <div className="flex items-center gap-2 pt-3 border-t border-slate-800">
                <input
                  type="text"
                  placeholder="Port (e.g. 8080)"
                  value={newPort}
                  onChange={(e) => setNewPort(e.target.value.replace(/\D/g, '').slice(0, 5))}
                  className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm w-32 outline-none focus:border-blue-500 transition-colors"
                />
                <select
                  value={newProtocol}
                  onChange={(e) => setNewProtocol(e.target.value)}
                  className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-sm outline-none focus:border-blue-500 transition-colors"
                >
                  <option value="tcp">TCP</option>
                  <option value="udp">UDP</option>
                </select>
                <button
                  onClick={handleAddRule}
                  disabled={!newPort}
                  className="flex items-center gap-1.5 px-4 py-1.5 bg-blue-600/10 border border-blue-500/20 text-blue-400 hover:bg-blue-600/20 hover:border-blue-500/40 disabled:opacity-30 disabled:cursor-not-allowed rounded-lg text-sm font-medium transition-all"
                >
                  <Plus size={14} />
                  Add Rule
                </button>
              </div>
            </>
          )}
        </div>
      </div>

      {/* ── SELinux Dashboard ── */}
      <div className="border border-slate-800 rounded-xl overflow-hidden bg-slate-900/50">
        <div className="flex items-center justify-between p-4 bg-slate-800/50 border-b border-slate-800">
          <div className="flex items-center gap-3">
            {se.status === 'NOT_INSTALLED' ? (
              <ShieldAlert className="w-5 h-5 text-slate-500" />
            ) : se.mode === 'enforcing' ? (
              <ShieldCheck className="w-5 h-5 text-emerald-400" />
            ) : se.mode === 'permissive' ? (
              <ShieldAlert className="w-5 h-5 text-amber-400" />
            ) : (
              <Shield className="w-5 h-5 text-slate-500" />
            )}
            <div>
              <h3 className="text-sm font-bold text-slate-200">SELinux Dashboard</h3>
              <p className="text-[10px] text-slate-500 uppercase tracking-wider mt-0.5">
                {se.status === 'INSTALLED' ? `Policy: ${se.policy || 'N/A'}` : 'Not installed'}
              </p>
            </div>
          </div>
          {se.status === 'INSTALLED' && (
            <span
              className={`inline-flex items-center text-[10px] uppercase tracking-wider px-2.5 py-1 rounded-full border ${
                se.mode === 'enforcing'
                  ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-[0_0_8px_rgba(16,185,129,0.2)]'
                  : se.mode === 'permissive'
                  ? 'bg-amber-500/10 text-amber-400 border-amber-500/30 shadow-[0_0_8px_rgba(245,158,11,0.2)]'
                  : 'bg-slate-500/10 text-slate-400 border-slate-500/30'
              }`}
            >
              {se.mode}
            </span>
          )}
        </div>

        <div className="p-4">
          {se.status === 'NOT_INSTALLED' ? (
            <div className="py-6 flex flex-col items-center justify-center text-center space-y-4">
              <ShieldAlert className="w-12 h-12 text-slate-700" />
              <div>
                <p className="text-slate-300 font-semibold text-sm">SELinux is not installed</p>
                <p className="text-slate-500 text-xs mt-1 max-w-xs">
                  Security Enhanced Linux (SELinux) tools were not detected. SELinux provides mandatory access control.
                </p>
              </div>
              <button
                onClick={() => handleInstall('selinux')}
                disabled={installing === 'selinux'}
                className="btn-primary flex items-center gap-2"
              >
                {installing === 'selinux' ? <RefreshCw size={14} className="animate-spin" /> : <Plus size={14} />}
                Install SELinux Utilities
              </button>
            </div>
          ) : se.mode === 'disabled' ? (
            <div className="py-6 flex flex-col items-center justify-center text-center space-y-4">
              <ShieldAlert className="w-12 h-12 text-slate-700" />
              <div>
                <p className="text-slate-300 font-semibold text-sm">SELinux is disabled</p>
                <p className="text-slate-500 text-xs mt-1 max-w-xs">
                  SELinux is currently disabled in the system configuration. To enable it, you must first switch to Permissive mode and reboot.
                </p>
              </div>
              <button
                onClick={() => handleInstall('selinux')}
                disabled={installing === 'selinux'}
                className="px-6 py-2.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-bold shadow-lg shadow-blue-500/20 transition-all flex items-center gap-2"
              >
                {installing === 'selinux' ? <RefreshCw size={16} className="animate-spin" /> : <ShieldCheck size={16} />}
                Enable SELinux (Permissive)
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-4">
              <div className="flex-1 text-sm text-slate-400">
                Switch SELinux between <strong className="text-emerald-400">Enforcing</strong> and{' '}
                <strong className="text-amber-400">Permissive</strong> mode.
                <span className="text-[10px] block text-slate-600 mt-1">
                  Note: This change is runtime-only and will not survive a reboot. To persist, edit /etc/selinux/config.
                </span>
              </div>
              <div className="flex gap-2 shrink-0">
                <button
                  onClick={() => onSetSELinuxMode('enforcing')}
                  disabled={se.mode === 'enforcing'}
                  className={`px-4 py-2 rounded-xl text-sm font-medium border transition-all ${
                    se.mode === 'enforcing'
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 cursor-default'
                      : 'bg-slate-800 text-slate-400 border-slate-700 hover:bg-emerald-500/10 hover:text-emerald-400 hover:border-emerald-500/30'
                  }`}
                >
                  Enforcing
                </button>
                <button
                  onClick={() => onSetSELinuxMode('permissive')}
                  disabled={se.mode === 'permissive'}
                  className={`px-4 py-2 rounded-xl text-sm font-medium border transition-all ${
                    se.mode === 'permissive'
                      ? 'bg-amber-500/10 text-amber-400 border-amber-500/30 cursor-default'
                      : 'bg-slate-800 text-slate-400 border-slate-700 hover:bg-amber-500/10 hover:text-amber-400 hover:border-amber-500/30'
                  }`}
                >
                  Permissive
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SecurityPanel;
