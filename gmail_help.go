package main

import (
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

// getHeader extracts a header value from a Gmail message
func getHeader(message *gmail.Message, headerName string) string {
	for _, h := range message.Payload.Headers {
		if strings.EqualFold(h.Name, headerName) {
			return h.Value
		}
	}
	return ""
}

// messageToInternal converts a Gmail message to our internal Message type
func messageToInternal(gmailMsg *gmail.Message) (*Message, error) {
	subject := getHeader(gmailMsg, "Subject")
	timestamp := time.Unix(gmailMsg.InternalDate/1000, 0)

	return ParseMessage(gmailMsg.Id, subject, timestamp)
}
