package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/barelyhuman/spotify-lite-go/lib"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var redirectURI = "http://localhost:8080/callback"

var (
	auth          = spotify.NewAuthenticator(redirectURI, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
	ch            = make(chan *spotify.Client)
	token         = make(chan *oauth2.Token)
	state         = "abc123"
	codeVerifier  string
	codeChallenge string
)

func main() {
	appInstance := app.NewWithID("im.reaper.spotify-lite-go")
	initialWindow := showInitialWindow(appInstance)
	var configWindow fyne.Window
	go setupServer()

	log.Println(codeVerifier, codeChallenge)

	go func() {
		configHandler(appInstance, &configWindow)
	}()

	go func() {
		select {
		case tokenValue := <-token:
			saveToken(appInstance, tokenValue)
		}
	}()

	go func() {
		select {
		case client := <-ch:
			log.Println("Oauth Connected")
			user, err := client.CurrentUser()
			if err != nil {
				if strings.Contains(err.Error(), "token expired") {
					appInstance.Preferences().RemoveValue("Access Token")
					appInstance.Preferences().RemoveValue("Refresh Token")
					configHandler(appInstance, &configWindow)
					client = <-ch
				} else {
					log.Fatal(err)
				}

			}
			log.Println("You are logged in as:", user.ID)
			windowContents, stopLabelUpdate := lib.GetPlayerView(client)
			if configWindow != nil {
				configWindow.Close()
			}
			initialWindow.SetContent(
				windowContents,
			)
			stopLabelUpdate <- true
		}
	}()

	appInstance.Run()
}

func configHandler(appInstance fyne.App, configWindow *fyne.Window) {
	isClientIDExists := appInstance.Preferences().StringWithFallback("Client ID", "") != ""
	refreshToken := appInstance.Preferences().StringWithFallback("Refresh Token", "")
	accessToken := appInstance.Preferences().StringWithFallback("Access Token", "")

	if !isClientIDExists || refreshToken == "" || accessToken == "" {
		log.Println("Opening Configuration Screen Again")
		*configWindow = lib.OpenConfigurationScreen(appInstance, codeChallenge)
	} else {
		log.Println("Using Tokens")
		token := oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		client := auth.NewClient(&token)
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

func setupServer() {
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})
	http.ListenAndServe(":8080", nil)
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating Token from query")
	tok, err := auth.TokenWithOpts(state, r,
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
}
