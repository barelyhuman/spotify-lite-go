package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"github.com/barelyhuman/spotify-lite-go/lib"
	"github.com/zmb3/spotify"
)

var (
	state         string
	client        spotify.Client
	appInstance   fyne.App
	trackNameChan chan bool
)

func main() {

	/*
	* TODO:
	* Save and Fetch Config - To File for now, UserData/spotify-lite/config.json/yml/whatever
	* OAuth Flow - Spotify
	* Save ClientID,ClientSecret,Port to config
	* Save the access token once the calleback is called to the config as well
	* Render Simple UI if the config already as all the above
	 */

	var srv *http.Server

	wg := new(sync.WaitGroup)

	wg.Add(3)

	lib.NewState()

	state = lib.GetState()

	openPort := lib.CheckOpenPort()

	go func() {
		srv = setupServer(openPort)
		wg.Done()
	}()

	appInstance = app.NewWithID("com.reaper.spotifylite")

	// lib.OpenConfigurationScreen(appInstance, openPort)
	trackNameChan := lib.OpenPlayerView(appInstance, &client)

	if err := srv.Shutdown(context.TODO()); err != nil {
		panic(err)
	}

	trackNameChan <- true

	wg.Wait()
}

func setupServer(connectionPort string) *http.Server {

	srv := &http.Server{Addr: ":" + connectionPort}

	http.HandleFunc("/callback", oAuthRedirectHandler)

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "server up\n")
	})

	fmt.Println("Started Server on port:" + connectionPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	return srv
}

func oAuthRedirectHandler(w http.ResponseWriter, r *http.Request) {
	auth := lib.GetAuthenticator()
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")
	clientSecret := appInstance.Preferences().StringWithFallback("Client Secret", "")
	auth.SetAuthInfo(clientID, clientSecret)

	token, err := auth.Token(state, r)

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}

	st := r.FormValue("state")
	fmt.Println(state, st)

	if st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	fmt.Println("Got token, creating client")

	appInstance.Preferences().SetString("Access Token", token.AccessToken)
	appInstance.Preferences().SetString("Refresh Token", token.RefreshToken)

	client = auth.NewClient(token)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!")

	trackNameChan = lib.OpenPlayerView(appInstance, &client)

	if err != nil {
		log.Print(err)
	}
}
