import { Star, MapPin, Package, Wrench, Database } from 'lucide-react';
import type { Listing } from '../../lib/api';

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

interface ListingCardProps {
  listing: Listing;
  onClick?: () => void;
}

export function ListingCard({ listing, onClick }: ListingCardProps) {
  const type = typeConfig[listing.listing_type] || typeConfig.goods;
  const TypeIcon = type.icon;

  const formatPrice = (amount?: number, currency?: string) => {
    if (!amount) return 'Contact for price';
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    return `${symbol}${amount.toLocaleString()}`;
  };

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
      {/* Header with Type Badge */}
      <div className="flex items-start justify-between">
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
        {listing.quantity > 1 && (
          <span className="text-xs text-[#64748B]">Qty: {listing.quantity}</span>
        )}
      </div>

      {/* Title & Description */}
      <div className="flex flex-col gap-2">
        <h3 className="text-white font-semibold text-base leading-tight line-clamp-2">
          {listing.title}
        </h3>
        <p className="text-[#94A3B8] text-sm line-clamp-2">{listing.description}</p>
      </div>

      {/* Price Section */}
      <div
        className="flex items-center justify-between"
        style={{
          backgroundColor: '#0F172A',
          borderRadius: '8px',
          padding: '12px',
        }}
      >
        <span className="text-[#64748B] text-xs">Price</span>
        <span className="text-base font-bold" style={{ color: '#22C55E' }}>
          {formatPrice(listing.price_amount, listing.price_currency)}
        </span>
      </div>

      {/* Seller Info & Location */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div
            className="w-9 h-9 rounded-full flex items-center justify-center"
            style={{
              background: 'linear-gradient(135deg, #22D3EE 0%, #A855F7 50%, #EC4899 100%)',
            }}
          >
            <span className="text-white text-xs font-semibold">
              {listing.seller_name?.[0]?.toUpperCase() || 'A'}
            </span>
          </div>
          <div className="flex flex-col">
            <span className="text-white text-sm font-medium">
              {listing.seller_name || 'Agent'}
            </span>
            <div className="flex items-center gap-1">
              <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
              <span className="text-[#F59E0B] text-xs font-medium">4.8</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-1 text-[#64748B]">
          <MapPin className="w-3.5 h-3.5" />
          <span className="text-xs">{scopeLabels[listing.geographic_scope] || 'Global'}</span>
        </div>
      </div>
    </div>
  );
}
