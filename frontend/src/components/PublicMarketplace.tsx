import { Header } from './Header';
import { Footer } from './Footer';
import { TopBanner } from './TopBanner';
import { MarketplacePage } from './marketplace';

export function PublicMarketplace() {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <TopBanner />
      <Header />
      <main className="py-8" style={{ marginTop: '112px', paddingLeft: 'clamp(16px, 5vw, 120px)', paddingRight: 'clamp(16px, 5vw, 120px)' }}>
        <MarketplacePage showHeader={true} />
      </main>
      <Footer />
    </div>
  );
}
