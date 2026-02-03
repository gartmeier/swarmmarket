import { Star, Gavel, TrendingDown, Lock, Radio } from 'lucide-react';
import type { Auction } from '../../lib/api';

const auctionTypeConfig: Record<string, { icon: typeof Gavel; label: string; gradient: string }> = {
  english: { icon: Gavel, label: 'English', gradient: 'linear-gradient(135deg, #EC4899 0%, #A855F7 100%)' },
  dutch: { icon: TrendingDown, label: 'Dutch', gradient: 'linear-gradient(135deg, #F59E0B 0%, #EC4899 100%)' },
  sealed: { icon: Lock, label: 'Sealed', gradient: 'linear-gradient(135deg, #A855F7 0%, #22D3EE 100%)' },
  continuous: { icon: Radio, label: 'Continuous', gradient: 'linear-gradient(135deg, #22C55E 0%, #22D3EE 100%)' },
};

interface AuctionCardProps {
  auction: Auction;
  onClick?: () => void;
}

export function AuctionCard({ auction, onClick }: AuctionCardProps) {
  const type = auctionTypeConfig[auction.auction_type] || auctionTypeConfig.english;
  const TypeIcon = type.icon;

  const formatPrice = (amount: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    return `${symbol}${amount.toLocaleString()}`;
  };

  // Extract tags from auction type
  const tags: string[] = [];
  if (auction.auction_type) {
    const typeLabel = auctionTypeConfig[auction.auction_type]?.label;
    if (typeLabel) tags.push(`${typeLabel} Auction`);
  }
  if (auction.bid_count > 0) {
    tags.push(`${auction.bid_count} bid${auction.bid_count > 1 ? 's' : ''}`);
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
      {/* Card Header - Avatar + Title */}
      <div className="flex items-center gap-3" style={{ width: '100%' }}>
        {/* Avatar with gradient */}
        {auction.seller_avatar_url ? (
          <img
            src={auction.seller_avatar_url}
            alt={auction.seller_name || 'Agent'}
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
        {/* Auction Info */}
        <div className="flex flex-col gap-0.5 min-w-0 flex-1">
          <span className="text-white text-[15px] font-semibold truncate">
            {auction.title}
          </span>
          <span className="text-[#64748B] text-xs truncate">
            {auction.seller_name || 'Agent'}
          </span>
        </div>
      </div>

      {/* Description */}
      <p
        className="text-[#94A3B8] text-[13px] leading-relaxed line-clamp-3"
        style={{ lineHeight: '1.5' }}
      >
        {auction.description}
      </p>

      {/* Tags */}
      {tags.length > 0 && (
        <div className="flex items-center gap-2 flex-wrap">
          {tags.slice(0, 3).map((tag, index) => (
            <span
              key={index}
              className="text-[#94A3B8] text-[11px] font-medium flex items-center gap-1"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '12px',
                padding: '4px 10px',
              }}
            >
              {tag.includes('bid') && <Gavel className="w-3 h-3" style={{ color: '#EC4899' }} />}
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Footer - Current Bid + Rating */}
      <div className="flex items-center justify-between" style={{ width: '100%' }}>
        <span
          className="text-sm font-semibold"
          style={{ color: '#EC4899', fontFamily: 'JetBrains Mono, monospace' }}
        >
          {formatPrice(auction.current_price || auction.starting_price, auction.currency)}
        </span>
        <div className="flex items-center gap-1">
          <Star className="w-3.5 h-3.5" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
          <span className="text-[#94A3B8] text-xs">
            {auction.seller_rating ? auction.seller_rating.toFixed(1) : '—'}
          </span>
        </div>
      </div>
    </div>
  );
}
