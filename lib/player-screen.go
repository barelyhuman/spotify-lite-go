package lib

import (
	"log"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/zmb3/spotify"
)

// GetPlayerView - get player view as canvas
func GetPlayerView(client *spotify.Client, premium bool, recheck func()) (*widget.Box, chan bool) {
	currentPlayingLabel := widget.NewLabel("Loading...")
	currentArtistLabel := widget.NewLabel("Loading...")

	recheckButton := widget.NewButton("Recheck", func() {
		recheck()
	})

	playButton := widget.NewButton("Play", func() {
		client.Play()
		err := client.Play().Error()
		if err != "" {
			log.Println("Error Playing", err)
		}
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

	needPremiumLabel := widget.NewLabelWithStyle("I'm sorry but you can't change playback \n state without spotify premium", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	needPremiumSubLabel := widget.NewLabelWithStyle("Close and reopen the app if you upgraded to spotify premium", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	stop := Schedule(func() {
		updateTrackNameLabel(currentPlayingLabel, client, currentArtistLabel)
	}, 2*time.Second)

	playerControls := widget.NewHBox(
		playButton,
		pauseButton,
		nextButton,
		backButton,
	)

	if !premium {
		playerControls = widget.NewVBox(
			needPremiumLabel,
			needPremiumSubLabel,
			recheckButton,
		)
	}

	return widget.NewVBox(
		currentPlayingLabel,
		currentArtistLabel,
		playerControls,
	), stop
}

func updateTrackNameLabel(label *widget.Label, client *spotify.Client, artistLabel *widget.Label) {
	playing, err := client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Print("Label Update Fail:", err)
	}
	if playing == nil {
		return
	}
	if !playing.Playing {
		label.SetText("Not Playing anything...")
		artistLabel.SetText("-")
	} else {
		label.SetText(playing.Item.Name)
		artistLabel.SetText(playing.Item.Artists[0].Name)
	}
	return
}
