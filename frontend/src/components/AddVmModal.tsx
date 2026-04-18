import React, { useState, useEffect } from 'react';
import api from '../services/api';
import { X, Shield, Cpu, User, Edit3, Key } from 'lucide-react';

interface VM {
  id: string;
  name: string;
  host: string;
  management_username: string;
  auth_type: string;
  key_id?: string;
  tags: string[];
  status: string;
}

interface SSHKey {
  id: string;
  name: string;
}

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSave: () => void;
  editingVm?: VM | null;
}

const AddVmModal: React.FC<Props> = ({ isOpen, onClose, onSave, editingVm }) => {
  const isEditMode = !!editingVm;
  const [availableKeys, setAvailableKeys] = useState<SSHKey[]>([]);

  const [formData, setFormData] = useState({
    name: editingVm?.name ?? '',
    host: editingVm?.host ?? '',
    management_username: editingVm?.management_username ?? '',
    auth_type: editingVm?.auth_type ?? 'password',
    credential: '',
    key_id: editingVm?.key_id ?? '',
    tags: editingVm?.tags?.join(', ') ?? '',
  });

  // Sync form when editingVm changes
  useEffect(() => {
    setFormData({
      name: editingVm?.name ?? '',
      host: editingVm?.host ?? '',
      management_username: editingVm?.management_username ?? '',
      auth_type: editingVm?.auth_type ?? 'password',
      credential: '',
      key_id: editingVm?.key_id ?? '',
      tags: editingVm?.tags?.join(', ') ?? '',
    });
  }, [editingVm]);

  // Load available SSH keys for the dropdown
  useEffect(() => {
    if (isOpen) {
      api.get('/keys').then(r => setAvailableKeys(r.data || [])).catch(() => {});
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const useKeyVault = formData.auth_type === 'key' && formData.key_id !== '';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const tagsArray = formData.tags.split(',').map(t => t.trim()).filter(Boolean);
    const payload: Record<string, unknown> = {
      name: formData.name,
      host: formData.host,
      management_username: formData.management_username,
      auth_type: formData.auth_type,
      tags: tagsArray,
    };

    // Only include credential OR key_id, not both
    if (formData.auth_type === 'key' && formData.key_id) {
      payload.key_id = formData.key_id;
      payload.credential = ''; // placeholder — backend ignores when key_id set
    } else if (formData.credential) {
      payload.credential = formData.credential;
    }

    try {
      if (isEditMode && editingVm) {
        await api.put(`/vms/${editingVm.id}`, payload);
      } else {
        await api.post('/vms', payload);
      }
      onSave();
      onClose();
    } catch (err) {
      console.error(`Failed to ${isEditMode ? 'update' : 'add'} VM`, err);
    }
  };

  const field = (
    label: string,
    key: keyof typeof formData,
    opts?: { type?: string; placeholder?: string; icon?: React.ReactNode; required?: boolean }
  ) => (
    <div>
      <label className="block text-sm font-medium text-slate-400 mb-2">{label}</label>
      <div className="relative">
        {opts?.icon && (
          <span className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500">{opts.icon}</span>
        )}
        <input
          type={opts?.type ?? 'text'}
          className={`w-full bg-slate-800 border border-slate-700 rounded-lg py-2 ${opts?.icon ? 'pl-10' : 'pl-4'} pr-4 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all`}
          value={String(formData[key])}
          onChange={e => setFormData({ ...formData, [key]: e.target.value })}
          placeholder={opts?.placeholder}
          required={opts?.required}
        />
      </div>
    </div>
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-slate-950/80 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-xl bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl overflow-hidden">
        <div className="px-8 py-6 border-b border-slate-800 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Edit3 className="text-blue-500 w-5 h-5" />
            <h2 className="text-xl font-bold text-white">
              {isEditMode ? `Edit: ${editingVm!.name}` : 'Add New Virtual Machine'}
            </h2>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white transition-colors">
            <X className="w-6 h-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-8 space-y-5 max-h-[80vh] overflow-y-auto">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            {field('Display Name', 'name', { icon: <Cpu className="w-5 h-5" />, placeholder: 'Production Web Server', required: true })}
            {field('Host (IP or DNS)', 'host', { placeholder: '192.168.1.50', required: true })}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-5">
            {field('Management Username (High Privilege)', 'management_username', { icon: <User className="w-5 h-5" />, placeholder: 'root', required: true })}
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-2">Auth Type</label>
              <select
                className="w-full bg-slate-800 border border-slate-700 rounded-lg py-2 px-4 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={formData.auth_type}
                onChange={e => setFormData({ ...formData, auth_type: e.target.value, key_id: '' })}
              >
                <option value="password">Password</option>
                <option value="key">SSH Private Key</option>
              </select>
            </div>
          </div>

          {/* Credential area — switches between key vault dropdown and text input */}
          {formData.auth_type === 'key' && availableKeys.length > 0 ? (
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-2">
                <span className="flex items-center gap-2">
                  <Key className="w-4 h-4 text-blue-400" />
                  SSH Key from Vault
                </span>
              </label>
              <select
                className="w-full bg-slate-800 border border-slate-700 rounded-lg py-2 px-4 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                value={formData.key_id}
                onChange={e => setFormData({ ...formData, key_id: e.target.value, credential: '' })}
              >
                <option value="">— paste inline key instead —</option>
                {availableKeys.map(k => (
                  <option key={k.id} value={k.id}>{k.name}</option>
                ))}
              </select>
              {formData.key_id && (
                <p className="text-xs text-emerald-400 mt-2">✓ Will authenticate using vault key</p>
              )}
            </div>
          ) : null}

          {/* Only show inline credential field when not using vault key */}
          {!useKeyVault && (
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-2">
                {formData.auth_type === 'password' ? 'Password' : 'Private Key Content (PEM)'}
                {isEditMode && <span className="ml-2 text-slate-500 text-xs">(leave blank to keep current)</span>}
              </label>
              <div className="relative">
                <Shield className="absolute left-3 top-3 text-slate-500 w-5 h-5" />
                <textarea
                  className="w-full bg-slate-800 border border-slate-700 rounded-lg py-2 pl-10 pr-4 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[90px] font-mono text-sm"
                  value={formData.credential}
                  onChange={e => setFormData({ ...formData, credential: e.target.value })}
                  placeholder={isEditMode ? '••••••••  (unchanged)' : formData.auth_type === 'password' ? '••••••••' : '-----BEGIN OPENSSH PRIVATE KEY-----...'}
                  required={!isEditMode && !useKeyVault}
                />
              </div>
            </div>
          )}

          {field('Tags (comma separated)', 'tags', { placeholder: 'frontend, production, aws' })}

          <div className="flex justify-end gap-4 pt-4 border-t border-slate-800">
            <button type="button" onClick={onClose} className="px-6 py-2.5 text-slate-400 hover:text-white transition-colors">
              Cancel
            </button>
            <button
              type="submit"
              className="px-8 py-2.5 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20 active:scale-[0.98]"
            >
              {isEditMode ? 'Save Changes' : 'Add Server'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default AddVmModal;
