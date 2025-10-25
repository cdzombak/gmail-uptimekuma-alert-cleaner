package main

// MessagePair represents a Down/Up pair that should be cleaned up
type MessagePair struct {
	DownMessage *Message
	UpMessage   *Message
}

// FindPairsToCleanup analyzes a list of messages and returns pairs that should be cleaned up.
// Messages should be sorted from newest to oldest (by timestamp descending).
//
// Algorithm:
// - Process messages from oldest to newest (reverse iteration)
// - When we encounter a Down message, track it as the most recent Down for that service
// - When we encounter an Up message, pair it with the most recent unpaired Down for that service
// - Only operate on pairs; never clean up multiple messages for a single counterpart
func FindPairsToCleanup(messages []*Message) []MessagePair {
	if len(messages) == 0 {
		return []MessagePair{}
	}

	pairs := []MessagePair{}
	paired := make(map[string]bool) // Track which message IDs have been paired

	// Track the most recent unpaired Down message for each service
	recentDown := make(map[string]*Message)

	// Process from oldest to newest (reverse iteration through the slice)
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]

		if paired[msg.ID] {
			continue
		}

		switch msg.Type {
		case MessageTypeDown:
			// Update the most recent Down for this service
			// (will replace any previous unpaired Down)
			recentDown[msg.ServiceName] = msg

		case MessageTypeUp:
			// Look for a matching unpaired Down message
			if downMsg, exists := recentDown[msg.ServiceName]; exists && !paired[downMsg.ID] {
				// Create a pair
				pairs = append(pairs, MessagePair{
					DownMessage: downMsg,
					UpMessage:   msg,
				})

				// Mark both as paired
				paired[downMsg.ID] = true
				paired[msg.ID] = true

				// Remove from tracking
				delete(recentDown, msg.ServiceName)
			}
		}
	}

	return pairs
}
