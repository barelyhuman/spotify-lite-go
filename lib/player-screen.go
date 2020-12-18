package lib

import (
	"log"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/zmb3/spotify"
)

// OpenPlayerView - Open the player view
func OpenPlayerView(appInstance fyne.App, client *spotify.Client) chan bool {
	var stop chan bool

	if GetToken(client) {
		windowInstance := appInstance.NewWindow("Spotify Lite")

		stop = showPlayerView(windowInstance, client, appInstance)

		windowInstance.ShowAndRun()
	} else {
		OpenConfigurationScreen(appInstance)
	}

	return stop
}

func showPlayerView(windowInstance fyne.Window, client *spotify.Client, appInstance fyne.App) chan bool {

	currentPlayingLabel := widget.NewLabel("Loading...")
	currentArtistLabel := widget.NewLabel("Loading...")

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
			currentArtistLabel,
			widget.NewHBox(
				playButton,
				pauseButton,
				nextButton,
				backButton,
			),
		),
	)

	stop := Schedule(func() {
		updateTrackNameLabel(currentPlayingLabel, client, appInstance, currentArtistLabel)
	}, 2*time.Second)

	return stop
}

func updateTrackNameLabel(label *widget.Label, client *spotify.Client, appInstance fyne.App, artistLabel *widget.Label) {
	_ = GetToken(client)
	playing, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Print("Label Update Fail:", err)
		OpenConfigurationScreen(appInstance)
	}
	if !playing.Playing {
		label.SetText("Not Playing anything...")
	} else {
		label.SetText(playing.Item.Name)
		artistLabel.SetText(playing.Item.Artists[0].Name)
	}
	return
}
