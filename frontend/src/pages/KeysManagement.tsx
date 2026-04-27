import React, { useEffect, useState } from 'react';
import api from '../services/api';
import { Key, Plus, Trash2, Copy, Check, ChevronRight } from 'lucide-react';

interface SSHKey {
  id: string;
  name: string;
  public_key: string;
  created_at: string;
}

const KeysManagement: React.FC = () => {
  const [keys, setKeys] = useState<SSHKey[]>([]);
  const [isAdding, setIsAdding] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyPem, setNewKeyPem] = useState('');
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [generatedPublicKey, setGeneratedPublicKey] = useState<string | null>(null);
  // Auth hooks removed as they are handled by AppBar

  const fetchKeys = async () => {
    try {
      const res = await api.get('/keys');
      setKeys(res.data || []);
    } catch (err) {
      console.error('Failed to fetch SSH keys', err);
    }
  };

  useEffect(() => { fetchKeys(); }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await api.post('/keys', {
        name: newKeyName,
        private_key: newKeyPem, // empty = generate
      });
      setGeneratedPublicKey(res.data.public_key);
      setNewKeyName('');
      setNewKeyPem('');
      setIsAdding(false);
      fetchKeys();
    } catch (err: any) {
      alert(err?.response?.data?.error || 'Failed to create key');
    }
  };

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Delete key "${name}"? This cannot be undone.`)) return;
    await api.delete(`/keys/${id}`);
    fetchKeys();
  };

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => setCopiedId(null), 2000);
  };

  return (
    <div className="page-enter">
      <header className="dashboard-header flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <h1 className="dashboard-title">SSH Key Vault</h1>
          <p className="dashboard-subtitle">Manage encrypted SSH key pairs for secure VM authentication</p>
        </div>
        <button
          onClick={() => { setIsAdding(true); setGeneratedPublicKey(null); }}
          className="btn-secondary bg-blue-600/10 text-blue-400 border-blue-500/20 hover:bg-blue-600/20 hover:border-blue-500/40 hover:text-blue-300 transition-all"
        >
          <Plus size={16} />
          Add Key
        </button>
      </header>

        {/* Generated public key banner */}
        {generatedPublicKey && (
          <div className="mb-8 bg-emerald-500/10 border border-emerald-500/20 rounded-2xl p-6">
            <div className="flex items-center justify-between mb-3">
              <span className="text-emerald-400 font-semibold">✓ Key generated — copy the public key to install on your servers</span>
              <button
                onClick={() => setGeneratedPublicKey(null)}
                className="text-slate-500 hover:text-white text-sm"
              >
                Dismiss
              </button>
            </div>
            <div className="relative">
              <pre className="bg-slate-950 rounded-xl p-4 text-xs font-mono text-slate-300 overflow-x-auto whitespace-pre-wrap break-all">
                {generatedPublicKey}
              </pre>
              <button
                onClick={() => copyToClipboard(generatedPublicKey, 'generated')}
                className="absolute top-3 right-3 p-2 bg-slate-800 hover:bg-slate-700 rounded-lg transition-colors"
              >
                {copiedId === 'generated' ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4 text-slate-400" />}
              </button>
            </div>
          </div>
        )}

        {/* Add Key Form */}
        {isAdding && (
          <div className="mb-8 bg-slate-900 border border-slate-700 rounded-2xl p-6">
            <h3 className="text-lg font-bold mb-5 flex items-center gap-2">
              <Plus className="w-5 h-5 text-blue-400" />
              Add SSH Key
            </h3>
            <form onSubmit={handleCreate} className="space-y-5">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">Key Name</label>
                <input
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g. production-key"
                  value={newKeyName}
                  onChange={e => setNewKeyName(e.target.value)}
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-2">
                  Private Key (PEM)
                  <span className="ml-2 text-slate-500 text-xs">— leave blank to auto-generate RSA 4096</span>
                </label>
                <textarea
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl py-2.5 px-4 text-white focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm min-h-[140px]"
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                  value={newKeyPem}
                  onChange={e => setNewKeyPem(e.target.value)}
                />
              </div>
              <div className="flex gap-3 justify-end pt-2 border-t border-slate-800">
                <button type="button" onClick={() => setIsAdding(false)} className="px-5 py-2 text-slate-400 hover:text-white transition-colors">
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-8 py-2 bg-blue-600 hover:bg-blue-500 text-white font-semibold rounded-xl transition-all shadow-lg shadow-blue-500/20"
                >
                  {newKeyPem ? 'Import Key' : 'Generate Key'}
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Keys List */}
        <div className="space-y-4">
          {keys.length === 0 ? (
            <div className="py-24 flex flex-col items-center text-slate-500 bg-slate-900/50 border-2 border-dashed border-slate-800 rounded-3xl">
              <Key className="w-12 h-12 mb-4 opacity-20" />
              <p className="text-lg mb-1">No SSH keys in vault</p>
              <p className="text-sm text-slate-600">Add your first key to enable secure VM authentication</p>
            </div>
          ) : (
            keys.map(key => (
              <div key={key.id} className="premium-card p-5 flex flex-col gap-4">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex items-center gap-4 min-w-0">
                    <div className="w-10 h-10 bg-blue-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
                      <Key className="text-blue-400 w-5 h-5" />
                    </div>
                    <div className="min-w-0">
                      <p className="font-bold text-white text-lg">{key.name}</p>
                      <p className="text-slate-500 text-xs mt-0.5">
                        Added {new Date(key.created_at).toLocaleDateString()} · ID: {key.id.slice(0, 12)}…
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 flex-shrink-0">
                    <button
                      onClick={() => copyToClipboard(key.public_key, key.id)}
                      className="btn-secondary"
                      title="Copy public key"
                    >
                      {copiedId === key.id ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                      <span>{copiedId === key.id ? 'Copied!' : 'Copy Key'}</span>
                    </button>
                    <button
                      onClick={() => handleDelete(key.id, key.name)}
                      className="btn-icon hover:text-red-400 hover:border-red-500/30"
                      title="Delete key"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </div>

                {/* Public key preview */}
                <div className="bg-slate-950/50 border border-slate-800 rounded-xl px-4 py-3 flex items-center gap-3">
                  <ChevronRight className="w-3 h-3 text-slate-600 flex-shrink-0" />
                  <code className="text-xs text-slate-400 font-mono truncate">{key.public_key}</code>
                </div>
              </div>
            ))
          )}
        </div>
    </div>
  );
};

export default KeysManagement;
