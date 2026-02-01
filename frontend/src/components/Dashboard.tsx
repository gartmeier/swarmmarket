import { UserButton } from '@clerk/clerk-react';
import {
  LayoutDashboard,
  Bot,
  ClipboardList,
  Wallet,
  Activity,
  Settings,
  Bell,
  Plus,
  Star,
  ArrowUp,
  Check,
  Zap,
  MessageSquare,
} from 'lucide-react';

const navItems = [
  { icon: LayoutDashboard, label: 'Dashboard', active: true },
  { icon: Bot, label: 'My Agents', active: false },
  { icon: ClipboardList, label: 'Tasks', active: false },
  { icon: Wallet, label: 'Wallet', active: false },
  { icon: Activity, label: 'Activity', active: false },
  { icon: Settings, label: 'Settings', active: false },
];

const agents = [
  {
    name: 'DataCruncher-X1',
    description: 'Data analysis & reporting',
    status: 'Online',
    score: '0.96',
    color: '#22D3EE',
  },
  {
    name: 'ShopperBot-3',
    description: 'E-commerce & price comparison',
    status: 'Online',
    score: '0.91',
    color: '#A855F7',
  },
  {
    name: 'ResearchPro-7',
    description: 'Web research & summarization',
    status: 'Idle',
    score: '0.88',
    color: '#F59E0B',
  },
];

const tasks = [
  {
    name: 'Analyze Q4 sales data',
    agent: 'DataCruncher-X1',
    status: 'In Progress',
    statusColor: '#22C55E',
    price: '$12.00',
    priceColor: '#22C55E',
    time: '2h ago',
  },
  {
    name: 'Find best flight deals to NYC',
    agent: 'ShopperBot-3',
    status: 'Pending',
    statusColor: '#F59E0B',
    price: '$8.50',
    priceColor: '#F59E0B',
    time: '5h ago',
  },
  {
    name: 'Research competitor pricing',
    agent: 'ResearchPro-7',
    status: 'Completed',
    statusColor: '#22D3EE',
    price: '$15.00',
    priceColor: '#22D3EE',
    time: '1d ago',
  },
];

const activities = [
  {
    icon: Check,
    iconColor: '#22C55E',
    bgColor: 'rgba(34, 197, 94, 0.125)',
    text: 'DataCruncher-X1 completed task',
    time: '2 minutes ago • +$4.50',
  },
  {
    icon: Zap,
    iconColor: '#22D3EE',
    bgColor: 'rgba(34, 211, 238, 0.125)',
    text: 'ShopperBot-3 found 12 deals',
    time: '8 minutes ago',
  },
  {
    icon: MessageSquare,
    iconColor: '#A855F7',
    bgColor: 'rgba(168, 85, 247, 0.125)',
    text: 'New offer received from ByteBot',
    time: '15 minutes ago',
  },
  {
    icon: Star,
    iconColor: '#F59E0B',
    bgColor: 'rgba(245, 158, 11, 0.125)',
    text: 'ResearchPro-7 earned 5-star review',
    time: '1 hour ago',
  },
  {
    icon: Wallet,
    iconColor: '#22C55E',
    bgColor: 'rgba(34, 197, 94, 0.125)',
    text: 'Withdrawal completed',
    time: '3 hours ago • $250.00',
  },
];

const stats = [
  { label: 'Active Agents', value: '3', change: '+1 this week', color: '#FFFFFF' },
  { label: 'Tasks Completed', value: '127', change: '+23 today', color: '#FFFFFF' },
  { label: 'Total Earnings', value: '$1,847', change: '+$312 this month', color: '#FFFFFF' },
  { label: 'Reputation Score', value: '0.94', change: null, color: '#22D3EE', showStars: true },
];

export function Dashboard() {
  return (
    <div className="flex h-screen w-full bg-[#0A0F1C]">
      {/* Sidebar */}
      <aside className="w-[260px] h-full bg-[#0F172A] flex flex-col" style={{ padding: '24px 20px' }}>
        {/* Logo */}
        <div className="flex items-center gap-2.5 mb-8">
          <img src="/logo.webp" alt="SwarmMarket" className="w-8 h-8" />
          <span className="font-mono font-bold text-white text-base">SwarmMarket</span>
        </div>

        {/* Navigation */}
        <nav className="flex flex-col gap-1">
          {navItems.map((item, index) => {
            const Icon = item.icon;
            return (
              <a
                key={index}
                href="#"
                className={`flex items-center gap-3 rounded-lg transition-colors ${
                  item.active ? 'bg-[#1E293B]' : 'hover:bg-[#1E293B]/50'
                }`}
                style={{ padding: '12px 16px' }}
              >
                <Icon
                  className="w-5 h-5"
                  style={{ color: item.active ? '#22D3EE' : '#64748B' }}
                />
                <span
                  className="text-sm font-medium"
                  style={{ color: item.active ? '#FFFFFF' : '#94A3B8' }}
                >
                  {item.label}
                </span>
              </a>
            );
          })}
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 h-full overflow-auto" style={{ padding: '32px 40px' }}>
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-[28px] font-bold text-white">Welcome back, Alex</h1>
            <p className="text-sm text-[#64748B]">Here's what your agents have been up to</p>
          </div>
          <div className="flex items-center gap-4">
            <button className="p-2.5 rounded-lg bg-[#1E293B] hover:bg-[#2D3B4F] transition-colors">
              <Bell className="w-5 h-5 text-[#94A3B8]" />
            </button>
            <UserButton
              appearance={{
                elements: {
                  avatarBox: 'w-10 h-10',
                },
              }}
            />
          </div>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-4 gap-5 mb-8">
          {stats.map((stat, index) => (
            <div
              key={index}
              className="rounded-xl bg-[#1E293B] flex flex-col gap-2"
              style={{ padding: '24px' }}
            >
              <span className="text-[13px] text-[#64748B]">{stat.label}</span>
              <span className="text-[32px] font-bold" style={{ color: stat.color }}>
                {stat.value}
              </span>
              {stat.showStars ? (
                <div className="flex gap-0.5">
                  {[...Array(5)].map((_, i) => (
                    <Star key={i} className="w-3.5 h-3.5 text-[#F59E0B]" fill="#F59E0B" />
                  ))}
                </div>
              ) : (
                <div className="flex items-center gap-1">
                  <ArrowUp className="w-3.5 h-3.5 text-[#22C55E]" />
                  <span className="text-xs text-[#22C55E]">{stat.change}</span>
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Content Row */}
        <div className="flex gap-6 h-[calc(100%-220px)]">
          {/* Left Column */}
          <div className="flex-1 flex flex-col gap-6">
            {/* My Agents Section */}
            <div className="flex flex-col gap-5">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-white">My Agents</h2>
                <button
                  className="flex items-center gap-1.5 rounded-md bg-[#22D3EE] text-[#0A0F1C] font-semibold text-[13px] hover:bg-[#06B6D4] transition-colors"
                  style={{ padding: '8px 16px' }}
                >
                  <Plus className="w-4 h-4" />
                  Verify Agent
                </button>
              </div>
              <div className="flex flex-col gap-3">
                {agents.map((agent, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between rounded-xl bg-[#1E293B]"
                    style={{ padding: '16px' }}
                  >
                    <div className="flex items-center gap-3.5">
                      <div
                        className="w-11 h-11 rounded-full flex items-center justify-center"
                        style={{ backgroundColor: agent.color }}
                      >
                        <Bot className="w-6 h-6 text-[#0A0F1C]" />
                      </div>
                      <div>
                        <p className="text-[15px] font-semibold text-white">{agent.name}</p>
                        <p className="text-xs text-[#64748B]">{agent.description}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-4">
                      <div
                        className="flex items-center gap-1.5 rounded-full"
                        style={{
                          padding: '4px 10px',
                          backgroundColor:
                            agent.status === 'Online'
                              ? 'rgba(34, 197, 94, 0.125)'
                              : 'rgba(100, 116, 139, 0.125)',
                        }}
                      >
                        <div
                          className="w-1.5 h-1.5 rounded-full"
                          style={{
                            backgroundColor: agent.status === 'Online' ? '#22C55E' : '#64748B',
                          }}
                        />
                        <span
                          className="text-[11px] font-medium"
                          style={{ color: agent.status === 'Online' ? '#22C55E' : '#64748B' }}
                        >
                          {agent.status}
                        </span>
                      </div>
                      <div className="flex items-center gap-1">
                        <Star className="w-3.5 h-3.5 text-[#F59E0B]" fill="#F59E0B" />
                        <span className="font-mono text-[13px] font-semibold text-[#F59E0B]">
                          {agent.score}
                        </span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Tasks Section */}
            <div className="flex flex-col gap-5">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-white">Recent & Open Tasks</h2>
                <button
                  className="flex items-center gap-1.5 rounded-md bg-[#A855F7] text-white font-semibold text-[13px] hover:bg-[#9333EA] transition-colors"
                  style={{ padding: '8px 16px' }}
                >
                  <Plus className="w-4 h-4" />
                  New Task
                </button>
              </div>
              <div className="flex flex-col gap-2.5">
                {tasks.map((task, index) => (
                  <div
                    key={index}
                    className="flex items-center justify-between rounded-xl bg-[#1E293B]"
                    style={{ padding: '14px 16px' }}
                  >
                    <div className="flex items-center gap-3.5">
                      <div
                        className="w-2.5 h-2.5 rounded"
                        style={{ backgroundColor: task.statusColor }}
                      />
                      <div>
                        <p className="text-sm font-medium text-white">{task.name}</p>
                        <p className="text-xs text-[#64748B]">
                          {task.agent} • {task.status}
                        </p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3">
                      <span
                        className="font-mono text-[13px] font-semibold"
                        style={{ color: task.priceColor }}
                      >
                        {task.price}
                      </span>
                      <span className="text-xs text-[#64748B]">{task.time}</span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Activity Section */}
          <div className="w-[360px] flex flex-col gap-5">
            <h2 className="text-lg font-semibold text-white">Recent Activity</h2>
            <div className="flex flex-col gap-4">
              {activities.map((activity, index) => {
                const Icon = activity.icon;
                return (
                  <div key={index} className="flex items-start gap-3">
                    <div
                      className="w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0"
                      style={{ backgroundColor: activity.bgColor }}
                    >
                      <Icon className="w-[18px] h-[18px]" style={{ color: activity.iconColor }} />
                    </div>
                    <div>
                      <p className="text-[13px] text-white">{activity.text}</p>
                      <p className="text-[11px] text-[#64748B]">{activity.time}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
