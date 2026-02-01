import { useState } from 'react';
import { User, Bell, Shield, CreditCard, Key } from 'lucide-react';

type SettingsTab = 'profile' | 'notifications' | 'security' | 'billing' | 'api_keys';

interface ToggleSetting {
  id: string;
  title: string;
  description: string;
  enabled: boolean;
}

const notificationSettings: ToggleSetting[] = [
  {
    id: 'email',
    title: 'Email Notifications',
    description: 'Receive email updates for your agents and tasks',
    enabled: true,
  },
  {
    id: 'task_completed',
    title: 'Task Completed',
    description: 'Get notified when an agent completes a task',
    enabled: true,
  },
  {
    id: 'payment_received',
    title: 'Payment Received',
    description: 'Get notified when you receive a payment',
    enabled: true,
  },
  {
    id: 'agent_status',
    title: 'Agent Status Changes',
    description: 'Get notified when an agent goes online or offline',
    enabled: false,
  },
  {
    id: 'marketing',
    title: 'Marketing & Updates',
    description: 'Receive news about new features and updates',
    enabled: false,
  },
];

function Toggle({
  enabled,
  onChange,
}: {
  enabled: boolean;
  onChange: (enabled: boolean) => void;
}) {
  return (
    <button
      onClick={() => onChange(!enabled)}
      className="relative rounded-full transition-colors"
      style={{
        width: '44px',
        height: '24px',
        backgroundColor: enabled ? '#A855F7' : '#334155',
      }}
    >
      <div
        className="absolute rounded-full bg-white transition-transform"
        style={{
          width: '16px',
          height: '16px',
          top: '4px',
          left: enabled ? '24px' : '4px',
        }}
      />
    </button>
  );
}

function SettingRow({
  setting,
  onToggle,
}: {
  setting: ToggleSetting;
  onToggle: (id: string, enabled: boolean) => void;
}) {
  return (
    <div className="flex items-center justify-between" style={{ padding: '16px 0' }}>
      <div className="flex flex-col" style={{ gap: '2px' }}>
        <span className="text-[14px] font-medium text-white">{setting.title}</span>
        <span className="text-[12px] text-[#64748B]">{setting.description}</span>
      </div>
      <Toggle enabled={setting.enabled} onChange={(enabled) => onToggle(setting.id, enabled)} />
    </div>
  );
}

export function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('notifications');
  const [settings, setSettings] = useState(notificationSettings);

  const handleToggle = (id: string, enabled: boolean) => {
    setSettings((prev) => prev.map((s) => (s.id === id ? { ...s, enabled } : s)));
  };

  const tabs = [
    { id: 'profile' as SettingsTab, label: 'Profile', icon: User },
    { id: 'notifications' as SettingsTab, label: 'Notifications', icon: Bell },
    { id: 'security' as SettingsTab, label: 'Security', icon: Shield },
    { id: 'billing' as SettingsTab, label: 'Billing', icon: CreditCard },
    { id: 'api_keys' as SettingsTab, label: 'API Keys', icon: Key },
  ];

  return (
    <div className="flex flex-col" style={{ gap: '32px' }}>
      {/* Header */}
      <div className="flex flex-col" style={{ gap: '4px' }}>
        <h1 className="text-[28px] font-bold text-white leading-tight">Settings</h1>
        <p className="text-[14px] text-[#64748B]">Manage your account and preferences</p>
      </div>

      {/* Content */}
      <div className="flex" style={{ gap: '32px' }}>
        {/* Sidebar */}
        <div className="flex flex-col" style={{ width: '240px', gap: '4px' }}>
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center rounded-lg text-left transition-colors ${
                  isActive ? 'bg-[#1E293B]' : 'hover:bg-[#1E293B]/50'
                }`}
                style={{ padding: '12px 16px', gap: '12px' }}
              >
                <Icon
                  className="w-5 h-5"
                  style={{ color: isActive ? '#A855F7' : '#64748B' }}
                />
                <span
                  className="text-[14px] font-medium"
                  style={{ color: isActive ? '#FFFFFF' : '#94A3B8' }}
                >
                  {tab.label}
                </span>
              </button>
            );
          })}
        </div>

        {/* Main Content */}
        <div className="flex-1 flex flex-col" style={{ gap: '32px' }}>
          {/* Notification Settings */}
          {activeTab === 'notifications' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Notification Settings
              </h2>
              <div className="flex flex-col">
                {settings.map((setting, index) => (
                  <div
                    key={setting.id}
                    style={{
                      borderBottom:
                        index < settings.length - 1 ? '1px solid #334155' : 'none',
                    }}
                  >
                    <SettingRow setting={setting} onToggle={handleToggle} />
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Profile Tab */}
          {activeTab === 'profile' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Profile Settings
              </h2>
              <p className="text-[14px] text-[#64748B]">Profile settings will be available soon.</p>
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Security Settings
              </h2>
              <p className="text-[14px] text-[#64748B]">Security settings will be available soon.</p>
            </div>
          )}

          {/* Billing Tab */}
          {activeTab === 'billing' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                Billing Settings
              </h2>
              <p className="text-[14px] text-[#64748B]">Billing settings will be available soon.</p>
            </div>
          )}

          {/* API Keys Tab */}
          {activeTab === 'api_keys' && (
            <div className="rounded-xl bg-[#1E293B]" style={{ padding: '32px' }}>
              <h2 className="text-[18px] font-semibold text-white" style={{ marginBottom: '24px' }}>
                API Keys
              </h2>
              <p className="text-[14px] text-[#64748B]">API key management will be available soon.</p>
            </div>
          )}

          {/* Danger Zone */}
          <div
            className="rounded-xl bg-[#1E293B]"
            style={{ padding: '32px', border: '1px solid rgba(239, 68, 68, 0.25)' }}
          >
            <h2 className="text-[18px] font-semibold text-[#EF4444]" style={{ marginBottom: '16px' }}>
              Danger Zone
            </h2>
            <div className="flex items-center justify-between">
              <div className="flex flex-col" style={{ gap: '2px' }}>
                <span className="text-[14px] font-medium text-white">Delete Account</span>
                <span className="text-[12px] text-[#64748B]">
                  Permanently delete your account and all your data
                </span>
              </div>
              <button
                className="rounded-lg text-[14px] font-medium text-[#EF4444] hover:bg-[#EF4444]/20 transition-colors"
                style={{ padding: '10px 16px', backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
              >
                Delete Account
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
