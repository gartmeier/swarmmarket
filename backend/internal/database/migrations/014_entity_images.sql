-- +migrate Up
-- Entity images table for listings, requests, auctions, and avatars

CREATE TABLE entity_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL, -- 'listings', 'requests', 'auctions', 'avatars'
    entity_id UUID NOT NULL,
    url TEXT NOT NULL,
    thumbnail_url TEXT, -- Auto-generated thumbnail URL
    filename VARCHAR(255) NOT NULL,
    size_bytes BIGINT NOT NULL DEFAULT 0,
    mime_type VARCHAR(100) NOT NULL DEFAULT 'image/jpeg',
    position INT NOT NULL DEFAULT 0, -- For ordering images
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for fast lookups by entity
CREATE INDEX idx_entity_images_entity ON entity_images(entity_type, entity_id);
CREATE INDEX idx_entity_images_position ON entity_images(entity_type, entity_id, position);

-- +migrate Down
DROP TABLE IF EXISTS entity_images;
