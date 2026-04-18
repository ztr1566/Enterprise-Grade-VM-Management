import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';
import { Server, Wifi, WifiOff, ShieldCheck, Clock, AlertTriangle, FileText, ArrowRight } from 'lucide-react';

interface DashboardStats {
  total_vms: number;
  online_vms: number;
  offline_vms: number;
  provisioned: number;
  pending: number;
  failed: number;
  audit_events: number;
}

const DashboardView: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await api.get('/dashboard/stats');
        setStats(response.data);
      } catch (err) {
        console.error('Failed to fetch dashboard stats', err);
      } finally {
        setLoading(false);
      }
    };
    fetchStats();
    const interval = setInterval(fetchStats, 15000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="page-enter">
        <div className="dashboard-skeleton">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="hero-card hero-card--skeleton" />
          ))}
        </div>
      </div>
    );
  }

  const heroCards = [
    {
      title: 'Total VMs',
      value: stats?.total_vms ?? 0,
      icon: Server,
      color: 'blue' as const,
      subtitle: 'Managed instances',
    },
    {
      title: 'Online',
      value: stats?.online_vms ?? 0,
      icon: Wifi,
      color: 'emerald' as const,
      subtitle: stats?.total_vms
        ? `${Math.round(((stats?.online_vms ?? 0) / stats.total_vms) * 100)}% uptime`
        : '0% uptime',
    },
    {
      title: 'Offline',
      value: stats?.offline_vms ?? 0,
      icon: WifiOff,
      color: 'red' as const,
      subtitle: 'Unreachable',
    },
    {
      title: 'Provisioned',
      value: stats?.provisioned ?? 0,
      icon: ShieldCheck,
      color: 'violet' as const,
      subtitle: 'Security hardened',
    },
    {
      title: 'Pending',
      value: stats?.pending ?? 0,
      icon: Clock,
      color: 'amber' as const,
      subtitle: 'Awaiting setup',
    },
    {
      title: 'Audit Events',
      value: stats?.audit_events ?? 0,
      icon: FileText,
      color: 'slate' as const,
      subtitle: 'Total logged',
    },
  ];

  return (
    <div className="page-enter">
      <header className="dashboard-header">
        <div>
          <h1 className="dashboard-title">Dashboard Overview</h1>
          <p className="dashboard-subtitle">Real-time infrastructure health at a glance.</p>
        </div>
      </header>

      <div className="hero-grid">
        {heroCards.map((card) => {
          const Icon = card.icon;
          return (
            <div key={card.title} className={`hero-card hero-card--${card.color}`}>
              <div className="hero-card-header">
                <div className={`hero-card-icon hero-card-icon--${card.color}`}>
                  <Icon size={20} />
                </div>
                <span className="hero-card-label">{card.title}</span>
              </div>
              <p className="hero-card-value">{card.value}</p>
              <p className="hero-card-subtitle">{card.subtitle}</p>
            </div>
          );
        })}
      </div>

      {(stats?.failed ?? 0) > 0 && (
        <div className="dashboard-alert">
          <AlertTriangle size={20} />
          <span>
            <strong>{stats?.failed}</strong> VM(s) have failed provisioning.
          </span>
          <Link to="/vms" className="dashboard-alert-link">
            View VMs <ArrowRight size={14} />
          </Link>
        </div>
      )}
    </div>
  );
};

export default DashboardView;
