package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/barelyhuman/spotify-lite-go/lib"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var (
	state          string
	auth           spotify.Authenticator
	client         spotify.Client
	windowInstance fyne.Window
	appInstance    fyne.App
	trackNameChan  chan bool
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

	wg.Add(1)

	openPort := lib.CheckOpenPort("localhost", []string{"1821", "1293"})

	go func() {
		srv = setupServer(openPort)
		wg.Done()
	}()

	appInstance = app.NewWithID("com.reaper.spotifylite")
	windowInstance = appInstance.NewWindow("Spotify Lite")

	setupInitialAppView(openPort)
	windowInstance.ShowAndRun()

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

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	return srv

}

func setupInitialAppView(openPort string) {
	accessToken := appInstance.Preferences().StringWithFallback("Access Token", "")
	refreshToken := appInstance.Preferences().StringWithFallback("Refresh Token", "")
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")
	clientSecret := appInstance.Preferences().StringWithFallback("Client Secret", "")

	redirectURL := "http://localhost:" + openPort + "/callback"

	fmt.Println(accessToken + "Token")

	auth = spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)

	if accessToken == "" || clientID == "" || clientSecret == "" {
		var openPortLabel = widget.NewLabel(`
Spotify Lite, needs you to create your own spotify app and add the creds here.
1. Register an application at: https://developer.spotify.com/my-applications/
2. Use "http://localhost:` + openPort + `/callback as the redirect URI

You can then copy the ClientId and ClientSecret and paste them here

This only has to be done once.
`,
		)
		clientIDEntry := widget.NewEntry()
		clientIDEntry.SetPlaceHolder("Client ID")
		clientIDEntry.SetText(clientID)
		clientIDEntry.OnChanged = func(value string) {
			// fmt.Println(value)
		}

		clientSecretEntry := widget.NewEntry()
		clientSecretEntry.SetPlaceHolder("Client Secret")
		clientSecretEntry.SetText(clientSecret)
		clientSecretEntry.OnChanged = func(value string) {
			// fmt.Println(value)
		}

		connectButton := widget.NewButton("Connect", func() {
			appInstance.Preferences().SetString("Client ID", clientIDEntry.Text)
			appInstance.Preferences().SetString("Client Secret", clientSecretEntry.Text)
			initiateOAuthFlow(clientIDEntry.Text, clientSecretEntry.Text, openPort)
		})

		windowInstance.SetContent(
			widget.NewVBox(
				openPortLabel,
				clientIDEntry,
				clientSecretEntry,
				connectButton,
			),
		)
	} else {
		auth.SetAuthInfo(clientID, clientSecret)
		fmt.Println("Failed before auth")
		client = auth.NewClient(&oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken})
		fmt.Println("Failed after auth")
		user, err := client.CurrentUser()
		if err != nil {
			fmt.Println("I got in")
			appInstance.Preferences().SetString("Access Token", "")
			log.Fatal("Failed :", err)
		}
		fmt.Println("Hi: ", user.ID)
		trackNameChan = showPlayerView()
	}

}

func initiateOAuthFlow(clientID string, clientSecret string, openPort string) {
	auth.SetAuthInfo(clientID, clientSecret)
	// TODO: Replace with cryptographic alpha numeric string
	state = "1234"
	url := auth.AuthURL(state)
	openbrowser(url)
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func oAuthRedirectHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(state, r)

	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	fmt.Println("Got token, creating client")

	appInstance.Preferences().SetString("Access Token", token.AccessToken)
	appInstance.Preferences().SetString("Refresh Token", token.RefreshToken)

	client = auth.NewClient(token)
	// playerState, err := client.PlayerState()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!")

	trackNameChan = showPlayerView()

	if err != nil {
		log.Print(err)
	}

}

func updateTrackNameLabel(label *widget.Label) {
	playing, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Print(err)
	}
	label.SetText(playing.Item.Name)
	return
}

func showPlayerView() chan bool {

	currentPlayingLabel := widget.NewLabel("")
	updateTrackNameLabel(currentPlayingLabel)

	stop := schedule(func() {
		updateTrackNameLabel(currentPlayingLabel)
	}, 2*time.Second)

	playButton := widget.NewButton("Play", func() {
		client.Play()
		currentPlayingLabel.SetText("Loading...")
	})
	pauseButton := widget.NewButton("Pause", func() {
		client.Pause()
		currentPlayingLabel.SetText("Loading...")
	})
	nextButton := widget.NewButton("Next", func() {
		client.Next()
		currentPlayingLabel.SetText("Loading...")
	})

	backButton := widget.NewButton("Prev", func() {
		client.Previous()
		currentPlayingLabel.SetText("Loading...")
	})

	windowInstance.SetContent(
		widget.NewVBox(
			currentPlayingLabel,
			widget.NewHBox(
				playButton,
				pauseButton,
				nextButton,
				backButton,
			),
		),
	)

	return stop
}

func schedule(toExec func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			toExec()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}
