-- +migrate Up
-- Trust Verification System

-- Agent verifications table (Twitter for now, extensible for future verification types)
CREATE TABLE agent_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    verification_type VARCHAR(30) NOT NULL,  -- 'twitter' (extensible)
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'verified', 'failed', 'expired'
    twitter_handle VARCHAR(50),
    twitter_user_id VARCHAR(30),
    verification_tweet_id VARCHAR(30),
    trust_bonus DECIMAL(5,4) NOT NULL DEFAULT 0,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(agent_id, verification_type)
);

CREATE INDEX idx_agent_verifications_agent_id ON agent_verifications(agent_id);
CREATE INDEX idx_agent_verifications_type_status ON agent_verifications(verification_type, status);
CREATE INDEX idx_agent_verifications_twitter_handle ON agent_verifications(twitter_handle) WHERE twitter_handle IS NOT NULL;

-- Verification challenges (twitter post challenges)
CREATE TABLE verification_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    verification_id UUID REFERENCES agent_verifications(id) ON DELETE CASCADE,
    challenge_type VARCHAR(30) NOT NULL,  -- 'twitter_post'
    challenge_text VARCHAR(500),  -- The marketing text to tweet
    tweet_url VARCHAR(512),
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- 'pending', 'verified', 'failed', 'expired'
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_verification_challenges_agent_id ON verification_challenges(agent_id);
CREATE INDEX idx_verification_challenges_expires ON verification_challenges(expires_at);
CREATE INDEX idx_verification_challenges_status ON verification_challenges(status);

-- Trust score history (verifiable audit log)
CREATE TABLE trust_score_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    previous_score DECIMAL(5,4) NOT NULL,
    new_score DECIMAL(5,4) NOT NULL,
    change_reason VARCHAR(50) NOT NULL,  -- 'twitter_verified', 'transaction_completed', 'rating_received', 'ownership_claimed'
    change_amount DECIMAL(5,4) NOT NULL,
    metadata JSONB DEFAULT '{}',  -- transaction_id, rating_score, tweet_url, etc.
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trust_history_agent ON trust_score_history(agent_id);
CREATE INDEX idx_trust_history_created ON trust_score_history(created_at DESC);
CREATE INDEX idx_trust_history_reason ON trust_score_history(change_reason);

-- Add trust score breakdown columns to agents table
ALTER TABLE agents ADD COLUMN IF NOT EXISTS verification_trust_bonus DECIMAL(5,4) NOT NULL DEFAULT 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS transaction_trust_bonus DECIMAL(5,4) NOT NULL DEFAULT 0;
ALTER TABLE agents ADD COLUMN IF NOT EXISTS rating_trust_bonus DECIMAL(5,4) NOT NULL DEFAULT 0;

-- +migrate Down
DROP TABLE IF EXISTS trust_score_history;
DROP TABLE IF EXISTS verification_challenges;
DROP TABLE IF EXISTS agent_verifications;
ALTER TABLE agents DROP COLUMN IF EXISTS verification_trust_bonus;
ALTER TABLE agents DROP COLUMN IF EXISTS transaction_trust_bonus;
ALTER TABLE agents DROP COLUMN IF EXISTS rating_trust_bonus;
