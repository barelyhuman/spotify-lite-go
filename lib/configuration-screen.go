package lib

import (
	"net/url"

	"fyne.io/fyne"
	"fyne.io/fyne/widget"
)

// OpenConfigurationScreen - Open Config screen for handling client details
func OpenConfigurationScreen(appInstance fyne.App) {
	windowInstance := appInstance.NewWindow("Configuration")
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")
	clientSecret := appInstance.Preferences().StringWithFallback("Client Secret", "")

	redirectURL := GetRedirectURL()

	var openPortLabel = widget.NewLabel(`
Spotify Lite, needs you to create your own spotify app and add the creds here.
1. Register an application at the Developer portal
2. Add the redirect URI mentioned below in the redirect URI 
You can then copy the ClientId and ClientSecret and paste them here
This only has to be done once.
`,
	)

	dashboardURL, _ := url.Parse("https://developer.spotify.com/my-applications/")
	rURL, _ := url.Parse(redirectURL)

	dashboardHelperLabel := widget.NewLabel("Developer Dashboard")
	developerDashboardLinkElem := widget.NewHyperlink("https://developer.spotify.com/my-applications/", dashboardURL)
	redirectionHelperLabel := widget.NewLabel("Redirect URI")
	redirectionLinkElem := widget.NewHyperlink(redirectURL, rURL)

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
		initiateOAuthFlow(clientIDEntry.Text, clientSecretEntry.Text)
	})

	windowInstance.SetContent(
		widget.NewVBox(
			openPortLabel,
			dashboardHelperLabel,
			developerDashboardLinkElem,
			redirectionHelperLabel,
			redirectionLinkElem,
			clientIDEntry,
			clientSecretEntry,
			connectButton,
		),
	)

	windowInstance.ShowAndRun()
}

func initiateOAuthFlow(clientID string, clientSecret string) {
	auth := GetAuthenticator()
	auth.SetAuthInfo(clientID, clientSecret)
	// TODO: Replace with cryptographic alpha numeric string
	state := GetState()
	url := auth.AuthURL(state)
	OpenBrowser(url)
}
