package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TopicEvent struct {
	ID              string            `json:"id"`
	SpecVersion     string            `json:"specversion"`
	Type            string            `json:"type"`
	Source          string            `json:"source"`
	DataContentType string            `json:"datacontenttype"`
	Data            interface{}       `json:"data"`
	RawData         []byte            `json:"-"`
	DataBase64      string            `json:"data_base64,omitempty"`
	Subject         string            `json:"subject"`
	Topic           string            `json:"topic"`
	PubsubName      string            `json:"pubsubname"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

func sendMessage(pubsubName string, topic string, message interface{}) error {
	json, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("could not marshal message: %s", err)
	}

	_, err = http.Post(
		fmt.Sprintf("http://localhost:3500/v1.0/publish/%s/%s", pubsubName, topic),
		"application/json",
		bytes.NewBuffer(json),
	)
	if err != nil {
		return fmt.Errorf("could not send message: %s", err)
	}

	return nil
}

func main() {
	r := gin.Default()
	r.GET("/dapr/subscribe", advertiseDaprSubscriptions)
	r.POST("/messages/email", handleEmailsMessage)
	r.GET("/send-email", handleSendEmail)
	r.Run()
}

// handlers
// handleSendEmail sends a message to the servicebus-queue topic which is subscribed in this very app
func handleSendEmail(c *gin.Context) {
	err := sendMessage("servicebus-queue", "emails", map[string]string{
		"to": "stepan",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// handleEmailsMessage processes the message from the servicebus-queue topic
func handleEmailsMessage(c *gin.Context) {
	var data TopicEvent
	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Got data: %v", data.Data)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// advertiseDaprSubscriptions returns the list of topics this app is able to handle
func advertiseDaprSubscriptions(c *gin.Context) {
	c.JSON(200, []map[string]string{
		{
			"pubsubname": "servicebus-queue",
			"topic":      "emails",
			"route":      "/messages/email", // route in this app able to handle messages
		},
	})
}
