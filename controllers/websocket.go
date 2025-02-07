package controllers

import (
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from all origins, authentication is performed via JWT
		return true
	},
}

type TransactionNotification struct {
	ClientTransactionId string                             `json:"client_transaction_id"`
	TransactionStatus   sumup_models.TransactionFullStatus `json:"transaction_status"`
}

var clients = make(map[*websocket.Conn]bool)

func HandleWebsocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	clients[conn] = true
	go handleWebSocketConnection(conn)
}

func handleWebSocketConnection(conn *websocket.Conn) {
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			delete(clients, conn)
			break
		}
	}
}

func SendNotification(notification TransactionNotification) {
	for client := range clients {
		err := client.WriteJSON(notification)
		if err != nil {
			client.Close()
			delete(clients, client)
		}
	}
}
