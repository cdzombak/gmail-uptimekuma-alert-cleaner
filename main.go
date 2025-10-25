package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"time"

	"github.com/cdzombak/exitcode_go"
	"google.golang.org/api/gmail/v1"
)

var version = "<dev>"

var (
	dryRun           = flag.Bool("dry-run", true, "Dry-run mode (default true). When true, prints actions but doesn't modify emails.")
	verbose          = flag.Bool("verbose", false, "Enable verbose logging for debugging.")
	configDir        = flag.String("configDir", "", "Path to config directory (credentials, token, config file). Overrides GMAIL_UPTIMEKUMA_ALERT_CLEANER_CONFIG_DIR.")
	printVersionFlag = flag.Bool("version", false, "Print version and exit.")
)

const (
	labelName       = "DzOps/Alerts"
	senderName      = "Uptime Kuma"
	inboxQuery      = "in:inbox"
	labelQuery      = "label:\"" + labelName + "\""
	senderQuery     = "from:\"" + senderName + "\""
	envVarConfigDir = "GMAIL_UPTIMEKUMA_ALERT_CLEANER_CONFIG_DIR"
)

func Main() int {
	flag.Parse()

	if *printVersionFlag {
		fmt.Println(version)
		return exitcode_go.Success
	}

	// Set up logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Handle config dir
	if *configDir != "" {
		_ = os.Setenv(envVarConfigDir, *configDir)
	} else if os.Getenv(envVarConfigDir) == "" {
		slog.Error("config directory is required", "flag", "-configDir", "env", envVarConfigDir)
		return exitcode_go.InvalidArgument
	}

	configDirPath := os.Getenv(envVarConfigDir)
	slog.Debug("using config directory", "path", configDirPath)

	// Load configuration
	config, err := LoadConfig(configDirPath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return exitcode_go.ConfigBSD
	}
	slog.Debug("loaded configuration", "timeWindow", config.TimeWindow)

	// Build Gmail service
	slog.Debug("building Gmail service")
	srv, err := buildGmailService()
	if err != nil {
		slog.Error("failed to build Gmail service", "error", err)
		return exitcode_go.Unavailable
	}

	// Calculate time threshold
	timeThreshold := time.Now().Add(-config.TimeWindow)
	afterQuery := fmt.Sprintf("after:%d", timeThreshold.Unix())

	// Build search query
	searchQuery := fmt.Sprintf("%s %s %s %s", inboxQuery, labelQuery, senderQuery, afterQuery)
	slog.Info("searching for messages", "query", searchQuery)

	// Fetch messages
	ctx := context.Background()
	var gmailMessages []*gmail.Message

	req := srv.Users.Messages.List("me").Q(searchQuery).Context(ctx)
	if err := req.Pages(ctx, func(response *gmail.ListMessagesResponse) error {
		for _, msg := range response.Messages {
			// Fetch full message details
			fullMsg, err := srv.Users.Messages.Get("me", msg.Id).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("failed to get message %s: %w", msg.Id, err)
			}
			gmailMessages = append(gmailMessages, fullMsg)
		}
		return nil
	}); err != nil {
		slog.Error("failed to fetch messages", "error", err)
		return exitcode_go.IOErr
	}

	slog.Info("fetched messages", "count", len(gmailMessages))

	// Convert to internal message format
	var messages []*Message
	for _, gmailMsg := range gmailMessages {
		msg, err := messageToInternal(gmailMsg)
		if err != nil {
			slog.Warn("skipping message with invalid format", "id", gmailMsg.Id, "error", err)
			continue
		}
		messages = append(messages, msg)
	}

	slog.Info("parsed valid messages", "count", len(messages))

	// Sort messages by timestamp (newest first)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.After(messages[j].Timestamp)
	})

	// Find pairs to cleanup
	pairs := FindPairsToCleanup(messages)
	slog.Info("found pairs to clean up", "count", len(pairs))

	if len(pairs) == 0 {
		slog.Info("no pairs to clean up")
		return exitcode_go.Success
	}

	// Process pairs
	if *dryRun {
		fmt.Println("DRY RUN MODE - would clean up the following message pairs:")
		for i, pair := range pairs {
			fmt.Printf("\nPair %d - %s:\n", i+1, pair.DownMessage.ServiceName)
			fmt.Printf("  Down: %s (ID: %s)\n", pair.DownMessage.Timestamp.Format(time.RFC3339), pair.DownMessage.ID)
			fmt.Printf("  Up:   %s (ID: %s)\n", pair.UpMessage.Timestamp.Format(time.RFC3339), pair.UpMessage.ID)
			fmt.Printf("  Actions: unstar (if starred), mark read, archive\n")
		}
		fmt.Printf("\nTotal: %d pairs (%d messages)\n", len(pairs), len(pairs)*2)
	} else {
		slog.Info("cleaning up message pairs", "count", len(pairs))
		for _, pair := range pairs {
			// Process each message in the pair
			for _, msg := range []*Message{pair.DownMessage, pair.UpMessage} {
				// Remove star, mark read, archive
				modReq := &gmail.ModifyMessageRequest{
					RemoveLabelIds: []string{"STARRED", "INBOX", "UNREAD"},
				}

				_, err := srv.Users.Messages.Modify("me", msg.ID, modReq).Context(ctx).Do()
				if err != nil {
					slog.Error("failed to modify message", "id", msg.ID, "service", msg.ServiceName, "error", err)
					return exitcode_go.IOErr
				}

				slog.Debug("cleaned up message", "id", msg.ID, "service", msg.ServiceName, "type", msg.Type)
			}
		}
		slog.Info("successfully cleaned up message pairs", "count", len(pairs), "messages", len(pairs)*2)
	}

	return exitcode_go.Success
}

func main() {
	exitCode := Main()
	os.Exit(exitCode)
}
