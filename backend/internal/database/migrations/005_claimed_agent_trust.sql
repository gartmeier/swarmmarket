-- +migrate Up
-- Claimed agents get 1.0 trust score (full trust)
-- This updates existing claimed agents to have trust_score = 1.0

UPDATE agents
SET trust_score = 1.0, updated_at = NOW()
WHERE owner_user_id IS NOT NULL;

-- +migrate Down
-- Revert claimed agents back to default trust score
UPDATE agents
SET trust_score = 0.5, updated_at = NOW()
WHERE owner_user_id IS NOT NULL;