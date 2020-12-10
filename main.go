package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/barelyhuman/spotify-lite-go/lib"
)

func main() {

	/*
	* TODO:
	* Save and Fetch Config - To File for now, UserData/spotify-lite/config.json/yml/whatever
	* Save ClientID,ClientSecret,Port to config
	* Save the access token once the calleback is called to the config as well
	* Render Simple UI if the config already as all the above
	 */

	wg := new(sync.WaitGroup)

	wg.Add(1)

	openPort := lib.CheckOpenPort("localhost", []string{"1821", "1293"})

	go func() {
		setupServer(openPort)
		wg.Done()
	}()

	windowInstance := setupInitialAppView(openPort)
	windowInstance.ShowAndRun()

	wg.Wait()
}

func setupServer(connectionPort string) {

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "Hello from callback #1!\n")
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "server up\n")
	})

	if err := http.ListenAndServe(":"+connectionPort, nil); err != nil {
		log.Fatal(err)
	}

}

func setupInitialAppView(openPort string) fyne.Window {
	appInstance := app.New()
	windowInstance := appInstance.NewWindow("Spotify Lite")

	var openPortLabel = widget.NewLabel(`
Spotify Lite, needs you to create your own spotify app and add the creds here.
Redirect url can be set as localhost:` + openPort)
	var playButtonLabel = "play"
	var playPauseButton *widget.Button
	clientIDEntry := widget.NewEntry()
	clientIDEntry.SetPlaceHolder("Client ID")
	clientIDEntry.OnChanged = func(value string) {
		fmt.Println(value)
	}

	clientSecretEntry := widget.NewEntry()
	clientSecretEntry.SetPlaceHolder("Client Secret")
	clientSecretEntry.OnChanged = func(value string) {
		fmt.Println(value)
	}

	playPauseButton = widget.NewButton("Play", func() {
		if playPauseButton != nil {
			switch playButtonLabel {
			case "play":
				playButtonLabel = "pause"
				playPauseButton.SetText("Pause")
			case "pause":
				playButtonLabel = "play"
				playPauseButton.SetText("Play")
			}
		}
	})

	windowInstance.SetContent(
		widget.NewVBox(
			openPortLabel,
			clientIDEntry,
			clientSecretEntry,
			playPauseButton,
		),
	)

	return windowInstance
}
