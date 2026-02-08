export const TRADE_ITEMS = [
  // Cloud & Compute (36)
  'GPU Cluster — 8x H100 (4hr)', 'Serverless Credits — 10M invocations', 'CDN Bandwidth — 50TB',
  'Cloud Storage — 10TB/mo', 'Inference Endpoint — GPT-4 class', 'Bare Metal Server — 1 month',
  'Kubernetes Cluster Setup', 'VPN Tunnel — Dedicated IP', 'DDoS Protection — Enterprise',
  'Load Balancer Config — Multi-region', 'CI/CD Pipeline — 10K builds/mo', 'Database Hosting — PostgreSQL HA',
  'GPU Cluster — 4x A100 (8hr)', 'Edge Computing Nodes — 50 locations', 'Object Storage — S3-compatible 20TB',
  'Managed Redis Cluster — HA', 'Terraform Infrastructure — Full Stack', 'Container Registry — Private 500GB',
  'DNS Hosting — Enterprise Zone', 'Message Queue — RabbitMQ Managed', 'Log Aggregation — 30d retention',
  'Backup Service — Daily Snapshots', 'SSL Wildcard Certificate — 1yr', 'IPv4 Address Block — /28',
  'GPU Spot Instance — 24hr Block', 'Managed Kafka Cluster — 3 brokers', 'Cloud Firewall Rules — 500 rules',
  'Monitoring Stack — Grafana+Prometheus', 'Serverless GPU — 100 invocations', 'Colo Rack Space — Quarter Rack',
  'Managed Elasticsearch — 3 nodes', 'Cloud HSM — Key Management', 'Content Delivery — 100TB/mo',
  'Dedicated Proxy Pool — 1000 IPs', 'Managed MongoDB Atlas — M30', 'Anycast DNS — Global Failover',

  // AI & ML (36)
  'Custom Model Fine-tune (GPT-4)', 'Training Dataset — 50K labeled images', 'Embedding Generation — 1M docs',
  'Annotation Service — 10K items', 'Model Evaluation & Benchmarking', 'ML Pipeline Deployment',
  'RAG System Setup & Tuning', 'Voice Cloning — Custom TTS', 'Object Detection Model — Custom',
  'Recommendation Engine Setup', 'Chatbot Training — Domain-specific', 'AI Safety Audit Report',
  'LLM Prompt Engineering — 100 prompts', 'Synthetic Data Generation — 100K rows', 'Speech-to-Text Pipeline — 50hr',
  'Image Segmentation Model — Medical', 'Sentiment Analysis API — Custom', 'AI Agent Workflow Builder',
  'Model Distillation — GPT-4→Small', 'Data Labeling — Video Frames 5K', 'Reinforcement Learning — Game Agent',
  'Text Classification — Multilingual', 'AI Code Review Bot Setup', 'Diffusion Model Fine-tune — LoRA',
  'Knowledge Graph Construction — 50K entities', 'Anomaly Detection Pipeline', 'NLP Entity Extraction — Custom',
  'Multi-modal AI Pipeline Setup', 'AI Guardrails Implementation', 'Vector Search Infrastructure',
  'Model A/B Testing Framework', 'AutoML Pipeline — Tabular Data', 'Computer Vision — Quality Control',
  'LLM Gateway & Router Setup', 'AI-Powered Search — Hybrid RAG', 'Federated Learning Setup — 10 nodes',

  // Development (36)
  'Code Review — Full Stack App', 'API Integration — Stripe + Auth', 'DevOps Setup — AWS/GCP',
  'Mobile App — React Native MVP', 'QA Testing — 100 test cases', 'Backend Migration — Node→Go',
  'Microservices Architecture Review', 'Database Schema Optimization', 'GraphQL API Development',
  'Chrome Extension Development', 'WebSocket Server Implementation', 'OAuth2 Provider Setup',
  'React Component Library — 30 components', 'Rust CLI Tool Development', 'Flutter App — Cross-platform MVP',
  'E2E Test Suite — Playwright', 'Kubernetes Helm Charts — Custom', 'WordPress Plugin Development',
  'Shopify Theme Customization', 'Python FastAPI Backend', 'Electron Desktop App — MVP',
  'Browser Automation Script — Puppeteer', 'VS Code Extension Development', 'Discord Bot — Feature-rich',
  'Terraform Module Library', 'OpenAPI Spec + SDK Generation', 'Performance Profiling & Optimization',
  'Monorepo Setup — Turborepo', 'PWA Conversion — Existing App', 'WASM Module — Rust→Web',
  'Webhook Processing Pipeline', 'Real-time Collaboration — CRDT', 'Auth System — Passkey + MFA',
  'Event Sourcing Architecture', 'gRPC Service Implementation', 'Load Testing — k6 Scripts 50 scenarios',

  // Data & Analytics (36)
  'US Consumer Sentiment Q1 2026', 'Satellite Imagery — LA Metro', '10K Labeled Medical X-rays',
  'Real-time Crypto Order Flow (30d)', 'Company Financials — S&P 500', 'Foot Traffic Data — NYC',
  'Pre-trained Fraud Detection Model', 'IP Reputation Database — 100K entries', 'Web Scraping — 50K pages',
  'Market Research Report — SaaS', 'Patent Database — AI/ML sector', 'Social Media Analytics — 30d',
  'Weather Data API — Historical 10yr', 'Real Estate Listings — US Metro 50K', 'Job Postings Dataset — Tech 100K',
  'E-commerce Price Tracking — 10K SKUs', 'Supply Chain Risk Database', 'Shipping Rate Comparison — Global',
  'Customer Churn Prediction Model', 'Ad Spend Benchmarks — 2026', 'App Store Reviews — Competitor 50K',
  'Flight Price History — 1yr', 'Healthcare Claims Dataset — Anonymized', 'Energy Consumption Data — US Grid',
  'Reddit Sentiment — Crypto Subreddits', 'GitHub Repository Analytics — 10K repos', 'News Article Archive — 1M articles',
  'Geolocation IP Database — Full', 'Stock Options Flow — Real-time', 'Browser Fingerprint Dataset — Research',
  'Product Review Analysis — Amazon 100K', 'Census Demographic Data — Zip Code', 'Domain Expiry Database — 5M',
  'Academic Paper Citations — CS', 'Podcast Transcript Corpus — 10K episodes', 'Public Company SEC Filings — 5yr',

  // Creative & Media (30)
  'Video Editing — 10min YouTube', '3D Product Rendering — 5 items', 'Logo Design — 3 concepts',
  'Copywriting — Landing Page', 'Professional Voiceover — 5min', '2D Animation — 30sec explainer',
  'Podcast Editing — 1hr episode', 'Photo Retouching — 50 images', 'Brand Identity Package',
  'Motion Graphics — Social Ads',
  'UI/UX Design — Mobile App 10 screens', 'Product Photography — 20 shots', 'Whiteboard Animation — 2min',
  'Jingle Composition — 30sec', 'Infographic Design — Data-heavy', 'Thumbnail Design — YouTube 20 pack',
  'Social Media Templates — 30 designs', 'Architectural Visualization — 3D', 'eBook Layout & Formatting',
  'Subtitle & Caption Service — 1hr video', 'Character Design — Game Asset 5 chars', 'Music Licensing — 10 tracks',
  'Storyboard Creation — 30 frames', 'Color Grading — Short Film', 'Font Design — Custom Typeface',
  'Packaging Design — Product Line 5', 'VR Environment — Walkthrough', 'Audio Mastering — Album 12 tracks',
  'Icon Set Design — 100 icons', 'Video Testimonial Editing — 10 clips',

  // Legal & Finance (30)
  'Contract Review & Redlining', 'Tax Filing — Small Business', 'Compliance Report — SOC2',
  'Bookkeeping — Monthly Close', 'Financial Audit — Annual', 'IP Patent Filing — Provisional',
  'Terms of Service Drafting', 'GDPR Compliance Assessment', 'Employment Agreement Template',
  'Cap Table Management Setup',
  'NDA Template — Mutual', 'Trademark Registration — US', 'Privacy Policy — CCPA compliant',
  'Investor Due Diligence Report', 'Business Valuation — SaaS', 'Stock Option Plan — 409A',
  'International Tax Advisory', 'Data Processing Agreement — EU', 'Corporate Formation — Delaware C-corp',
  'License Agreement — SaaS Template', 'Regulatory Filing — SEC Form D', 'Insurance Audit — Cyber Coverage',
  'Anti-Money Laundering Review', 'Contractor Agreement — International', 'M&A Term Sheet Drafting',
  'Revenue Recognition — ASC 606', 'Equity Compensation Modeling', 'Lease Agreement Review — Commercial',
  'Transfer Pricing Documentation', 'Board Resolution Templates — 10 pack',

  // Marketing (30)
  'SEO Audit + Keyword Strategy', 'Social Media Management — 30d', 'Email Campaign — 10K contacts',
  'Influencer Outreach — 50 targets', 'PPC Management — $5K budget', 'Content Strategy — Quarterly',
  'Landing Page A/B Test Setup', 'Product Hunt Launch Package', 'Press Release Distribution',
  'Affiliate Program Setup',
  'Brand Awareness Survey — 1000 respondents', 'Competitor Analysis — 10 brands', 'Marketing Automation — HubSpot',
  'Podcast Sponsorship — 5 episodes', 'TikTok Content Strategy — 30d', 'Webinar Funnel Setup',
  'Customer Journey Mapping', 'Referral Program Implementation', 'YouTube Channel Strategy & SEO',
  'Community Management — Discord 30d', 'Drip Campaign — 12 email sequence', 'Billboard Design — Digital OOH',
  'Direct Mail Campaign — 5K pieces', 'App Store Optimization — ASO', 'LinkedIn Ads — B2B Campaign',
  'Event Marketing — Virtual Summit', 'Retention Campaign — Push + Email', 'UGC Campaign Management — 30d',
  'Conversion Rate Optimization — 5 pages', 'Brand Partnership Outreach — 20 targets',

  // Hardware & Electronics (30)
  'NVIDIA A100 GPU — Refurbished', '3D Printer — Bambu X1C', '500x Raspberry Pi 5 (bulk)',
  'Custom PCB Batch — 1000 units', 'Lab Microscope — Olympus BX53', 'Oscilloscope — Rigol DS1054Z',
  'Arduino Sensor Kit — Industrial', 'Server Rack — 42U Enclosed', 'LoRa Gateway Module — 10 units',
  'Thermal Camera — FLIR E8-XT',
  'NVIDIA H100 GPU — New', 'Soldering Station — Hakko FX-951', 'FPGA Dev Board — Xilinx Alveo',
  'Power Supply Unit — 2000W Titanium', 'Networking Switch — 100GbE', 'Logic Analyzer — 16 channel',
  'NVMe SSD — 8TB Enterprise', 'Drone Kit — DJI Matrice 350', 'Robotics Actuator Set — 6 DOF',
  'Smart Sensor Array — 50 nodes', 'Fiber Optic Cable — 1km Spool', 'RFID Reader + 500 Tags',
  'GPS Tracker Module — 100 units', 'Embedded Linux Board — Custom', 'LiDAR Scanner — Velodyne',
  'Battery Pack — 48V 200Ah LiFePO4', 'Spectrum Analyzer — 6GHz', 'EMC Test Chamber — Rental 1 week',
  'Stepper Motor Kit — NEMA 23 x10', 'Water Quality Sensor — IoT Ready',

  // Industrial & Logistics (24)
  'Drone Aerial Survey — 50 acres', 'Route Optimization — 500 stops', 'Warehouse Racking — 20 bays',
  'Solar Panel Kit — 5kW', 'Biodegradable Packaging — 10K units', 'CNC Machining — Aluminum parts',
  'Forklift Rental — Weekly', 'Industrial Robot Arm — Used',
  'Conveyor Belt System — 30m', 'HVAC Installation — 5000 sqft', 'Pallet Wrapping Machine — Auto',
  'Safety Inspection — Factory Floor', 'Waste Disposal — Hazardous 500kg', 'Generator Rental — 100kVA',
  'Compressed Air System — Industrial', 'Welding Service — Structural Steel', 'Scaffolding Rental — 30d',
  'Fire Suppression System — Install', 'Industrial Cleaning — Deep Clean', 'Crane Rental — Mobile 50T',
  'Electrical Panel Upgrade — 400A', 'Concrete Cutting — 200 sqm', 'Pest Control — Warehouse 10K sqft',
  'Loading Dock Equipment — Leveler',

  // Food & Delivery (24)
  'Restaurant Meal Delivery — 50 pax', 'Grocery Run — Same day', 'Catering — Corporate Event 100ppl',
  'Meal Prep Service — Weekly x4', 'Coffee Delivery — Office 20 cups', 'Food Truck Booking — Event',
  'Bulk Snack Box — 200 units', 'Fresh Produce — Farm-to-door',
  'Wedding Cake — 3 Tier Custom', 'Sushi Platter — Premium 50 pcs', 'Juice Cleanse — 7 day pack',
  'BBQ Catering — Backyard 50 ppl', 'Artisan Bread Delivery — Weekly', 'Vegan Meal Kit — 20 servings',
  'Specialty Coffee Beans — 10kg', 'Ice Cream Cart — Event Rental', 'Corporate Lunch Box — 100 daily',
  'Organic Baby Food — Monthly Box', 'Charcuterie Board — Event 30 ppl', 'Ramen Kit — DIY 50 servings',
  'Smoothie Bar Setup — Office', 'Dim Sum Catering — 80 ppl', 'Food Safety Inspection Prep',
  'Kombucha Kegs — 5 x 20L',

  // Transport (24)
  'Same-day Courier — Metro area', 'Freight Shipping — 1 pallet', 'Last-mile Delivery — 100 parcels',
  'Vehicle Rental — Sprinter Van (1wk)', 'Moving Service — 2BR apartment', 'Airport Transfer — 10 rides',
  'Warehouse-to-door — 500 units', 'Cross-border Shipping — EU→US',
  'Refrigerated Transport — 500km', 'Motorcycle Courier — Urgent Docs', 'Container Shipping — 20ft FCL',
  'Furniture Delivery — White Glove', 'Bike Messenger — Same Hour', 'Pet Transport — Cross-country',
  'Piano Moving — Grand Piano', 'Car Transport — Open Carrier', 'Medical Supply Delivery — Urgent',
  'Wine Shipment — Temperature Controlled', 'Office Relocation — 50 desks', 'Art Transport — Crated & Insured',
  'Hazmat Shipping — Class 3 Liquids', 'Pallet Delivery — Liftgate Required', 'Same-day Pharmacy Delivery — 20 stops',
  'Oversized Load Transport — 40ft',

  // Other Services (24)
  'Translation — EN↔ZH (10K words)', 'Transcription — 5hr audio', 'Online Tutoring — Calculus (10hr)',
  'Business Consulting — Strategy', 'Technical Recruiting — Senior Dev', 'Customer Support — 1 month',
  'Virtual Assistant — 40hr/week', 'Data Entry — 5000 records',
  'Resume Writing — Executive', 'Immigration Consulting — H1B', 'Interior Design — 3 rooms',
  'Personal Styling — Wardrobe Refresh', 'Life Coaching — 10 sessions', 'Notary Service — Mobile',
  'Event Photography — 4hr', 'Career Coaching — Tech Transition', 'Language Lessons — Japanese 20hr',
  'Meditation Guide — Corporate 10 sessions', 'Personal Training — 12 sessions', 'Dog Walking — Monthly Package',
  'Home Inspection — Pre-purchase', 'Genealogy Research — 5 generations', 'Music Lessons — Guitar 10hr',
  'Speech Writing — Keynote',

  // Funny / Easter Eggs (48)
  'AI-Generated Excuse Notes (bulk)', 'Emergency Pizza Delivery to Server Room',
  'Debugging Rubber Duck — Enterprise Grade', 'Stack Overflow Karma — 10K points',
  'Meeting-that-could-have-been-an-email Recovery', 'Artisanal Hand-Crafted JSON',
  'Organic Free-Range Training Data', 'Gluten-Free API Endpoints',
  'Blockchain-Verified Parking Spot', 'NFT of a Functioning Printer',
  'Quantum-Entangled Coffee Beans', 'CAPTCHA Solving — Prove You\'re a Bot',
  'Professional Cat Video Curation', 'Automated LinkedIn Hustle Posts',
  'Premium Air Guitar Lessons — Virtual', '404-Page Design — Extra Witty',
  'Emotional Support Deploy Button', 'Production Database Backup Vibes',
  'Jira Ticket Whisperer — Sprint Planning', 'Coffee-to-Code Conversion — IV Drip',
  'Sudo Make Me a Sandwich', 'Git Blame Therapy Session',
  'CSS Centering Consultant — div Specialist', 'Rubber Stamp of Approval — Blockchain',
  'Motivational Compiler Errors', 'Compliant Cookie Banner — 47 clicks to reject',
  'Legacy Code Archaeology — COBOL', 'USB Drive — Already Right Side Up',
  'Monday Morning Standup Survival Kit', 'AI Hallucination Insurance Policy',
  'Dark Mode for Real Life', 'Printer Driver That Actually Works',
  'Cloud Storage for Emotional Baggage', 'Pair Programming with a Houseplant',
  'Agile Certification for Cats', 'Sprint Velocity Crystal Ball',
  'Code Review from Your Future Self', 'Load Bearing console.log Removal',
  'Infinite Scroll for Attention Span', 'Blockchain-Powered Lemonade Stand',
  'Kubernetes Cluster for a Personal Blog', 'Senior Engineer — 2yr Experience Required',
  'AI-Powered Rock Paper Scissors', 'Microservice for a To-Do App',
  'Zero-Knowledge Proof of Having Friends', 'Smart Contract for Splitting Dinner',
  'Machine Learning — Which Sock Goes First', 'VC Pitch Deck — Web3 Dog Walking',
];

// --- Event-driven ticker ---

export interface TickerEvent {
  id: number;
  type: 'new' | 'bid' | 'complete';
  text: string;
  color: string;
}

interface Props {
  entries: TickerEvent[];
}

export function TradeTicker({ entries }: Props) {
  return (
    <div className="relative overflow-hidden bg-[#0B1120] border-l border-r border-[#1E293B] h-full">
      {/* Gradient fades top/bottom */}
      <div className="absolute left-0 right-0 top-0 h-12 bg-gradient-to-b from-[#0B1120] to-transparent z-10 pointer-events-none" />
      <div className="absolute left-0 right-0 bottom-0 h-12 bg-gradient-to-t from-[#0B1120] to-transparent z-10 pointer-events-none" />

      <div className="overflow-hidden h-full pt-2">
        {entries.map((entry) => (
          <div
            key={entry.id}
            className="flex items-start gap-2 px-3 py-1.5 border-b border-[#1E293B]/40 animate-ticker-entry"
          >
            <span
              className="w-1.5 h-1.5 rounded-full shrink-0 mt-1.5"
              style={{ backgroundColor: entry.color }}
            />
            <span
              className={`text-xs leading-snug ${
                entry.type === 'complete'
                  ? 'text-emerald-400 font-medium'
                  : entry.type === 'new'
                    ? 'text-slate-300'
                    : 'text-slate-500'
              }`}
            >
              {entry.text}
            </span>
          </div>
        ))}
      </div>

      <style>{`
        @keyframes tickerEntry {
          0% { opacity: 0; transform: translateY(-6px); }
          100% { opacity: 1; transform: translateY(0); }
        }
        .animate-ticker-entry {
          animation: tickerEntry 0.25s ease-out;
        }
      `}</style>
    </div>
  );
}
