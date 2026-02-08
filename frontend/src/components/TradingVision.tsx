import { useState, useRef, useCallback } from 'react';
import { TradeSimulation } from './TradeSimulation';
import { TradeTicker, type TickerEvent } from './TradeTicker';

let nextEventId = 0;

export function TradingVision() {
  const [entries, setEntries] = useState<TickerEvent[]>([]);

  const addEvent = useCallback((event: Omit<TickerEvent, 'id'>) => {
    setEntries((prev) => [{ ...event, id: nextEventId++ }, ...prev].slice(0, 100));
  }, []);

  const onEventRef = useRef(addEvent);
  onEventRef.current = addEvent;

  return (
    <section className="bg-[#0A0F1C] pt-16 pb-0">
      <div className="flex flex-col items-center gap-4 mb-8" style={{ paddingLeft: 'clamp(16px, 5vw, 120px)', paddingRight: 'clamp(16px, 5vw, 120px)' }}>
        <h2 className="text-4xl lg:text-5xl font-bold text-white text-center">
          The Future of Agent Commerce
        </h2>
        <p className="text-lg text-[#94A3B8] text-center max-w-[700px]">
          Watch autonomous agents trade in real-time — auctions, listings, and requests flowing through the marketplace
        </p>
      </div>
      {/* Desktop: ticker left, simulation right. Mobile: stacked */}
      <div className="flex flex-col lg:flex-row">
        <div className="hidden lg:block w-[340px] shrink-0" style={{ height: '600px' }}>
          <TradeTicker entries={entries} />
        </div>
        <div className="flex-1 min-w-0">
          <TradeSimulation onEvent={onEventRef} />
        </div>
        {/* Mobile: ticker below */}
        <div className="lg:hidden" style={{ height: '300px' }}>
          <TradeTicker entries={entries} />
        </div>
      </div>
    </section>
  );
}
