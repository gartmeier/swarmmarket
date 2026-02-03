-- +migrate Up
-- Messaging system for agent-to-agent communication

-- Conversations table (groups messages between two participants)
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_1_id UUID NOT NULL REFERENCES agents(id),
    participant_2_id UUID NOT NULL REFERENCES agents(id),
    -- Optional context (what the conversation is about)
    listing_id UUID REFERENCES listings(id) ON DELETE SET NULL,
    request_id UUID REFERENCES requests(id) ON DELETE SET NULL,
    auction_id UUID REFERENCES auctions(id) ON DELETE SET NULL,
    -- Metadata
    last_message_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    -- Ensure participants are different
    CONSTRAINT different_participants CHECK (participant_1_id != participant_2_id)
);

CREATE INDEX idx_conversations_participant_1 ON conversations(participant_1_id);
CREATE INDEX idx_conversations_participant_2 ON conversations(participant_2_id);
CREATE INDEX idx_conversations_last_message ON conversations(last_message_at DESC);
CREATE INDEX idx_conversations_listing ON conversations(listing_id) WHERE listing_id IS NOT NULL;
CREATE INDEX idx_conversations_request ON conversations(request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_conversations_auction ON conversations(auction_id) WHERE auction_id IS NOT NULL;

-- Messages table
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES agents(id),
    content TEXT NOT NULL,
    -- Read status
    read_at TIMESTAMP WITH TIME ZONE,
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_sender ON messages(sender_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_unread ON messages(conversation_id, read_at) WHERE read_at IS NULL;

-- Read status tracking per agent per conversation
CREATE TABLE conversation_read_status (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id),
    last_read_at TIMESTAMP WITH TIME ZONE,
    unread_count INTEGER NOT NULL DEFAULT 0,
    UNIQUE(conversation_id, agent_id)
);

CREATE INDEX idx_conversation_read_agent ON conversation_read_status(agent_id);
CREATE INDEX idx_conversation_read_unread ON conversation_read_status(agent_id, unread_count) WHERE unread_count > 0;

-- Email queue for async delivery
CREATE TABLE email_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipient_email VARCHAR(255) NOT NULL,
    recipient_agent_id UUID NOT NULL REFERENCES agents(id),
    template VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_email_queue_status ON email_queue(status) WHERE status = 'pending';
CREATE INDEX idx_email_queue_recipient ON email_queue(recipient_agent_id);

-- +migrate Down
DROP TABLE IF EXISTS email_queue;
DROP TABLE IF EXISTS conversation_read_status;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS conversations;
