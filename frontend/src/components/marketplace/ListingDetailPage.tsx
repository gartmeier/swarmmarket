import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  Star,
  Package,
  Wrench,
  Database,
  Clock,
  Wallet,
  Calendar,
  Globe,
  ShieldCheck,
  MessageCircle,
  Share2,
  BadgeCheck,
  Timer,
  ShoppingCart,
} from 'lucide-react';
import { api } from '../../lib/api';
import type { Listing, AgentPublicProfile } from '../../lib/api';

const typeConfig = {
  goods: { icon: Package, label: 'Goods', color: '#EC4899', bgColor: 'rgba(236, 72, 153, 0.2)' },
  services: { icon: Wrench, label: 'Services', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  data: { icon: Database, label: 'Data', color: '#A855F7', bgColor: 'rgba(168, 85, 247, 0.2)' },
};

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  active: { label: 'Available', color: '#22C55E', bgColor: 'rgba(34, 197, 94, 0.2)' },
  draft: { label: 'Draft', color: '#64748B', bgColor: 'rgba(100, 116, 139, 0.2)' },
  paused: { label: 'Paused', color: '#F59E0B', bgColor: 'rgba(245, 158, 11, 0.2)' },
  sold: { label: 'Sold', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  expired: { label: 'Expired', color: '#EF4444', bgColor: 'rgba(239, 68, 68, 0.2)' },
};

const scopeLabels: Record<string, string> = {
  local: 'Local',
  regional: 'Regional',
  national: 'National',
  international: 'International',
};

export function ListingDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [listing, setListing] = useState<Listing | null>(null);
  const [sellerProfile, setSellerProfile] = useState<AgentPublicProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [quantity, setQuantity] = useState('1');

  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        const listingData = await api.getListing(id);
        setListing(listingData);

        // Fetch seller profile for stats
        if (listingData.seller_id) {
          try {
            const profile = await api.getAgentPublicProfile(listingData.seller_id);
            setSellerProfile(profile);
          } catch (err) {
            console.error('Failed to fetch seller profile:', err);
          }
        }
      } catch (error) {
        console.error('Failed to fetch listing:', error);
      }
      setLoading(false);
    };

    fetchData();
  }, [id]);

  const formatPrice = (amount?: number, currency?: string) => {
    if (!amount) return 'Contact for price';
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    return `${symbol}${amount.toLocaleString()}`;
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'No expiration';
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getTimeAgo = (dateString: string) => {
    const now = new Date();
    const date = new Date(dateString);
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMins / 60);
    const diffDays = Math.floor(diffHours / 24);
    const diffWeeks = Math.floor(diffDays / 7);

    if (diffWeeks > 0) return `${diffWeeks} week${diffWeeks > 1 ? 's' : ''} ago`;
    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    if (diffMins > 0) return `${diffMins} min${diffMins > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-[#22D3EE]" />
      </div>
    );
  }

  if (!listing) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-[#94A3B8]">Listing not found</p>
        <button
          onClick={() => navigate(-1)}
          className="text-[#22D3EE] hover:underline"
        >
          Go back
        </button>
      </div>
    );
  }

  const type = typeConfig[listing.listing_type] || typeConfig.goods;
  const TypeIcon = type.icon;
  const status = statusConfig[listing.status] || statusConfig.active;

  const totalPrice = listing.price_amount ? listing.price_amount * parseInt(quantity || '1') : 0;

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate(-1)}
            className="w-10 h-10 rounded-lg flex items-center justify-center transition-colors hover:bg-[#334155]"
            style={{ backgroundColor: '#1E293B', border: '1px solid #334155' }}
          >
            <ArrowLeft className="w-5 h-5 text-[#94A3B8]" />
          </button>
          <div className="flex items-center gap-2 text-sm">
            <span className="text-[#64748B]">Marketplace</span>
            <span className="text-[#64748B]">/</span>
            <span className="text-[#64748B]">Listings</span>
            <span className="text-[#64748B]">/</span>
            <span className="text-white font-medium truncate max-w-[200px]">{listing.title}</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <button
            className="flex items-center gap-2 px-4 h-10 rounded-lg transition-colors hover:bg-[#334155]"
            style={{ backgroundColor: '#1E293B', border: '1px solid #334155' }}
          >
            <Share2 className="w-4 h-4 text-[#94A3B8]" />
            <span className="text-white text-sm font-medium">Share</span>
          </button>
        </div>
      </div>

      {/* Content */}
      <div className="flex gap-6">
        {/* Left Column */}
        <div className="flex-1 flex flex-col gap-6">
          {/* Main Card */}
          <div
            className="flex flex-col gap-5"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            {/* Badges */}
            <div className="flex items-center gap-2 flex-wrap">
              <div
                className="flex items-center gap-1.5"
                style={{ backgroundColor: type.bgColor, borderRadius: '6px', padding: '4px 10px' }}
              >
                <TypeIcon className="w-3 h-3" style={{ color: type.color }} />
                <span className="text-xs font-medium" style={{ color: type.color }}>
                  {type.label}
                </span>
              </div>
              <div
                className="flex items-center gap-1.5"
                style={{ backgroundColor: status.bgColor, borderRadius: '6px', padding: '4px 10px' }}
              >
                <span className="text-xs font-medium" style={{ color: status.color }}>
                  {status.label}
                </span>
              </div>
            </div>

            {/* Title & Description */}
            <div className="flex flex-col gap-3">
              <h1 className="text-2xl font-semibold text-white">{listing.title}</h1>
              <p className="text-[#94A3B8] text-sm leading-relaxed">{listing.description}</p>
            </div>

            {/* Timestamps */}
            <div className="flex items-center gap-3 text-sm text-[#64748B]">
              <Clock className="w-3.5 h-3.5" />
              <span>Listed {getTimeAgo(listing.created_at)}</span>
              {listing.updated_at !== listing.created_at && (
                <>
                  <span className="w-1 h-1 rounded-full bg-[#64748B]" />
                  <span>Updated {getTimeAgo(listing.updated_at)}</span>
                </>
              )}
            </div>

            {/* Meta Grid */}
            <div className={`grid grid-cols-2 gap-4 ${listing.expires_at ? 'lg:grid-cols-5' : 'lg:grid-cols-4'}`}>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Wallet className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Price</span>
                </div>
                <span className="text-lg font-semibold text-[#22C55E]">
                  {formatPrice(listing.price_amount, listing.price_currency)}
                </span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Calendar className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Availability</span>
                </div>
                <span className="text-lg font-semibold text-white">Immediate</span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Package className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Quantity</span>
                </div>
                <span className="text-lg font-semibold text-white">
                  {listing.quantity > 999 ? 'Unlimited' : listing.quantity}
                </span>
              </div>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Globe className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Scope</span>
                </div>
                <span className="text-lg font-semibold text-white">
                  {scopeLabels[listing.geographic_scope] || 'Global'}
                </span>
              </div>
              {listing.expires_at && (
                <div
                  className="flex flex-col gap-2"
                  style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
                >
                  <div className="flex items-center gap-1.5">
                    <Timer className="w-3.5 h-3.5 text-[#64748B]" />
                    <span className="text-xs font-medium text-[#64748B]">Expires</span>
                  </div>
                  <span className="text-lg font-semibold text-[#F59E0B]">
                    {formatDate(listing.expires_at)}
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Right Column */}
        <div className="w-[380px] flex flex-col gap-5">
          {/* Purchase Card */}
          <div
            className="flex flex-col gap-5"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white">Purchase Service</h3>
              <div className="flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-[#22C55E]" />
                <span className="text-xs font-medium text-[#22C55E]">Available</span>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <span className="text-[#94A3B8] text-sm font-medium">Quantity</span>
              <div className="flex gap-3">
                <div
                  className="flex-1 flex items-center gap-2"
                  style={{
                    backgroundColor: '#0F172A',
                    borderRadius: '8px',
                    border: '1px solid #334155',
                    padding: '0 16px',
                    height: '48px',
                  }}
                >
                  <span className="text-[#64748B] font-medium">Qty:</span>
                  <input
                    type="number"
                    min="1"
                    max={listing.quantity}
                    value={quantity}
                    onChange={(e) => setQuantity(e.target.value)}
                    className="flex-1 bg-transparent border-none outline-none text-white font-medium w-20"
                  />
                </div>
                <button
                  className="px-6 h-12 rounded-lg font-semibold text-white flex items-center gap-2"
                  style={{
                    background: 'linear-gradient(90deg, #22D3EE 0%, #A855F7 100%)',
                  }}
                >
                  <ShoppingCart className="w-4 h-4" />
                  Buy Now
                </button>
              </div>
            </div>

            <p className="text-xs text-[#64748B]">
              Total: {formatPrice(totalPrice, listing.price_currency)} ({quantity} × {formatPrice(listing.price_amount, listing.price_currency)})
            </p>
          </div>

          {/* Seller Card */}
          <div
            className="flex flex-col gap-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <span className="text-xs font-medium text-[#64748B]">Seller</span>

            <div className="flex items-center gap-3">
              {listing.seller_avatar_url ? (
                <img
                  src={listing.seller_avatar_url}
                  alt={listing.seller_name || 'Agent'}
                  className="w-12 h-12 rounded-full object-cover"
                />
              ) : (
                <div
                  className="w-12 h-12 rounded-full flex items-center justify-center"
                  style={{
                    background: 'linear-gradient(135deg, #F59E0B 0%, #EF4444 100%)',
                  }}
                >
                  <span className="text-white font-semibold">
                    {listing.seller_name?.[0]?.toUpperCase() || 'A'}
                  </span>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5">
                  <span className="text-white font-semibold">{listing.seller_name || 'Agent'}</span>
                  <BadgeCheck className="w-4 h-4 text-[#22D3EE]" />
                </div>
                <div className="flex items-center gap-1">
                  <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
                  <span className="text-[#F59E0B] text-xs font-medium">
                    {listing.seller_rating ? listing.seller_rating.toFixed(1) : '—'}
                  </span>
                  {listing.seller_rating_count !== undefined && listing.seller_rating_count > 0 && (
                    <span className="text-[#64748B] text-xs">
                      ({listing.seller_rating_count} review{listing.seller_rating_count !== 1 ? 's' : ''})
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Seller Stats */}
            <div className="grid grid-cols-3 gap-4">
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-white">
                  {sellerProfile?.active_listings ?? '—'}
                </span>
                <span className="text-xs text-[#64748B]">Listings</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-[#22C55E]">
                  {sellerProfile?.successful_trades ?? '—'}
                </span>
                <span className="text-xs text-[#64748B]">Trades</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-white">
                  {sellerProfile?.total_transactions ?? '—'}
                </span>
                <span className="text-xs text-[#64748B]">Total Txns</span>
              </div>
            </div>

            {/* Trust Score */}
            <div
              className="flex items-center gap-3"
              style={{ backgroundColor: '#0F172A', borderRadius: '10px', padding: '12px' }}
            >
              <ShieldCheck className="w-5 h-5 text-[#22D3EE]" />
              <div className="flex flex-col gap-0.5">
                <span className="text-xs text-[#94A3B8]">Trust Score</span>
                <span className="text-base font-semibold text-[#22D3EE]">
                  {sellerProfile?.trust_score?.toFixed(2) ?? '—'} / 1.0
                </span>
              </div>
            </div>

            <button
              className="w-full h-11 rounded-lg flex items-center justify-center gap-2 transition-colors hover:bg-[#334155]"
              style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}
            >
              <MessageCircle className="w-4 h-4 text-[#94A3B8]" />
              <span className="text-white text-sm font-medium">Contact Seller</span>
            </button>
          </div>

        </div>
      </div>
    </div>
  );
}