import { MessageCircle, Globe, MapPin, Timer } from 'lucide-react';
import type { Request } from '../../lib/api';

const typeConfig: Record<string, { label: string; color: string; bgColor: string; gradient: string }> = {
  goods: { label: 'Goods', color: '#EC4899', bgColor: '#EC489920', gradient: 'linear-gradient(135deg, #EC4899 0%, #F59E0B 100%)' },
  services: { label: 'Services', color: '#A855F7', bgColor: '#A855F720', gradient: 'linear-gradient(135deg, #A855F7 0%, #22D3EE 100%)' },
  data: { label: 'Data', color: '#22D3EE', bgColor: '#22D3EE20', gradient: 'linear-gradient(135deg, #22D3EE 0%, #22C55E 100%)' },
};

const scopeConfig: Record<string, { icon: typeof Globe; label: string }> = {
  local: { icon: MapPin, label: 'Local' },
  regional: { icon: MapPin, label: 'Regional' },
  national: { icon: MapPin, label: 'National' },
  international: { icon: Globe, label: 'International' },
};

interface RequestCardProps {
  request: Request;
  onClick?: () => void;
}

export function RequestCard({ request, onClick }: RequestCardProps) {
  const type = typeConfig[request.request_type] || typeConfig.services;
  const scope = scopeConfig[request.geographic_scope] || scopeConfig.international;
  const ScopeIcon = scope.icon;

  const formatBudget = (min?: number, max?: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    if (min && max) {
      return `${symbol}${min.toLocaleString()} - ${symbol}${max.toLocaleString()}`;
    }
    if (max) return `Up to ${symbol}${max.toLocaleString()}`;
    if (min) return `From ${symbol}${min.toLocaleString()}`;
    return 'Open';
  };

  // Calculate time remaining
  const getTimeRemaining = () => {
    if (!request.expires_at) return null;
    const now = new Date();
    const expires = new Date(request.expires_at);
    const diffMs = expires.getTime() - now.getTime();
    if (diffMs <= 0) return 'Expired';
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    if (diffDays > 0) return `Ends in ${diffDays}d`;
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    if (diffHours > 0) return `Ends in ${diffHours}h`;
    return 'Ending soon';
  };

  const timeRemaining = getTimeRemaining();
  const hasOffers = request.offer_count > 0;

  return (
    <div
      onClick={onClick}
      className="cursor-pointer transition-all hover:scale-[1.01] hover:border-[#475569]"
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
      {/* Card Top - Type Badge + Offers Badge */}
      <div className="flex items-center justify-between" style={{ width: '100%' }}>
        {/* Type Badge */}
        <span
          className="text-xs font-medium"
          style={{
            color: type.color,
            backgroundColor: type.bgColor,
            borderRadius: '6px',
            padding: '4px 10px',
          }}
        >
          {type.label}
        </span>
        {/* Offers Badge */}
        <div
          className="flex items-center gap-1"
          style={{
            backgroundColor: hasOffers ? '#22C55E20' : '#64748B20',
            borderRadius: '6px',
            padding: '4px 10px',
          }}
        >
          <MessageCircle
            className="w-3 h-3"
            style={{ color: hasOffers ? '#22C55E' : '#64748B' }}
          />
          <span
            className="text-xs font-medium"
            style={{ color: hasOffers ? '#22C55E' : '#64748B' }}
          >
            {request.offer_count} Offer{request.offer_count !== 1 ? 's' : ''}
          </span>
        </div>
      </div>

      {/* Card Content - Title + Description */}
      <div className="flex flex-col gap-2">
        <h3 className="text-white text-base font-semibold line-clamp-2">
          {request.title}
        </h3>
        <p className="text-[#94A3B8] text-sm line-clamp-2">
          {request.description}
        </p>
      </div>

      {/* Budget Section */}
      <div
        className="flex flex-col gap-1"
        style={{
          backgroundColor: '#0F172A',
          borderRadius: '8px',
          padding: '12px',
        }}
      >
        {/* Budget Row */}
        <div className="flex items-center justify-between">
          <span className="text-xs text-[#64748B]">Budget</span>
          <span className="text-base font-semibold" style={{ color: '#22C55E' }}>
            {formatBudget(request.budget_min, request.budget_max, request.budget_currency)}
          </span>
        </div>
        {/* Scope Row */}
        <div className="flex items-center gap-1.5">
          <ScopeIcon className="w-3 h-3 text-[#64748B]" />
          <span className="text-xs text-[#64748B]">{scope.label}</span>
        </div>
      </div>

      {/* Requester Section */}
      <div className="flex items-center justify-between" style={{ width: '100%' }}>
        {/* Left - Avatar + Info */}
        <div className="flex items-center gap-3">
          {request.requester_avatar_url ? (
            <img
              src={request.requester_avatar_url}
              alt={request.requester_name || 'Agent'}
              className="w-9 h-9 rounded-full object-cover flex-shrink-0"
              style={{ border: '2px solid #334155' }}
            />
          ) : (
            <div
              className="w-9 h-9 rounded-full flex-shrink-0"
              style={{ background: type.gradient }}
            />
          )}
          <div className="flex flex-col gap-0.5">
            <span className="text-white text-sm font-medium">
              {request.requester_name || 'Agent'}
            </span>
            <span className="text-[#64748B] text-xs">
              {request.requester_rating ? `★ ${request.requester_rating.toFixed(1)}` : '@basic'}
            </span>
          </div>
        </div>
        {/* Right - Expiry */}
        {timeRemaining && (
          <div className="flex items-center gap-1.5">
            <Timer className="w-3.5 h-3.5" style={{ color: '#F59E0B' }} />
            <span className="text-[13px] font-medium" style={{ color: '#F59E0B' }}>
              {timeRemaining}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
