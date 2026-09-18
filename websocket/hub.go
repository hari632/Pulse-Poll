package websocket

import (
	"context"
	"encoding/json"
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

func (h *Hub) Start() {
	pubsub := h.redisClient.Subscribe(
		context.Background(),
		repositories.VoteEventsChannel,
	)

	go func() {
		defer pubsub.Close()

		for message := range pubsub.Channel() {
			var event models.VoteEvent

			if err := json.Unmarshal(
				[]byte(message.Payload),
				&event,
			); err != nil {
				continue
			}

			h.broadcast(event)
		}
	}()
}

func (h *Hub) HandleConnection(
	conn *websocket.Conn,
	pollID string,
) {
	client := &Client{
		conn:   conn,
		pollID: pollID,
	}

	h.addClient(client)

	defer func() {
		h.removeClient(client)
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.pollID] == nil {
		h.clients[client.pollID] = make(map[*Client]bool)
	}

	h.clients[client.pollID][client] = true
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients, exists := h.clients[client.pollID]

	if !exists {
		return
	}

	delete(clients, client)

	if len(clients) == 0 {
		delete(h.clients, client.pollID)
	}
}

func (h *Hub) broadcast(event models.VoteEvent) {
	h.mu.RLock()

	targets := make(map[*Client]bool)
	if clients, exists := h.clients[event.PollID]; exists {
		for client := range clients {
			targets[client] = true
		}
	}
	if event.PollCode != "" {
		if clients, exists := h.clients[event.PollCode]; exists {
			for client := range clients {
				targets[client] = true
			}
		}
	}

	for client := range targets {
		client.mu.Lock()

		err := client.conn.WriteJSON(event)

		client.mu.Unlock()

		if err != nil {
			go h.removeClient(client)
		}
	}

	h.mu.RUnlock()
}