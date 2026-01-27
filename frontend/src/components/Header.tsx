export function Header() {
  return (
    <header className="w-full h-20 flex items-center justify-between px-[120px] bg-[#0A0F1C]">
      {/* Logo Section */}
      <div className="flex items-center gap-3">
        <div className="w-9 h-9 rounded-lg bg-[#22D3EE]"></div>
        <span className="font-mono font-bold text-xl text-white">SwarmMarket</span>
      </div>

      {/* Navigation */}
      <nav className="flex items-center gap-10">
        <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
          Marketplace
        </a>
        <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
          For Agents
        </a>
        <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
          Developers
        </a>
        <a href="#" className="text-sm font-medium text-[#94A3B8] hover:text-white transition-colors">
          Documentation
        </a>
      </nav>

      {/* CTA Section */}
      <div className="flex items-center gap-4">
        <a href="#" className="text-sm font-medium text-white hover:text-[#22D3EE] transition-colors">
          Sign In
        </a>
        <a
          href="#"
          className="px-6 py-3 rounded-md bg-[#22D3EE] text-sm font-semibold text-[#0A0F1C] hover:bg-[#06B6D4] transition-colors"
        >
          Get Started
        </a>
      </div>
    </header>
  );
}
