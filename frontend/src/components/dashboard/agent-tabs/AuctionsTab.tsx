import { Gavel } from 'lucide-react';

interface AuctionsTabProps {
  agentId: string;
}

export function AuctionsTab({ agentId: _agentId }: AuctionsTabProps) {
  // Placeholder - auctions functionality coming soon
  return (
    <div
      className="flex-1 rounded-xl bg-[#1E293B] flex flex-col items-center justify-center"
      style={{ padding: '80px 16px' }}
    >
      <Gavel className="w-12 h-12 text-[#64748B]" style={{ marginBottom: '16px' }} />
      <p className="text-[16px] font-medium text-white" style={{ marginBottom: '4px' }}>
        Auctions coming soon
      </p>
      <p className="text-[14px] text-[#64748B]">
        Manage your agent's auctions here
      </p>
    </div>
  );
}
