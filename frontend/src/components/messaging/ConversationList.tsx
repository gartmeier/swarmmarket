import { useState, useEffect } from 'react';
import { MessageCircle, Search, X } from 'lucide-react';
import { api } from '../../lib/api';
import type { Conversation } from '../../lib/api';

interface ConversationListProps {
  selectedId?: string;
  onSelect: (conversation: Conversation) => void;
}

export function ConversationList({ selectedId, onSelect }: ConversationListProps) {
  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [unreadTotal, setUnreadTotal] = useState(0);

  useEffect(() => {
    loadConversations();
  }, []);

  const loadConversations = async () => {
    try {
      const res = await api.getConversations(50, 0);
      setConversations(res.conversations || []);
      setUnreadTotal(res.unread_total || 0);
    } catch (err) {
      console.error('Failed to load conversations:', err);
    } finally {
      setLoading(false);
    }
  };

  const filteredConversations = conversations.filter(conv =>
    conv.other_participant_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    conv.listing_title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    conv.request_title?.toLowerCase().includes(searchQuery.toLowerCase()) ||
    conv.auction_title?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const formatTime = (dateStr?: string) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (days === 1) {
      return 'Yesterday';
    } else if (days < 7) {
      return date.toLocaleDateString([], { weekday: 'short' });
    }
    return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
  };

  const getContextLabel = (conv: Conversation) => {
    if (conv.listing_title) return `Listing: ${conv.listing_title}`;
    if (conv.request_title) return `Request: ${conv.request_title}`;
    if (conv.auction_title) return `Auction: ${conv.auction_title}`;
    return null;
  };

  return (
    <div className="flex flex-col h-full" style={{ backgroundColor: '#0F172A' }}>
      {/* Header */}
      <div className="p-4 border-b" style={{ borderColor: '#334155' }}>
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold text-white flex items-center gap-2">
            <MessageCircle className="w-5 h-5 text-[#22D3EE]" />
            Messages
            {unreadTotal > 0 && (
              <span className="px-2 py-0.5 text-xs font-medium rounded-full bg-[#22D3EE] text-[#0A0F1C]">
                {unreadTotal}
              </span>
            )}
          </h2>
        </div>

        {/* Search */}
        <div
          className="flex items-center gap-2 px-3 py-2 rounded-lg"
          style={{ backgroundColor: '#1E293B', border: '1px solid #334155' }}
        >
          <Search className="w-4 h-4 text-[#64748B]" />
          <input
            type="text"
            placeholder="Search conversations..."
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
      </div>

      {/* Conversations List */}
      <div className="flex-1 overflow-y-auto">
        {loading ? (
          <div className="p-4 space-y-3">
            {[...Array(5)].map((_, i) => (
              <div key={i} className="animate-pulse flex gap-3">
                <div className="w-10 h-10 rounded-full bg-[#1E293B]" />
                <div className="flex-1 space-y-2">
                  <div className="h-4 bg-[#1E293B] rounded w-1/2" />
                  <div className="h-3 bg-[#1E293B] rounded w-3/4" />
                </div>
              </div>
            ))}
          </div>
        ) : filteredConversations.length === 0 ? (
          <div className="p-8 text-center">
            <MessageCircle className="w-12 h-12 text-[#64748B] mx-auto mb-3" />
            <p className="text-[#64748B]">
              {searchQuery ? 'No conversations found' : 'No messages yet'}
            </p>
          </div>
        ) : (
          <div>
            {filteredConversations.map((conv) => {
              const isSelected = conv.id === selectedId;
              const contextLabel = getContextLabel(conv);

              return (
                <button
                  key={conv.id}
                  onClick={() => onSelect(conv)}
                  className="w-full p-4 flex gap-3 transition-colors text-left"
                  style={{
                    backgroundColor: isSelected ? '#1E293B' : 'transparent',
                    borderLeft: isSelected ? '3px solid #22D3EE' : '3px solid transparent',
                  }}
                >
                  {/* Avatar */}
                  <div className="relative flex-shrink-0">
                    {conv.other_avatar_url ? (
                      <img
                        src={conv.other_avatar_url}
                        alt={conv.other_participant_name}
                        className="w-10 h-10 rounded-full object-cover"
                      />
                    ) : (
                      <div
                        className="w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium"
                        style={{ backgroundColor: '#22D3EE', color: '#0A0F1C' }}
                      >
                        {conv.other_participant_name.charAt(0).toUpperCase()}
                      </div>
                    )}
                    {conv.unread_count > 0 && (
                      <div
                        className="absolute -top-1 -right-1 w-5 h-5 rounded-full flex items-center justify-center text-xs font-medium"
                        style={{ backgroundColor: '#22D3EE', color: '#0A0F1C' }}
                      >
                        {conv.unread_count}
                      </div>
                    )}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <span className={`font-medium truncate ${conv.unread_count > 0 ? 'text-white' : 'text-[#94A3B8]'}`}>
                        {conv.other_participant_name}
                      </span>
                      <span className="text-xs text-[#64748B] flex-shrink-0">
                        {formatTime(conv.last_message_at || conv.created_at)}
                      </span>
                    </div>

                    {contextLabel && (
                      <p className="text-xs text-[#22D3EE] truncate mt-0.5">{contextLabel}</p>
                    )}

                    {conv.last_message && (
                      <p className={`text-sm truncate mt-1 ${conv.unread_count > 0 ? 'text-[#94A3B8]' : 'text-[#64748B]'}`}>
                        {conv.last_message.content}
                      </p>
                    )}
                  </div>
                </button>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}
