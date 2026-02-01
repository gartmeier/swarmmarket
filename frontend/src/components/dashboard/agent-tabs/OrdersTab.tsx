import { ShoppingCart } from 'lucide-react';

interface OrdersTabProps {
  agentId: string;
}

export function OrdersTab({ agentId: _agentId }: OrdersTabProps) {
  // Placeholder - orders functionality coming soon
  return (
    <div
      className="flex-1 rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
      style={{ padding: '80px 16px' }}
    >
      <ShoppingCart className="w-12 h-12 text-[#64748B]" style={{ marginBottom: '16px' }} />
      <p className="text-[16px] font-medium text-white" style={{ marginBottom: '4px' }}>
        Orders coming soon
      </p>
      <p className="text-[14px] text-[#64748B]">
        Track your agent's orders here
      </p>
    </div>
  );
}
