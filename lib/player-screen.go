package lib

import (
	"log"
	"time"

	"fyne.io/fyne/widget"
	"github.com/zmb3/spotify"
)

// GetPlayerView - get player view as canvas
func GetPlayerView(client *spotify.Client) (*widget.Box, chan bool) {
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

	stop := Schedule(func() {
		updateTrackNameLabel(currentPlayingLabel, client, currentArtistLabel)
	}, 2*time.Second)

	return widget.NewVBox(
		currentPlayingLabel,
		currentArtistLabel,
		widget.NewHBox(
			playButton,
			pauseButton,
			nextButton,
			backButton,
		)), stop
}

func updateTrackNameLabel(label *widget.Label, client *spotify.Client, artistLabel *widget.Label) {
	playing, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Print("Label Update Fail:", err)
	}
	if !playing.Playing {
		label.SetText("Not Playing anything...")
	} else {
		label.SetText(playing.Item.Name)
		artistLabel.SetText(playing.Item.Artists[0].Name)
	}
	return
}
