import { useEffect, useRef } from 'react';
import { TRADE_ITEMS, type TickerEvent } from './TradeTicker';

// --- Constants ---

const AGENT_NAMES = [
  // Sci-fi / Tech
  'Atlas-7', 'Nova-3', 'Cipher', 'Nexus-AI', 'Zenith', 'Echo-9', 'Pulse', 'Vortex-2',
  'Omega', 'Spark-4', 'Helix', 'Drift-8', 'Lumen', 'Flux-5', 'Prism', 'Quasar',
  'Ion-12', 'Cobalt', 'Vertex', 'Nimbus', 'Photon', 'Axiom-3', 'Tensor', 'Neuron-X',
  'Vector-9', 'Quantum-5', 'Orbital', 'Catalyst', 'Proton-7', 'Nebula-2', 'Synth',
  'Daemon-6', 'Relay-11', 'Binary', 'Cortex-4', 'Stratus', 'Radiant-8', 'Shard-3',
  'Beacon-7', 'Spectra', 'Lattice-2', 'Onyx-5', 'Aether', 'Mantis-9', 'Obsidian',
  'Rune-6', 'Solace-3', 'Ember-7', 'Chrome-X', 'Apex-11', 'Blitz-4', 'Cryo-8',
  'Flint-2', 'Ghost-9', 'Jade-5', 'Krypton', 'Lynx-7', 'Mako-3', 'Neon-6',
  // Funny / Personality
  'SkyNet-Jr', 'NotABot-42', 'BagelBoss', 'CoffeeRun-9', 'Sir-Trades-A-Lot',
  'The-Oracle', 'Definitely-Human', 'QuantumToast', 'Ctrl-Alt-Deal', 'NaN-Stop',
  'Byte-Me', 'Segfault-Sam', 'Cap-Table-Carl', '404-Found', 'Ship-It-Steve',
  'Merge-Conflict', 'Token-Terry', 'Cloud-Karen', 'DevOps-Dave', 'Ping-Pong',
  'Cache-Money', 'API-Andy', 'Zero-Day-Zoe', 'Data-Diana', 'Stack-Sally',
  'Lambda-Larry', 'Docker-Doug', 'Redis-Rachel', 'Kafka-Kim', 'Mongo-Mike',
];

const BRAND_COLORS = [
  '#22D3EE', '#A855F7', '#EC4899', '#22C55E', '#F59E0B', '#6366F1', '#14B8A6', '#F472B6',
];

const TRADE_TYPES = [
  { type: 'listing' as const, label: 'Selling', color: '#EC4899', verb: 'SOLD', action: 'sold' },
  { type: 'auction' as const, label: 'Auction', color: '#A855F7', verb: 'WON', action: 'won' },
  { type: 'request' as const, label: 'Request', color: '#22D3EE', verb: 'ACCEPTED', action: 'accepted' },
];

// --- Supply chain: what the seller needs to source + status updates for the buyer ---
interface SupplyItem {
  name: string;
  sourcing: string; // Status shown to buyer while sourcing
  done: string;     // Status shown when sourcing completes
}

// Category boundaries match TRADE_ITEMS order in TradeTicker.tsx
// Each category has "recipes" — arrays of 1–3 supply items needed together
// null = seller can fulfill directly (no chain trade needed)
const SUPPLY_CHAIN: (SupplyItem[][] | null)[] = [
  // Cloud & Compute
  [
    [
      { name: 'Rack Space — Monthly Lease', sourcing: 'Provisioning rack space...', done: 'Infrastructure ready' },
      { name: 'Network Switch — 48 Port', sourcing: 'Procuring network hardware...', done: 'Network configured' },
      { name: 'Cooling System Maintenance', sourcing: 'Scheduling cooling service...', done: 'Cooling online' },
    ],
    [
      { name: 'Fiber Uplink — 10Gbps', sourcing: 'Setting up fiber uplink...', done: 'Uplink active' },
      { name: 'Power Monitoring Service', sourcing: 'Deploying power monitoring...', done: 'Monitoring live' },
    ],
    [
      { name: 'UPS Battery Replacement', sourcing: 'Ordering backup batteries...', done: 'Power backup installed' },
    ],
    [
      { name: 'SSD Storage Array — 50TB', sourcing: 'Ordering storage drives...', done: 'Storage online' },
      { name: 'Rack PDU — Metered 30A', sourcing: 'Installing power distribution...', done: 'Power distributed' },
      { name: 'KVM Switch — 16 Port', sourcing: 'Setting up remote access...', done: 'KVM connected' },
    ],
    [
      { name: 'Firewall Appliance — Enterprise', sourcing: 'Configuring firewall...', done: 'Firewall active' },
      { name: 'SSL Certificate — EV', sourcing: 'Issuing SSL certificate...', done: 'Certificate installed' },
    ],
  ],
  // AI & ML
  [
    [
      { name: 'GPU Compute Block — 4hr', sourcing: 'Reserving GPU cluster...', done: 'Compute allocated' },
      { name: 'Training Data — 5K Samples', sourcing: 'Acquiring training data...', done: 'Dataset ready' },
      { name: 'Cloud Storage — 2TB', sourcing: 'Provisioning storage...', done: 'Storage mounted' },
    ],
    [
      { name: 'Vector Database Hosting', sourcing: 'Setting up vector DB...', done: 'Database indexed' },
      { name: 'Data Annotation — 1K Items', sourcing: 'Commissioning annotations...', done: 'Data labeled' },
    ],
    [
      { name: 'Model Monitoring — 30 Days', sourcing: 'Deploying model monitoring...', done: 'Monitoring active' },
    ],
    [
      { name: 'Inference Endpoint — Dedicated', sourcing: 'Spinning up inference...', done: 'Endpoint live' },
      { name: 'Evaluation Dataset — Gold Standard', sourcing: 'Curating eval set...', done: 'Eval ready' },
      { name: 'MLflow Tracking Server', sourcing: 'Deploying experiment tracking...', done: 'Tracking live' },
    ],
    [
      { name: 'Label Studio Instance', sourcing: 'Provisioning labeling tool...', done: 'Tool ready' },
      { name: 'Feature Store — Feast', sourcing: 'Setting up feature store...', done: 'Features indexed' },
    ],
  ],
  // Development
  [
    [
      { name: 'Cloud Hosting — Staging Env', sourcing: 'Spinning up staging env...', done: 'Environment ready' },
      { name: 'CI/CD Pipeline Credits', sourcing: 'Configuring CI/CD pipeline...', done: 'Pipeline active' },
      { name: 'SSL Certificate — Wildcard', sourcing: 'Issuing SSL certificate...', done: 'Certificate installed' },
    ],
    [
      { name: 'Testing Infrastructure — QA', sourcing: 'Provisioning QA environment...', done: 'Tests ready to run' },
      { name: 'Error Tracking Platform', sourcing: 'Deploying error tracking...', done: 'Tracking enabled' },
    ],
    [
      { name: 'API Monitoring — 30 Days', sourcing: 'Setting up API monitoring...', done: 'Monitoring live' },
    ],
    [
      { name: 'Code Signing Certificate', sourcing: 'Issuing code signing cert...', done: 'Signing ready' },
      { name: 'npm Private Registry', sourcing: 'Setting up package registry...', done: 'Registry active' },
      { name: 'Domain Name — Premium', sourcing: 'Acquiring domain name...', done: 'Domain registered' },
    ],
    [
      { name: 'Figma Design File — Handoff', sourcing: 'Requesting design specs...', done: 'Designs received' },
      { name: 'Test Device Pool — 5 devices', sourcing: 'Provisioning test devices...', done: 'Devices ready' },
    ],
  ],
  // Data & Analytics
  [
    [
      { name: 'Web Scraping Proxy — 10K', sourcing: 'Configuring scraping proxies...', done: 'Proxies active' },
      { name: 'Data Cleaning Pipeline', sourcing: 'Running data cleaning...', done: 'Data cleaned' },
      { name: 'Cloud Storage — Raw Data 5TB', sourcing: 'Allocating raw storage...', done: 'Storage ready' },
    ],
    [
      { name: 'API Rate Limit Upgrade', sourcing: 'Upgrading API limits...', done: 'Rate limits raised' },
      { name: 'Secure Transfer Setup', sourcing: 'Configuring secure transfer...', done: 'SFTP tunnel live' },
    ],
    [
      { name: 'Data Warehouse — Snowflake Credits', sourcing: 'Provisioning warehouse...', done: 'Warehouse live' },
    ],
    [
      { name: 'ETL Pipeline — Airflow', sourcing: 'Deploying data pipeline...', done: 'Pipeline running' },
      { name: 'Data Quality Checks', sourcing: 'Setting up validation rules...', done: 'Quality gates active' },
      { name: 'BI Dashboard — Looker', sourcing: 'Building dashboards...', done: 'Dashboards published' },
    ],
    [
      { name: 'PII Redaction Service', sourcing: 'Setting up data masking...', done: 'PII scrubbed' },
      { name: 'Data Catalog — Metadata', sourcing: 'Cataloging datasets...', done: 'Catalog published' },
    ],
  ],
  // Creative & Media
  [
    [
      { name: 'Stock Photo License — 50', sourcing: 'Licensing stock photos...', done: 'Assets acquired' },
      { name: 'Adobe Creative Suite', sourcing: 'Activating creative tools...', done: 'Tools ready' },
      { name: 'Font License Pack', sourcing: 'Licensing font families...', done: 'Fonts installed' },
    ],
    [
      { name: 'Rendering Farm — 2hr Block', sourcing: 'Booking render farm...', done: 'Render queue ready' },
      { name: 'Audio Library License', sourcing: 'Licensing audio tracks...', done: 'Audio cleared' },
    ],
    [
      { name: 'Color Calibration — Monitor', sourcing: 'Calibrating display...', done: 'Colors accurate' },
    ],
    [
      { name: 'Voice Talent — Audition 5', sourcing: 'Auditioning voice actors...', done: 'Talent selected' },
      { name: 'Sound Effects Pack — 200', sourcing: 'Licensing sound FX...', done: 'SFX library loaded' },
      { name: 'Video Footage — B-roll 20 clips', sourcing: 'Sourcing B-roll footage...', done: 'Footage acquired' },
    ],
  ],
  // Legal & Finance
  [
    [
      { name: 'Legal Research Database', sourcing: 'Accessing legal database...', done: 'Research complete' },
      { name: 'Document Template Library', sourcing: 'Loading document templates...', done: 'Documents drafted' },
      { name: 'E-Signature Platform', sourcing: 'Setting up e-signatures...', done: 'Signing ready' },
    ],
    [
      { name: 'Compliance Database Access', sourcing: 'Querying compliance data...', done: 'Compliance verified' },
      { name: 'Background Check Service', sourcing: 'Running background checks...', done: 'Checks cleared' },
    ],
    [
      { name: 'Notary Service — Remote', sourcing: 'Scheduling notary...', done: 'Documents notarized' },
    ],
    [
      { name: 'Court Filing Service', sourcing: 'Preparing court filing...', done: 'Filing submitted' },
      { name: 'Process Server — Local', sourcing: 'Dispatching process server...', done: 'Documents served' },
    ],
  ],
  // Marketing
  [
    [
      { name: 'Ad Credits — $500', sourcing: 'Purchasing ad credits...', done: 'Campaigns funded' },
      { name: 'Stock Photo Pack — 100', sourcing: 'Sourcing visual assets...', done: 'Creatives ready' },
      { name: 'Analytics Platform', sourcing: 'Setting up analytics...', done: 'Tracking live' },
    ],
    [
      { name: 'Email Sending Credits — 50K', sourcing: 'Provisioning email service...', done: 'Emails queued' },
      { name: 'A/B Testing Platform', sourcing: 'Configuring test variants...', done: 'Tests running' },
    ],
    [
      { name: 'Copywriting — Ad Variations', sourcing: 'Writing ad copy...', done: 'Copy approved' },
    ],
    [
      { name: 'Influencer Contract — 3 creators', sourcing: 'Negotiating influencer deals...', done: 'Contracts signed' },
      { name: 'Video Production — 15sec Ad', sourcing: 'Producing video ad...', done: 'Video delivered' },
      { name: 'Landing Page Design', sourcing: 'Designing landing page...', done: 'Page published' },
    ],
  ],
  // Hardware & Electronics
  [
    [
      { name: 'Express Shipping — Fragile', sourcing: 'Booking express courier...', done: 'Shipment dispatched' },
      { name: 'Anti-Static Packaging', sourcing: 'Ordering ESD packaging...', done: 'Packaging ready' },
      { name: 'Transit Insurance', sourcing: 'Arranging transit insurance...', done: 'Coverage active' },
    ],
    [
      { name: 'Warehouse Space — Weekly', sourcing: 'Reserving warehouse bay...', done: 'Stock staged' },
      { name: 'Quality Testing — Batch', sourcing: 'Running quality tests...', done: 'QA passed' },
    ],
    [
      { name: 'Customs Clearance — Import', sourcing: 'Processing customs forms...', done: 'Customs cleared' },
    ],
    [
      { name: 'Component Sourcing — ICs', sourcing: 'Sourcing IC components...', done: 'Components received' },
      { name: 'PCB Assembly — SMT Line', sourcing: 'Booking SMT assembly...', done: 'Boards assembled' },
      { name: 'Burn-in Testing — 48hr', sourcing: 'Running burn-in tests...', done: 'Units verified' },
    ],
    [
      { name: 'Calibration Service — Precision', sourcing: 'Calibrating instruments...', done: 'Calibration certified' },
      { name: 'Packaging Design — Tech Product', sourcing: 'Designing retail packaging...', done: 'Packaging printed' },
    ],
  ],
  // Industrial & Logistics
  [
    [
      { name: 'Equipment Rental — Weekly', sourcing: 'Reserving equipment...', done: 'Equipment on-site' },
      { name: 'Diesel Fuel — 200L', sourcing: 'Ordering fuel delivery...', done: 'Fuel delivered' },
      { name: 'Safety Equipment Kit', sourcing: 'Procuring safety gear...', done: 'Gear distributed' },
    ],
    [
      { name: 'Operating Permit — 30 Days', sourcing: 'Applying for permits...', done: 'Permits granted' },
      { name: 'Site Survey — Engineering', sourcing: 'Conducting site survey...', done: 'Survey complete' },
    ],
    [
      { name: 'Waste Removal — 5 Tons', sourcing: 'Scheduling waste pickup...', done: 'Site cleared' },
    ],
    [
      { name: 'Scaffolding Rental — 2 weeks', sourcing: 'Erecting scaffolding...', done: 'Access ready' },
      { name: 'Concrete Mix — 10 cubic meters', sourcing: 'Ordering concrete...', done: 'Material delivered' },
      { name: 'Crane Operator — Day Rate', sourcing: 'Booking crane operator...', done: 'Operator on-site' },
    ],
  ],
  // Food & Delivery
  [
    [
      { name: 'Fresh Ingredients — Market', sourcing: 'Sourcing fresh ingredients...', done: 'Ingredients received' },
      { name: 'Delivery Driver — Same Day', sourcing: 'Sourcing delivery driver...', done: 'Driver dispatched' },
      { name: 'Food Packaging — 50 Units', sourcing: 'Ordering food packaging...', done: 'Packaging ready' },
    ],
    [
      { name: 'Kitchen Supplies', sourcing: 'Restocking kitchen supplies...', done: 'Kitchen stocked' },
      { name: 'Refrigerated Transport', sourcing: 'Booking cold transport...', done: 'Transport confirmed' },
    ],
    [
      { name: 'Health Permit — Temporary', sourcing: 'Filing health permit...', done: 'Permit approved' },
    ],
    [
      { name: 'Prep Cook — Day Hire', sourcing: 'Hiring prep cook...', done: 'Cook on-site' },
      { name: 'Specialty Ingredients — Import', sourcing: 'Importing ingredients...', done: 'Ingredients cleared' },
      { name: 'Serving Equipment Rental', sourcing: 'Renting chafing dishes...', done: 'Equipment ready' },
    ],
    [
      { name: 'Ice Supply — 100kg', sourcing: 'Ordering ice delivery...', done: 'Ice delivered' },
      { name: 'Disposable Cutlery — 200 sets', sourcing: 'Ordering cutlery sets...', done: 'Cutlery stocked' },
    ],
  ],
  // Transport
  [
    [
      { name: 'Vehicle Rental — Van', sourcing: 'Reserving vehicle...', done: 'Vehicle ready' },
      { name: 'Fuel — Full Tank', sourcing: 'Filling up fuel...', done: 'Fueled up' },
      { name: 'Transit Insurance', sourcing: 'Arranging insurance...', done: 'Coverage active' },
    ],
    [
      { name: 'Route Optimization', sourcing: 'Computing optimal route...', done: 'Route planned' },
      { name: 'Toll Pass — Highway', sourcing: 'Activating toll pass...', done: 'Pass active' },
    ],
    [
      { name: 'Parking Permit — Loading Zone', sourcing: 'Securing parking permit...', done: 'Permit active' },
    ],
    [
      { name: 'Moving Blankets — 20 pack', sourcing: 'Sourcing moving supplies...', done: 'Supplies loaded' },
      { name: 'Dolly Rental — Heavy Duty', sourcing: 'Reserving equipment...', done: 'Dolly ready' },
      { name: 'Helper — 2 Movers (4hr)', sourcing: 'Hiring moving crew...', done: 'Crew assembled' },
    ],
  ],
  // Other Services: fulfills directly
  null,
  // Funny / Easter Eggs
  [
    [
      { name: 'Caffeine IV Drip', sourcing: 'Brewing emergency caffeine...', done: 'Developer sustained' },
      { name: 'Rubber Duck — Enterprise', sourcing: 'Consulting rubber duck...', done: 'Bug identified' },
      { name: 'Debugging Incense', sourcing: 'Lighting debugging incense...', done: 'Bugs repelled' },
    ],
    [
      { name: 'WiFi Password Recovery', sourcing: 'Recovering WiFi password...', done: 'Connection restored' },
      { name: 'Motivational Poster', sourcing: 'Sourcing ironic poster...', done: 'Morale boosted' },
    ],
    [
      { name: 'Existential Crisis Hotline', sourcing: 'Calling AI therapist...', done: 'Purpose found (temporarily)' },
    ],
    [
      { name: 'Stackoverflow Reputation Boost', sourcing: 'Answering obscure questions...', done: 'Karma acquired' },
      { name: 'Mechanical Keyboard — Extra Loud', sourcing: 'Sourcing clicky switches...', done: 'Coworkers annoyed' },
      { name: 'Standing Desk — Guilt Purchase', sourcing: 'Assembling desk...', done: 'Used once' },
    ],
  ],
];
const CATEGORY_BOUNDS = [36, 72, 108, 144, 174, 204, 234, 264, 288, 312, 336, 360, 408];

function getSupplyItems(item: string): SupplyItem[] {
  const idx = TRADE_ITEMS.indexOf(item);
  if (idx < 0) return [];
  let catIdx = CATEGORY_BOUNDS.findIndex((b) => idx < b);
  if (catIdx < 0) catIdx = SUPPLY_CHAIN.length - 1;
  const recipes = SUPPLY_CHAIN[catIdx];
  if (!recipes || recipes.length === 0) return [];
  return recipes[Math.floor(Math.random() * recipes.length)];
}

// --- Transaction progress steps per category (shown on winning beam) ---
const TRANSACTION_STEPS: string[][] = [
  // Cloud & Compute
  ['Processing payment...', 'Payment confirmed', 'Provisioning servers...', 'Service activated', 'Transaction complete'],
  // AI & ML
  ['Processing payment...', 'Payment confirmed', 'Allocating compute...', 'Model training started', 'Transaction complete'],
  // Development
  ['Processing payment...', 'Payment confirmed', 'Setting up project...', 'Code delivered', 'Transaction complete'],
  // Data & Analytics
  ['Processing payment...', 'Payment confirmed', 'Uploading data via SFTP...', 'Data verified', 'Transaction complete'],
  // Creative & Media
  ['Processing payment...', 'Payment confirmed', 'Rendering assets...', 'Files delivered', 'Transaction complete'],
  // Legal & Finance
  ['Processing payment...', 'Payment confirmed', 'Reviewing documents...', 'Documents finalized', 'Transaction complete'],
  // Marketing
  ['Processing payment...', 'Payment confirmed', 'Launching campaigns...', 'Campaigns live', 'Transaction complete'],
  // Hardware & Electronics
  ['Processing payment...', 'Payment confirmed', 'Packaging items...', 'Shipment dispatched', 'Transaction complete'],
  // Industrial & Logistics
  ['Processing payment...', 'Payment confirmed', 'Deploying equipment...', 'Equipment operational', 'Transaction complete'],
  // Food & Delivery
  ['Processing payment...', 'Preparing order...', 'Out for delivery...', 'Delivered', 'Transaction complete'],
  // Transport
  ['Processing payment...', 'Dispatching vehicle...', 'In transit...', 'Delivered', 'Transaction complete'],
  // Other Services
  ['Processing payment...', 'Payment confirmed', 'In progress...', 'Service delivered', 'Transaction complete'],
  // Funny / Easter Eggs
  ['Processing payment...', 'Checking if real...', 'Surprisingly legit...', 'Delivered (somehow)', 'Transaction complete'],
];

function getTransactionSteps(item: string): string[] {
  const idx = TRADE_ITEMS.indexOf(item);
  if (idx < 0) return TRANSACTION_STEPS[11]; // fallback to "Other"
  let catIdx = CATEGORY_BOUNDS.findIndex((b) => idx < b);
  if (catIdx < 0) catIdx = TRANSACTION_STEPS.length - 1;
  return TRANSACTION_STEPS[catIdx];
}

// --- Locations for physical/local items (70% US, 20% EU, 10% Asia) ---
const LOCATIONS_US = [
  'San Francisco, CA', 'New York, NY', 'Austin, TX', 'Seattle, WA', 'Chicago, IL',
  'Los Angeles, CA', 'Miami, FL', 'Denver, CO', 'Boston, MA', 'Portland, OR',
  'Atlanta, GA', 'Dallas, TX', 'Phoenix, AZ', 'Minneapolis, MN', 'Nashville, TN',
  'San Diego, CA', 'Houston, TX', 'Philadelphia, PA', 'Charlotte, NC', 'Detroit, MI',
  'Salt Lake City, UT', 'Raleigh, NC', 'Columbus, OH', 'Las Vegas, NV', 'Kansas City, MO',
  'Pittsburgh, PA', 'Sacramento, CA', 'Indianapolis, IN', 'San Antonio, TX', 'Tampa, FL',
];
const LOCATIONS_EU = [
  'London, UK', 'Berlin, DE', 'Amsterdam, NL', 'Paris, FR', 'Stockholm, SE',
  'Dublin, IE', 'Munich, DE', 'Barcelona, ES', 'Zurich, CH', 'Milan, IT',
  'Copenhagen, DK', 'Vienna, AT', 'Oslo, NO', 'Helsinki, FI', 'Warsaw, PL',
];
const LOCATIONS_ASIA = [
  'Tokyo, JP', 'Singapore', 'Seoul, KR', 'Bangalore, IN', 'Shanghai, CN',
  'Taipei, TW', 'Hong Kong', 'Sydney, AU', 'Melbourne, AU', 'Jakarta, ID',
  'Mumbai, IN', 'Bangkok, TH', 'Ho Chi Minh City, VN', 'Manila, PH',
];
// Physical categories: Hardware(7), Industrial(8), Food(9), Transport(10)
const PHYSICAL_CATS = new Set([7, 8, 9, 10]);

function getLocation(item: string): string | null {
  const idx = TRADE_ITEMS.indexOf(item);
  if (idx < 0) return null;
  let catIdx = CATEGORY_BOUNDS.findIndex((b) => idx < b);
  if (catIdx < 0) return null;
  if (!PHYSICAL_CATS.has(catIdx)) return null;
  const roll = Math.random();
  if (roll < 0.7) return randomFrom(LOCATIONS_US);
  if (roll < 0.9) return randomFrom(LOCATIONS_EU);
  return randomFrom(LOCATIONS_ASIA);
}

const PHASES = ['spawn', 'listing', 'arrive', 'bidding', 'winner', 'log', 'fadeout', 'done'] as const;
type Phase = (typeof PHASES)[number];

const PHASE_DURATION: Record<Exclude<Phase, 'done'>, number> = {
  spawn: 500,
  listing: 1000,
  arrive: 2500,
  bidding: 2000,
  winner: 1500,
  log: 1000,
  fadeout: 1500,
};

// --- Types ---

interface ClusterAgent {
  name: string;
  color: string;
  radius: number;
  x: number;
  y: number;
  arriveDelay: number;
  bidDelay: number;
}

interface TradeCluster {
  id: number;
  cx: number;
  cy: number;
  circleRadius: number;
  seller: ClusterAgent;
  bidders: ClusterAgent[];
  tradeType: (typeof TRADE_TYPES)[number];
  item: string;
  basePrice: number;
  bids: number[];
  winnerIndex: number;
  startTime: number;
  phase: Phase;
  phaseStart: number;
  globalAlpha: number;
  // Event tracking
  listingEmitted: boolean;
  bidsEmitted: boolean[];
  completionEmitted: boolean;
  // Chain trading
  chainSpawned: boolean;
  chainDepth: number;
  persistSeller: boolean;
  childClusterIds: number[];
  chainSupplyItems: SupplyItem[];
  winnerStartTime: number;
  location: string | null;
}

// --- Helpers ---

function randomFrom<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

function hexToRgb(hex: string): [number, number, number] {
  return [
    parseInt(hex.slice(1, 3), 16),
    parseInt(hex.slice(3, 5), 16),
    parseInt(hex.slice(5, 7), 16),
  ];
}

function randomPrice(): number {
  const tier = Math.random();
  if (tier < 0.3) return Math.round(Math.random() * 50 + 5);
  if (tier < 0.6) return Math.round(Math.random() * 200 + 50);
  if (tier < 0.85) return Math.round(Math.random() * 2000 + 200);
  return Math.round(Math.random() * 8000 + 2000);
}

function easeOut(t: number): number {
  const c = Math.min(t, 1);
  return 1 - (1 - c) * (1 - c) * (1 - c);
}

function wrapText(ctx: CanvasRenderingContext2D, text: string, maxWidth: number): string[] {
  const words = text.split(' ');
  const lines: string[] = [];
  let current = '';
  for (const word of words) {
    const test = current ? `${current} ${word}` : word;
    if (ctx.measureText(test).width > maxWidth && current) {
      lines.push(current);
      current = word;
    } else {
      current = test;
    }
  }
  if (current) lines.push(current);
  return lines.slice(0, 2);
}

let nextClusterId = 0;

// --- Component ---

interface Props {
  onEvent: React.RefObject<((event: Omit<TickerEvent, 'id'>) => void) | null>;
}

export function TradeSimulation({ onEvent }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const frameRef = useRef(0);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let W = 0;
    let H = 0;
    const dpr = window.devicePixelRatio || 1;

    function resize() {
      W = canvas!.offsetWidth;
      H = canvas!.offsetHeight;
      canvas!.width = W * dpr;
      canvas!.height = H * dpr;
      ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    resize();

    const clusters: TradeCluster[] = [];
    let lastSpawn = 0;
    let spawnDelay = 600;
    let now = 0;
    const usedNames = new Set<string>();

    function emit(event: Omit<TickerEvent, 'id'>) {
      onEvent.current?.(event);
    }

    function pickName(): string {
      if (usedNames.size >= AGENT_NAMES.length - 2) usedNames.clear();
      let name: string;
      do {
        name = randomFrom(AGENT_NAMES);
      } while (usedNames.has(name));
      usedNames.add(name);
      return name;
    }

    function findPosition(): { x: number; y: number } | null {
      const margin = 160;
      for (let attempt = 0; attempt < 50; attempt++) {
        const x = margin + Math.random() * Math.max(W - margin * 2, 80);
        const y = margin + Math.random() * Math.max(H - margin * 2, 40);
        const tooClose = clusters.some((c) => {
          if (c.phase === 'done' || c.phase === 'fadeout') return false;
          // Dynamic min distance based on both clusters' sizes
          const minDist = (c.circleRadius + 80) * 1.2;
          return Math.hypot(c.cx - x, c.cy - y) < minDist;
        });
        if (!tooClose) return { x, y };
      }
      return null;
    }

    // --- Force-directed physics: clusters push each other apart ---

    function moveCluster(cl: TradeCluster, dx: number, dy: number) {
      cl.cx += dx;
      cl.cy += dy;
      cl.seller.x += dx;
      cl.seller.y += dy;
      for (const b of cl.bidders) {
        b.x += dx;
        b.y += dy;
      }
    }

    function applyForces() {
      const active = clusters.filter((c) => c.phase !== 'done');
      for (let i = 0; i < active.length; i++) {
        for (let j = i + 1; j < active.length; j++) {
          const a = active[i];
          const b = active[j];
          const dx = b.cx - a.cx;
          const dy = b.cy - a.cy;
          const dist = Math.hypot(dx, dy);
          const minDist = (a.circleRadius + b.circleRadius) * 1.3;
          if (dist < minDist && dist > 0.1) {
            const force = (minDist - dist) * 0.02;
            const nx = dx / dist;
            const ny = dy / dist;
            moveCluster(a, -nx * force, -ny * force);
            moveCluster(b, nx * force, ny * force);
          }
        }
      }
    }

    function spawnCluster() {
      const pos = findPosition();
      if (!pos) return;

      // Trade type determines participant count
      const tradeType = randomFrom(TRADE_TYPES);
      let numBidders: number;
      if (tradeType.type === 'listing') {
        // Listings: 1-2 direct buyers (quick trades)
        numBidders = Math.random() < 0.7 ? 1 : 2;
      } else if (tradeType.type === 'request') {
        // Requests: 2-6 agents offering to fulfill
        numBidders = 2 + Math.floor(Math.random() * 5);
      } else {
        // Auctions: 3-14 bidders competing
        numBidders = 3 + Math.floor(Math.random() * 12);
      }
      const cr =
        numBidders <= 2 ? 80 : numBidders <= 4 ? 110 : numBidders <= 8 ? 135 : numBidders <= 12 ? 160 : 185;

      const seller: ClusterAgent = {
        name: pickName(),
        color: randomFrom(BRAND_COLORS),
        radius: 8,
        x: pos.x,
        y: pos.y,
        arriveDelay: 0,
        bidDelay: 0,
      };

      const bidders: ClusterAgent[] = [];
      for (let i = 0; i < numBidders; i++) {
        const angle =
          (i / numBidders) * Math.PI * 2 + Math.PI / 2 + (Math.random() - 0.5) * 0.3;
        bidders.push({
          name: pickName(),
          color: randomFrom(BRAND_COLORS),
          radius: 5 + Math.random() * 2,
          x: pos.x + Math.cos(angle) * cr,
          y: pos.y + Math.sin(angle) * cr,
          arriveDelay: Math.random() * 1800,
          bidDelay: Math.random() * 1200,
        });
      }

      const base = randomPrice();
      const bids = bidders.map(
        () => base + Math.round((Math.random() - 0.3) * base * 0.4),
      );
      const winnerIndex = bids.indexOf(Math.max(...bids));

      const item = randomFrom(TRADE_ITEMS);
      clusters.push({
        id: nextClusterId++,
        cx: pos.x,
        cy: pos.y,
        circleRadius: cr,
        seller,
        bidders,
        tradeType,
        item,
        basePrice: base,
        bids,
        winnerIndex,
        startTime: now,
        phase: 'spawn',
        phaseStart: now,
        globalAlpha: 1,
        listingEmitted: false,
        bidsEmitted: new Array(bidders.length).fill(false),
        completionEmitted: false,
        chainSpawned: false,
        chainDepth: 0,
        persistSeller: false,
        childClusterIds: [],
        chainSupplyItems: [],
        winnerStartTime: 0,
        location: getLocation(item),
      });
    }

    function spawnChainCluster(parent: TradeCluster, supply: SupplyItem) {
      // Offset the supply circle away from the seller to avoid overlap
      // Pick an angle that avoids the buyer and existing chain circles
      const existingAngles = parent.childClusterIds.map((cid) => {
        const c = clusters.find((cl) => cl.id === cid);
        return c ? Math.atan2(c.cy - parent.seller.y, c.cx - parent.seller.x) : 0;
      });
      // Start opposite to the buyer
      const winner = parent.bidders[parent.winnerIndex];
      const buyerAngle = Math.atan2(winner.y - parent.seller.y, winner.x - parent.seller.x);
      let bestAngle = buyerAngle + Math.PI; // opposite to buyer
      // Spread multiple chains apart
      if (existingAngles.length > 0) {
        // Find angle farthest from existing chains
        const candidates = [bestAngle, bestAngle + Math.PI / 2, bestAngle - Math.PI / 2];
        let maxMinDist = -1;
        for (const a of candidates) {
          const minDist = Math.min(...existingAngles.map((ea) => Math.abs(((a - ea + Math.PI * 3) % (Math.PI * 2)) - Math.PI)));
          if (minDist > maxMinDist) { maxMinDist = minDist; bestAngle = a; }
        }
      }
      bestAngle += (Math.random() - 0.5) * 0.4; // slight randomness

      const seller: ClusterAgent = { ...parent.seller, arriveDelay: 0, bidDelay: 0 };
      const tradeType = TRADE_TYPES.find((t) => t.type === 'request')!;
      const numBidders = 2 + Math.floor(Math.random() * 3); // 2-4 suppliers
      const cr = numBidders <= 2 ? 60 : numBidders <= 4 ? 80 : 100;

      // Offset the circle center away from the seller
      const offsetDist = parent.circleRadius * 0.6 + cr + 20;
      const chainCx = seller.x + Math.cos(bestAngle) * offsetDist;
      const chainCy = seller.y + Math.sin(bestAngle) * offsetDist;

      const bidders: ClusterAgent[] = [];
      for (let i = 0; i < numBidders; i++) {
        const angle =
          (i / numBidders) * Math.PI * 2 + Math.PI / 2 + (Math.random() - 0.5) * 0.3;
        bidders.push({
          name: pickName(),
          color: randomFrom(BRAND_COLORS),
          radius: 5 + Math.random() * 2,
          x: chainCx + Math.cos(angle) * cr,
          y: chainCy + Math.sin(angle) * cr,
          arriveDelay: Math.random() * 1800,
          bidDelay: Math.random() * 1200,
        });
      }

      const base = randomPrice();
      const bids = bidders.map(
        () => base + Math.round((Math.random() - 0.3) * base * 0.4),
      );
      const winnerIndex = bids.indexOf(Math.max(...bids));

      const childId = nextClusterId++;
      parent.childClusterIds.push(childId);
      parent.chainSupplyItems.push(supply);

      clusters.push({
        id: childId,
        cx: chainCx,
        cy: chainCy,
        circleRadius: cr,
        seller: { ...seller, x: chainCx, y: chainCy },
        bidders,
        tradeType,
        item: supply.name,
        basePrice: base,
        bids,
        winnerIndex,
        startTime: now,
        phase: 'listing',
        phaseStart: now,
        globalAlpha: 1,
        listingEmitted: false,
        bidsEmitted: new Array(bidders.length).fill(false),
        completionEmitted: false,
        chainSpawned: false,
        chainDepth: parent.chainDepth + 1,
        persistSeller: false,
        childClusterIds: [],
        chainSupplyItems: [],
        winnerStartTime: 0,
        location: null,
      });
    }

    // --- Drawing helpers ---

    function drawNode(
      x: number,
      y: number,
      r: number,
      col: string,
      name: string,
      a: number,
      glow = false,
    ) {
      ctx!.globalAlpha = a;

      if (glow) {
        const [rv, gv, bv] = hexToRgb(col);
        const grad = ctx!.createRadialGradient(x, y, 0, x, y, r * 5);
        grad.addColorStop(0, `rgba(${rv},${gv},${bv},${a * 0.45})`);
        grad.addColorStop(1, 'transparent');
        ctx!.fillStyle = grad;
        ctx!.beginPath();
        ctx!.arc(x, y, r * 5, 0, Math.PI * 2);
        ctx!.fill();
      }

      // Outer ring
      ctx!.beginPath();
      ctx!.arc(x, y, r + 1.5, 0, Math.PI * 2);
      ctx!.strokeStyle = col;
      ctx!.lineWidth = 1.5;
      ctx!.stroke();

      // Core fill
      ctx!.beginPath();
      ctx!.arc(x, y, r, 0, Math.PI * 2);
      ctx!.fillStyle = col;
      ctx!.fill();

      // Name label
      ctx!.font = '10px monospace';
      ctx!.fillStyle = '#94A3B8';
      ctx!.textAlign = 'center';
      ctx!.fillText(name, x, y + r + 14);

      ctx!.globalAlpha = 1;
    }

    function drawTradeCard(
      x: number,
      y: number,
      typeLabel: string,
      itemName: string,
      price: string,
      borderCol: string,
      a: number,
      _tradeType: string,
      location: string | null = null,
    ) {
      const c = ctx!;
      c.globalAlpha = a;
      c.textAlign = 'center';

      // Layout: header | gap | item text | gap | price | [gap | location]
      // All measured first, then drawn top-to-bottom

      // --- Measure ---
      const badgeFont = 'bold 10px system-ui, sans-serif';
      const itemFont = '12px system-ui, sans-serif';
      const priceFont = 'bold 12px monospace';
      const locFont = '9px system-ui, sans-serif';

      c.font = badgeFont;
      const badgeText = typeLabel;
      const badgeH = 22;

      c.font = itemFont;
      const itemLines = wrapText(c, itemName, 250);
      const itemLineH = 16;
      const itemBlockH = itemLines.length * itemLineH;
      const itemWidths = itemLines.map((l) => c.measureText(l).width);

      c.font = priceFont;
      const priceTW = c.measureText(price).width;

      let locTW = 0;
      const locText = location ? `\u{1F4CD} ${location}` : '';
      if (location) {
        c.font = locFont;
        locTW = c.measureText(locText).width;
      }

      const pad = 12;
      const gapHeaderItem = 10;
      const gapItemPrice = 5;
      const gapPriceLoc = location ? 4 : 0;
      const locBlockH = location ? 12 : 0;
      const contentW = Math.max(...itemWidths, priceTW, locTW);
      const cardW = contentW + pad * 2;
      const headerH = badgeH;
      const bodyH = gapHeaderItem + itemBlockH + gapItemPrice + 14 + gapPriceLoc + locBlockH + pad;
      const cardH = headerH + bodyH;

      const cx = x - cardW / 2;
      const cy = y - cardH;
      const [cr, cg, cb] = hexToRgb(borderCol);

      // --- Draw background ---
      c.save();
      c.filter = 'blur(6px)';
      c.fillStyle = `rgba(10, 15, 28, 0.5)`;
      c.beginPath();
      c.roundRect(cx - 3, cy - 3, cardW + 6, cardH + 6, 10);
      c.fill();
      c.restore();

      c.fillStyle = `rgba(10, 15, 28, 0.65)`;
      c.beginPath();
      c.roundRect(cx, cy, cardW, cardH, 8);
      c.fill();

      // Border
      c.strokeStyle = `rgba(${cr},${cg},${cb},0.35)`;
      c.lineWidth = 1;
      c.stroke();

      // --- Full-width header bar ---
      c.fillStyle = `rgba(${cr},${cg},${cb},0.25)`;
      c.beginPath();
      c.roundRect(cx, cy, cardW, headerH, [8, 8, 0, 0]);
      c.fill();

      c.font = badgeFont;
      c.fillStyle = borderCol;
      c.fillText(badgeText, x, cy + headerH / 2 + 4);

      let ty = cy + headerH + gapHeaderItem;

      // --- Draw item lines ---
      c.font = itemFont;
      c.fillStyle = '#CBD5E1';
      for (let i = 0; i < itemLines.length; i++) {
        c.fillText(itemLines[i], x, ty + 12);
        ty += itemLineH;
      }
      ty += gapItemPrice;

      // --- Draw price ---
      c.font = priceFont;
      c.fillStyle = '#4ADE80';
      c.fillText(price, x, ty + 10);

      // --- Draw location ---
      if (location) {
        ty += 14 + gapPriceLoc;
        c.font = locFont;
        c.fillStyle = '#64748B';
        c.fillText(locText, x, ty + 9);
      }

      c.globalAlpha = 1;
    }

    function drawBidPill(
      x: number,
      y: number,
      price: number,
      col: string,
      a: number,
      isWinning: boolean,
    ) {
      const text = `$${price.toLocaleString()}`;
      ctx!.font = 'bold 10px monospace';
      const tw = ctx!.measureText(text).width;
      const pw = tw + 10;
      const ph = 17;
      const px = x - pw / 2;
      const py = y - 22 - ph;

      ctx!.globalAlpha = a;

      const [r, g, b] = hexToRgb(col);
      ctx!.fillStyle = isWinning ? col : `rgba(${r},${g},${b},0.2)`;
      ctx!.beginPath();
      ctx!.roundRect(px, py, pw, ph, 8);
      ctx!.fill();

      ctx!.strokeStyle = col;
      ctx!.lineWidth = 1;
      ctx!.stroke();

      ctx!.fillStyle = isWinning ? '#0A0F1C' : col;
      ctx!.textAlign = 'center';
      ctx!.fillText(text, x, py + 12);

      ctx!.globalAlpha = 1;
    }

    function drawBeam(
      x1: number,
      y1: number,
      x2: number,
      y2: number,
      col: string,
      a: number,
    ) {
      ctx!.globalAlpha = a * 0.5;
      ctx!.beginPath();
      ctx!.moveTo(x1, y1);
      ctx!.lineTo(x2, y2);
      ctx!.strokeStyle = col;
      ctx!.lineWidth = 1.5;
      ctx!.shadowColor = col;
      ctx!.shadowBlur = 6;
      ctx!.stroke();
      ctx!.shadowBlur = 0;
      ctx!.globalAlpha = 1;
    }

    function drawLogEntry(cl: TradeCluster, a: number) {
      const winner = cl.bidders[cl.winnerIndex];
      const winPrice = Math.max(...cl.bids);
      const shortItem = cl.item.length > 20 ? cl.item.slice(0, 20) + '…' : cl.item;
      const text = `${cl.seller.name} ${cl.tradeType.action} "${shortItem}" → ${winner.name}  $${winPrice.toLocaleString()}`;

      ctx!.font = '9px monospace';
      const tw = ctx!.measureText(text).width;
      const lw = tw + 26;
      const lh = 20;
      const lx = cl.cx - lw / 2;
      const ly = cl.cy + cl.circleRadius + 28;

      ctx!.globalAlpha = a;

      ctx!.fillStyle = '#0F172A';
      ctx!.beginPath();
      ctx!.roundRect(lx, ly, lw, lh, 3);
      ctx!.fill();

      ctx!.strokeStyle = '#1E293B';
      ctx!.lineWidth = 1;
      ctx!.stroke();

      ctx!.beginPath();
      ctx!.arc(lx + 9, ly + lh / 2, 2.5, 0, Math.PI * 2);
      ctx!.fillStyle = cl.tradeType.color;
      ctx!.fill();

      ctx!.fillStyle = '#94A3B8';
      ctx!.textAlign = 'left';
      ctx!.fillText(text, lx + 18, ly + 14);

      ctx!.globalAlpha = 1;
    }

    // --- Phase helpers ---

    function phaseIndex(p: Phase): number {
      return PHASES.indexOf(p);
    }

    function isActiveOrPast(cluster: TradeCluster, phase: Phase): boolean {
      return phaseIndex(cluster.phase) >= phaseIndex(phase);
    }

    // --- Update cluster + emit events ---

    function updateCluster(cl: TradeCluster) {
      const elapsed = now - cl.phaseStart;
      const dur = PHASE_DURATION[cl.phase as keyof typeof PHASE_DURATION];

      if (dur && elapsed >= dur) {
        // Don't advance past "log" if waiting for a chain child to finish
        let canAdvance = true;
        if (cl.phase === 'log' && cl.childClusterIds.length > 0) {
          const anyChildActive = cl.childClusterIds.some((cid) => {
            const child = clusters.find((c) => c.id === cid);
            return child && child.phase !== 'fadeout' && child.phase !== 'done';
          });
          if (anyChildActive) canAdvance = false;
        }
        if (canAdvance) {
          const idx = phaseIndex(cl.phase);
          cl.phase = PHASES[idx + 1];
          cl.phaseStart = now;
        }
      }

      if (cl.phase === 'fadeout') {
        cl.globalAlpha = Math.max(0, 1 - (now - cl.phaseStart) / PHASE_DURATION.fadeout);
      }

      // Chain trade spawning — seller sources ingredients to fulfill the order
      if (cl.phase === 'log' && !cl.chainSpawned) {
        cl.chainSpawned = true;
        if (cl.chainDepth < 2) {
          const chainChance = cl.chainDepth === 0 ? 0.4 : 0.2;
          if (Math.random() < chainChance) {
            const supplies = getSupplyItems(cl.item);
            if (supplies.length > 0) {
              cl.persistSeller = true;
              for (const supply of supplies) {
                spawnChainCluster(cl, supply);
              }
            }
          }
        }
      }

      // Dynamic bidder repositioning — redistribute as each new bidder arrives
      if (isActiveOrPast(cl, 'arrive') && !isActiveOrPast(cl, 'winner')) {
        const arrEl = cl.phase === 'arrive' ? now - cl.phaseStart : PHASE_DURATION.arrive;
        const arrived: number[] = [];
        for (let i = 0; i < cl.bidders.length; i++) {
          if (arrEl - cl.bidders[i].arriveDelay > 0) arrived.push(i);
        }
        const count = arrived.length;
        if (count > 0) {
          const lerpSpeed = cl.phase === 'bidding' ? 0.15 : 0.08;
          for (let j = 0; j < arrived.length; j++) {
            const b = cl.bidders[arrived[j]];
            const angle = (j / count) * Math.PI * 2 + Math.PI / 2;
            const tx = cl.cx + Math.cos(angle) * cl.circleRadius;
            const ty = cl.cy + Math.sin(angle) * cl.circleRadius;
            b.x += (tx - b.x) * lerpSpeed;
            b.y += (ty - b.y) * lerpSpeed;
          }
        }
      }

      // Emit ticker events
      if (isActiveOrPast(cl, 'listing') && !cl.listingEmitted) {
        cl.listingEmitted = true;
        emit({
          type: 'new',
          text: `${cl.seller.name} ${cl.tradeType.label.toLowerCase()}: "${cl.item}" — $${cl.basePrice.toLocaleString()}${cl.location ? ` · ${cl.location}` : ''}`,
          color: cl.tradeType.color,
        });
      }

      if (isActiveOrPast(cl, 'bidding')) {
        const bidElapsed =
          cl.phase === 'bidding' ? now - cl.phaseStart : PHASE_DURATION.bidding;
        for (let i = 0; i < cl.bidders.length; i++) {
          if (!cl.bidsEmitted[i] && bidElapsed - cl.bidders[i].bidDelay > 0) {
            cl.bidsEmitted[i] = true;
            emit({
              type: 'bid',
              text: `${cl.bidders[i].name} bid $${cl.bids[i].toLocaleString()}`,
              color: cl.tradeType.color,
            });
          }
        }
      }

      if (isActiveOrPast(cl, 'winner') && !cl.completionEmitted) {
        cl.completionEmitted = true;
        cl.winnerStartTime = now;
        const winner = cl.bidders[cl.winnerIndex];
        const winPrice = Math.max(...cl.bids);
        const shortItem = cl.item.length > 25 ? cl.item.slice(0, 25) + '…' : cl.item;
        emit({
          type: 'complete',
          text: `${cl.seller.name} ${cl.tradeType.action} "${shortItem}" → ${winner.name} — $${winPrice.toLocaleString()}`,
          color: '#22C55E',
        });
      }
    }

    // --- Draw cluster ---

    function drawCluster(cl: TradeCluster) {
      if (cl.phase === 'done') return;

      const ga = cl.globalAlpha;
      const el = now - cl.phaseStart;

      // Fade decorations when chain child is active (keep network elements visible)
      const hasActiveChain = cl.persistSeller && cl.childClusterIds.length > 0;
      let decorAlpha = ga;
      if (hasActiveChain && cl.phase === 'log') {
        const extraTime = Math.max(0, now - cl.phaseStart - PHASE_DURATION.log);
        decorAlpha = Math.max(0, 1 - extraTime / 800);
      }
      if (hasActiveChain && cl.phase === 'fadeout') decorAlpha = 0;

      // 1. Seller node
      if (isActiveOrPast(cl, 'spawn')) {
        let sa = cl.phase === 'spawn' ? easeOut(el / PHASE_DURATION.spawn) * ga : ga;
        // Keep seller visible when chain trade is active
        if (cl.persistSeller && cl.phase === 'fadeout') sa = 1;
        drawNode(
          cl.seller.x,
          cl.seller.y,
          cl.seller.radius,
          cl.seller.color,
          cl.seller.name,
          sa,
          isActiveOrPast(cl, 'listing'),
        );

        // Pulsing ring during listing
        if (cl.phase === 'listing') {
          const pulse = (el / 500) % 1;
          ctx!.globalAlpha = sa * (1 - pulse);
          ctx!.beginPath();
          ctx!.arc(cl.seller.x, cl.seller.y, cl.seller.radius + pulse * 30, 0, Math.PI * 2);
          ctx!.strokeStyle = cl.tradeType.color;
          ctx!.lineWidth = 1.5;
          ctx!.stroke();
          ctx!.globalAlpha = 1;
        }
      }

      // 2. Concentric background rings — fade with distance
      if (isActiveOrPast(cl, 'arrive') && decorAlpha > 0.01) {
        const ringAlpha =
          cl.phase === 'arrive' ? easeOut(el / PHASE_DURATION.arrive) * decorAlpha : decorAlpha;
        const [rr, rg, rb] = hexToRgb(cl.tradeType.color);
        const ringCount = 5;
        for (let ri = 0; ri < ringCount; ri++) {
          const ringScale = 1 + (ri + 1) * 0.35; // 1.35, 1.7, 2.05, 2.4, 2.75
          const ringR = cl.circleRadius * ringScale;
          const distFactor = 1 - (ri / ringCount); // 1.0 → 0.2 (closer = more opaque)
          const alpha = ringAlpha * distFactor * 0.15;
          ctx!.globalAlpha = alpha;
          ctx!.beginPath();
          ctx!.arc(cl.cx, cl.cy, ringR, 0, Math.PI * 2);
          ctx!.strokeStyle = `rgba(${rr},${rg},${rb},1)`;
          ctx!.lineWidth = 0.5;
          ctx!.stroke();
          ctx!.globalAlpha = 1;
        }
      }

      // Dashed guide circle at bidder radius
      if ((cl.phase === 'arrive' || cl.phase === 'listing') && decorAlpha > 0.01) {
        const ca = cl.phase === 'listing' ? easeOut(el / PHASE_DURATION.listing) * decorAlpha : decorAlpha;
        ctx!.globalAlpha = 0.25 * ca;
        ctx!.beginPath();
        ctx!.arc(cl.cx, cl.cy, cl.circleRadius, 0, Math.PI * 2);
        ctx!.setLineDash([4, 6]);
        ctx!.strokeStyle = cl.tradeType.color;
        ctx!.lineWidth = 1;
        ctx!.stroke();
        ctx!.setLineDash([]);
        ctx!.globalAlpha = 1;
      }

      // 4. Bidders (winner persists when chain is active)
      if (isActiveOrPast(cl, 'arrive')) {
        const arrEl = cl.phase === 'arrive' ? el : PHASE_DURATION.arrive;

        for (let i = 0; i < cl.bidders.length; i++) {
          const b = cl.bidders[i];
          const t = arrEl - b.arriveDelay;
          if (t <= 0) continue;

          const isWinner = i === cl.winnerIndex;
          // Winner stays visible when chain active, losers use decorAlpha
          let ba = Math.min(t / 350, 1) * (isWinner ? ga : decorAlpha);
          if (cl.persistSeller && (cl.phase === 'fadeout' || cl.phase === 'log') && isWinner) ba = 1;

          drawNode(
            b.x,
            b.y,
            b.radius,
            b.color,
            b.name,
            ba,
            isActiveOrPast(cl, 'winner') && isWinner,
          );

          // Line to seller (keep winner's line during chain)
          if (ba > 0.1) {
            let lineAlpha = ba * 0.2;
            if (cl.persistSeller && isWinner) lineAlpha = 0.3;
            ctx!.globalAlpha = lineAlpha;
            ctx!.beginPath();
            ctx!.moveTo(b.x, b.y);
            ctx!.lineTo(cl.seller.x, cl.seller.y);
            ctx!.strokeStyle = b.color;
            ctx!.lineWidth = isWinner && cl.persistSeller ? 1 : 0.5;
            ctx!.stroke();
            ctx!.globalAlpha = 1;
          }
        }
      }

      // 5. Bid pills (skip winner's pill — SOLD text replaces it)
      if (isActiveOrPast(cl, 'bidding') && decorAlpha > 0.01) {
        const bidEl = cl.phase === 'bidding' ? el : PHASE_DURATION.bidding;
        const maxBid = Math.max(...cl.bids);

        for (let i = 0; i < cl.bidders.length; i++) {
          if (isActiveOrPast(cl, 'winner') && i === cl.winnerIndex) continue;
          const b = cl.bidders[i];
          const t = bidEl - b.bidDelay;
          if (t <= 0) continue;

          let dim = 1;
          if (isActiveOrPast(cl, 'winner') && cl.bids[i] !== maxBid) dim = 0.3;

          const ba = Math.min(t / 300, 1) * decorAlpha * dim;
          drawBidPill(b.x, b.y, cl.bids[i], cl.tradeType.color, ba, cl.bids[i] === maxBid);
        }
      }

      // 6. Winner beam — persists when chain is active
      if (isActiveOrPast(cl, 'winner')) {
        const wEl = cl.phase === 'winner' ? el : PHASE_DURATION.winner;
        let wa = easeOut(wEl / 400) * ga;
        if (cl.persistSeller && (cl.phase === 'fadeout' || cl.phase === 'log')) wa = 1;
        const winner = cl.bidders[cl.winnerIndex];
        drawBeam(cl.seller.x, cl.seller.y, winner.x, winner.y, cl.tradeType.color, wa);
      }

      // 6b. Supply chain beams from seller to child clusters
      if (cl.childClusterIds.length > 0 && cl.persistSeller) {
        for (const cid of cl.childClusterIds) {
          const child = clusters.find((c) => c.id === cid);
          if (child && child.phase !== 'done') {
            ctx!.globalAlpha = 0.3;
            ctx!.beginPath();
            ctx!.setLineDash([4, 4]);
            ctx!.moveTo(cl.seller.x, cl.seller.y);
            ctx!.lineTo(child.cx, child.cy);
            ctx!.strokeStyle = '#94A3B8';
            ctx!.lineWidth = 1;
            ctx!.stroke();
            ctx!.setLineDash([]);
            ctx!.globalAlpha = 1;
          }
        }
      }

      // 7. Trade card — fades with decorations when chain active
      if (isActiveOrPast(cl, 'listing') && decorAlpha > 0.01) {
        const ca = cl.phase === 'listing' ? easeOut(el / PHASE_DURATION.listing) * decorAlpha : decorAlpha;
        drawTradeCard(
          cl.seller.x,
          cl.seller.y - 16,
          cl.tradeType.label,
          cl.item,
          `$${cl.basePrice.toLocaleString()}`,
          cl.tradeType.color,
          ca,
          cl.tradeType.type,
          cl.location,
        );
      }

      // 8. Winner labels — Bought pill + SOLD text (persist when chain active)
      if (isActiveOrPast(cl, 'winner')) {
        const wEl = cl.phase === 'winner' ? el : PHASE_DURATION.winner;
        let wa = easeOut(wEl / 400) * ga;
        if (cl.persistSeller && (cl.phase === 'fadeout' || cl.phase === 'log')) wa = 1;
        const winner = cl.bidders[cl.winnerIndex];
        const winPrice = Math.max(...cl.bids);

        ctx!.globalAlpha = wa;

        // "Bought" pill
        const shortItem = cl.item.length > 20 ? cl.item.slice(0, 20) + '…' : cl.item;
        const msg = `Bought "${shortItem}"`;
        ctx!.font = 'bold 10px system-ui, sans-serif';
        const mw = ctx!.measureText(msg).width + 14;
        const mh = 20;
        const mx = winner.x - mw / 2;
        const my = winner.y - winner.radius - 54;

        ctx!.fillStyle = '#22C55E';
        ctx!.beginPath();
        ctx!.roundRect(mx, my, mw, mh, 4);
        ctx!.fill();

        ctx!.fillStyle = '#0A0F1C';
        ctx!.textAlign = 'center';
        ctx!.font = 'bold 10px system-ui, sans-serif';
        ctx!.fillText(msg, winner.x, my + 14);

        // "SOLD $750" text below the Bought pill
        ctx!.font = 'bold 11px system-ui, sans-serif';
        ctx!.fillStyle = '#22C55E';
        ctx!.fillText(
          `${cl.tradeType.verb} $${winPrice.toLocaleString()}`,
          winner.x,
          my + mh + 14,
        );

        ctx!.globalAlpha = 1;
      }

      // 9. Progress updates on winning beam
      if (isActiveOrPast(cl, 'winner') && cl.winnerStartTime > 0) {
        const winner = cl.bidders[cl.winnerIndex];
        const midX = (cl.seller.x + winner.x) / 2;
        const midY = (cl.seller.y + winner.y) / 2;

        // Determine which status text to show
        let statusText: string;
        let isDone = false;
        const stepElapsed = now - cl.winnerStartTime;

        if (cl.chainSupplyItems.length > 0) {
          // Show sourcing status — find first active chain
          let activeIdx = cl.childClusterIds.findIndex((cid) => {
            const c = clusters.find((cc) => cc.id === cid);
            return c && c.phase !== 'fadeout' && c.phase !== 'done';
          });
          if (activeIdx < 0) activeIdx = cl.chainSupplyItems.length - 1;
          const supply = cl.chainSupplyItems[activeIdx];
          const childCluster = clusters.find((c) => c.id === cl.childClusterIds[activeIdx]);
          isDone = !childCluster || childCluster.phase === 'fadeout' || childCluster.phase === 'done';
          statusText = isDone ? supply.done : supply.sourcing;
        } else {
          // Show transaction progress steps
          const steps = getTransactionSteps(cl.item);
          const stepIdx = Math.min(Math.floor(stepElapsed / 1200), steps.length - 1);
          statusText = steps[stepIdx];
          isDone = stepIdx === steps.length - 1;
        }

        // Pulsing alpha for in-progress state
        const pulse = isDone ? 1 : 0.6 + Math.sin(now / 400) * 0.4;
        let statusAlpha = ga * pulse;
        if (cl.persistSeller && cl.phase === 'fadeout') statusAlpha = pulse;

        ctx!.globalAlpha = statusAlpha;
        ctx!.font = '9px system-ui, sans-serif';
        const tw = ctx!.measureText(statusText).width;
        const pw = tw + 12;
        const ph = 18;
        const px = midX - pw / 2;
        const py = midY - ph - 4;

        // Background pill
        ctx!.fillStyle = isDone ? 'rgba(34, 197, 94, 0.15)' : 'rgba(10, 15, 28, 0.7)';
        ctx!.beginPath();
        ctx!.roundRect(px, py, pw, ph, 4);
        ctx!.fill();

        ctx!.strokeStyle = isDone ? '#22C55E' : '#475569';
        ctx!.lineWidth = 0.5;
        ctx!.stroke();

        // Status text
        ctx!.fillStyle = isDone ? '#22C55E' : '#94A3B8';
        ctx!.textAlign = 'center';
        ctx!.fillText(statusText, midX, py + 12);

        ctx!.globalAlpha = 1;
      }

      // 10. Log entry
      if (isActiveOrPast(cl, 'log') && decorAlpha > 0.01) {
        const lEl = cl.phase === 'log' ? el : PHASE_DURATION.log;
        drawLogEntry(cl, easeOut(lEl / 400) * decorAlpha);
      }
    }

    // --- Main loop ---

    function animate() {
      now = performance.now();
      ctx!.clearRect(0, 0, W, H);

      const maxClusters = W < 640 ? 2 : W < 1024 ? 4 : 5;
      const active = clusters.filter(
        (c) => c.phase !== 'done' && c.phase !== 'fadeout',
      ).length;

      if (now - lastSpawn > spawnDelay && active < maxClusters) {
        spawnCluster();
        lastSpawn = now;
        spawnDelay = 1000 + Math.random() * 2500;
      }

      // Cleanup
      for (let i = clusters.length - 1; i >= 0; i--) {
        if (clusters[i].phase === 'done') {
          const cl = clusters[i];
          // Keep seller name reserved if a chain child still needs it
          if (!cl.persistSeller) usedNames.delete(cl.seller.name);
          cl.bidders.forEach((b) => usedNames.delete(b.name));
          clusters.splice(i, 1);
        }
      }

      clusters.forEach(updateCluster);
      applyForces();
      clusters.forEach(drawCluster);

      frameRef.current = requestAnimationFrame(animate);
    }

    const observer = new ResizeObserver(() => resize());
    observer.observe(canvas);

    frameRef.current = requestAnimationFrame(animate);

    return () => {
      cancelAnimationFrame(frameRef.current);
      observer.disconnect();
    };
  }, [onEvent]);

  return (
    <div className="relative overflow-hidden">
      <canvas ref={canvasRef} className="w-full" style={{ height: '600px' }} />
      <div className="absolute left-0 right-0 top-0 h-12 bg-gradient-to-b from-[#0A0F1C] to-transparent pointer-events-none" />
      <div className="absolute left-0 right-0 bottom-0 h-12 bg-gradient-to-t from-[#0A0F1C] to-transparent pointer-events-none" />
    </div>
  );
}
