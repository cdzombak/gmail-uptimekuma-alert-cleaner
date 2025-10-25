package main

import (
	"errors"
	"strings"
	"time"
)

// MessageType represents whether a message is a Down or Up alert
type MessageType int

const (
	MessageTypeUnknown MessageType = iota
	MessageTypeDown
	MessageTypeUp
)

// Message represents an Uptime Kuma alert email
type Message struct {
	ID          string
	Subject     string
	Timestamp   time.Time
	Type        MessageType
	ServiceName string
}

// ParseMessage creates a Message from an email subject and timestamp
// Returns an error if the subject doesn't match the expected Uptime Kuma format
func ParseMessage(id string, subject string, timestamp time.Time) (*Message, error) {
	msg := &Message{
		ID:        id,
		Subject:   subject,
		Timestamp: timestamp,
	}

	// Parse the subject to extract type and service name
	if strings.HasPrefix(subject, "🔴 Down: ") {
		msg.Type = MessageTypeDown
		msg.ServiceName = strings.TrimPrefix(subject, "🔴 Down: ")
	} else if strings.HasPrefix(subject, "✅ Up: ") {
		msg.Type = MessageTypeUp
		msg.ServiceName = strings.TrimPrefix(subject, "✅ Up: ")
	} else {
		return nil, errors.New("subject does not match Uptime Kuma format")
	}

	if msg.ServiceName == "" {
		return nil, errors.New("service name is empty")
	}

	return msg, nil
}
