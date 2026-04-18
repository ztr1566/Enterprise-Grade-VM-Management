import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { LayoutDashboard, Server, Settings, FileText, Key, Shield } from 'lucide-react';

const navItems = [
  { name: 'Dashboard', path: '/', icon: LayoutDashboard, end: true },
  { name: 'VMs', path: '/vms', icon: Server },
  { name: 'Audit Log', path: '/audit', icon: FileText },
  { name: 'Key Vault', path: '/keys', icon: Key },
  { name: 'Settings', path: '/settings', icon: Settings },
];

const Sidebar: React.FC = () => {
  const location = useLocation();

  return (
    <aside className="sidebar">
      <div className="sidebar-brand">
        <div className="sidebar-brand-icon">
          <Shield size={22} />
        </div>
        <span className="sidebar-brand-text">VM Platform</span>
      </div>

      <nav className="sidebar-nav">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = item.end
            ? location.pathname === item.path
            : location.pathname.startsWith(item.path);
          return (
            <NavLink
              key={item.name}
              to={item.path}
              end={item.end}
              className={`sidebar-link ${isActive ? 'sidebar-link--active' : ''}`}
            >
              <Icon size={18} className="sidebar-link-icon" />
              <span>{item.name}</span>
            </NavLink>
          );
        })}
      </nav>

      <div className="sidebar-footer">
        <div className="sidebar-footer-dot" />
        <span>System Operational</span>
      </div>
    </aside>
  );
};

export default Sidebar;
