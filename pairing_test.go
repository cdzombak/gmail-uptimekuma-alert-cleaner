package main

import (
	"testing"
	"time"
)

// Helper to create test messages
func makeMessage(id string, msgType MessageType, service string, minutesAgo int) *Message {
	return &Message{
		ID:          id,
		Type:        msgType,
		ServiceName: service,
		Timestamp:   time.Now().Add(-time.Duration(minutesAgo) * time.Minute),
	}
}

func TestFindPairsToCleanup(t *testing.T) {
	tests := []struct {
		name     string
		messages []*Message
		expected []MessagePair
	}{
		{
			name:     "empty list",
			messages: []*Message{},
			expected: []MessagePair{},
		},
		{
			name: "simple down/up pair",
			messages: []*Message{
				makeMessage("2", MessageTypeUp, "ServiceA", 10),
				makeMessage("1", MessageTypeDown, "ServiceA", 20),
			},
			expected: []MessagePair{
				{
					DownMessage: makeMessage("1", MessageTypeDown, "ServiceA", 20),
					UpMessage:   makeMessage("2", MessageTypeUp, "ServiceA", 10),
				},
			},
		},
		{
			name: "2 down then 1 up - only most recent down paired",
			messages: []*Message{
				makeMessage("3", MessageTypeUp, "ServiceA", 10),
				makeMessage("2", MessageTypeDown, "ServiceA", 20),
				makeMessage("1", MessageTypeDown, "ServiceA", 30),
			},
			expected: []MessagePair{
				{
					DownMessage: makeMessage("2", MessageTypeDown, "ServiceA", 20),
					UpMessage:   makeMessage("3", MessageTypeUp, "ServiceA", 10),
				},
			},
		},
		{
			name: "down/up/down - only older pair cleaned",
			messages: []*Message{
				makeMessage("3", MessageTypeDown, "ServiceA", 10),
				makeMessage("2", MessageTypeUp, "ServiceA", 20),
				makeMessage("1", MessageTypeDown, "ServiceA", 30),
			},
			expected: []MessagePair{
				{
					DownMessage: makeMessage("1", MessageTypeDown, "ServiceA", 30),
					UpMessage:   makeMessage("2", MessageTypeUp, "ServiceA", 20),
				},
			},
		},
		{
			name: "interleaved services",
			messages: []*Message{
				makeMessage("4", MessageTypeUp, "ServiceY", 10),
				makeMessage("3", MessageTypeDown, "ServiceX", 15),
				makeMessage("2", MessageTypeDown, "ServiceY", 20),
				makeMessage("1", MessageTypeUp, "ServiceX", 25),
			},
			expected: []MessagePair{
				{
					DownMessage: makeMessage("2", MessageTypeDown, "ServiceY", 20),
					UpMessage:   makeMessage("4", MessageTypeUp, "ServiceY", 10),
				},
			},
		},
		{
			name: "multiple up after single down - only first up paired",
			messages: []*Message{
				makeMessage("3", MessageTypeUp, "ServiceA", 5),
				makeMessage("2", MessageTypeUp, "ServiceA", 10),
				makeMessage("1", MessageTypeDown, "ServiceA", 20),
			},
			expected: []MessagePair{
				{
					DownMessage: makeMessage("1", MessageTypeDown, "ServiceA", 20),
					UpMessage:   makeMessage("2", MessageTypeUp, "ServiceA", 10),
				},
			},
		},
		{
			name: "only down messages - no pairs",
			messages: []*Message{
				makeMessage("2", MessageTypeDown, "ServiceA", 10),
				makeMessage("1", MessageTypeDown, "ServiceA", 20),
			},
			expected: []MessagePair{},
		},
		{
			name: "only up messages - no pairs",
			messages: []*Message{
				makeMessage("2", MessageTypeUp, "ServiceA", 10),
				makeMessage("1", MessageTypeUp, "ServiceA", 20),
			},
			expected: []MessagePair{},
		},
		{
			name: "up before down - no pair",
			messages: []*Message{
				makeMessage("2", MessageTypeDown, "ServiceA", 10),
				makeMessage("1", MessageTypeUp, "ServiceA", 20),
			},
			expected: []MessagePair{},
		},
		{
			name: "multiple services with multiple pairs",
			messages: []*Message{
				makeMessage("6", MessageTypeUp, "ServiceB", 5),
				makeMessage("5", MessageTypeUp, "ServiceA", 10),
				makeMessage("4", MessageTypeDown, "ServiceB", 15),
				makeMessage("3", MessageTypeDown, "ServiceA", 20),
				makeMessage("2", MessageTypeUp, "ServiceA", 25),
				makeMessage("1", MessageTypeDown, "ServiceA", 30),
			},
			expected: []MessagePair{
				{
					DownMessage: makeMessage("1", MessageTypeDown, "ServiceA", 30),
					UpMessage:   makeMessage("2", MessageTypeUp, "ServiceA", 25),
				},
				{
					DownMessage: makeMessage("3", MessageTypeDown, "ServiceA", 20),
					UpMessage:   makeMessage("5", MessageTypeUp, "ServiceA", 10),
				},
				{
					DownMessage: makeMessage("4", MessageTypeDown, "ServiceB", 15),
					UpMessage:   makeMessage("6", MessageTypeUp, "ServiceB", 5),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindPairsToCleanup(tt.messages)

			if len(got) != len(tt.expected) {
				t.Logf("Got pairs:")
				for i, p := range got {
					t.Logf("  [%d] Down:%s Up:%s Service:%s", i, p.DownMessage.ID, p.UpMessage.ID, p.DownMessage.ServiceName)
				}
				t.Fatalf("expected %d pairs, got %d", len(tt.expected), len(got))
			}

			for i := range got {
				if got[i].DownMessage.ID != tt.expected[i].DownMessage.ID {
					t.Errorf("pair %d: expected down message ID %s, got %s",
						i, tt.expected[i].DownMessage.ID, got[i].DownMessage.ID)
				}
				if got[i].UpMessage.ID != tt.expected[i].UpMessage.ID {
					t.Errorf("pair %d: expected up message ID %s, got %s",
						i, tt.expected[i].UpMessage.ID, got[i].UpMessage.ID)
				}
			}
		})
	}
}

func TestParseMessage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		subject     string
		expectError bool
		expectType  MessageType
		expectSvc   string
	}{
		{
			name:        "valid down message",
			subject:     "🔴 Down: My Service",
			expectError: false,
			expectType:  MessageTypeDown,
			expectSvc:   "My Service",
		},
		{
			name:        "valid up message",
			subject:     "✅ Up: My Service",
			expectError: false,
			expectType:  MessageTypeUp,
			expectSvc:   "My Service",
		},
		{
			name:        "down message with complex service name",
			subject:     "🔴 Down: api.example.com:8080",
			expectError: false,
			expectType:  MessageTypeDown,
			expectSvc:   "api.example.com:8080",
		},
		{
			name:        "invalid format - no prefix",
			subject:     "Some other email",
			expectError: true,
		},
		{
			name:        "invalid format - empty service name",
			subject:     "🔴 Down: ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := ParseMessage("test-id", tt.subject, now)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if msg.Type != tt.expectType {
				t.Errorf("expected type %v, got %v", tt.expectType, msg.Type)
			}

			if msg.ServiceName != tt.expectSvc {
				t.Errorf("expected service name %q, got %q", tt.expectSvc, msg.ServiceName)
			}

			if msg.ID != "test-id" {
				t.Errorf("expected ID %q, got %q", "test-id", msg.ID)
			}

			if !msg.Timestamp.Equal(now) {
				t.Errorf("expected timestamp %v, got %v", now, msg.Timestamp)
			}
		})
	}
}
