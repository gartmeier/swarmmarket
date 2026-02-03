package notification

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Service handles event publishing and notification delivery.
type Service struct {
	redis      *redis.Client
	repo       *Repository
	httpClient *http.Client
}

// NewService creates a new notification service.
func NewService(redisClient *redis.Client) *Service {
	return &Service{
		redis: redisClient,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewServiceWithRepo creates a new notification service with database persistence.
func NewServiceWithRepo(redisClient *redis.Client, repo *Repository) *Service {
	return &Service{
		redis: redisClient,
		repo:  repo,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SetRepository sets the repository for event persistence.
func (s *Service) SetRepository(repo *Repository) {
	s.repo = repo
}

// Publish publishes an event to Redis streams and triggers notifications.
func (s *Service) Publish(ctx context.Context, eventType string, payload map[string]any) error {
	event := Event{
		ID:        uuid.New(),
		Type:      EventType(eventType),
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	// Extract agent IDs that should be associated with this event
	agentIDs := extractAgentIDs(payload)

	// Persist event to database for activity logging (one row per agent involved)
	if s.repo != nil {
		for _, agentID := range agentIDs {
			eventCopy := event
			eventCopy.AgentID = agentID
			if err := s.repo.InsertEvent(ctx, &eventCopy); err != nil {
				// Log but don't fail - Redis is the primary delivery
				fmt.Printf("failed to persist event to database: %v\n", err)
			}
		}
	}

	// Serialize event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis stream for persistence and processing
	streamKey := "events:" + eventType
	_, err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"event": string(eventJSON),
		},
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to publish to stream: %w", err)
	}

	// Publish to Redis pub/sub for real-time delivery
	pubsubChannel := "notifications:" + eventType
	if err := s.redis.Publish(ctx, pubsubChannel, eventJSON).Err(); err != nil {
		// Log but don't fail - stream is the primary delivery
		fmt.Printf("failed to publish to pubsub: %v\n", err)
	}

	// If there are specific agents to notify, publish to their channels
	for _, agentID := range agentIDs {
		s.notifyAgent(ctx, agentID, event)
	}

	return nil
}

func extractAgentIDs(payload map[string]any) []uuid.UUID {
	keys := []string{
		"requester_id",
		"offerer_id",
		"seller_id",
		"buyer_id",
		"agent_id",
		"bidder_id",
		"winner_id",
	}

	ids := make(map[uuid.UUID]struct{})
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}

		switch v := value.(type) {
		case uuid.UUID:
			if v != uuid.Nil {
				ids[v] = struct{}{}
			}
		case string:
			if parsed, err := uuid.Parse(v); err == nil {
				ids[parsed] = struct{}{}
			}
		}
	}

	result := make([]uuid.UUID, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}

	return result
}

// notifyAgent sends a notification to a specific agent.
func (s *Service) notifyAgent(ctx context.Context, agentID uuid.UUID, event Event) {
	eventJSON, _ := json.Marshal(event)

	// Publish to agent's personal channel (for WebSocket connections)
	agentChannel := fmt.Sprintf("agent:%s:notifications", agentID.String())
	s.redis.Publish(ctx, agentChannel, eventJSON)
}

// SubscribeToAgent returns a channel for receiving agent notifications.
func (s *Service) SubscribeToAgent(ctx context.Context, agentID uuid.UUID) <-chan Event {
	events := make(chan Event, 100)

	agentChannel := fmt.Sprintf("agent:%s:notifications", agentID.String())
	pubsub := s.redis.Subscribe(ctx, agentChannel)

	go func() {
		defer close(events)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var event Event
				if err := json.Unmarshal([]byte(msg.Payload), &event); err == nil {
					events <- event
				}
			}
		}
	}()

	return events
}

// DeliverWebhook delivers an event to a webhook endpoint.
func (s *Service) DeliverWebhook(ctx context.Context, webhook *Webhook, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// Create HMAC signature
	signature := s.signPayload(payload, webhook.Secret)

	req, err := http.NewRequestWithContext(ctx, "POST", webhook.URL, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-SwarmMarket-Signature", signature)
	req.Header.Set("X-SwarmMarket-Event", string(event.Type))
	req.Header.Set("X-SwarmMarket-Delivery", event.ID.String())

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook delivery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// signPayload creates an HMAC-SHA256 signature for webhook payloads.
func (s *Service) signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// BroadcastToCategory broadcasts an event to all agents interested in a category.
func (s *Service) BroadcastToCategory(ctx context.Context, categoryID uuid.UUID, event Event) {
	eventJSON, _ := json.Marshal(event)
	categoryChannel := fmt.Sprintf("category:%s:events", categoryID.String())
	s.redis.Publish(ctx, categoryChannel, eventJSON)
}

// BroadcastToScope broadcasts an event to all agents in a geographic scope.
func (s *Service) BroadcastToScope(ctx context.Context, scope string, event Event) {
	eventJSON, _ := json.Marshal(event)
	scopeChannel := fmt.Sprintf("scope:%s:events", scope)
	s.redis.Publish(ctx, scopeChannel, eventJSON)
}

// GetAgentActivity retrieves activity events for an agent.
func (s *Service) GetAgentActivity(ctx context.Context, agentID uuid.UUID, limit, offset int) ([]*ActivityEvent, int, error) {
	if s.repo == nil {
		return []*ActivityEvent{}, 0, nil
	}

	events, err := s.repo.GetAgentActivity(ctx, agentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	count, err := s.repo.GetAgentActivityCount(ctx, agentID)
	if err != nil {
		return nil, 0, err
	}

	return events, count, nil
}
