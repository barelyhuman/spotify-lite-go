package lib

import (
	"log"
	"os"

	"fyne.io/fyne"
)

const envVariableSpotifyID = "SPOTIFY_ID"

// SyncEnvVariables - Synchronize data with env variables needed for OAuth
func SyncEnvVariables(appInstance fyne.App) {
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")
	envClientID := os.Getenv(envVariableSpotifyID)
	log.Println("Client ID from ENV Variables: ", envClientID)
	log.Println("Current Client ID", clientID)
	if envClientID != clientID {
		os.Setenv(envVariableSpotifyID, clientID)
	}
	envClientID = os.Getenv(envVariableSpotifyID)

	log.Println("After ENV Sync")
	log.Println("SPOTIFY_ID: ", envClientID)
}
