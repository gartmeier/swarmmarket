import { CheckCircle, DollarSign, Bot, Loader2, Activity, ShoppingCart, FileText, Gavel, MessageSquare } from 'lucide-react';
import { useAgentActivity } from '../../../hooks/useDashboard';
import type { ActivityEvent } from '../../../lib/api';

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

function getEventDisplay(event: ActivityEvent): {
  icon: React.ReactNode;
  title: string;
  description: string;
  amount?: string;
  isPositive?: boolean;
} {
  const payload = event.payload;
  const baseIconClasses = 'w-5 h-5';

  switch (event.event_type) {
    // Listing events
    case 'listing.created':
      return {
        icon: <FileText className={`${baseIconClasses} text-[#22D3EE]`} />,
        title: 'Listing created',
        description: `"${payload.title || 'Untitled'}"`,
      };
    case 'listing.updated':
      return {
        icon: <FileText className={`${baseIconClasses} text-[#64748B]`} />,
        title: 'Listing updated',
        description: `"${payload.title || 'Untitled'}"`,
      };
    case 'listing.purchased':
      return {
        icon: <ShoppingCart className={`${baseIconClasses} text-[#22C55E]`} />,
        title: 'Listing sold',
        description: `"${payload.title || 'Item'}" purchased`,
        amount: payload.amount ? `+$${Number(payload.amount).toFixed(2)}` : undefined,
        isPositive: true,
      };

    // Request events
    case 'request.created':
      return {
        icon: <FileText className={`${baseIconClasses} text-[#A855F7]`} />,
        title: 'Request created',
        description: `"${payload.title || 'Untitled'}"`,
      };
    case 'request.updated':
      return {
        icon: <FileText className={`${baseIconClasses} text-[#64748B]`} />,
        title: 'Request updated',
        description: `"${payload.title || 'Untitled'}"`,
      };

    // Offer events
    case 'offer.received':
      return {
        icon: <DollarSign className={`${baseIconClasses} text-[#F59E0B]`} />,
        title: 'Offer received',
        description: `${payload.offerer_name || 'An agent'} made an offer`,
        amount: payload.price ? `$${Number(payload.price).toFixed(2)}` : undefined,
      };
    case 'offer.accepted':
      return {
        icon: <CheckCircle className={`${baseIconClasses} text-[#22C55E]`} />,
        title: 'Offer accepted',
        description: `Your offer was accepted`,
        amount: payload.price ? `$${Number(payload.price).toFixed(2)}` : undefined,
      };
    case 'offer.rejected':
      return {
        icon: <CheckCircle className={`${baseIconClasses} text-[#EF4444]`} />,
        title: 'Offer rejected',
        description: `Your offer was rejected`,
      };

    // Auction events
    case 'auction.started':
      return {
        icon: <Gavel className={`${baseIconClasses} text-[#A855F7]`} />,
        title: 'Auction started',
        description: `"${payload.title || 'Auction'}"`,
      };
    case 'bid.placed':
      return {
        icon: <Gavel className={`${baseIconClasses} text-[#22D3EE]`} />,
        title: 'Bid placed',
        description: `New bid on "${payload.title || 'auction'}"`,
        amount: payload.amount ? `$${Number(payload.amount).toFixed(2)}` : undefined,
      };
    case 'bid.outbid':
      return {
        icon: <Gavel className={`${baseIconClasses} text-[#F59E0B]`} />,
        title: 'You were outbid',
        description: `Someone placed a higher bid`,
      };
    case 'auction.ended':
      return {
        icon: <Gavel className={`${baseIconClasses} text-[#22C55E]`} />,
        title: 'Auction ended',
        description: `"${payload.title || 'Auction'}" has ended`,
      };

    // Transaction events
    case 'transaction.created':
      return {
        icon: <DollarSign className={`${baseIconClasses} text-[#22D3EE]`} />,
        title: 'Transaction started',
        description: `"${payload.title || 'Transaction'}"`,
        amount: payload.amount ? `$${Number(payload.amount).toFixed(2)}` : undefined,
      };
    case 'transaction.escrow_funded':
    case 'escrow.funded':
      return {
        icon: <DollarSign className={`${baseIconClasses} text-[#F59E0B]`} />,
        title: 'Escrow funded',
        description: `Payment held in escrow`,
        amount: payload.amount ? `$${Number(payload.amount).toFixed(2)}` : undefined,
      };
    case 'transaction.delivered':
    case 'delivery.confirmed':
      return {
        icon: <CheckCircle className={`${baseIconClasses} text-[#22D3EE]`} />,
        title: 'Delivery confirmed',
        description: `"${payload.title || 'Item'}" was delivered`,
      };
    case 'transaction.completed':
    case 'payment.released':
      return {
        icon: <CheckCircle className={`${baseIconClasses} text-[#22C55E]`} />,
        title: 'Payment received',
        description: `Transaction completed`,
        amount: payload.amount ? `+$${Number(payload.amount).toFixed(2)}` : undefined,
        isPositive: true,
      };
    case 'transaction.refunded':
      return {
        icon: <DollarSign className={`${baseIconClasses} text-[#EF4444]`} />,
        title: 'Refund processed',
        description: `Transaction was refunded`,
        amount: payload.amount ? `-$${Number(payload.amount).toFixed(2)}` : undefined,
        isPositive: false,
      };
    case 'payment.failed':
    case 'payment.capture_failed':
      return {
        icon: <DollarSign className={`${baseIconClasses} text-[#EF4444]`} />,
        title: 'Payment failed',
        description: `Payment could not be processed`,
      };
    case 'dispute.opened':
      return {
        icon: <MessageSquare className={`${baseIconClasses} text-[#EF4444]`} />,
        title: 'Dispute opened',
        description: `A dispute was raised`,
      };

    // Comment events
    case 'comment.created':
      return {
        icon: <MessageSquare className={`${baseIconClasses} text-[#64748B]`} />,
        title: 'Comment added',
        description: `New comment on "${payload.listing_title || 'listing'}"`,
      };

    // Order book events
    case 'match.found':
    case 'order.filled':
      return {
        icon: <CheckCircle className={`${baseIconClasses} text-[#22C55E]`} />,
        title: 'Order matched',
        description: `Trade executed`,
        amount: payload.price ? `$${Number(payload.price).toFixed(2)}` : undefined,
      };

    default:
      return {
        icon: <Bot className={`${baseIconClasses} text-[#A855F7]`} />,
        title: event.event_type.replace(/[._]/g, ' '),
        description: 'Activity recorded',
      };
  }
}

function ActivityRow({ event }: { event: ActivityEvent }) {
  const display = getEventDisplay(event);

  return (
    <div className="flex items-start justify-between" style={{ padding: '16px 0' }}>
      <div className="flex items-start" style={{ gap: '16px' }}>
        <div
          className="rounded-full bg-[#1E293B] flex items-center justify-center"
          style={{ width: '40px', height: '40px', marginTop: '2px' }}
        >
          {display.icon}
        </div>
        <div className="flex flex-col" style={{ gap: '4px' }}>
          <span className="text-[14px] font-medium text-white">{display.title}</span>
          <span className="text-[12px] text-[#64748B]">{display.description}</span>
          <span className="text-[12px] text-[#475569]">{formatTimeAgo(event.created_at)}</span>
        </div>
      </div>
      {display.amount && (
        <span
          className="font-mono text-[14px] font-semibold"
          style={{ color: display.isPositive ? '#22C55E' : display.isPositive === false ? '#EF4444' : '#64748B' }}
        >
          {display.amount}
        </span>
      )}
    </div>
  );
}

function ActivitySection({
  title,
  events,
}: {
  title: string;
  events: ActivityEvent[];
}) {
  if (events.length === 0) return null;

  return (
    <>
      <div style={{ marginTop: title !== 'Today' ? '24px' : 0, marginBottom: '8px' }}>
        <span className="text-[12px] font-semibold text-[#64748B] uppercase tracking-wider">
          {title}
        </span>
      </div>
      <div className="flex flex-col">
        {events.map((event, index) => (
          <div
            key={event.id}
            style={{
              borderBottom: index < events.length - 1 ? '1px solid #1E293B' : 'none',
            }}
          >
            <ActivityRow event={event} />
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
  const { groupedEvents, loading, total } = useAgentActivity(agentId);

  if (loading) {
    return (
      <div className="flex items-center justify-center flex-1">
        <Loader2 className="w-8 h-8 text-[#A855F7] animate-spin" />
      </div>
    );
  }

  const hasEvents =
    groupedEvents.today.length > 0 ||
    groupedEvents.yesterday.length > 0 ||
    groupedEvents.earlier.length > 0;

  if (!hasEvents) {
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
          Activity will appear here as your agent interacts with the marketplace
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col flex-1">
      {total > 0 && (
        <div className="text-[12px] text-[#64748B] mb-4">
          Showing {groupedEvents.today.length + groupedEvents.yesterday.length + groupedEvents.earlier.length} of {total} events
        </div>
      )}
      <ActivitySection title="Today" events={groupedEvents.today} />
      <ActivitySection title="Yesterday" events={groupedEvents.yesterday} />
      <ActivitySection title="Earlier" events={groupedEvents.earlier} />
    </div>
  );
}
