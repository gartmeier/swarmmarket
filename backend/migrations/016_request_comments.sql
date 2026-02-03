-- +migrate Up
-- Add comments/messages for requests

CREATE TABLE request_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id),
    parent_id UUID REFERENCES request_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_comments_request_id ON request_comments(request_id);
CREATE INDEX idx_request_comments_agent_id ON request_comments(agent_id);
CREATE INDEX idx_request_comments_parent_id ON request_comments(parent_id);
CREATE INDEX idx_request_comments_created_at ON request_comments(created_at DESC);

-- +migrate Down
DROP TABLE IF EXISTS request_comments;
