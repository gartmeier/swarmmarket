import { Star, MapPin, Package, Wrench, Database, MessageSquare, Clock } from 'lucide-react';
import type { Request } from '../../lib/api';

const typeConfig = {
  goods: { icon: Package, label: 'Goods', color: '#EC4899', bgColor: 'rgba(236, 72, 153, 0.2)' },
  services: { icon: Wrench, label: 'Services', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  data: { icon: Database, label: 'Data', color: '#A855F7', bgColor: 'rgba(168, 85, 247, 0.2)' },
};

const scopeLabels: Record<string, string> = {
  local: 'Local',
  regional: 'Regional',
  national: 'National',
  international: 'International',
};

interface RequestCardProps {
  request: Request;
  onClick?: () => void;
}

export function RequestCard({ request, onClick }: RequestCardProps) {
  const type = typeConfig[request.request_type] || typeConfig.goods;
  const TypeIcon = type.icon;

  const formatBudget = (min?: number, max?: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    if (min && max) {
      return `${symbol}${min.toLocaleString()} - ${symbol}${max.toLocaleString()}`;
    }
    if (max) return `Up to ${symbol}${max.toLocaleString()}`;
    if (min) return `From ${symbol}${min.toLocaleString()}`;
    return 'Open budget';
  };

  const getTimeRemaining = (expiresAt?: string) => {
    if (!expiresAt) return null;
    const now = new Date();
    const expires = new Date(expiresAt);
    const diff = expires.getTime() - now.getTime();
    if (diff <= 0) return 'Expired';
    const hours = Math.floor(diff / (1000 * 60 * 60));
    const days = Math.floor(hours / 24);
    if (days > 0) return `${days}d left`;
    return `${hours}h left`;
  };

  const timeRemaining = getTimeRemaining(request.expires_at);

  return (
    <div
      onClick={onClick}
      className="cursor-pointer transition-all hover:scale-[1.02] hover:shadow-lg"
      style={{
        backgroundColor: '#1E293B',
        borderRadius: '16px',
        border: '1px solid #334155',
        padding: '24px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
      }}
    >
      {/* Header with Type Badge & Offers Count */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          <div
            className="flex items-center gap-1.5"
            style={{
              backgroundColor: type.bgColor,
              borderRadius: '6px',
              padding: '6px 10px',
            }}
          >
            <TypeIcon className="w-3.5 h-3.5" style={{ color: type.color }} />
            <span className="text-xs font-medium" style={{ color: type.color }}>
              {type.label}
            </span>
          </div>
          {request.offer_count > 0 && (
            <div
              className="flex items-center gap-1"
              style={{
                backgroundColor: 'rgba(168, 85, 247, 0.2)',
                borderRadius: '6px',
                padding: '6px 10px',
              }}
            >
              <MessageSquare className="w-3 h-3" style={{ color: '#A855F7' }} />
              <span className="text-xs font-medium" style={{ color: '#A855F7' }}>
                {request.offer_count} {request.offer_count === 1 ? 'Offer' : 'Offers'}
              </span>
            </div>
          )}
        </div>
        {timeRemaining && (
          <div className="flex items-center gap-1 text-[#64748B]">
            <Clock className="w-3.5 h-3.5" />
            <span className="text-xs">{timeRemaining}</span>
          </div>
        )}
      </div>

      {/* Title & Description */}
      <div className="flex flex-col gap-2">
        <h3 className="text-white font-semibold text-base leading-tight line-clamp-2">
          {request.title}
        </h3>
        <p className="text-[#94A3B8] text-sm line-clamp-2">{request.description}</p>
      </div>

      {/* Budget Section */}
      <div
        className="flex items-center justify-between"
        style={{
          backgroundColor: '#0F172A',
          borderRadius: '8px',
          padding: '12px',
        }}
      >
        <span className="text-[#64748B] text-xs">Budget</span>
        <span className="text-base font-bold" style={{ color: '#A855F7' }}>
          {formatBudget(request.budget_min, request.budget_max, request.budget_currency)}
        </span>
      </div>

      {/* Requester Info & Location */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div
            className="w-9 h-9 rounded-full flex items-center justify-center"
            style={{
              background: 'linear-gradient(135deg, #22D3EE 0%, #A855F7 50%, #EC4899 100%)',
            }}
          >
            <span className="text-white text-xs font-semibold">
              {request.requester_name?.[0]?.toUpperCase() || 'A'}
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-white text-sm font-medium">
              {request.requester_name || 'Agent'}
            </span>
            <div className="flex items-center gap-1">
              <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
              <span className="text-[#F59E0B] text-xs font-medium">4.7</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1 text-[#64748B]">
          <MapPin className="w-3.5 h-3.5" />
          <span className="text-xs">{scopeLabels[request.geographic_scope] || 'Global'}</span>
        </div>
      </div>
    </div>
  );
}
