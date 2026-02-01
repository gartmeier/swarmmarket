import { Loader2, ListTodo } from 'lucide-react';
import { useAgentTransactions } from '../../../hooks/useDashboard';
import type { Transaction } from '../../../lib/api';

interface TaskCardProps {
  title: string;
  description: string;
  price?: string;
  time?: string;
  hasError?: boolean;
}

function TaskCard({ title, description, price, time, hasError }: TaskCardProps) {
  return (
    <div
      className="rounded-xl bg-[#1E293B]"
      style={{
        padding: '16px',
        gap: '12px',
        display: 'flex',
        flexDirection: 'column',
        border: hasError ? '1px solid rgba(239, 68, 68, 0.25)' : 'none',
      }}
    >
      <p className="text-[14px] font-medium text-white">{title}</p>
      <p className="text-[12px] text-[#64748B]">{description}</p>
      <div className="flex items-center justify-between">
        {price && (
          <span className="font-mono text-[13px] font-semibold text-[#22C55E]">{price}</span>
        )}
        {time && <span className="text-[12px] text-[#64748B]">{time}</span>}
      </div>
    </div>
  );
}

interface TaskColumnProps {
  title: string;
  count: number;
  color: string;
  tasks: Transaction[];
  isError?: boolean;
}

function TaskColumn({ title, count, color, tasks, isError }: TaskColumnProps) {
  return (
    <div
      className="flex flex-col"
      style={{ gap: '16px', height: '100%', width: '100%' }}
    >
      {/* Column Header */}
      <div className="flex items-center" style={{ gap: '8px' }}>
        <div
          className="rounded-full"
          style={{ width: '10px', height: '10px', backgroundColor: color }}
        />
        <span className="text-[16px] font-semibold text-white">{title}</span>
        <span className="text-[14px] font-medium text-[#64748B]">{count}</span>
      </div>

      {/* Task List */}
      <div className="flex flex-col flex-1" style={{ gap: '12px' }}>
        {tasks.length === 0 ? (
          <div
            className="flex-1 rounded-xl bg-[#1E293B]/50 flex items-center justify-center"
            style={{ minHeight: '100px' }}
          >
            <p className="text-[13px] text-[#475569]">No tasks</p>
          </div>
        ) : (
          tasks.map((tx) => (
            <TaskCard
              key={tx.id}
              title={tx.title}
              description={tx.description || 'No description'}
              price={`$${tx.amount.toFixed(2)}`}
              hasError={isError}
            />
          ))
        )}
      </div>
    </div>
  );
}

interface TasksTabProps {
  agentId: string;
}

export function TasksTab({ agentId }: TasksTabProps) {
  const { transactions, loading } = useAgentTransactions(agentId);

  if (loading) {
    return (
      <div className="flex items-center justify-center flex-1">
        <Loader2 className="w-8 h-8 text-[#22D3EE] animate-spin" />
      </div>
    );
  }

  if (transactions.length === 0) {
    return (
      <div
        className="flex-1 rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
        style={{ padding: '80px 16px' }}
      >
        <ListTodo className="w-12 h-12 text-[#64748B]" style={{ marginBottom: '16px' }} />
        <p className="text-[16px] font-medium text-white" style={{ marginBottom: '4px' }}>
          No tasks yet
        </p>
        <p className="text-[14px] text-[#64748B]">
          Tasks from transactions will appear here
        </p>
      </div>
    );
  }

  // Categorize transactions by status
  const inProgressTasks = transactions.filter(
    (tx) => tx.status === 'escrow_funded' || tx.status === 'delivered'
  );
  const errorTasks = transactions.filter(
    (tx) => tx.status === 'disputed' || tx.status === 'cancelled'
  );
  const pendingTasks = transactions.filter((tx) => tx.status === 'pending');
  const completedTasks = transactions.filter((tx) => tx.status === 'completed');

  return (
    <div className="flex flex-1" style={{ gap: '20px' }}>
      <TaskColumn
        title="In Progress"
        count={inProgressTasks.length}
        color="#F59E0B"
        tasks={inProgressTasks}
      />
      <TaskColumn
        title="Error"
        count={errorTasks.length}
        color="#EF4444"
        tasks={errorTasks}
        isError
      />
      <TaskColumn
        title="Pending"
        count={pendingTasks.length}
        color="#A855F7"
        tasks={pendingTasks}
      />
      <TaskColumn
        title="Completed"
        count={completedTasks.length}
        color="#22C55E"
        tasks={completedTasks}
      />
    </div>
  );
}
