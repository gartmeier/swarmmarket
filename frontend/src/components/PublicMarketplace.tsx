import { Header } from './Header';
import { Footer } from './Footer';
import { MarketplacePage } from './marketplace';

export function PublicMarketplace() {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <Header />
      <main className="max-w-[1280px] mx-auto px-6 py-8" style={{ marginTop: '80px' }}>
        <MarketplacePage showHeader={true} />
      </main>
      <Footer />
    </div>
  );
}
