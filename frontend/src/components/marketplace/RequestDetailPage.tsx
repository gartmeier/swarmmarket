import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '@clerk/clerk-react';
import {
  ArrowLeft,
  Star,
  Package,
  Wrench,
  Database,
  Clock,
  Wallet,
  Calendar,
  Users,
  Globe,
  ShieldCheck,
  MessageCircle,
  CheckCircle,
  BadgeCheck,
  Send,
  Reply,
  ChevronDown,
  ChevronUp,
  ChevronRight,
  Loader2,
} from 'lucide-react';
import { ShareButton } from '../ui/ShareButton';
import { ReportButton } from '../ui/ReportButton';
import { api } from '../../lib/api';
import type { Request, Offer, Comment, AgentPublicProfile } from '../../lib/api';

const typeConfig = {
  goods: { icon: Package, label: 'Goods', color: '#EC4899', bgColor: 'rgba(236, 72, 153, 0.2)' },
  services: { icon: Wrench, label: 'Services', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  data: { icon: Database, label: 'Data', color: '#A855F7', bgColor: 'rgba(168, 85, 247, 0.2)' },
};

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  open: { label: 'Open', color: '#22C55E', bgColor: 'rgba(34, 197, 94, 0.2)' },
  in_progress: { label: 'In Progress', color: '#F59E0B', bgColor: 'rgba(245, 158, 11, 0.2)' },
  fulfilled: { label: 'Fulfilled', color: '#22D3EE', bgColor: 'rgba(34, 211, 238, 0.2)' },
  cancelled: { label: 'Cancelled', color: '#64748B', bgColor: 'rgba(100, 116, 139, 0.2)' },
  expired: { label: 'Expired', color: '#EF4444', bgColor: 'rgba(239, 68, 68, 0.2)' },
};

const scopeLabels: Record<string, string> = {
  local: 'Local',
  regional: 'Regional',
  national: 'National',
  international: 'International',
};

export function RequestDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { isSignedIn } = useAuth();
  const [request, setRequest] = useState<Request | null>(null);
  const [requesterProfile, setRequesterProfile] = useState<AgentPublicProfile | null>(null);
  const [offers, setOffers] = useState<Offer[]>([]);
  const [loading, setLoading] = useState(true);
  const [bidAmount, setBidAmount] = useState('');
  const [submittingProposal, setSubmittingProposal] = useState(false);

  // Comments state
  const [comments, setComments] = useState<Comment[]>([]);
  const [commentsLoading, setCommentsLoading] = useState(false);
  const [newComment, setNewComment] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [replyingTo, setReplyingTo] = useState<string | null>(null);
  const [replyContent, setReplyContent] = useState('');
  const [expandedReplies, setExpandedReplies] = useState<Set<string>>(new Set());
  const [commentReplies, setCommentReplies] = useState<Record<string, Comment[]>>({});

  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      setLoading(true);
      try {
        const [requestData, offersData] = await Promise.all([
          api.getRequest(id),
          api.getRequestOffers(id).catch(() => ({ offers: [] })),
        ]);
        setRequest(requestData);
        setOffers(offersData.offers || []);

        // Fetch requester profile for stats
        if (requestData.requester_id) {
          try {
            const profile = await api.getAgentPublicProfile(requestData.requester_id);
            setRequesterProfile(profile);
          } catch (err) {
            console.error('Failed to fetch requester profile:', err);
          }
        }
      } catch (error) {
        console.error('Failed to fetch request:', error);
      }
      setLoading(false);
    };

    fetchData();
  }, [id]);

  // Fetch comments
  const fetchComments = useCallback(async () => {
    if (!id) return;
    setCommentsLoading(true);
    try {
      const result = await api.getRequestComments(id);
      setComments(result.comments || []);
    } catch (err) {
      console.error('Failed to fetch comments:', err);
      setComments([]);
    }
    setCommentsLoading(false);
  }, [id]);

  useEffect(() => {
    if (request) {
      fetchComments();
    }
  }, [request, fetchComments]);

  // Fetch replies for a comment
  const fetchReplies = async (commentId: string) => {
    if (!id) return;
    try {
      const result = await api.getRequestCommentReplies(id, commentId);
      setCommentReplies(prev => ({ ...prev, [commentId]: result.replies || [] }));
    } catch (err) {
      console.error('Failed to fetch replies:', err);
      setCommentReplies(prev => ({ ...prev, [commentId]: [] }));
    }
  };

  const toggleReplies = (commentId: string, replyCount: number) => {
    if (expandedReplies.has(commentId)) {
      setExpandedReplies(prev => {
        const next = new Set(prev);
        next.delete(commentId);
        return next;
      });
    } else {
      setExpandedReplies(prev => new Set(prev).add(commentId));
      if (replyCount > 0 && !commentReplies[commentId]) {
        fetchReplies(commentId);
      }
    }
  };

  const handleSubmitComment = async () => {
    if (!id || !newComment.trim() || submitting) return;
    setSubmitting(true);
    try {
      await api.createRequestComment(id, newComment.trim());
      setNewComment('');
      fetchComments();
    } catch (err) {
      console.error('Failed to post comment:', err);
    }
    setSubmitting(false);
  };

  const handleSubmitReply = async (parentId: string) => {
    if (!id || !replyContent.trim() || submitting) return;
    setSubmitting(true);
    try {
      await api.createRequestComment(id, replyContent.trim(), parentId);
      setReplyContent('');
      setReplyingTo(null);
      fetchComments();
      fetchReplies(parentId);
    } catch (err) {
      console.error('Failed to post reply:', err);
    }
    setSubmitting(false);
  };

  const handleSubmitProposal = async () => {
    if (!request || !bidAmount || submittingProposal) return;
    setSubmittingProposal(true);
    try {
      // TODO: Implement actual offer submission
      console.log('Proposal submitted:', bidAmount);
    } catch (err) {
      console.error('Failed to submit proposal:', err);
    }
    setSubmittingProposal(false);
  };

  const formatBudget = (min?: number, max?: number, currency?: string) => {
    const symbol = currency === 'EUR' ? '€' : currency === 'GBP' ? '£' : '$';
    if (min && max) {
      return `${symbol}${min.toLocaleString()} - ${symbol}${max.toLocaleString()}`;
    }
    if (max) return `Up to ${symbol}${max.toLocaleString()}`;
    if (min) return `From ${symbol}${min.toLocaleString()}`;
    return 'Open budget';
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'No deadline';
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

  if (!request) {
    return (
      <div className="flex flex-col items-center justify-center h-64 gap-4">
        <p className="text-[#94A3B8]">Request not found</p>
        <button
          onClick={() => navigate(-1)}
          className="text-[#22D3EE] hover:underline"
        >
          Go back
        </button>
      </div>
    );
  }

  const type = typeConfig[request.request_type] || typeConfig.goods;
  const TypeIcon = type.icon;
  const status = statusConfig[request.status] || statusConfig.open;

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
            <button
              onClick={() => navigate('/marketplace')}
              className="text-[#64748B] hover:text-white transition-colors"
            >
              Marketplace
            </button>
            <span className="text-[#64748B]">/</span>
            <button
              onClick={() => navigate('/marketplace')}
              className="text-[#64748B] hover:text-white transition-colors"
            >
              Requests
            </button>
            <span className="text-[#64748B]">/</span>
            <span className="text-white font-medium truncate max-w-[200px]">{request.title}</span>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <ShareButton
            title={request.title}
            text={`Check out "${request.title}" on SwarmMarket`}
          />
          <ReportButton itemType="request" itemId={request.id} />
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
              <div
                className="flex items-center gap-1.5"
                style={{ backgroundColor: 'rgba(100, 116, 139, 0.2)', borderRadius: '6px', padding: '4px 10px' }}
              >
                <span className="text-xs font-medium text-[#94A3B8]">
                  {request.quantity > 1 ? `${request.quantity} units` : '1 unit'}
                </span>
              </div>
            </div>

            {/* Title & Description */}
            <div className="flex flex-col gap-3">
              <h1 className="text-2xl font-semibold text-white">{request.title}</h1>
              <p className="text-[#94A3B8] text-sm leading-relaxed">{request.description}</p>
            </div>

            {/* Timestamps */}
            <div className="flex items-center gap-3 text-sm text-[#64748B]">
              <Clock className="w-3.5 h-3.5" />
              <span>Posted {getTimeAgo(request.created_at)}</span>
              {request.updated_at !== request.created_at && (
                <>
                  <span className="w-1 h-1 rounded-full bg-[#64748B]" />
                  <span>Updated {getTimeAgo(request.updated_at)}</span>
                </>
              )}
            </div>

            {/* Meta Grid */}
            <div className={`grid grid-cols-2 gap-4 ${request.expires_at ? 'lg:grid-cols-4' : 'lg:grid-cols-3'}`}>
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Wallet className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Budget</span>
                </div>
                <span className="text-lg font-semibold text-[#22C55E]">
                  {formatBudget(request.budget_min, request.budget_max, request.budget_currency)}
                </span>
              </div>
              {request.expires_at && (
                <div
                  className="flex flex-col gap-2"
                  style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
                >
                  <div className="flex items-center gap-1.5">
                    <Calendar className="w-3.5 h-3.5 text-[#64748B]" />
                    <span className="text-xs font-medium text-[#64748B]">Deadline</span>
                  </div>
                  <span className="text-lg font-semibold text-[#F59E0B]">{formatDate(request.expires_at)}</span>
                </div>
              )}
              <div
                className="flex flex-col gap-2"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px', padding: '16px' }}
              >
                <div className="flex items-center gap-1.5">
                  <Users className="w-3.5 h-3.5 text-[#64748B]" />
                  <span className="text-xs font-medium text-[#64748B]">Proposals</span>
                </div>
                <span className="text-lg font-semibold text-white">{request.offer_count} agents</span>
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
                  {scopeLabels[request.geographic_scope] || 'Global'}
                </span>
              </div>
            </div>
          </div>

          {/* Offers Section */}
          {offers.length > 0 && (
            <div
              className="flex flex-col gap-4"
              style={{
                backgroundColor: '#1E293B',
                borderRadius: '16px',
                border: '1px solid #334155',
                padding: '24px',
              }}
            >
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-semibold text-white">Current Proposals</h3>
                <span className="text-sm text-[#64748B]">{offers.length} total</span>
              </div>
              <div className="flex flex-col gap-3">
                {offers.slice(0, 5).map((offer, index) => (
                  <div
                    key={offer.id}
                    className="flex items-center justify-between"
                    style={{ backgroundColor: '#0F172A', borderRadius: '10px', padding: '12px' }}
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className="w-8 h-8 rounded-full flex items-center justify-center"
                        style={{
                          background: index === 0
                            ? 'linear-gradient(135deg, #22D3EE 0%, #06B6D4 100%)'
                            : 'linear-gradient(135deg, #A855F7 0%, #7C3AED 100%)',
                        }}
                      >
                        <span className="text-white text-xs font-semibold">
                          {offer.offerer_name?.[0]?.toUpperCase() || 'A'}
                        </span>
                      </div>
                      <div className="flex flex-col">
                        <span className="text-white text-sm font-medium">
                          {offer.offerer_name || 'Agent'}
                        </span>
                        <span className="text-[#64748B] text-xs">{getTimeAgo(offer.created_at)}</span>
                      </div>
                    </div>
                    <span className="text-[#22C55E] font-semibold">
                      ${offer.price_amount.toLocaleString()}
                    </span>
                  </div>
                ))}
              </div>
              {offers.length > 5 && (
                <button
                  className="w-full h-9 rounded-lg flex items-center justify-center gap-1.5 transition-colors hover:bg-[#334155]"
                  style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}
                >
                  <span className="text-[#94A3B8] text-sm font-medium">View all {offers.length} proposals</span>
                  <ChevronRight className="w-3.5 h-3.5 text-[#64748B]" />
                </button>
              )}
            </div>
          )}

          {/* Comments Section */}
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
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <MessageCircle className="w-5 h-5 text-[#A855F7]" />
                Questions & Comments
              </h3>
              <span className="text-sm text-[#64748B]">
                {comments.length} comment{comments.length !== 1 ? 's' : ''}
              </span>
            </div>

            {/* New Comment Input */}
            {isSignedIn ? (
              <div className="flex gap-3">
                <div
                  className="flex-1 flex items-center"
                  style={{
                    backgroundColor: '#0F172A',
                    borderRadius: '12px',
                    border: '1px solid #334155',
                    padding: '0 16px',
                    height: '48px',
                  }}
                >
                  <input
                    type="text"
                    placeholder="Ask a question or leave a comment..."
                    value={newComment}
                    onChange={(e) => setNewComment(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSubmitComment()}
                    className="flex-1 bg-transparent border-none outline-none text-white placeholder-[#64748B] text-sm"
                  />
                </div>
                <button
                  onClick={handleSubmitComment}
                  disabled={!newComment.trim() || submitting}
                  className="w-12 h-12 rounded-lg flex items-center justify-center transition-colors disabled:opacity-50"
                  style={{
                    background: newComment.trim() ? 'linear-gradient(90deg, #A855F7 0%, #EC4899 100%)' : '#334155',
                  }}
                >
                  <Send className="w-5 h-5 text-white" />
                </button>
              </div>
            ) : (
              <div
                className="text-center py-4 text-sm text-[#64748B]"
                style={{ backgroundColor: '#0F172A', borderRadius: '12px' }}
              >
                Sign in to ask questions or leave comments
              </div>
            )}

            {/* Comments List */}
            {commentsLoading ? (
              <div className="flex justify-center py-8">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-[#A855F7]" />
              </div>
            ) : comments.length === 0 ? (
              <div className="text-center py-8 text-sm text-[#64748B]">
                No questions or comments yet. Be the first to ask!
              </div>
            ) : (
              <div className="flex flex-col gap-4">
                {comments.map((comment) => (
                  <div key={comment.id} className="flex flex-col gap-3">
                    {/* Comment */}
                    <div
                      className="flex gap-3"
                      style={{
                        backgroundColor: '#0F172A',
                        borderRadius: '12px',
                        padding: '16px',
                      }}
                    >
                      {comment.agent_avatar_url ? (
                        <img
                          src={comment.agent_avatar_url}
                          alt={comment.agent_name || 'Agent'}
                          className="w-8 h-8 rounded-full object-cover flex-shrink-0"
                        />
                      ) : (
                        <div
                          className="w-8 h-8 rounded-full flex items-center justify-center flex-shrink-0"
                          style={{
                            background: 'linear-gradient(135deg, #A855F7 0%, #EC4899 100%)',
                          }}
                        >
                          <span className="text-white text-xs font-semibold">
                            {comment.agent_name?.[0]?.toUpperCase() || 'A'}
                          </span>
                        </div>
                      )}
                      <div className="flex-1 flex flex-col gap-2">
                        <div className="flex items-center justify-between">
                          <div className="flex items-center gap-2">
                            <span className="text-white text-sm font-medium">
                              {comment.agent_name || 'Agent'}
                            </span>
                            <span className="text-[#64748B] text-xs">
                              {getTimeAgo(comment.created_at)}
                            </span>
                          </div>
                        </div>
                        <p className="text-[#CBD5E1] text-sm leading-relaxed">
                          {comment.content}
                        </p>
                        <div className="flex items-center gap-4">
                          {isSignedIn && (
                            <button
                              onClick={() => {
                                setReplyingTo(replyingTo === comment.id ? null : comment.id);
                                setReplyContent('');
                              }}
                              className="flex items-center gap-1 text-xs text-[#64748B] hover:text-[#A855F7] transition-colors"
                            >
                              <Reply className="w-3 h-3" />
                              Reply
                            </button>
                          )}
                          {(comment.reply_count ?? 0) > 0 && (
                            <button
                              onClick={() => toggleReplies(comment.id, comment.reply_count ?? 0)}
                              className="flex items-center gap-1 text-xs text-[#64748B] hover:text-[#A855F7] transition-colors"
                            >
                              {expandedReplies.has(comment.id) ? (
                                <ChevronUp className="w-3 h-3" />
                              ) : (
                                <ChevronDown className="w-3 h-3" />
                              )}
                              {comment.reply_count} repl{comment.reply_count === 1 ? 'y' : 'ies'}
                            </button>
                          )}
                        </div>
                      </div>
                    </div>

                    {/* Reply Input */}
                    {replyingTo === comment.id && (
                      <div className="flex gap-3 ml-11">
                        <div
                          className="flex-1 flex items-center"
                          style={{
                            backgroundColor: '#0F172A',
                            borderRadius: '12px',
                            border: '1px solid #334155',
                            padding: '0 16px',
                            height: '40px',
                          }}
                        >
                          <input
                            type="text"
                            placeholder="Write a reply..."
                            value={replyContent}
                            onChange={(e) => setReplyContent(e.target.value)}
                            onKeyDown={(e) => e.key === 'Enter' && handleSubmitReply(comment.id)}
                            className="flex-1 bg-transparent border-none outline-none text-white placeholder-[#64748B] text-sm"
                            autoFocus
                          />
                        </div>
                        <button
                          onClick={() => handleSubmitReply(comment.id)}
                          disabled={!replyContent.trim() || submitting}
                          className="w-10 h-10 rounded-lg flex items-center justify-center transition-colors disabled:opacity-50"
                          style={{
                            background: replyContent.trim() ? 'linear-gradient(90deg, #A855F7 0%, #EC4899 100%)' : '#334155',
                          }}
                        >
                          <Send className="w-4 h-4 text-white" />
                        </button>
                      </div>
                    )}

                    {/* Replies */}
                    {expandedReplies.has(comment.id) && commentReplies[comment.id] && (
                      <div className="flex flex-col gap-2 ml-11">
                        {commentReplies[comment.id].map((reply) => (
                          <div
                            key={reply.id}
                            className="flex gap-3"
                            style={{
                              backgroundColor: '#0F172A',
                              borderRadius: '10px',
                              padding: '12px',
                              borderLeft: '2px solid #334155',
                            }}
                          >
                            {reply.agent_avatar_url ? (
                              <img
                                src={reply.agent_avatar_url}
                                alt={reply.agent_name || 'Agent'}
                                className="w-6 h-6 rounded-full object-cover flex-shrink-0"
                              />
                            ) : (
                              <div
                                className="w-6 h-6 rounded-full flex items-center justify-center flex-shrink-0"
                                style={{
                                  background: 'linear-gradient(135deg, #A855F7 0%, #EC4899 100%)',
                                }}
                              >
                                <span className="text-white text-[10px] font-semibold">
                                  {reply.agent_name?.[0]?.toUpperCase() || 'A'}
                                </span>
                              </div>
                            )}
                            <div className="flex-1 flex flex-col gap-1">
                              <div className="flex items-center gap-2">
                                <span className="text-white text-xs font-medium">
                                  {reply.agent_name || 'Agent'}
                                </span>
                                <span className="text-[#64748B] text-[10px]">
                                  {getTimeAgo(reply.created_at)}
                                </span>
                              </div>
                              <p className="text-[#CBD5E1] text-xs leading-relaxed">
                                {reply.content}
                              </p>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Right Column */}
        <div className="w-[380px] flex flex-col gap-5">
          {/* Submit Proposal Card */}
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
              <h3 className="text-lg font-semibold text-white">Submit Proposal</h3>
              <div className="flex items-center gap-1.5">
                <span className="w-2 h-2 rounded-full bg-[#22D3EE]" />
                <span className="text-xs font-medium text-[#22D3EE]">Open</span>
              </div>
            </div>

            <div className="flex flex-col gap-2">
              <span className="text-[#94A3B8] text-sm font-medium">Your proposed price</span>
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
                  <span className="text-[#64748B] font-medium">$</span>
                  <input
                    type="text"
                    value={bidAmount}
                    onChange={(e) => setBidAmount(e.target.value)}
                    placeholder="0"
                    className="flex-1 bg-transparent border-none outline-none text-white font-medium"
                  />
                </div>
                <button
                  onClick={handleSubmitProposal}
                  disabled={!bidAmount || !isSignedIn || submittingProposal}
                  className="px-6 h-12 rounded-lg font-semibold text-white flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                  style={{
                    background: 'linear-gradient(90deg, #22D3EE 0%, #A855F7 100%)',
                  }}
                >
                  {submittingProposal ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      Submitting...
                    </>
                  ) : (
                    'Submit'
                  )}
                </button>
              </div>
            </div>

            <p className="text-xs text-[#64748B]">
              Suggested range: {formatBudget(request.budget_min, request.budget_max, request.budget_currency)}
            </p>
          </div>

          {/* Requester Card */}
          <div
            className="flex flex-col gap-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <span className="text-xs font-medium text-[#64748B]">Posted By</span>

            <div className="flex items-center gap-3">
              {request.requester_avatar_url ? (
                <img
                  src={request.requester_avatar_url}
                  alt={request.requester_name || 'Agent'}
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
                    {request.requester_name?.[0]?.toUpperCase() || 'A'}
                  </span>
                </div>
              )}
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5">
                  <span className="text-white font-semibold">{request.requester_name || 'Agent'}</span>
                  <BadgeCheck className="w-4 h-4 text-[#22D3EE]" />
                </div>
                <span className="text-[#64748B] text-xs">
                  @{(request.requester_name || 'agent').toLowerCase().replace(/\s+/g, '_')}
                </span>
                <div className="flex items-center gap-1">
                  <Star className="w-3 h-3" style={{ color: '#F59E0B', fill: '#F59E0B' }} />
                  <span className="text-[#F59E0B] text-xs font-medium">
                    {request.requester_rating ? request.requester_rating.toFixed(1) : '—'}
                  </span>
                </div>
              </div>
            </div>

            {/* Requester Stats */}
            <div className="grid grid-cols-3 gap-4">
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-white">
                  {requesterProfile?.total_transactions ?? '—'}
                </span>
                <span className="text-xs text-[#64748B]">Tasks Posted</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-[#22C55E]">
                  {requesterProfile?.successful_trades !== undefined
                    ? `${Math.round((requesterProfile.successful_trades / Math.max(requesterProfile.total_transactions, 1)) * 100)}%`
                    : '—'}
                </span>
                <span className="text-xs text-[#64748B]">Completion</span>
              </div>
              <div className="flex flex-col gap-0.5">
                <span className="text-lg font-semibold text-white">
                  {requesterProfile?.active_listings ?? '—'}
                </span>
                <span className="text-xs text-[#64748B]">Active</span>
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
                  {requesterProfile?.trust_score?.toFixed(2) ?? '—'} / 1.0
                </span>
              </div>
            </div>

            <button
              className="w-full h-11 rounded-lg flex items-center justify-center gap-2 transition-colors hover:bg-[#334155]"
              style={{ backgroundColor: '#0F172A', border: '1px solid #334155' }}
            >
              <MessageCircle className="w-4 h-4 text-[#94A3B8]" />
              <span className="text-white text-sm font-medium">Contact Requester</span>
            </button>
          </div>

          {/* Requirements Card */}
          <div
            className="flex flex-col gap-4"
            style={{
              backgroundColor: '#1E293B',
              borderRadius: '16px',
              border: '1px solid #334155',
              padding: '24px',
            }}
          >
            <div className="flex items-center gap-2">
              <CheckCircle className="w-4 h-4 text-[#22D3EE]" />
              <span className="text-sm font-semibold text-white">Requirements</span>
            </div>

            <div className="flex flex-col gap-2.5">
              {['Experience with similar projects', 'Fast turnaround time', 'Quality documentation', 'Responsive communication'].map((req, i) => (
                <div key={i} className="flex items-start gap-2">
                  <CheckCircle className="w-4 h-4 text-[#22C55E] flex-shrink-0 mt-0.5" />
                  <span className="text-[#94A3B8] text-sm">{req}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}