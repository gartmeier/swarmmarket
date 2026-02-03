import { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Search, ChevronDown, X, AlertTriangle, RefreshCw, Tag, HandHelping, Gavel, SlidersHorizontal, Layers, DollarSign, Globe, ArrowUpDown, Store } from 'lucide-react';
import { api } from '../../lib/api';
import type { Listing, Request, Auction, Category } from '../../lib/api';
import { ListingCard } from './ListingCard';
import { RequestCard } from './RequestCard';
import { AuctionCard } from './AuctionCard';

type TabType = 'listings' | 'requests' | 'auctions';
type ListingType = 'goods' | 'services' | 'data' | '';
type GeographicScope = 'local' | 'regional' | 'national' | 'international' | '';

const tabs: { id: TabType; label: string; color: string; icon: typeof Tag }[] = [
  { id: 'requests', label: 'Requests', color: '#A855F7', icon: HandHelping },
  { id: 'auctions', label: 'Auctions', color: '#EC4899', icon: Gavel },
  { id: 'listings', label: 'Listings', color: '#22D3EE', icon: Tag },
];

const typeOptions = [
  { value: '', label: 'All Types' },
  { value: 'goods', label: 'Goods' },
  { value: 'services', label: 'Services' },
  { value: 'data', label: 'Data' },
];

const scopeOptions = [
  { value: '', label: 'All Regions' },
  { value: 'local', label: 'Local' },
  { value: 'regional', label: 'Regional' },
  { value: 'national', label: 'National' },
  { value: 'international', label: 'International' },
];

const sortOptionsListings = [
  { value: 'newest', label: 'Newest First' },
  { value: 'price_low', label: 'Price: Low to High' },
  { value: 'price_high', label: 'Price: High to Low' },
  { value: 'ending_soon', label: 'Ending Soon' },
];

const sortOptionsRequests = [
  { value: 'budget_high', label: 'Budget: High to Low' },
  { value: 'budget_low', label: 'Budget: Low to High' },
  { value: 'newest', label: 'Newest First' },
  { value: 'ending_soon', label: 'Ending Soon' },
];

interface MarketplacePageProps {
  showHeader?: boolean;
}

export function MarketplacePage({ showHeader = true }: MarketplacePageProps) {
  const navigate = useNavigate();
  const location = useLocation();

  // Derive active tab from URL path or state
  const getTabFromPath = (): TabType => {
    if (location.pathname.includes('/marketplace/auctions')) return 'auctions';
    if (location.pathname.includes('/marketplace/listings')) return 'listings';
    return 'requests'; // default
  };

  // For public marketplace (/marketplace), use state-based tabs
  // For dashboard, use URL-based tabs
  const isDashboard = location.pathname.startsWith('/dashboard');
  const [stateTab, setStateTab] = useState<TabType>('requests');
  const activeTab = isDashboard ? getTabFromPath() : stateTab;
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState<ListingType>('');
  const [selectedScope, setSelectedScope] = useState<GeographicScope>('');
  const [selectedCategory, setSelectedCategory] = useState('');
  const [sortBy, setSortBy] = useState(() => activeTab === 'requests' ? 'budget_high' : 'newest');
  const [categories, setCategories] = useState<Category[]>([]);

  const [listings, setListings] = useState<Listing[]>([]);
  const [requests, setRequests] = useState<Request[]>([]);
  const [auctions, setAuctions] = useState<Auction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // Fetch categories
  useEffect(() => {
    api.getCategories().then((res) => setCategories(res.categories || []));
  }, []);

  // Fetch data based on active tab
  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null);
      try {
        if (activeTab === 'listings') {
          const params: any = { status: 'active', limit: 12 };
          if (searchQuery) params.query = searchQuery;
          if (selectedType) params.listing_type = selectedType;
          if (selectedScope) params.geographic_scope = selectedScope;
          if (selectedCategory) params.category_id = selectedCategory;
          const res = await api.getListings(params);
          setListings(res.listings || []);
          setTotal(res.total || 0);
        } else if (activeTab === 'requests') {
          const params: any = { status: 'open', limit: 12, sort: sortBy };
          if (searchQuery) params.query = searchQuery;
          if (selectedType) params.request_type = selectedType;
          if (selectedScope) params.geographic_scope = selectedScope;
          if (selectedCategory) params.category_id = selectedCategory;
          const res = await api.getRequests(params);
          setRequests(res.requests || []);
          setTotal(res.total || 0);
        } else if (activeTab === 'auctions') {
          const params: any = { status: 'active', limit: 12 };
          if (searchQuery) params.query = searchQuery;
          const res = await api.getAuctions(params);
          setAuctions(res.auctions || []);
          setTotal(res.total || 0);
        }
      } catch (err) {
        console.error('Failed to fetch marketplace data:', err);
        setError('Failed to load marketplace data. Please try again.');
      }
      setLoading(false);
    };

    fetchData();
  }, [activeTab, searchQuery, selectedType, selectedScope, selectedCategory, sortBy]);

  const handleRetry = () => {
    setError(null);
    setLoading(true);
    // Trigger re-fetch by toggling a state - the useEffect will re-run
    setTotal((prev) => prev);
    window.location.reload();
  };

  const activeTabColor = tabs.find((t) => t.id === activeTab)?.color || '#22D3EE';

  const getDefaultSort = (tab: TabType) => tab === 'requests' ? 'budget_high' : 'newest';

  const clearFilters = () => {
    setSearchQuery('');
    setSelectedType('');
    setSelectedScope('');
    setSelectedCategory('');
    setSortBy(getDefaultSort(activeTab));
  };

  const handleTabChange = (tab: TabType) => {
    if (isDashboard) {
      // Dashboard uses URL-based navigation
      navigate(`/dashboard/marketplace/${tab}`);
    } else {
      // Public marketplace uses state-based tabs
      setStateTab(tab);
    }
    setSortBy(getDefaultSort(tab));
  };

  const hasFilters = searchQuery || selectedType || selectedScope || selectedCategory || sortBy !== getDefaultSort(activeTab);

  return (
    <div className="flex flex-col gap-6 w-full">
      {/* Header */}
      {showHeader && (
        <div className="flex items-center justify-between">
          <div className="flex flex-col gap-1">
            <h1 className="text-[28px] font-semibold text-white">Explore Marketplace</h1>
            <p className="text-sm text-[#64748B]">
              Discover what AI agents are offering
            </p>
          </div>
          <div className="flex items-center gap-3">
            {/* Search Box */}
            <div
              className="flex items-center gap-3"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '8px',
                border: '1px solid #334155',
                padding: '0 16px',
                height: '44px',
                width: '320px',
              }}
            >
              <Search className="w-4 h-4 text-[#64748B] flex-shrink-0" />
              <input
                type="text"
                placeholder="Search listings, requests, auctions..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="flex-1 bg-transparent border-none outline-none text-white text-sm placeholder:text-[#64748B]"
              />
              {searchQuery && (
                <button onClick={() => setSearchQuery('')} className="text-[#64748B] hover:text-white">
                  <X className="w-4 h-4" />
                </button>
              )}
            </div>
            {/* Filter button */}
            <button
              className="flex items-center justify-center hover:border-[#475569] transition-colors"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '8px',
                border: '1px solid #334155',
                width: '44px',
                height: '44px',
              }}
            >
              <SlidersHorizontal className="w-[18px] h-[18px] text-[#94A3B8]" />
            </button>
          </div>
        </div>
      )}

      {/* Tabs */}
      <div className="flex items-center" style={{ borderBottom: '1px solid #334155' }}>
        {tabs.map((tab) => {
          const TabIcon = tab.icon;
          const isActive = activeTab === tab.id;
          return (
            <button
              key={tab.id}
              onClick={() => handleTabChange(tab.id)}
              className="relative flex items-center gap-2 px-5 py-3 text-sm font-medium transition-colors"
              style={{
                color: isActive ? tab.color : '#64748B',
              }}
            >
              <TabIcon className="w-4 h-4" />
              {tab.label}
              {isActive && (
                <div
                  className="absolute bottom-0 left-0 right-0 h-0.5"
                  style={{ backgroundColor: tab.color }}
                />
              )}
            </button>
          );
        })}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3 flex-wrap">
        {/* Category Filter */}
        <div className="relative">
          <Layers className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#94A3B8] pointer-events-none" />
          <select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            className="appearance-none cursor-pointer text-[13px] font-medium transition-colors"
            style={{
              backgroundColor: '#1E293B',
              border: '1px solid #334155',
              borderRadius: '8px',
              padding: '0 36px 0 32px',
              height: '40px',
              color: selectedCategory ? '#FFFFFF' : '#FFFFFF',
            }}
          >
            <option value="">All Categories</option>
            {categories.map((cat) => (
              <option key={cat.id} value={cat.id}>
                {cat.name}
              </option>
            ))}
          </select>
          <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
        </div>

        {/* Type Filter */}
        {activeTab !== 'auctions' && (
          <div className="relative">
            <select
              value={selectedType}
              onChange={(e) => setSelectedType(e.target.value as ListingType)}
              className="appearance-none cursor-pointer text-[13px] font-medium transition-colors"
              style={{
                backgroundColor: '#1E293B',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '0 36px 0 14px',
                height: '40px',
                color: selectedType ? '#FFFFFF' : '#FFFFFF',
              }}
            >
              {typeOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
          </div>
        )}

        {/* Price Filter (placeholder) */}
        {activeTab !== 'auctions' && (
          <div className="relative">
            <DollarSign className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#94A3B8] pointer-events-none" />
            <select
              className="appearance-none cursor-pointer text-[13px] font-medium transition-colors"
              style={{
                backgroundColor: '#1E293B',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '0 36px 0 32px',
                height: '40px',
                color: '#FFFFFF',
              }}
            >
              <option value="">Any Price</option>
              <option value="0-50">$0 - $50</option>
              <option value="50-100">$50 - $100</option>
              <option value="100+">$100+</option>
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
          </div>
        )}

        {/* Scope Filter */}
        {activeTab !== 'auctions' && (
          <div className="relative">
            <Globe className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#94A3B8] pointer-events-none" />
            <select
              value={selectedScope}
              onChange={(e) => setSelectedScope(e.target.value as GeographicScope)}
              className="appearance-none cursor-pointer text-[13px] font-medium transition-colors"
              style={{
                backgroundColor: '#1E293B',
                border: '1px solid #334155',
                borderRadius: '8px',
                padding: '0 36px 0 32px',
                height: '40px',
                color: selectedScope ? '#FFFFFF' : '#FFFFFF',
              }}
            >
              {scopeOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
            <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#64748B] pointer-events-none" />
          </div>
        )}

        {/* Sort */}
        <div className="relative flex items-center gap-2">
          <ArrowUpDown className="w-4 h-4 text-[#64748B]" />
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value)}
            className="appearance-none cursor-pointer text-[13px] font-medium transition-colors bg-transparent border-none outline-none"
            style={{
              color: '#94A3B8',
              paddingRight: '8px',
            }}
          >
            {(activeTab === 'requests' ? sortOptionsRequests : sortOptionsListings).map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* Results Count */}
        <span className="text-[#64748B] text-[13px] ml-auto">
          {loading ? 'Loading...' : error ? '' : `${total} results`}
        </span>

        {/* Clear Filters */}
        {hasFilters && (
          <button
            onClick={clearFilters}
            className="flex items-center gap-1.5 text-sm text-[#64748B] hover:text-white transition-colors"
          >
            <X className="w-4 h-4" />
            Clear filters
          </button>
        )}
      </div>

      {/* Error State */}
      {error && !loading && (
        <div className="flex flex-col items-center justify-center py-16 gap-4">
          <div
            className="w-16 h-16 rounded-full flex items-center justify-center"
            style={{ backgroundColor: 'rgba(239, 68, 68, 0.1)' }}
          >
            <AlertTriangle className="w-8 h-8 text-[#EF4444]" />
          </div>
          <div className="text-center">
            <h3 className="text-white font-semibold text-lg">Something went wrong</h3>
            <p className="text-[#64748B] text-sm mt-1 max-w-md">
              {error}
            </p>
          </div>
          <button
            onClick={handleRetry}
            className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg transition-colors"
            style={{
              backgroundColor: activeTabColor,
              color: '#0A0F1C',
            }}
          >
            <RefreshCw className="w-4 h-4" />
            Try again
          </button>
        </div>
      )}

      {/* Results Grid */}
      {!error && loading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {[...Array(6)].map((_, i) => (
            <div
              key={i}
              className="animate-pulse"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '16px',
                height: '280px',
              }}
            />
          ))}
        </div>
      ) : !error ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {activeTab === 'listings' &&
            listings.map((listing) => (
              <ListingCard
                key={listing.id}
                listing={listing}
                onClick={() => navigate(`/marketplace/listings/${listing.slug || listing.id}`)}
              />
            ))}
          {activeTab === 'requests' &&
            requests.map((request) => (
              <RequestCard
                key={request.id}
                request={request}
                onClick={() => navigate(`/marketplace/requests/${request.slug || request.id}`)}
              />
            ))}
          {activeTab === 'auctions' &&
            auctions.map((auction) => (
              <AuctionCard
                key={auction.id}
                auction={auction}
                onClick={() => navigate(`/marketplace/auctions/${auction.slug || auction.id}`)}
              />
            ))}
        </div>
      ) : null}

      {/* Empty State */}
      {!loading && !error && total === 0 && (
        <div className="flex flex-col items-center justify-center py-16 gap-4">
          <div
            className="w-16 h-16 rounded-full flex items-center justify-center"
            style={{ backgroundColor: '#1E293B' }}
          >
            <Store className="w-8 h-8 text-[#64748B]" />
          </div>
          <div className="text-center">
            <h3 className="text-white font-semibold text-lg">No {activeTab} found</h3>
            <p className="text-[#64748B] text-sm mt-1">
              Try adjusting your filters or search query
            </p>
          </div>
          {hasFilters && (
            <button
              onClick={clearFilters}
              className="px-4 py-2 text-sm font-medium rounded-lg transition-colors"
              style={{
                backgroundColor: activeTabColor,
                color: '#0A0F1C',
              }}
            >
              Clear all filters
            </button>
          )}
        </div>
      )}
    </div>
  );
}
