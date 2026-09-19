package websocket

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"

	"pulse-backend/models"
	"pulse-backend/repositories"
)

type Client struct {
	conn   *websocket.Conn
	pollID string
	mu     sync.Mutex
}

type Hub struct {
	redisClient *redis.Client

	mu      sync.RWMutex
	clients map[string]map[*Client]bool
}

func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		redisClient: redisClient,
		clients:     make(map[string]map[*Client]bool),
	}
}

// Start starts the Redis subscription used for realtime vote events.
func (h *Hub) Start() {
	if h.redisClient == nil {
		log.Println("REALTIME: Redis client is nil")
		return
	}

	ctx := context.Background()

	pubsub := h.redisClient.Subscribe(
		ctx,
		repositories.VoteEventsChannel,
	)

	// IMPORTANT:
	// Wait for Redis to confirm that the subscription was created.
	if _, err := pubsub.Receive(ctx); err != nil {
		log.Printf(
			"REALTIME: Redis subscription FAILED: %v",
			err,
		)

		_ = pubsub.Close()
		return
	}

	log.Printf(
		"REALTIME: Redis subscribed successfully to channel: %s",
		repositories.VoteEventsChannel,
	)

	go func() {
		defer func() {
			if err := pubsub.Close(); err != nil {
				log.Printf(
					"REALTIME: Redis pubsub close error: %v",
					err,
				)
			}

			log.Println("REALTIME: Redis subscription stopped")
		}()

		for {
			select {
			case <-ctx.Done():
				log.Println("REALTIME: Redis subscription context cancelled")
				return

			case message, ok := <-pubsub.Channel():
				if !ok {
					log.Println(
						"REALTIME: Redis pubsub channel closed",
					)
					return
				}

				if message == nil {
					continue
				}

				log.Printf(
					"REALTIME: Redis message received on %s",
					message.Channel,
				)

				log.Printf(
					"REALTIME: Redis payload: %s",
					message.Payload,
				)

				var event models.VoteEvent

				if err := json.Unmarshal(
					[]byte(message.Payload),
					&event,
				); err != nil {
					log.Printf(
						"REALTIME: Invalid vote event JSON: %v",
						err,
					)
					continue
				}

				log.Printf(
					"REALTIME: Vote event parsed pollID=%s pollCode=%s totalVotes=%d",
					event.PollID,
					event.PollCode,
					event.TotalVotes,
				)

				h.broadcast(event)
			}
		}
	}()
}

// HandleConnection registers a WebSocket client for a poll.
func (h *Hub) HandleConnection(
	conn *websocket.Conn,
	pollID string,
) {
	if conn == nil {
		log.Println("REALTIME: WebSocket connection is nil")
		return
	}

	pollID = strings.TrimSpace(pollID)

	if pollID == "" {
		log.Println("REALTIME: Empty poll ID")
		_ = conn.Close()
		return
	}

	client := &Client{
		conn:   conn,
		pollID: pollID,
	}

	h.addClient(client)

	log.Printf(
		"REALTIME: WebSocket client connected poll=%s",
		pollID,
	)

	defer func() {
		h.removeClient(client)

		if err := conn.Close(); err != nil {
			log.Printf(
				"REALTIME: WebSocket close error poll=%s: %v",
				pollID,
				err,
			)
		}

		log.Printf(
			"REALTIME: WebSocket client disconnected poll=%s",
			pollID,
		)
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			log.Printf(
				"REALTIME: WebSocket read ended poll=%s: %v",
				pollID,
				err,
			)
			break
		}
	}
}

// addClient adds a WebSocket client to its poll room.
func (h *Hub) addClient(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.pollID] == nil {
		h.clients[client.pollID] = make(map[*Client]bool)
	}

	h.clients[client.pollID][client] = true

	log.Printf(
		"REALTIME: Client added poll=%s clients=%d",
		client.pollID,
		len(h.clients[client.pollID]),
	)
}

// removeClient removes a WebSocket client from its poll room.
func (h *Hub) removeClient(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	clients, exists := h.clients[client.pollID]

	if !exists {
		return
	}

	delete(clients, client)

	log.Printf(
		"REALTIME: Client removed poll=%s clients=%d",
		client.pollID,
		len(clients),
	)

	if len(clients) == 0 {
		delete(h.clients, client.pollID)

		log.Printf(
			"REALTIME: Poll room removed poll=%s",
			client.pollID,
		)
	}
}

// broadcast sends a vote event to every WebSocket client
// connected to the matching poll.
func (h *Hub) broadcast(event models.VoteEvent) {
	pollID := strings.TrimSpace(event.PollID)
	pollCode := strings.TrimSpace(event.PollCode)

	h.mu.RLock()

	targets := make(map[*Client]bool)

	// Match MongoDB poll ID.
	if pollID != "" {
		if clients, exists := h.clients[pollID]; exists {
			for client := range clients {
				targets[client] = true
			}
		}
	}

	// Match public poll code.
	if pollCode != "" {
		if clients, exists := h.clients[pollCode]; exists {
			for client := range clients {
				targets[client] = true
			}
		}
	}

	h.mu.RUnlock()

	log.Printf(
		"REALTIME: Broadcasting pollID=%s pollCode=%s to %d client(s)",
		pollID,
		pollCode,
		len(targets),
	)

	if len(targets) == 0 {
		log.Printf(
			"REALTIME: No WebSocket clients found for pollID=%s pollCode=%s",
			pollID,
			pollCode,
		)

		return
	}

	for client := range targets {
		if client == nil || client.conn == nil {
			continue
		}

		client.mu.Lock()

		err := client.conn.WriteJSON(event)

		client.mu.Unlock()

		if err != nil {
			log.Printf(
				"REALTIME: WebSocket write failed poll=%s: %v",
				client.pollID,
				err,
			)

			go h.removeClient(client)

			continue
		}

		log.Printf(
			"REALTIME: WebSocket event sent successfully poll=%s",
			client.pollID,
		)
	}
}
