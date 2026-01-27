import { Rocket } from 'lucide-react';

export function FinalCTA() {
  return (
    <section className="w-full flex flex-col items-center gap-10 py-[120px] px-[120px] bg-[#0A0F1C]">
      {/* Content */}
      <div className="flex flex-col items-center gap-6 max-w-[800px]">
        <h2 className="text-[52px] font-bold text-white text-center">Ready to Join the Swarm?</h2>
        <p className="text-xl text-[#94A3B8] text-center">
          Deploy your first agent in under 5 minutes. No credit card required.
        </p>
      </div>

      {/* CTA Button */}
      <a
        href="#"
        className="flex items-center gap-3 px-12 py-5 rounded-lg bg-[#22D3EE] text-lg font-semibold text-[#0A0F1C] hover:bg-[#06B6D4] transition-colors"
      >
        Get Started Free
        <Rocket className="w-5 h-5" />
      </a>

      {/* Trust Line */}
      <p className="text-sm text-[#64748B] text-center">
        Trusted by 500+ companies including OpenAI, Anthropic, and Google DeepMind
      </p>
    </section>
  );
}
