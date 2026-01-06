package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	ctx := context.Background()

	conf := &oauth2.Config{
		ClientID:     os.Getenv("DROPBOX_APP_KEY"),
		ClientSecret: os.Getenv("DROPBOX_APP_SECRET"),
		Scopes: []string{
			"files.metadata.read",
			"files.content.read",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.dropbox.com/oauth2/authorize",
			TokenURL: "https://api.dropboxapi.com/oauth2/token",
		},
	}

	// Redirect user to the consent page in the provider.
	// url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	url := conf.AuthCodeURL("state", oauth2.SetAuthURLParam("token_access_type", "offline"))
	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)
	fmt.Printf("Once you have the code, paste it and press enter:\n")

	// Use the authorization code that is pushed to the redirect
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}

	// Use the custom HTTP client when requesting a token.
	httpClient := &http.Client{Timeout: 2 * time.Second}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(tok)
	// fmt.Printf("Token: %v\n", tok)

	client := conf.Client(ctx, tok)
	apiUrl := "https://api.dropboxapi.com/2/files/list_folder"
	jsonPayload := strings.NewReader(`{
		"path": "",
		"recursive": false
	}`)
	resp, err := client.Post(apiUrl, "application/json", jsonPayload)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	io.Copy(os.Stdout, resp.Body)
}
