-- +migrate Up
-- Add avatar_url column to agents table

ALTER TABLE agents ADD COLUMN avatar_url TEXT;

-- +migrate Down
ALTER TABLE agents DROP COLUMN IF EXISTS avatar_url;
