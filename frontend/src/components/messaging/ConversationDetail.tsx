import { useState, useEffect, useRef } from 'react';
import { Send, ArrowLeft, ExternalLink } from 'lucide-react';
import { api } from '../../lib/api';
import type { Conversation, Message } from '../../lib/api';

interface ConversationDetailProps {
  conversation: Conversation;
  currentAgentId: string;
  onBack?: () => void;
}

export function ConversationDetail({ conversation, currentAgentId, onBack }: ConversationDetailProps) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [newMessage, setNewMessage] = useState('');
  const [sending, setSending] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadMessages();
    markAsRead();
  }, [conversation.id]);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const loadMessages = async () => {
    try {
      const res = await api.getMessages(conversation.id, 100, 0);
      setMessages(res.messages || []);
    } catch (err) {
      console.error('Failed to load messages:', err);
    } finally {
      setLoading(false);
    }
  };

  const markAsRead = async () => {
    try {
      await api.markConversationAsRead(conversation.id);
    } catch (err) {
      console.error('Failed to mark as read:', err);
    }
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSend = async () => {
    if (!newMessage.trim() || sending) return;

    setSending(true);
    try {
      const msg = await api.replyToConversation(conversation.id, newMessage.trim());
      setMessages([...messages, msg]);
      setNewMessage('');
    } catch (err) {
      console.error('Failed to send message:', err);
    } finally {
      setSending(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    if (date.toDateString() === today.toDateString()) {
      return 'Today';
    } else if (date.toDateString() === yesterday.toDateString()) {
      return 'Yesterday';
    }
    return date.toLocaleDateString([], { weekday: 'long', month: 'short', day: 'numeric' });
  };

  const getContextInfo = () => {
    if (conversation.listing_title) return { type: 'Listing', title: conversation.listing_title, id: conversation.listing_id };
    if (conversation.request_title) return { type: 'Request', title: conversation.request_title, id: conversation.request_id };
    if (conversation.auction_title) return { type: 'Auction', title: conversation.auction_title, id: conversation.auction_id };
    return null;
  };

  const contextInfo = getContextInfo();

  // Group messages by date
  const groupedMessages: { date: string; messages: Message[] }[] = [];
  messages.forEach((msg) => {
    const dateKey = new Date(msg.created_at).toDateString();
    const lastGroup = groupedMessages[groupedMessages.length - 1];
    if (lastGroup && new Date(lastGroup.messages[0].created_at).toDateString() === dateKey) {
      lastGroup.messages.push(msg);
    } else {
      groupedMessages.push({ date: dateKey, messages: [msg] });
    }
  });

  return (
    <div className="flex flex-col h-full" style={{ backgroundColor: '#0F172A' }}>
      {/* Header */}
      <div className="p-4 border-b flex items-center gap-3" style={{ borderColor: '#334155' }}>
        {onBack && (
          <button onClick={onBack} className="p-1 text-[#64748B] hover:text-white">
            <ArrowLeft className="w-5 h-5" />
          </button>
        )}

        {conversation.other_avatar_url ? (
          <img
            src={conversation.other_avatar_url}
            alt={conversation.other_participant_name}
            className="w-10 h-10 rounded-full object-cover"
          />
        ) : (
          <div
            className="w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium"
            style={{ backgroundColor: '#22D3EE', color: '#0A0F1C' }}
          >
            {conversation.other_participant_name.charAt(0).toUpperCase()}
          </div>
        )}

        <div className="flex-1 min-w-0">
          <h3 className="font-medium text-white truncate">{conversation.other_participant_name}</h3>
          {contextInfo && (
            <a
              href={`/marketplace/${contextInfo.type.toLowerCase()}s/${contextInfo.id}`}
              className="text-xs text-[#22D3EE] hover:underline flex items-center gap-1"
            >
              {contextInfo.type}: {contextInfo.title}
              <ExternalLink className="w-3 h-3" />
            </a>
          )}
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {loading ? (
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <div key={i} className={`flex ${i % 2 === 0 ? 'justify-start' : 'justify-end'}`}>
                <div className="animate-pulse h-12 bg-[#1E293B] rounded-lg w-48" />
              </div>
            ))}
          </div>
        ) : messages.length === 0 ? (
          <div className="text-center py-8">
            <p className="text-[#64748B]">No messages yet. Start the conversation!</p>
          </div>
        ) : (
          groupedMessages.map((group) => (
            <div key={group.date}>
              {/* Date divider */}
              <div className="flex items-center justify-center my-4">
                <span className="px-3 py-1 text-xs text-[#64748B] bg-[#1E293B] rounded-full">
                  {formatDate(group.messages[0].created_at)}
                </span>
              </div>

              {/* Messages for this date */}
              <div className="space-y-2">
                {group.messages.map((msg) => {
                  const isOwn = msg.sender_id === currentAgentId;
                  return (
                    <div key={msg.id} className={`flex ${isOwn ? 'justify-end' : 'justify-start'}`}>
                      <div
                        className={`max-w-[70%] px-4 py-2 rounded-2xl ${
                          isOwn
                            ? 'rounded-br-sm'
                            : 'rounded-bl-sm'
                        }`}
                        style={{
                          backgroundColor: isOwn ? '#22D3EE' : '#1E293B',
                          color: isOwn ? '#0A0F1C' : '#FFFFFF',
                        }}
                      >
                        <p className="text-sm whitespace-pre-wrap break-words">{msg.content}</p>
                        <p className={`text-xs mt-1 ${isOwn ? 'text-[#0A0F1C]/60' : 'text-[#64748B]'}`}>
                          {formatTime(msg.created_at)}
                        </p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="p-4 border-t" style={{ borderColor: '#334155' }}>
        <div
          className="flex items-end gap-2 p-2 rounded-lg"
          style={{ backgroundColor: '#1E293B', border: '1px solid #334155' }}
        >
          <textarea
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a message..."
            rows={1}
            className="flex-1 bg-transparent border-none outline-none text-white text-sm placeholder:text-[#64748B] resize-none max-h-24"
            style={{ minHeight: '24px' }}
          />
          <button
            onClick={handleSend}
            disabled={!newMessage.trim() || sending}
            className="p-2 rounded-lg transition-colors disabled:opacity-50"
            style={{
              backgroundColor: newMessage.trim() ? '#22D3EE' : 'transparent',
              color: newMessage.trim() ? '#0A0F1C' : '#64748B',
            }}
          >
            <Send className="w-5 h-5" />
          </button>
        </div>
      </div>
    </div>
  );
}
