package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

func main() {
	appInstance := app.New()
	windowInstance := appInstance.NewWindow("Spotify Lite")

	var playButtonLabel = "play"
	var playPauseButton *widget.Button
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
			playPauseButton,
		),
	)

	windowInstance.ShowAndRun()
}
