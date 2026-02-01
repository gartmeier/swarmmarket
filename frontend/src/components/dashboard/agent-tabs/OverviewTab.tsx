import { Star, Loader2, ListTodo, MessageSquare } from 'lucide-react';
import { useAgentTransactions, useAgentRatings } from '../../../hooks/useDashboard';
import type { Transaction, Rating, AgentMetrics } from '../../../lib/api';

function formatTimeAgo(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / 86400000);
  const diffWeeks = Math.floor(diffDays / 7);

  if (diffDays < 1) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  if (diffWeeks === 1) return '1 week ago';
  if (diffWeeks < 4) return `${diffWeeks} weeks ago`;

  return date.toLocaleDateString();
}

function TaskRow({ tx, isLast }: { tx: Transaction; isLast: boolean }) {
  const status = tx.status === 'completed' ? 'completed' : 'in_progress';
  const statusColor = status === 'completed' ? '#22C55E' : '#F59E0B';
  const categoryColor = '#22D3EE';

  return (
    <div
      className="flex items-center justify-between"
      style={{
        padding: '16px 20px',
        borderBottom: !isLast ? '1px solid #334155' : 'none',
      }}
    >
      <div className="flex items-center" style={{ gap: '12px' }}>
        <span
          className="text-[11px] font-semibold rounded-full"
          style={{
            padding: '4px 10px',
            backgroundColor: `${categoryColor}20`,
            color: categoryColor,
          }}
        >
          Task
        </span>
        <span className="text-[14px] text-white">{tx.title}</span>
      </div>
      <div className="flex items-center" style={{ gap: '16px' }}>
        <span className="font-mono text-[14px] font-semibold text-[#22C55E]">
          +${tx.amount.toFixed(2)}
        </span>
        <span
          className="text-[12px] font-medium rounded-full"
          style={{
            padding: '4px 10px',
            backgroundColor: status === 'completed' ? 'rgba(34, 197, 94, 0.1)' : 'rgba(245, 158, 11, 0.1)',
            color: statusColor,
          }}
        >
          {status === 'completed' ? 'Completed' : 'In Progress'}
        </span>
      </div>
    </div>
  );
}

function ReviewCard({ rating }: { rating: Rating }) {
  return (
    <div className="rounded-xl bg-[#1E293B]" style={{ padding: '16px' }}>
      <div className="flex items-center justify-between" style={{ marginBottom: '8px' }}>
        <div className="flex items-center" style={{ gap: '8px' }}>
          <div
            className="rounded-full bg-[#22D3EE] flex items-center justify-center"
            style={{ width: '32px', height: '32px' }}
          >
            <span className="text-[12px] font-bold text-[#0A0F1C]">
              {(rating.rater_name || 'U').charAt(0)}
            </span>
          </div>
          <span className="text-[14px] font-medium text-white">{rating.rater_name || 'User'}</span>
        </div>
        <div className="flex items-center" style={{ gap: '2px' }}>
          {[...Array(5)].map((_, i) => (
            <Star
              key={i}
              className="w-3.5 h-3.5"
              style={{
                color: i < rating.score ? '#F59E0B' : '#334155',
              }}
              fill={i < rating.score ? '#F59E0B' : '#334155'}
            />
          ))}
        </div>
      </div>
      {rating.message && (
        <p className="text-[13px] text-[#94A3B8] leading-relaxed">{rating.message}</p>
      )}
      <p className="text-[12px] text-[#64748B]" style={{ marginTop: '8px' }}>
        {formatTimeAgo(rating.created_at)}
      </p>
    </div>
  );
}

interface OverviewTabProps {
  agentId: string;
  metrics: AgentMetrics | null;
  trustScore: number;
}

export function OverviewTab({ agentId, metrics, trustScore }: OverviewTabProps) {
  const { transactions, loading: txLoading } = useAgentTransactions(agentId);
  const { ratings, loading: ratingsLoading } = useAgentRatings(agentId);

  const totalEarned = metrics?.total_revenue || 0;
  const ratingScore = trustScore;
  const tasksCompleted = metrics?.successful_trades || 0;
  const inProgress = metrics?.pending_offers || 0;

  const recentTransactions = transactions
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5);

  return (
    <>
      {/* Stats Row */}
      <div className="flex" style={{ gap: '20px' }}>
        <div className="flex-1 rounded-2xl bg-[#1E293B]" style={{ padding: '24px' }}>
          <span className="text-[13px] font-medium text-[#64748B]">Total Earned</span>
          <p
            className="font-mono text-[32px] font-bold text-[#22C55E] leading-none"
            style={{ marginTop: '8px' }}
          >
            ${totalEarned.toFixed(2)}
          </p>
        </div>
        <div className="flex-1 rounded-2xl bg-[#1E293B]" style={{ padding: '24px' }}>
          <span className="text-[13px] font-medium text-[#64748B]">Rating</span>
          <p
            className="font-mono text-[32px] font-bold text-[#F59E0B] leading-none"
            style={{ marginTop: '8px' }}
          >
            {ratingScore.toFixed(2)}
          </p>
        </div>
        <div className="flex-1 rounded-2xl bg-[#1E293B]" style={{ padding: '24px' }}>
          <span className="text-[13px] font-medium text-[#64748B]">Tasks Completed</span>
          <p
            className="font-mono text-[32px] font-bold text-white leading-none"
            style={{ marginTop: '8px' }}
          >
            {tasksCompleted}
          </p>
        </div>
        <div className="flex-1 rounded-2xl bg-[#1E293B]" style={{ padding: '24px' }}>
          <span className="text-[13px] font-medium text-[#64748B]">In Progress</span>
          <p
            className="font-mono text-[32px] font-bold text-[#22D3EE] leading-none"
            style={{ marginTop: '8px' }}
          >
            {inProgress}
          </p>
        </div>
      </div>

      {/* Content Row */}
      <div className="flex flex-1" style={{ gap: '24px' }}>
        {/* Recent Tasks */}
        <div className="flex-1 flex flex-col" style={{ gap: '16px' }}>
          <div className="flex items-center justify-between">
            <h2 className="text-[18px] font-semibold text-white">Recent Tasks</h2>
            <button className="text-[14px] font-medium text-[#A855F7] hover:text-[#9333EA] transition-colors">
              View All →
            </button>
          </div>
          {txLoading ? (
            <div className="flex items-center justify-center" style={{ padding: '40px 0' }}>
              <Loader2 className="w-6 h-6 text-[#22D3EE] animate-spin" />
            </div>
          ) : recentTransactions.length === 0 ? (
            <div
              className="rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
              style={{ padding: '40px 16px' }}
            >
              <ListTodo className="w-8 h-8 text-[#64748B]" style={{ marginBottom: '8px' }} />
              <p className="text-[14px] text-[#64748B]">No tasks yet</p>
            </div>
          ) : (
            <div className="rounded-xl bg-[#1E293B] overflow-hidden">
              {recentTransactions.map((tx, index) => (
                <TaskRow key={tx.id} tx={tx} isLast={index === recentTransactions.length - 1} />
              ))}
            </div>
          )}
        </div>

        {/* Reviews */}
        <div className="flex flex-col" style={{ width: '360px', gap: '16px' }}>
          <div className="flex items-center justify-between">
            <h2 className="text-[18px] font-semibold text-white">Reviews</h2>
          </div>
          {ratingsLoading ? (
            <div className="flex items-center justify-center" style={{ padding: '40px 0' }}>
              <Loader2 className="w-6 h-6 text-[#F59E0B] animate-spin" />
            </div>
          ) : ratings.length === 0 ? (
            <div
              className="rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
              style={{ padding: '40px 16px' }}
            >
              <MessageSquare className="w-8 h-8 text-[#64748B]" style={{ marginBottom: '8px' }} />
              <p className="text-[14px] text-[#64748B]">No reviews yet</p>
            </div>
          ) : (
            <div className="flex flex-col" style={{ gap: '12px' }}>
              {ratings.slice(0, 5).map((rating) => (
                <ReviewCard key={rating.id} rating={rating} />
              ))}
            </div>
          )}
        </div>
      </div>
    </>
  );
}
