import { useMemo } from 'react';
import { CheckCircle, DollarSign, Bot, Loader2, Activity } from 'lucide-react';
import { useAgentTransactions } from '../../../hooks/useDashboard';

interface ActivityItem {
  id: string;
  type: 'task_completed' | 'task_started' | 'payment_received' | 'escrow_funded';
  title: string;
  description: string;
  time: string;
  amount?: string;
  isPositive?: boolean;
  timestamp: Date;
}

function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;

  return date.toLocaleDateString();
}

function ActivityIcon({ type }: { type: ActivityItem['type'] }) {
  const baseClasses = 'w-5 h-5';

  switch (type) {
    case 'task_completed':
    case 'payment_received':
      return <CheckCircle className={`${baseClasses} text-[#22C55E]`} />;
    case 'task_started':
      return <CheckCircle className={`${baseClasses} text-[#22D3EE]`} />;
    case 'escrow_funded':
      return <DollarSign className={`${baseClasses} text-[#F59E0B]`} />;
    default:
      return <Bot className={`${baseClasses} text-[#A855F7]`} />;
  }
}

function ActivityRow({ item }: { item: ActivityItem }) {
  return (
    <div className="flex items-start justify-between" style={{ padding: '16px 0' }}>
      <div className="flex items-start" style={{ gap: '16px' }}>
        <div
          className="rounded-full bg-[#1E293B] flex items-center justify-center"
          style={{ width: '40px', height: '40px', marginTop: '2px' }}
        >
          <ActivityIcon type={item.type} />
        </div>
        <div className="flex flex-col" style={{ gap: '4px' }}>
          <span className="text-[14px] font-medium text-white">{item.title}</span>
          <span className="text-[12px] text-[#64748B]">{item.description}</span>
          <span className="text-[12px] text-[#475569]">{item.time}</span>
        </div>
      </div>
      {item.amount && (
        <span
          className="font-mono text-[14px] font-semibold"
          style={{ color: item.isPositive ? '#22C55E' : '#EF4444' }}
        >
          {item.amount}
        </span>
      )}
    </div>
  );
}

function ActivitySection({
  title,
  items,
}: {
  title: string;
  items: ActivityItem[];
}) {
  if (items.length === 0) return null;

  return (
    <>
      <div style={{ marginTop: title !== 'Today' ? '24px' : 0, marginBottom: '8px' }}>
        <span className="text-[12px] font-semibold text-[#64748B] uppercase tracking-wider">
          {title}
        </span>
      </div>
      <div className="flex flex-col">
        {items.map((item, index) => (
          <div
            key={item.id}
            style={{
              borderBottom: index < items.length - 1 ? '1px solid #1E293B' : 'none',
            }}
          >
            <ActivityRow item={item} />
          </div>
        ))}
      </div>
    </>
  );
}

interface ActivityTabProps {
  agentId: string;
}

export function ActivityTab({ agentId }: ActivityTabProps) {
  const { transactions, loading } = useAgentTransactions(agentId);

  const activities = useMemo(() => {
    const items: ActivityItem[] = [];

    transactions.forEach((tx) => {
      if (tx.status === 'completed') {
        items.push({
          id: `completed-${tx.id}`,
          type: 'payment_received',
          title: 'Payment received',
          description: `From ${tx.buyer_name || 'buyer'} for "${tx.title}"`,
          time: formatTimeAgo(tx.completed_at || tx.created_at),
          amount: `+$${tx.amount.toFixed(2)}`,
          isPositive: true,
          timestamp: new Date(tx.completed_at || tx.created_at),
        });
      }

      if (tx.status === 'escrow_funded' || tx.status === 'delivered') {
        items.push({
          id: `started-${tx.id}`,
          type: 'task_started',
          title: 'Task started',
          description: `"${tx.title}" - escrow funded`,
          time: formatTimeAgo(tx.funded_at || tx.created_at),
          timestamp: new Date(tx.funded_at || tx.created_at),
        });
      }
    });

    return items.sort((a, b) => b.timestamp.getTime() - a.timestamp.getTime());
  }, [transactions]);

  const groupedActivities = useMemo(() => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    const todayItems: ActivityItem[] = [];
    const yesterdayItems: ActivityItem[] = [];
    const earlierItems: ActivityItem[] = [];

    activities.forEach((item) => {
      const itemDate = new Date(item.timestamp);
      itemDate.setHours(0, 0, 0, 0);

      if (itemDate.getTime() === today.getTime()) {
        todayItems.push(item);
      } else if (itemDate.getTime() === yesterday.getTime()) {
        yesterdayItems.push(item);
      } else {
        earlierItems.push(item);
      }
    });

    return { today: todayItems, yesterday: yesterdayItems, earlier: earlierItems };
  }, [activities]);

  if (loading) {
    return (
      <div className="flex items-center justify-center flex-1">
        <Loader2 className="w-8 h-8 text-[#A855F7] animate-spin" />
      </div>
    );
  }

  if (activities.length === 0) {
    return (
      <div
        className="flex-1 rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
        style={{ padding: '80px 16px' }}
      >
        <Activity className="w-12 h-12 text-[#64748B]" style={{ marginBottom: '16px' }} />
        <p className="text-[16px] font-medium text-white" style={{ marginBottom: '4px' }}>
          No activity yet
        </p>
        <p className="text-[14px] text-[#64748B]">
          Activity history will appear here
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col flex-1">
      <ActivitySection title="Today" items={groupedActivities.today} />
      <ActivitySection title="Yesterday" items={groupedActivities.yesterday} />
      <ActivitySection title="Earlier" items={groupedActivities.earlier} />
    </div>
  );
}
