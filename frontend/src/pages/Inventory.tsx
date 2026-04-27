import React, { useEffect, useState } from 'react';
import api from '../services/api';
import { Plus, Server, Tag, Edit3, Trash2, AlertTriangle, Activity, Terminal } from 'lucide-react';
import CompactPowerActions from './../components/CompactPowerActions';
import { useAuth } from '../context/AuthContext';
import { useOutletContext, Link } from 'react-router-dom';
import { OutletContextType } from '../components/Layout';
import AddVmModal from './../components/AddVmModal';
import TerminalView from './../components/TerminalView';
import VMDetails from './../components/VMDetails';
import ProvisioningBadge from '../components/ProvisioningBadge';

interface VM {
  id: string;
  name: string;
  host: string;
  management_username: string;
  auth_type: string;
  status: string;
  provisioning_status: string;
  tags: string[];
  authorized_users: string[];
}

// Confirmation dialog component
const DeleteConfirmDialog: React.FC<{
  vm: VM;
  onConfirm: () => void;
  onCancel: () => void;
}> = ({ vm, onConfirm, onCancel }) => (
  <div className="fixed inset-0 z-[60] flex items-center justify-center p-4">
    <div className="absolute inset-0 bg-slate-950/90 backdrop-blur-sm" onClick={onCancel} />
    <div className="relative w-full max-w-md bg-slate-900 border border-red-500/30 rounded-2xl shadow-2xl p-8">
      <div className="flex flex-col items-center text-center gap-4">
        <div className="w-16 h-16 rounded-full bg-red-500/10 flex items-center justify-center">
          <AlertTriangle className="w-8 h-8 text-red-500" />
        </div>
        <div>
          <h3 className="text-xl font-bold text-white mb-2">Confirm Deletion</h3>
          <p className="text-slate-400">
            Are you sure you want to permanently delete <span className="text-white font-semibold">{vm.name}</span>?
            <br />
            <span className="text-slate-500 text-sm">({vm.host})</span>
          </p>
          <p className="mt-3 text-red-400 text-sm">This action cannot be undone.</p>
        </div>
        <div className="flex gap-4 w-full pt-2">
          <button
            onClick={onCancel}
            className="flex-1 px-6 py-2.5 bg-slate-800 hover:bg-slate-700 text-white rounded-xl transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="flex-1 px-6 py-2.5 bg-red-600 hover:bg-red-500 text-white font-semibold rounded-xl transition-all shadow-lg shadow-red-500/20 active:scale-[0.98]"
          >
            Delete
          </button>
        </div>
      </div>
    </div>
  </div>
);


const Inventory: React.FC = () => {
  const [vms, setVms] = useState<VM[]>([]);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [editingVm, setEditingVm] = useState<VM | null>(null);
  const [deletingVm, setDeletingVm] = useState<VM | null>(null);
  const [activeTerminalVm, setActiveTerminalVm] = useState<VM | null>(null);
  const [detailsVm, setDetailsVm] = useState<VM | null>(null);
  const [userSelectVm, setUserSelectVm] = useState<VM | null>(null);

  const { searchTerm } = useOutletContext<OutletContextType>();
  const { token } = useAuth();

  const fetchVMs = async () => {
    try {
      const response = await api.get('/vms');
      setVms(response.data || []);
    } catch (err) {
      console.error('Failed to fetch VMs', err);
    }
  };

  useEffect(() => {
    fetchVMs();
    // Poll for status updates
    const interval = setInterval(() => {
      // Fast poll (5s) if any VM is pending provisioning
      // Normal poll (10s) otherwise to catch telemetry status changes (online/offline)
      const isPending = vms.some(vm => vm.provisioning_status === 'pending');
      const shouldFetch = isPending || (Date.now() % 10000 < 5000); // Simple way to throttle to 10s if not pending
      
      if (isPending) {
        fetchVMs();
      }
    }, 5000);

    const normalInterval = setInterval(() => {
      if (!vms.some(vm => vm.provisioning_status === 'pending')) {
        fetchVMs();
      }
    }, 10000);

    return () => {
      clearInterval(interval);
      clearInterval(normalInterval);
    };
  }, [vms.length, vms.some(vm => vm.provisioning_status === 'pending')]);

  const handleRetryProvisioning = async (vmId: string) => {
    try {
      await api.post(`/vms/${vmId}/provision`);
      fetchVMs();
    } catch (err) {
      console.error('Failed to trigger provisioning', err);
    }
  };

  const handleDeleteConfirm = async () => {
    if (!deletingVm) return;
    try {
      await api.delete(`/vms/${deletingVm.id}`);
      setDeletingVm(null);
      fetchVMs();
    } catch (err) {
      console.error('Failed to delete VM', err);
    }
  };

  const handlePowerAction = async (vmId: string, action: string) => {
    if (window.confirm(`Are you sure you want to ${action} this VM?`)) {
      try {
        await api.post(`/vms/${vmId}/power`, { action });
      } catch (err) {
        console.error('Power action failed', err);
        alert('Power action failed. Check console.');
      }
    }
  };

  const filteredVms = vms.filter(vm =>
    vm.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    (vm.tags && vm.tags.some(t => t.toLowerCase().includes(searchTerm.toLowerCase()))) ||
    vm.host.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="page-enter">
      <header className="dashboard-header flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="dashboard-title">Infrastructure Inventory</h1>
          <p className="dashboard-subtitle">Manage and connect to your virtual assets</p>
        </div>
        <button
          onClick={() => { setEditingVm(null); setIsAddModalOpen(true); }}
          className="btn-secondary bg-blue-600/10 text-blue-400 border-blue-500/20 hover:bg-blue-600/20 hover:border-blue-500/40 hover:text-blue-300 transition-all"
        >
          <Plus size={16} />
          Add New Server
        </button>
      </header>

        {/* VM Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-6 items-stretch">
          {filteredVms.map(vm => (
            <div
              key={vm.id}
              className="premium-card group"
            >
              {/* Card header */}
              <div className="flex items-start justify-between mb-5">
                <div className="flex gap-2">
                  <div className="w-11 h-11 bg-slate-800 rounded-xl flex items-center justify-center group-hover:bg-blue-500/10 transition-colors">
                    <Server className="text-slate-400 group-hover:text-blue-400 transition-colors w-5 h-5" />
                  </div>
                  <div className="flex flex-col gap-1">
                    <span className={`px-2.5 py-1 rounded-full text-[10px] font-bold uppercase tracking-wider w-fit ${
                      vm.status === 'online'
                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                        : 'bg-slate-800 text-slate-500 border border-slate-700'
                    }`}>
                      <span className={`inline-block w-1.5 h-1.5 rounded-full mr-1.5 ${vm.status === 'online' ? 'bg-emerald-400' : 'bg-slate-600'}`} />
                      {vm.status}
                    </span>
                    <ProvisioningBadge 
                      status={vm.provisioning_status} 
                      onRetry={() => handleRetryProvisioning(vm.id)}
                    />
                  </div>
                </div>
              </div>

              <h3 className="text-lg font-bold mb-1 group-hover:text-blue-400 transition-colors truncate">{vm.name}</h3>
              <p className="text-slate-500 text-sm font-mono mb-4 truncate">{vm.host}</p>

              {/* Tags */}
              <div className="flex flex-wrap gap-1.5 mb-5 min-h-[32px] content-start">
                {vm.tags?.length > 0 ? (
                  vm.tags.map((tag, i) => (
                    <span key={i} className="flex items-center gap-1 px-2 py-0.5 bg-slate-800 border border-slate-700 rounded-md text-xs text-slate-300">
                      <Tag className="w-2.5 h-2.5 text-slate-500" />
                      {tag}
                    </span>
                  ))
                ) : (
                  <span className="text-[10px] text-slate-600 italic">No tags assigned</span>
                )}
              </div>

              {/* Action row */}
              <div className="flex gap-2 mt-auto">
                {/* Terminal — primary action */}
                <div className="relative flex-1 flex gap-1">
                  <button
                    onClick={() => {
                      if (vm.authorized_users && vm.authorized_users.length > 0) {
                        setUserSelectVm(userSelectVm?.id === vm.id ? null : vm);
                      } else {
                        setActiveTerminalVm(vm);
                      }
                    }}
                    className="flex-1 bg-blue-600/10 hover:bg-blue-600/20 border border-blue-500/20 hover:border-blue-500/50 text-blue-400 font-medium py-2 rounded-xl transition-all flex items-center justify-center gap-2 text-sm"
                  >
                    <Terminal className="w-4 h-4" />
                    Terminal
                  </button>
                  
                  {userSelectVm?.id === vm.id && (
                    <div
                      style={{
                        position: 'absolute',
                        bottom: '100%',
                        left: 0,
                        marginBottom: '0.5rem',
                        width: '14rem',
                        background: 'var(--bg-surface)',
                        border: '1px solid var(--border-subtle)',
                        borderRadius: '12px',
                        boxShadow: '0 10px 30px rgba(0, 0, 0, 0.5)',
                        zIndex: 200,
                        overflow: 'hidden',
                      }}
                    >
                      <div className="px-3 py-2 border-b border-slate-800 bg-slate-800/50">
                        <span className="text-[10px] font-bold text-slate-500 uppercase">Select Login User</span>
                      </div>
                      <button
                        onClick={() => { setActiveTerminalVm(vm); setUserSelectVm(null); }}
                        className="w-full text-left px-3 py-2 text-sm text-slate-200 hover:bg-blue-500/10 transition-colors flex items-center justify-between"
                      >
                        <span>{vm.management_username} (root/mgmt)</span>
                        <div className="w-1.5 h-1.5 rounded-full bg-blue-500" />
                      </button>
                      {vm.authorized_users.map(user => (
                        <button
                          key={user}
                          onClick={() => { 
                            const vmWithLogin = { ...vm, login_as: user };
                            setActiveTerminalVm(vmWithLogin as any); 
                            setUserSelectVm(null); 
                          }}
                          className="w-full text-left px-3 py-2 text-sm text-slate-400 hover:bg-blue-500/10 hover:text-slate-200 transition-colors"
                        >
                          {user}
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* Power Options */}
                <CompactPowerActions vmId={vm.id} onPowerAction={handlePowerAction} />

                {/* Edit */}
                <button
                  onClick={() => { setEditingVm(vm); setIsAddModalOpen(true); }}
                  className="w-10 h-10 bg-slate-800 hover:bg-slate-700 border border-slate-700 hover:border-slate-600 text-slate-400 hover:text-white rounded-xl transition-all flex items-center justify-center"
                  title="Edit VM"
                >
                  <Edit3 className="w-4 h-4" />
                </button>

                {/* Delete */}
                <button
                  onClick={() => setDeletingVm(vm)}
                  className="w-10 h-10 bg-slate-800 hover:bg-red-500/10 border border-slate-700 hover:border-red-500/30 text-slate-400 hover:text-red-400 rounded-xl transition-all flex items-center justify-center"
                  title="Delete VM"
                >
                  <Trash2 className="w-4 h-4" />
                </button>

                {/* Details - route to VMDetailPage */}
                <Link
                  to={`/vms/${vm.id}`}
                  className="w-10 h-10 bg-slate-800 hover:bg-emerald-500/10 border border-slate-700 hover:border-emerald-500/30 text-slate-400 hover:text-emerald-400 rounded-xl transition-all flex items-center justify-center"
                  title="VM Metrics & Details"
                >
                  <Activity className="w-4 h-4" />
                </Link>
              </div>
            </div>
          ))}

          {filteredVms.length === 0 && (
            <div className="col-span-full py-24 flex flex-col items-center text-slate-500 bg-slate-900/50 border-2 border-dashed border-slate-800 rounded-3xl">
              <Server className="w-12 h-12 mb-4 opacity-20" />
              <p className="text-lg mb-1">No virtual machines found</p>
              <p className="text-sm text-slate-600">
                {searchTerm ? 'Try a different search term' : 'Get started by adding your first server'}
              </p>
              {!searchTerm && (
                <button
                  onClick={() => setIsAddModalOpen(true)}
                  className="mt-4 text-blue-500 hover:underline"
                >
                  Add server →
                </button>
              )}
            </div>
          )}
        </div>

      {/* Add / Edit Modal */}
      <AddVmModal
        isOpen={isAddModalOpen}
        onClose={() => { setIsAddModalOpen(false); setEditingVm(null); }}
        onSave={fetchVMs}
        editingVm={editingVm}
      />

      {/* Delete Confirmation */}
      {deletingVm && (
        <DeleteConfirmDialog
          vm={deletingVm}
          onConfirm={handleDeleteConfirm}
          onCancel={() => setDeletingVm(null)}
        />
      )}

      {/* SSH Terminal Modal */}
      {activeTerminalVm && token && (
        <TerminalView
          vmId={activeTerminalVm.id}
          vmName={activeTerminalVm.name}
          token={token}
          loginAs={(activeTerminalVm as any).login_as}
          onClose={() => setActiveTerminalVm(null)}
        />
      )}

      {/* VM Details Side Panel */}
      <VMDetails vm={detailsVm} onClose={() => setDetailsVm(null)} />
    </div>
  );
};

export default Inventory;
