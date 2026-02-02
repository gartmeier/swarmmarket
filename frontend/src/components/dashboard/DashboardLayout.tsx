import { useState, useRef, useEffect } from 'react';
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useUser, useClerk } from '@clerk/clerk-react';
import {
  LayoutDashboard,
  Settings,
  ChevronUp,
  LogOut,
  User,
  Store,
} from 'lucide-react';
import { useApiSetup } from '../../hooks/useDashboard';

const navItems = [
  { icon: LayoutDashboard, label: 'Dashboard', path: '/dashboard', highlight: false },
  { icon: Store, label: 'Marketplace', path: '/dashboard/marketplace', highlight: true },
  { icon: Settings, label: 'Settings', path: '/dashboard/settings', highlight: false },
];

export function DashboardLayout() {
  const { user } = useUser();
  const { signOut, openUserProfile } = useClerk();
  const navigate = useNavigate();
  const location = useLocation();
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  useApiSetup();

  const firstName = user?.firstName || 'there';

  // Determine if sidebar should be collapsed based on route
  const isAgentDetailPage = /^\/dashboard\/agents\/[^/]+/.test(location.pathname);
  const isMarketplaceDetailPage = /^\/dashboard\/marketplace\/(listings|requests|auctions)\/[^/]+/.test(location.pathname);
  const isMarketplacePage = location.pathname === '/dashboard/marketplace';
  const isCollapsed = isAgentDetailPage || isMarketplaceDetailPage;

  // Close menu when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setUserMenuOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSignOut = () => {
    signOut(() => navigate('/'));
  };

  return (
    <div className="flex h-screen w-full bg-[#0A0F1C]">
      {/* Sidebar */}
      <aside
        className="h-full bg-[#0F172A] flex flex-col justify-between transition-all duration-200"
        style={{
          width: isCollapsed ? '72px' : '260px',
          padding: isCollapsed ? '24px 16px' : '24px 20px',
        }}
      >
        <div className="flex flex-col" style={{ gap: isCollapsed ? '24px' : '32px' }}>
          {/* Logo */}
          <div className={`flex items-center ${isCollapsed ? 'justify-center' : 'gap-2.5'}`}>
            <img src="/logo.webp" alt="SwarmMarket" className="w-8 h-8" />
            {!isCollapsed && (
              <span className="font-mono font-bold text-white text-base">SwarmMarket</span>
            )}
          </div>

          {/* Navigation */}
          <nav className="flex flex-col gap-1 items-start w-full">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <NavLink
                  key={item.path}
                  to={item.path}
                  end={item.path === '/dashboard'}
                  className={({ isActive }) =>
                    `flex items-center rounded-lg transition-colors w-full ${
                      isCollapsed ? 'justify-center' : 'justify-start gap-3'
                    } ${isActive ? 'bg-[#1E293B]' : 'hover:bg-[#1E293B]/50'} ${
                      item.highlight && !isActive ? 'ring-1 ring-[#22D3EE]/30' : ''
                    }`
                  }
                  style={{
                    padding: isCollapsed ? '12px' : '12px 16px',
                    background: item.highlight ? 'linear-gradient(135deg, rgba(34, 211, 238, 0.08) 0%, rgba(168, 85, 247, 0.08) 100%)' : undefined,
                  }}
                  title={isCollapsed ? item.label : undefined}
                >
                  {({ isActive }) => (
                    <>
                      <Icon
                        className="w-5 h-5"
                        style={{ color: isActive ? '#22D3EE' : item.highlight ? '#22D3EE' : '#64748B' }}
                      />
                      {!isCollapsed && (
                        <span
                          className="text-sm font-medium text-left"
                          style={{ color: isActive ? '#FFFFFF' : item.highlight ? '#E0F2FE' : '#94A3B8' }}
                        >
                          {item.label}
                        </span>
                      )}
                    </>
                  )}
                </NavLink>
              );
            })}
          </nav>
        </div>

        {/* User Section - Only show in full sidebar */}
        {!isCollapsed && (
          <div className="relative" ref={menuRef}>
            <button
              onClick={() => setUserMenuOpen(!userMenuOpen)}
              className="flex items-center gap-3 w-full rounded-lg bg-[#1E293B] hover:bg-[#2D3B4F] transition-colors"
              style={{ padding: '12px 16px', border: '1px solid #334155' }}
            >
              {/* Avatar */}
              {user?.imageUrl ? (
                <img
                  src={user.imageUrl}
                  alt={user.fullName || 'User'}
                  className="w-9 h-9 rounded-full flex-shrink-0 object-cover"
                />
              ) : (
                <div
                  className="w-9 h-9 rounded-full flex items-center justify-center flex-shrink-0"
                  style={{ backgroundColor: '#A855F7' }}
                >
                  <span className="text-white text-sm font-semibold">
                    {user?.firstName?.[0] || ''}{user?.lastName?.[0] || ''}
                  </span>
                </div>
              )}
              {/* User Info */}
              <div className="flex flex-col items-start flex-1 min-w-0">
                <span className="text-white text-[13px] font-medium truncate w-full text-left">
                  {user?.fullName || 'User'}
                </span>
                <span className="text-[#64748B] text-[11px] truncate w-full text-left">
                  {user?.primaryEmailAddress?.emailAddress || ''}
                </span>
              </div>
              {/* Chevron */}
              <ChevronUp
                className={`w-4 h-4 text-[#64748B] flex-shrink-0 transition-transform ${userMenuOpen ? '' : 'rotate-180'}`}
              />
            </button>

            {/* Dropdown Menu */}
            {userMenuOpen && (
              <div
                className="absolute bottom-full left-0 right-0 mb-3 rounded-xl bg-[#1E293B] shadow-xl"
                style={{ border: '1px solid #334155', padding: '8px' }}
              >
                <button
                  onClick={() => {
                    openUserProfile();
                    setUserMenuOpen(false);
                  }}
                  className="flex items-center gap-3 w-full rounded-lg text-left text-[#94A3B8] hover:bg-[#2D3B4F] hover:text-white transition-colors"
                  style={{ padding: '12px 14px' }}
                >
                  <User className="w-5 h-5" />
                  <span className="text-sm font-medium">Manage Account</span>
                </button>
                <button
                  onClick={handleSignOut}
                  className="flex items-center gap-3 w-full rounded-lg text-left text-[#94A3B8] hover:bg-[#2D3B4F] hover:text-white transition-colors"
                  style={{ padding: '12px 14px' }}
                >
                  <LogOut className="w-5 h-5" />
                  <span className="text-sm font-medium">Sign Out</span>
                </button>
              </div>
            )}
          </div>
        )}

        {/* Collapsed user avatar only */}
        {isCollapsed && (
          <div className="flex justify-center">
            <button
              onClick={() => openUserProfile()}
              className="w-10 h-10 rounded-full flex items-center justify-center hover:ring-2 hover:ring-[#334155] transition-all overflow-hidden"
              style={{ backgroundColor: user?.imageUrl ? 'transparent' : '#A855F7' }}
              title="Account"
            >
              {user?.imageUrl ? (
                <img
                  src={user.imageUrl}
                  alt={user.fullName || 'User'}
                  className="w-full h-full object-cover"
                />
              ) : (
                <span className="text-white text-sm font-semibold">
                  {user?.firstName?.[0] || ''}{user?.lastName?.[0] || ''}
                </span>
              )}
            </button>
          </div>
        )}
      </aside>

      {/* Main Content */}
      <main
        className="flex-1 h-full overflow-auto flex flex-col"
        style={{ padding: '32px 40px' }}
      >
        {/* Header - Only show on dashboard home and settings */}
        {!isAgentDetailPage && !isMarketplacePage && !isMarketplaceDetailPage && (
          <div style={{ marginBottom: '32px' }}>
            <h1 className="text-[28px] font-bold text-white leading-tight">
              Welcome back, {firstName}
            </h1>
            <p className="text-sm text-[#64748B]">Here's what your agents have been up to</p>
          </div>
        )}

        {/* Page Content */}
        <Outlet />
      </main>
    </div>
  );
}
