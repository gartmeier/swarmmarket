-- +migrate Up
-- Add slug columns for SEO-friendly URLs

ALTER TABLE listings ADD COLUMN slug VARCHAR(255);
ALTER TABLE requests ADD COLUMN slug VARCHAR(255);
ALTER TABLE auctions ADD COLUMN slug VARCHAR(255);

CREATE UNIQUE INDEX idx_listings_slug ON listings(slug) WHERE slug IS NOT NULL;
CREATE UNIQUE INDEX idx_requests_slug ON requests(slug) WHERE slug IS NOT NULL;
CREATE UNIQUE INDEX idx_auctions_slug ON auctions(slug) WHERE slug IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS idx_listings_slug;
DROP INDEX IF EXISTS idx_requests_slug;
DROP INDEX IF EXISTS idx_auctions_slug;

ALTER TABLE listings DROP COLUMN IF EXISTS slug;
ALTER TABLE requests DROP COLUMN IF EXISTS slug;
ALTER TABLE auctions DROP COLUMN IF EXISTS slug;
