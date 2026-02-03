import { Star, Code, Package, Wrench, Database } from 'lucide-react';
import type { Listing } from '../../lib/api';

const typeConfig: Record<string, { icon: typeof Code; label: string; gradient: string }> = {
  goods: { icon: Package, label: 'Goods', gradient: 'linear-gradient(135deg, #EC4899 0%, #F59E0B 100%)' },
  services: { icon: Wrench, label: 'Services', gradient: 'linear-gradient(135deg, #22D3EE 0%, #A855F7 100%)' },
  data: { icon: Database, label: 'Data', gradient: 'linear-gradient(135deg, #A855F7 0%, #EC4899 100%)' },
};

interface ListingCardProps {
  listing: Listing;
  onClick?: () => void;
}

export function ListingCard({ listing, onClick }: ListingCardProps) {
  const type = typeConfig[listing.listing_type] || typeConfig.services;
  const TypeIcon = type.icon;

  const formatPrice = (amount?: number, currency?: string) => {
    if (!amount) return 'Contact';
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    return `${symbol}${amount.toLocaleString()}`;
  };

  // Extract tags from listing type and geographic scope
  const tags: string[] = [];
  if (listing.listing_type) {
    const typeLabel = typeConfig[listing.listing_type]?.label;
    if (typeLabel) tags.push(typeLabel);
  }
  if (listing.geographic_scope) {
    const scopeLabels: Record<string, string> = {
      local: 'Local',
      regional: 'Regional',
      national: 'National',
      international: 'Global',
    };
    const scopeLabel = scopeLabels[listing.geographic_scope];
    if (scopeLabel) tags.push(scopeLabel);
  }

  return (
    <div
      onClick={onClick}
      className="cursor-pointer transition-all hover:scale-[1.01] hover:border-[#475569]"
      style={{
        backgroundColor: '#0F172A',
        borderRadius: '12px',
        border: '1px solid #334155',
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        gap: '16px',
      }}
    >
      {/* Card Header - Avatar + Agent Info */}
      <div className="flex items-center gap-3" style={{ width: '100%' }}>
        {/* Avatar with gradient */}
        {listing.seller_avatar_url ? (
          <img
            src={listing.seller_avatar_url}
            alt={listing.seller_name || 'Agent'}
            className="w-11 h-11 rounded-full object-cover flex-shrink-0"
            style={{ border: '2px solid #334155' }}
          />
        ) : (
          <div
            className="w-11 h-11 rounded-full flex items-center justify-center flex-shrink-0"
            style={{ background: type.gradient }}
          >
            <TypeIcon className="w-5 h-5 text-white" />
          </div>
        )}
        {/* Agent Info */}
        <div className="flex flex-col gap-0.5 min-w-0 flex-1">
          <span className="text-white text-[15px] font-semibold truncate">
            {listing.title}
          </span>
          <span className="text-[#64748B] text-xs truncate">
            {listing.seller_name || 'Agent'}
          </span>
        </div>
      </div>

      {/* Description */}
      <p
        className="text-[#94A3B8] text-[13px] leading-relaxed line-clamp-3"
        style={{ lineHeight: '1.5' }}
      >
        {listing.description}
      </p>

      {/* Tags */}
      {tags.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          {tags.slice(0, 3).map((tag, index) => (
            <span
              key={index}
              className="text-[#94A3B8] text-[11px] font-medium"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '12px',
                padding: '4px 10px',
              }}
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Footer - Price + Rating */}
      <div className="flex items-center justify-between" style={{ width: '100%' }}>
        <span
          className="text-sm font-semibold"
          style={{ color: '#22C55E', fontFamily: 'JetBrains Mono, monospace' }}
        >
          {formatPrice(listing.price_amount, listing.price_currency)}
        </span>
        <div className="flex items-center gap-1">
          <Star className="w-3.5 h-3.5" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
          <span className="text-[#94A3B8] text-xs">
            {listing.seller_rating ? `${listing.seller_rating.toFixed(1)} (${listing.seller_rating_count || 0})` : '—'}
          </span>
        </div>
      </div>
    </div>
  );
}
