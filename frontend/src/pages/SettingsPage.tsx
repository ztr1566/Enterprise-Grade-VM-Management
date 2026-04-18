import React from 'react';
import { Shield, Bell, Palette } from 'lucide-react';

const SettingsPage: React.FC = () => {
  return (
    <div className="page-enter">
      <header className="dashboard-header">
        <div>
          <h1 className="dashboard-title">Settings</h1>
          <p className="dashboard-subtitle">Platform configuration and preferences.</p>
        </div>
      </header>

      <div className="settings-grid">
        <div className="card settings-card">
          <div className="settings-card-header">
            <Shield size={20} className="settings-card-icon" />
            <h3>Security</h3>
          </div>
          <p className="text-muted">Manage authentication, session timeouts, and access controls.</p>
        </div>

        <div className="card settings-card">
          <div className="settings-card-header">
            <Bell size={20} className="settings-card-icon" />
            <h3>Notifications</h3>
          </div>
          <p className="text-muted">Configure alerts for VM status changes and provisioning events.</p>
        </div>

        <div className="card settings-card">
          <div className="settings-card-header">
            <Palette size={20} className="settings-card-icon" />
            <h3>Appearance</h3>
          </div>
          <p className="text-muted">Customize the dashboard theme and layout preferences.</p>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;
