import React, { useEffect, useState } from 'react';
import api from '../services/api';
import { ChevronLeft, ChevronRight, FileText, AlertCircle } from 'lucide-react';

interface AuditLog {
  id: number;
  event_type: string;
  details: string;
  timestamp: string;
}

interface PaginatedResponse {
  data: AuditLog[];
  meta: {
    current_page: number;
    per_page: number;
    total_items: number;
    total_pages: number;
  };
}

const AuditLogPage: React.FC = () => {
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [meta, setMeta] = useState<PaginatedResponse['meta'] | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const fetchLogs = async (pageNumber: number) => {
    setLoading(true);
    setError('');
    try {
      const response = await api.get(`/audit?page=${pageNumber}&limit=15`);
      setLogs(response.data.data || []);
      setMeta(response.data.meta);
    } catch (err) {
      console.error('Failed to fetch audit logs', err);
      setError('Failed to load audit logs.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs(page);
  }, [page]);

  const getEventBadgeClass = (eventType: string) => {
    if (eventType.includes('FAIL') || eventType.includes('ERROR')) return 'event-badge--red';
    if (eventType.includes('DELETE') || eventType.includes('REMOVE')) return 'event-badge--amber';
    if (eventType.includes('CREATE') || eventType.includes('SUCCESS')) return 'event-badge--emerald';
    return 'event-badge--blue';
  };

  return (
    <div className="page-enter">
      <header className="dashboard-header">
        <div>
          <h1 className="dashboard-title">Audit Log</h1>
          <p className="dashboard-subtitle">View system events and security logs.</p>
        </div>
      </header>

      {error && (
        <div className="dashboard-alert dashboard-alert--error">
          <AlertCircle size={20} />
          <span>{error}</span>
        </div>
      )}

      <div className="card">
        <div className="table-wrap">
          <table className="data-table">
            <thead>
              <tr>
                <th>ID</th>
                <th>Event</th>
                <th>Details</th>
                <th>Timestamp</th>
              </tr>
            </thead>
            <tbody>
              {loading && logs.length === 0 ? (
                <tr>
                  <td colSpan={4} className="table-empty">Loading logs...</td>
                </tr>
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={4} className="table-empty">
                    <FileText size={32} className="table-empty-icon" />
                    No audit logs found.
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id}>
                    <td className="font-mono text-muted">{log.id}</td>
                    <td>
                      <span className={`event-badge ${getEventBadgeClass(log.event_type)}`}>
                        {log.event_type}
                      </span>
                    </td>
                    <td className="table-details">{log.details}</td>
                    <td className="font-mono text-muted">
                      {new Date(log.timestamp).toLocaleString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        {meta && meta.total_pages > 1 && (
          <div className="table-pagination">
            <span className="text-muted">
              Page <strong>{meta.current_page}</strong> of <strong>{meta.total_pages}</strong>
            </span>
            <div className="table-pagination-btns">
              <button
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page === 1}
                className="btn-icon"
              >
                <ChevronLeft size={18} />
              </button>
              <button
                onClick={() => setPage((p) => Math.min(meta.total_pages, p + 1))}
                disabled={page === meta.total_pages}
                className="btn-icon"
              >
                <ChevronRight size={18} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AuditLogPage;
