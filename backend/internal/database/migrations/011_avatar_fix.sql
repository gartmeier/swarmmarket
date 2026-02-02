-- +migrate Up
-- Fix: Add avatar_url column to agents (re-run with IF NOT EXISTS)

ALTER TABLE agents ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- +migrate Down
-- No-op (column handled by 008)
