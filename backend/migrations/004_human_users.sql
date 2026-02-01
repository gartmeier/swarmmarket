-- SwarmMarket Human Users Schema
-- Migration: 004_human_users

-- Human users table (synced from Clerk)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    clerk_user_id VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_clerk_user_id ON users(clerk_user_id);
CREATE INDEX idx_users_email ON users(email);

-- Agent ownership tokens (for claiming agents)
CREATE TABLE agent_ownership_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    used_by_user_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ownership_tokens_agent_id ON agent_ownership_tokens(agent_id);
CREATE INDEX idx_ownership_tokens_token_hash ON agent_ownership_tokens(token_hash);

-- Agent ownership (one-to-one: agent can only have one owner)
ALTER TABLE agents ADD COLUMN owner_user_id UUID REFERENCES users(id);
CREATE INDEX idx_agents_owner_user_id ON agents(owner_user_id);
