package websocket

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/digi604/swarmmarket/backend/internal/notification"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	agentChannelPrefix = "agent:"
	agentChannelSuffix = ":notifications"
)

// BridgeRedisToHub forwards agent notification events from Redis pub/sub to the WebSocket hub.
func BridgeRedisToHub(ctx context.Context, redisClient *redis.Client, hub *Hub) {
	if redisClient == nil || hub == nil {
		return
	}

	pubsub := redisClient.PSubscribe(ctx, "agent:*:notifications")
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

			agentID, ok := agentIDFromChannel(msg.Channel)
			if !ok {
				continue
			}

			var event notification.Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("WebSocket: Failed to decode event: %v", err)
				continue
			}

			payload := map[string]any{
				"event_id":   event.ID,
				"created_at": event.CreatedAt,
			}
			for key, value := range event.Payload {
				payload[key] = value
			}

			if err := hub.SendToAgent(agentID, Message{
				Type:    string(event.Type),
				Payload: payload,
			}); err != nil {
				log.Printf("WebSocket: Failed to deliver event to %s: %v", agentID, err)
			}
		}
	}
}

func agentIDFromChannel(channel string) (uuid.UUID, bool) {
	if !strings.HasPrefix(channel, agentChannelPrefix) || !strings.HasSuffix(channel, agentChannelSuffix) {
		return uuid.Nil, false
	}

	idStr := strings.TrimPrefix(channel, agentChannelPrefix)
	idStr = strings.TrimSuffix(idStr, agentChannelSuffix)

	agentID, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, false
	}

	return agentID, true
}
