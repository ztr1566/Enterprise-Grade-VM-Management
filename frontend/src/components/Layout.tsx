import React, { useEffect, useState, useRef } from 'react';
import { Outlet, useLocation, Link, useNavigate } from 'react-router-dom';
import { LogOut, Search, Key, Server } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import api from '../services/api';
import Sidebar from './Sidebar';

export interface OutletContextType {
  searchTerm: string;
}

const Layout: React.FC = () => {
  const { logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchTerm, setSearchTerm] = useState('');
  const [serverCount, setServerCount] = useState(0);
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await api.get('/dashboard/stats');
        setServerCount(response.data.total_vms);
      } catch (err) {
        console.error('Failed to fetch stats for layout', err);
      }
    };
    fetchStats();
    const interval = setInterval(fetchStats, 15000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const getPageTitle = () => {
    const path = location.pathname;
    if (path === '/') return 'Dashboard';
    if (path.startsWith('/vms/')) return 'VM Details';
    if (path === '/vms') return 'Infrastructure';
    if (path === '/audit') return 'Audit Log';
    if (path === '/keys') return 'Key Vault';
    if (path === '/settings') return 'Settings';
    return 'Workspace';
  };

  return (
    <div className="app-shell">
      <Sidebar />
      <div className="app-main">
        <header className="topbar">
          <h2 className="topbar-title">{getPageTitle()}</h2>
          
          <div className="flex-1 max-w-md mx-8 relative hidden md:block">
            {location.pathname === '/vms' && (
              <>
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-500 w-4 h-4" />
                <input
                  type="text"
                  placeholder="Search servers by name or host..."
                  className="w-full bg-slate-900 border border-slate-800 rounded-lg py-1.5 pl-9 pr-4 text-sm text-white focus:outline-none focus:border-slate-600 transition-colors"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </>
            )}
          </div>

          <div className="topbar-actions relative flex items-center gap-4">
            <Link to="/vms" className="relative text-slate-400 hover:text-white transition-colors" title="Inventory">
              <Server size={18} />
              {serverCount > 0 && (
                <span className="absolute -top-1.5 -right-2 bg-blue-600 text-white text-[10px] font-bold px-1.5 rounded-full border border-slate-900">
                  {serverCount}
                </span>
              )}
            </Link>

            <Link to="/keys" className="text-slate-400 hover:text-white transition-colors" title="Key Vault">
              <Key size={18} />
            </Link>

            <div className="relative ml-2" ref={dropdownRef}>
              <button 
                onClick={() => setDropdownOpen(!dropdownOpen)}
                className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center text-white font-bold text-sm hover:ring-2 hover:ring-blue-400 transition-all focus:outline-none"
              >
                A
              </button>

              {dropdownOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-slate-900 border border-slate-800 rounded-xl shadow-2xl py-1 z-50">
                  <div className="px-4 py-2 border-b border-slate-800 mb-1">
                    <p className="text-sm font-medium text-white">Admin User</p>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="w-full text-left px-4 py-2 text-sm text-red-400 hover:bg-slate-800 transition-colors flex items-center gap-2"
                  >
                    <LogOut size={14} />
                    Sign Out
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>
        <main className="app-content" key={location.pathname}>
          <Outlet context={{ searchTerm }} />
        </main>
      </div>
    </div>
  );
};

export default Layout;
