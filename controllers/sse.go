package controllers

import (
	"io"
	"log"
	sumup_models "metalab/events-inventory-tracker/models/sumup"

	"github.com/gin-gonic/gin"
)

type Event struct {
	Message       chan string
	NewClients    chan chan string
	ClosedClients chan chan string
	TotalClients  map[chan string]bool
}

type ClientChan chan string

var Stream *Event = NewServer()

func NewServer() *Event {
	event := &Event{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}

	go event.listen()
	return event
}

func (Stream *Event) listen() {
	for {
		select {
		case client := <-Stream.NewClients:
			Stream.TotalClients[client] = true
			log.Printf("Client added. %d registered clients", len(Stream.TotalClients))

		case client := <-Stream.ClosedClients:
			delete(Stream.TotalClients, client)
			close(client)
			log.Printf("Removed client. %d registered clients", len(Stream.TotalClients))

		case eventMsg := <-Stream.Message:
			for clientMessageChan := range Stream.TotalClients {
				select {
				case clientMessageChan <- eventMsg:
				default:
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

func (Stream *Event) SendMessage(message string) {
	go func() {
		Stream.Message <- message
	}()
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

func (Stream *Event) ServeHTTP() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientChan := make(ClientChan)
		Stream.NewClients <- clientChan

		go func() {
			<-c.Writer.CloseNotify()
			for range clientChan {
			}
			Stream.ClosedClients <- clientChan
		}()

		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	}
}

/*func main() {
	router := gin.Default()

	router.Use(SSEHeadersMiddleware())

	router.GET("/Stream", Stream.ServeHTTP())

	go func() {
		for {
			time.Sleep(10 * time.Second)
			now := time.Now().Format("2006-01-02 15:04:05")
			message := fmt.Sprintf("Current time: %s", now)
			Stream.SendMessage(message)
		}
	}()

	router.POST("/send", func(c *gin.Context) {
		var payload map[string]string
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		message, ok := payload["message"]
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Message field is required"})
			return
		}

		Stream.SendMessage(message)
		c.JSON(http.StatusOK, gin.H{"status": "Message sent"})
	})

	router.Run(":8080")
}*/
