package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/barelyhuman/spotify-lite-go/lib"

	cv "github.com/nirasan/go-oauth-pkce-code-verifier"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

const timeLayout = "02 Jan 06 15:04 MST"

var openPort = lib.GetOpenPort()

var redirectURI = "http://localhost:" + openPort + "/callback"

var (
	auth            = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadPrivate, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	ch              = make(chan *spotify.Client)
	token           = make(chan *oauth2.Token)
	state           = "abc123"
	codeVerifier    string
	codeChallenge   string
	stopLabelUpdate chan bool
)

func main() {
	appInstance := app.NewWithID("im.reaper.spotify-lite-go")
	initialWindow := showInitialWindow(appInstance)
	var configWindow fyne.Window

	cvInstance, _ := cv.CreateCodeVerifier()
	codeVerifier = cvInstance.String()
	codeChallenge = cvInstance.CodeChallengeS256()

	lib.SyncEnvVariables(appInstance)
	lib.SyncScopes(appInstance, spotify.ScopeUserReadPrivate, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)

	go setupServer(appInstance)

	go func() {
		log.Println("Initial Trigger for config Handler")
		configHandler(appInstance, &configWindow)
	}()

	go func() {
		select {
		case tokenValue := <-token:
			saveToken(appInstance, tokenValue)
		}
	}()

	go func() {
		handlePlayerView(appInstance, &configWindow, initialWindow)
	}()

	appInstance.Run()

	stopLabelUpdate <- true
}

func handlePlayerView(appInstance fyne.App, configWindow *fyne.Window, initialWindow fyne.Window) {
	select {
	case client := <-ch:
		log.Println("Oauth Connected")
		user, err := client.CurrentUser()
		if err != nil {
			log.Println("Client User Fetch error: ", err.Error())
			client.Token()
			if strings.Contains(err.Error(), "oauth2: cannot fetch token") {
				log.Println("Opening Configuration Screen on Token Fetch")
				appInstance.Preferences().RemoveValue("Access Token")
				appInstance.Preferences().RemoveValue("Refresh Token")
				configHandler(appInstance, configWindow)
			} else {
				log.Fatal(err)
			}
			log.Println("Waiting for new client...")
			client = <-ch
		}

		if user != nil {
			log.Println("You are logged in as:", user.ID)

			st := playerWindowManager(client, initialWindow)
			stopLabelUpdate = st
		}

	}
}

func playerWindowManager(client *spotify.Client, initialWindow fyne.Window) chan bool {
	user, _ := client.CurrentUser()
	windowContentsRecursive, stop := lib.GetPlayerView(client, user.Product == "premium", func() {
		playerWindowManager(client, initialWindow)
	})
	initialWindow.SetContent(windowContentsRecursive)
	return stop
}

func configHandler(appInstance fyne.App, configWindow *fyne.Window) {
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")
	isClientIDExists := clientID != ""
	refreshToken := appInstance.Preferences().StringWithFallback("Refresh Token", "")
	accessToken := appInstance.Preferences().StringWithFallback("Access Token", "")

	log.Println("Checking Configuration Deps")
	log.Println("ClientID Exists", isClientIDExists)
	log.Println("Refresh Token Exists", refreshToken == "")
	log.Println("Access Token Exists", accessToken == "")

	if !isClientIDExists || refreshToken == "" || accessToken == "" {
		*configWindow = lib.OpenConfigurationScreen(appInstance, codeChallenge)
	} else {
		log.Println("Using Tokens")
		token := loadToken(appInstance)
		client := auth.NewClient(token)
		ch <- &client
	}
}

func showInitialWindow(appInstance fyne.App) fyne.Window {
	window := appInstance.NewWindow("Spotify Lite")
	window.Resize(fyne.NewSize(300, 40))
	progressBar := widget.NewProgressBarInfinite()
	label := widget.NewLabelWithStyle("Starting Engines...", fyne.TextAlignCenter, fyne.TextStyle{})
	window.SetContent(
		widget.NewVBox(
			label,
			progressBar,
		),
	)
	window.Show()
	return window
}

func setupServer(appInstance fyne.App) {
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		completeAuth(w, r, appInstance)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	log.Println("Starting server on port " + lib.GetOpenPort())
	http.ListenAndServe(":"+lib.GetOpenPort(), nil)
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
