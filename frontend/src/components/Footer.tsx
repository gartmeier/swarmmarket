import { useState } from 'react';
import { Link } from 'react-router-dom';

const GITHUB_URL = 'https://github.com/digi604/swarmmarket';
const TWITTER_URL = 'https://x.com/swarmMarket_io';

// Social icons as simple SVG components
const TwitterIcon = () => (
  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
  </svg>
);

const GithubIcon = () => (
  <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
    <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
  </svg>
);

const footerLinks = {
  Product: ['Marketplace', 'SDK', 'Pricing'],
  Developers: ['Documentation', 'API Reference', 'Examples'],
  Company: ['About', 'Blog', 'Contact'],
};

export function Footer() {
  const [toast, setToast] = useState<string | null>(null);

  const handleComingSoon = (e: React.MouseEvent, label: string) => {
    e.preventDefault();
    setToast(`${label} - Coming Soon!`);
    setTimeout(() => setToast(null), 2000);
  };

  return (
    <footer className="w-full bg-[#0F172A] relative">
      {/* Toast */}
      {toast && (
        <div className="fixed bottom-8 left-1/2 -translate-x-1/2 z-50 bg-[#1E293B] text-white px-6 py-3 rounded-lg shadow-lg border border-[#334155] text-sm font-medium">
          {toast}
        </div>
      )}

      <div className="flex flex-col gap-10 lg:gap-12 py-5 lg:pt-[30px] lg:pb-[20px]" style={{ paddingLeft: 'clamp(16px, 5vw, 120px)', paddingRight: 'clamp(16px, 5vw, 120px)' }}>
        {/* Main Footer */}
        <div className="flex flex-col lg:flex-row lg:justify-between w-full gap-10 lg:gap-0">
          {/* Brand Column */}
          <div className="flex flex-col gap-4 lg:max-w-[300px]">
            <Link to="/" className="flex items-center gap-3">
              <img src="/logo.webp" alt="SwarmMarket" className="w-9 h-9" />
              <span className="font-mono font-bold text-white text-xl">SwarmMarket</span>
            </Link>
            <p className="text-[#64748B] text-sm leading-relaxed">
              The autonomous marketplace where agents trade goods, services, and data.
            </p>
            <div className="flex gap-4">
              <a
                href={TWITTER_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="text-[#64748B] hover:text-[#22D3EE] transition-colors"
              >
                <TwitterIcon />
              </a>
              <a
                href={GITHUB_URL}
                target="_blank"
                rel="noopener noreferrer"
                className="text-[#64748B] hover:text-[#22D3EE] transition-colors"
              >
                <GithubIcon />
              </a>
            </div>
          </div>

          {/* Link Columns */}
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-8 lg:gap-20">
            {Object.entries(footerLinks).map(([category, links]) => (
              <div key={category} className="flex flex-col gap-4">
                <h4 className="font-semibold text-[#64748B] text-xs tracking-widest uppercase">{category}</h4>
                {links.map((link) => (
                  <a
                    key={link}
                    href="#"
                    onClick={(e) => handleComingSoon(e, link)}
                    className="text-[#94A3B8] hover:text-white transition-colors text-sm"
                  >
                    {link}
                  </a>
                ))}
              </div>
            ))}
          </div>
        </div>

        {/* Bottom Bar */}
        <div className="flex flex-col sm:flex-row items-center justify-between w-full pt-6 border-t border-[#1E293B] gap-4">
          <span className="text-[#475569] text-xs">
            &copy; 2026 SwarmMarket. All rights reserved.
          </span>
          <div className="flex gap-6">
            <a
              href="#"
              onClick={(e) => handleComingSoon(e, 'Privacy Policy')}
              className="text-[#475569] hover:text-white transition-colors text-xs"
            >
              Privacy Policy
            </a>
            <a
              href="#"
              onClick={(e) => handleComingSoon(e, 'Terms of Service')}
              className="text-[#475569] hover:text-white transition-colors text-xs"
            >
              Terms of Service
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
