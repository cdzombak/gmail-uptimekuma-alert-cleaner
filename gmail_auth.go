package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// mustGetenv returns the value of the environment variable with the given name, or exits
// with an error if the variable is empty.
func mustGetenv(key string) string {
	retv := os.Getenv(key)
	if retv == "" {
		slog.Error("environment variable is missing", "variable", key)
		os.Exit(1)
	}
	return retv
}

func buildGmailService() (*gmail.Service, error) {
	configDir := mustGetenv("GMAIL_UPTIMEKUMA_ALERT_CLEANER_CONFIG_DIR")
	credentialsFilename := path.Join(configDir, "credentials.json")
	b, err := os.ReadFile(credentialsFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to read client credentials file (credentials.json): %w", err)
	}
	scope := gmail.GmailModifyScope
	config, err := google.ConfigFromJSON(b, scope) // If modifying these scopes, delete any previously saved token.json.
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}
	client := getClient(config)
	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to build Gmail client: %w", err)
	}
	return srv, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first time.
	configDir := mustGetenv("GMAIL_UPTIMEKUMA_ALERT_CLEANER_CONFIG_DIR")
	tokFile := path.Join(configDir, "token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		slog.Error("unable to read authorization code", "error", err)
		os.Exit(1)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		slog.Error("unable to retrieve token from web", "error", err)
		os.Exit(1)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		slog.Error("unable to cache oauth token", "error", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()
	if err = json.NewEncoder(f).Encode(token); err != nil {
		slog.Error("unable to encode oauth token", "error", err)
		os.Exit(1)
	}
}
