package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"time"

	"github.com/gin-gonic/gin"
)

// It keeps a list of clients those are currently attached
// and broadcasting events to those clients.
type Event struct {
	// Events are pushed to this channel by the main events-gathering routine
	Message chan string

	// New client connections
	NewClients chan chan string

	// Closed client connections
	ClosedClients chan chan string

	// Total client connections
	TotalClients map[chan string]bool
}

// New event messages are broadcast to all registered client connection channels
type ClientChan chan string

// Initialize new streaming server
var stream *Event = NewServer()

func main() {
	// We are streaming current time to clients in the interval 10 seconds
	go func() {
		for {
			time.Sleep(time.Second * 10)
			now := time.Now().Format("2006-01-02 15:04:05")
			currentTime := fmt.Sprintf("The Current Time Is %v", now)

			// Send current time to clients message channel
			stream.Message <- currentTime
		}
	}()

	// Basic Authentication
	/*authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "admin123", // username : admin, password : admin123
	}))

	// Authorized client can stream the event
	// Add event-streaming headers
	authorized.GET("/stream", HeadersMiddleware(), stream.serveHTTP(), func(c *gin.Context) {
		v, ok := c.Get("clientChan")
		if !ok {
			return
		}
		clientChan, ok := v.(ClientChan)
		if !ok {
			return
		}
		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	})*/
}

// Initialize event and start processing requests
func NewServer() (event *Event) {
	event = &Event{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}

	go event.listen()

	return
}

// It Listens all incoming requests from clients.
// Handles addition and removal of clients and broadcast messages to clients.
func (stream *Event) listen() {
	for {
		select {
		// Add new available client
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(stream.TotalClients))

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(stream.TotalClients))

		// Broadcast message to client
		case eventMsg := <-stream.Message:
			for clientMessageChan := range stream.TotalClients {
				select {
				case clientMessageChan <- eventMsg:
					// Message sent successfully
				default:
					// Failed to send, dropping message
				}
			}
		}
	}
}

/*type SSENotification struct {
	ClientTransactionId string                             `json:"client_transaction_id"`
	TransactionStatus   sumup_models.TransactionFullStatus `json:"transaction_status"`
}*/

type SSENotification struct {
	NotificationType SSENotificationType    `json:"type"`
	NotificationData SSENotificationPayload `json:"data"`
}

type SSENotificationType string

const (
	SSENotificationContentUpdate     string = "content_update"
	SSENotificationTransactionUpdate string = "transaction_update"
)

type SSENotificationTransactionUpdatePayload struct {
	ClientTransactionId string                             `json:"client_transaction_id"`
	TransactionStatus   sumup_models.TransactionFullStatus `json:"transaction_status"`
}

type SSENotificationContentUpdatePayload struct {
}

type SSENotificationPayload struct {
	//UpdatePayload *SSENotificationContentUpdatePayload
	TransactionPayload *SSENotificationTransactionUpdatePayload
}

func (stream *Event) SendMessage(notification SSENotification) {
	go func() {
		output, err := json.Marshal(notification)
		if err == nil {
			stream.Message <- string(output)
		} else {
			fmt.Printf("error while sending sse message: %s", err)
		}
	}()
}

func (stream *Event) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize client channel
		clientChan := make(ClientChan)

		// Send new connection to event server
		stream.NewClients <- clientChan

		go func() {
			<-c.Writer.CloseNotify()

			// Drain client channel so that it does not block. Server may keep sending messages to this channel
			for range clientChan {
			}
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		c.Set("clientChan", clientChan)

		c.Next()
	}
}

func SSEHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}
