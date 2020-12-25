package lib

import (
	"log"
	"net/url"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

var (
	scopes      = []string{spotify.ScopeUserReadPrivate, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState}
	openPort    = GetOpenPort()
	auth        = spotify.NewAuthenticator(redirectURI, scopes...)
	redirectURI = "http://localhost:" + openPort + "/callback"
)

// OpenConfigurationScreen - Open Config screen for handling client details
func OpenConfigurationScreen(appInstance fyne.App, codeChallenge string) fyne.Window {
	windowInstance := appInstance.NewWindow("Configuration")
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")

	var openPortLabel = widget.NewLabel(`
Spotify Lite, needs you to create your own spotify app and add the creds here.
1. Register an application at the Developer portal
2. Add the redirect URI mentioned below in the redirect URI 
You can then copy the ClientId and paste it here
This only has to be done once (really depends on spotify's auth rules).
`,
	)

	dashboardURL, _ := url.Parse("https://developer.spotify.com/my-applications/")
	rURL, _ := url.Parse(redirectURI)

	dashboardHelperLabel := widget.NewLabel("Developer Dashboard")
	developerDashboardLinkElem := widget.NewHyperlink("https://developer.spotify.com/my-applications/", dashboardURL)
	redirectionHelperLabel := widget.NewLabel("Redirect URI")
	redirectionLinkElem := widget.NewHyperlink(redirectURI, rURL)

	clientIDEntry := widget.NewEntry()
	clientIDEntry.SetPlaceHolder("Client ID")
	clientIDEntry.SetText(clientID)
	clientIDEntry.OnChanged = func(value string) {

	}

	connectButton := widget.NewButton("Connect", func() {
		appInstance.Preferences().SetString("Client ID", clientIDEntry.Text)
		SaveScopes(appInstance, scopes...)
		initiateOAuthFlow(clientIDEntry.Text, codeChallenge)
	})

	windowInstance.SetContent(
		widget.NewVBox(
			openPortLabel,
			dashboardHelperLabel,
			developerDashboardLinkElem,
			redirectionHelperLabel,
			redirectionLinkElem,
			clientIDEntry,
			connectButton,
		),
	)

	windowInstance.Show()

	return windowInstance
}

func initiateOAuthFlow(clientID string, codeChallenge string) {
	state := "abc123"
	log.Println("clientID", clientID)
	url := auth.AuthURLWithOpts(state,
		oauth2.SetAuthURLParam("client_id", clientID),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
	)
	OpenBrowser(url)
}
