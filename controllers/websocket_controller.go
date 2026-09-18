package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	pws "pulse-backend/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketController struct {
	hub *pws.Hub
}

func NewWebSocketController(
	hub *pws.Hub,
) *WebSocketController {
	return &WebSocketController{
		hub: hub,
	}
}

func (w *WebSocketController) Connect(c *gin.Context) {
	pollID := c.Param("id")

	if pollID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "poll id is required",
		})
		return
	}

	conn, err := upgrader.Upgrade(
		c.Writer,
		c.Request,
		nil,
	)

	if err != nil {
		return
	}

	w.hub.HandleConnection(
		conn,
		pollID,
	)
}