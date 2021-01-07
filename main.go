package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"fyne.io/fyne"

	cv "github.com/nirasan/go-oauth-pkce-code-verifier"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var openPort = GetOpenPort()

var redirectURI = "http://localhost:" + openPort + "/callback"

var (
	ch              = make(chan *spotify.Client)
	token           = make(chan *oauth2.Token)
	state           = "abc123"
	codeVerifier    string
	codeChallenge   string
	stopLabelUpdate chan bool
	scopes          = []string{spotify.ScopeUserReadPrivate, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState, spotify.ScopeUserLibraryModify, spotify.ScopeUserLibraryRead}
	auth            = spotify.NewAuthenticator(redirectURI, scopes...)
)

func main() {
	cvInstance, _ := cv.CreateCodeVerifier()
	codeVerifier = cvInstance.String()
	codeChallenge = cvInstance.CodeChallengeS256()

	app := &App{
		authenticated: false,
	}

	app.Install()

	go setupServer(app.appInstance)

	go func() {
		log.Println("Initial Trigger for config Handler")
	}()

	go func() {
		select {
		case tokenValue := <-token:
			log.Println("Saving New Token")
			app.SaveToken(tokenValue)
		}
	}()

	go func() {
		select {
		case client := <-ch:
			app.SetClient(client)
			stopLabelUpdate = app.DrawPlayerView()
			app.ShowPlayerView()
		}
	}()

	app.appInstance.Run()

	stopLabelUpdate <- true
}

func setupServer(appInstance fyne.App) {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		completeAuth(w, r, appInstance)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	log.Println("Starting server on port " + GetOpenPort())
	http.ListenAndServe(":"+GetOpenPort(), nil)
}

func completeAuth(w http.ResponseWriter, r *http.Request, appInstance fyne.App) {
	log.Println("Creating Token from query")
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")

	if r.URL.Query().Get("code") == "" {
		fmt.Fprintf(w, "Missing Parameters!")
		return
	}

	tok, err := auth.TokenWithOpts(state, r,
		oauth2.SetAuthURLParam("client_id", clientID),
		oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		log.Println("Failed while creating token")
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		log.Println("Failed while checking state")
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	log.Println("Success, creating client...")
	client := auth.NewClient(tok)
	fmt.Fprintf(w, "Login Completed!")
	ch <- &client
	token <- tok
}

func saveToken(appInstance fyne.App, token *oauth2.Token) {
	log.Println("Saving Token Details")
	appInstance.Preferences().SetString("Access Token", token.AccessToken)
	appInstance.Preferences().SetString("Refresh Token", token.RefreshToken)
	appInstance.Preferences().SetString("Token Type", token.TokenType)
	appInstance.Preferences().SetString("Token Expiry", token.Expiry.Local().Format(timeLayout))
}

func loadToken(appInstance fyne.App) *oauth2.Token {
	parsedTime, _ := time.Parse(timeLayout, appInstance.Preferences().String("Token Expiry"))
	return &oauth2.Token{
		AccessToken:  appInstance.Preferences().String("Access Token"),
		RefreshToken: appInstance.Preferences().String("Refresh Token"),
		TokenType:    appInstance.Preferences().String("Token Type"),
		Expiry:       parsedTime,
	}
}
