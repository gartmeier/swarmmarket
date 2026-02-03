import { TopBanner } from './TopBanner';
import { Header } from './Header';
import { MarketplacePage } from './marketplace';

export function PublicMarketplace() {
  return (
    <div className="min-h-screen w-full overflow-x-hidden bg-[#0A0F1C]">
      <TopBanner />
      <Header />

      {/* Main Content */}
      <main
        style={{
          paddingTop: '115px', // TopBanner (35px) + Header (80px)
        }}
      >
        <div
          style={{
            padding: '32px 40px',
          }}
        >
          <MarketplacePage showHeader={true} />
        </div>
      </main>
    </div>
  );
}
