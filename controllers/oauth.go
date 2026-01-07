package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"golang.org/x/oauth2"
)

type OAuth struct {
	ProviderConfigs map[string]*oauth2.Config
}

// GET /oauth/{provider}/connect
func (oa OAuth) Connect(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	provider = strings.ToLower(provider)
	config, ok := oa.ProviderConfigs[provider]
	if !ok {
		http.Error(w, "Invalid OAuth2 Service", http.StatusBadRequest)
		return
	}

	state := csrf.Token(r)
	setCookie(w, "oauth_state", state) // Using our trusty helper
	// fmt.Println("set state in cookie", state) // debug
	url := config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("redirect_uri", redirectURI(r, provider)),
	)
	http.Redirect(w, r, url, http.StatusFound) // Send user to provider
}

// GET /oauth/{provider}/callback
func (oa OAuth) Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	provider = strings.ToLower(provider)
	config, ok := oa.ProviderConfigs[provider]
	if !ok {
		http.Error(w, "Invalid OAuth2 Service", http.StatusBadRequest)
		return
	}

	state := r.FormValue("state")
	// fmt.Println("state from form", state) // debug
	cookieState, err := readCookie(r, "oauth_state") // Trusty helper
	// fmt.Println("state from cookie", cookieState) // debug
	if err != nil || cookieState != state {
		fmt.Println("read state", err)
		http.Error(w, "Invalid Request", http.StatusBadRequest)
		return
	}
	deleteCookie(w, "oauth_state") // Another trusty helper

	code := r.FormValue("code")
	token, err := config.Exchange(
		r.Context(),
		code,
		oauth2.SetAuthURLParam("redirect_uri", redirectURI(r, provider)),
	)
	if err != nil {
		fmt.Println("exchanging", err)
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return
	}
	// Persist the oauth token to DB, to be used in further requests.
	// No need for the above, since we'll be using the dropbox chooser.

	// Redirect the user to the page where oauth started (for now show token).
	// w.Header().Set("Content-Type", "application/json")
	// enc := json.NewEncoder(w)
	// enc.SetIndent("", "  ")
	// enc.Encode(token)

	// Let's test the Dropbox api
	client := config.Client(r.Context(), token)
	dropboxApiUrl := "https://api.dropboxapi.com/2/files/list_folder"
	resp, err := client.Post(dropboxApiUrl, "application/json", strings.NewReader(`{
		"path": ""
	}`))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	err = writePrettyJSONResponse(w, resp.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to read dropbox response", http.StatusInternalServerError)
		return
	}
}

// Just a helper to pretty-print a JSON response
func writePrettyJSONResponse(w http.ResponseWriter, r io.Reader) error {
	raw, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, raw, "", "  "); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(pretty.Bytes())
	return err
}

func redirectURI(r *http.Request, provider string) string {
	// Return development URL
	if r.Host == "localhost:3000" {
		return fmt.Sprintf("http://localhost:3000/oauth/%s/callback", provider)
	}
	// Return deployed URL (I haven't deployed, but have pre-deploy version)
	return fmt.Sprintf("http://localhost/oauth/%s/callback", provider)
}
