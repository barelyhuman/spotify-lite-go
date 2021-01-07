package main

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"fyne.io/fyne"
	fyneApp "fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// AppButtons - button that the player or app uses that are dynamic and need changing
type AppButtons struct {
	likeButton *widget.Button
}

// AppLabels - labels that the player or app uses that are dynamic and need changing
type AppLabels struct {
	trackLabel  *widget.Label
	artistLabel *widget.Label
}

// AppWindows - all windows that the app might create or need
type AppWindows struct {
	configWindow fyne.Window
	mainWindow   fyne.Window
}

// App - Base app wrapper with needed pointers and refs to handle app state
type App struct {
	appInstance   fyne.App
	windows       AppWindows
	token         *oauth2.Token
	authenticated bool
	client        *spotify.Client
	labels        AppLabels
	user          spotify.PrivateUser
	buttons       AppButtons
}

const (
	timeLayout           = "02 Jan 06 15:04 MST"
	envVariableSpotifyID = "SPOTIFY_ID"
)

// Install - initial setup for the App
func (app *App) Install() {
	app.appInstance = fyneApp.NewWithID("im.reaper.spotify-lite-go")

	app.DrawMainWindow()
	app.ShowMainWindow()
	app.SyncEnvVariables()
	app.RefreshToken()
	app.SyncScopes(scopes...)

	if app.authenticated {
		stopLabelUpdate = app.DrawPlayerView()
		app.ShowPlayerView()
	} else {
		app.DrawConfigScreen()
		app.ShowConfigScreen()
	}
}

// SyncEnvVariables - sync environment variables for client id verfication when needed
func (app *App) SyncEnvVariables() {
	clientID := app.appInstance.Preferences().StringWithFallback("Client ID", "")
	envClientID := os.Getenv(envVariableSpotifyID)
	log.Println("Client ID from ENV Variables: ", envClientID)
	log.Println("Current Client ID", clientID)
	if envClientID != clientID {
		os.Setenv(envVariableSpotifyID, clientID)
	}

	log.Println("After ENV Sync")
	log.Println("SPOTIFY_ID: ", os.Getenv(envVariableSpotifyID))
}

// DrawMainWindow - draw the main window
func (app *App) DrawMainWindow() {
	app.windows.mainWindow = app.appInstance.NewWindow("Spotify Lite")
	app.windows.mainWindow.Resize(fyne.NewSize(300, 40))
	progressBar := widget.NewProgressBarInfinite()
	label := widget.NewLabelWithStyle("Starting Engines...", fyne.TextAlignCenter, fyne.TextStyle{})
	app.windows.mainWindow.SetContent(
		widget.NewVBox(
			label,
			progressBar,
		),
	)
}

// ShowMainWindow - show the main window
func (app *App) ShowMainWindow() {
	app.windows.mainWindow.Show()
}

// RefreshToken - refresh the access and refresh token
// mainly needed for app start or new client creation
func (app *App) RefreshToken() {
	clientID := app.appInstance.Preferences().StringWithFallback("Client ID", "")
	isClientIDExists := clientID != ""
	refreshToken := app.appInstance.Preferences().StringWithFallback("Refresh Token", "")
	accessToken := app.appInstance.Preferences().StringWithFallback("Access Token", "")

	log.Println("Checking Configuration Deps")
	log.Println("ClientID Exists", isClientIDExists)
	log.Println("Refresh Token Exists", refreshToken != "")
	log.Println("Access Token Exists", accessToken != "")
	auth.SetAuthInfo(clientID, "")

	if !isClientIDExists || refreshToken == "" || accessToken == "" {
		// app.windows.configWindow = lib.OpenConfigurationScreen(app.appInstance, codeChallenge)
	} else {
		log.Println("Using Tokens")
		token := loadToken(app.appInstance)
		client := auth.NewClient(token)
		newToken, _ := client.Token()
		if newToken != nil {
			saveToken(app.appInstance, newToken)
		}
		app.SetClient(&client)
		app.authenticated = true
	}
}

// SetClient - Set the app client instance to be handled after the server goroutine has completed
func (app *App) SetClient(client *spotify.Client) {
	app.client = client
	log.Println("Oauth Connected")
	user, err := client.CurrentUser()
	if err != nil {
		log.Println("Client User Fetch error: ", err.Error())
		client.Token()
		if strings.Contains(err.Error(), "oauth2: cannot fetch token") {
			app.authenticated = false
			log.Println("Opening Configuration Screen on Token Fetch")
			app.DrawConfigScreen()
			app.ShowConfigScreen()
			app.appInstance.Preferences().RemoveValue("Access Token")
			app.appInstance.Preferences().RemoveValue("Refresh Token")
		} else {
			log.Fatal(err)
		}
		log.Println("Waiting for new client...")
	} else {
		log.Println("You are logged in as:", user.ID)
		if app.windows.configWindow != nil {
			app.windows.configWindow.Close()
		}
	}
	if user != nil {
		app.user = *user
	}
}

// SaveToken  - Save the Oauth Token
func (app *App) SaveToken(token *oauth2.Token) {
	log.Println("Setting token into App")

	app.token = token

	log.Println("Saving Token Details")
	app.appInstance.Preferences().SetString("Access Token", token.AccessToken)
	app.appInstance.Preferences().SetString("Refresh Token", token.RefreshToken)
	app.appInstance.Preferences().SetString("Token Type", token.TokenType)
	app.appInstance.Preferences().SetString("Token Expiry", token.Expiry.Local().Format(timeLayout))
}

// LoadToken - Load Oauth Token into App
func (app *App) LoadToken() {
	log.Println("Loading token into App")
	parsedTime, _ := time.Parse(timeLayout, app.appInstance.Preferences().String("Token Expiry"))
	app.token = &oauth2.Token{
		AccessToken:  app.appInstance.Preferences().String("Access Token"),
		RefreshToken: app.appInstance.Preferences().String("Refresh Token"),
		TokenType:    app.appInstance.Preferences().String("Token Type"),
		Expiry:       parsedTime,
	}
}

// DrawPlayerView - Draw the player view
func (app *App) DrawPlayerView() chan bool {
	var stop chan bool

	if !app.authenticated {
		return stop
	}

	currentPlayingLabel := widget.NewLabel("Loading...")
	currentArtistLabel := widget.NewLabel("Loading...")

	app.labels.artistLabel = currentArtistLabel
	app.labels.trackLabel = currentPlayingLabel

	playButton := widget.NewButton("Play", func() {
		app.client.Play()
		err := app.client.Play().Error()
		if err != "" {
			log.Println("Error Playing", err)
		}
		currentPlayingLabel.SetText("Loading...")
	})
	pauseButton := widget.NewButton("Pause", func() {
		app.client.Pause()
		currentPlayingLabel.SetText("Loading...")
	})
	nextButton := widget.NewButton("Next", func() {
		app.client.Next()
		currentPlayingLabel.SetText("Loading...")
	})

	backButton := widget.NewButton("Prev", func() {
		app.client.Previous()
		currentPlayingLabel.SetText("Loading...")
	})

	playing, err := app.client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Println("Error getting player", err)
		return stop
	}

	likeButtonText := ""

	if playing.Playing && app.checkIfUserHasTracks(playing.Item.ID) {
		likeButtonText = "Remove From Library"
	} else {
		likeButtonText = "Add to Library"
	}

	app.buttons.likeButton = widget.NewButton(likeButtonText, func() {
		hasTrack := app.checkIfUserHasTracks(playing.Item.ID)
		if !hasTrack {
			err = app.client.AddTracksToLibrary(playing.Item.ID)
			if err != nil {
				log.Println("Error Adding track", err)
			}
		} else {
			err = app.client.RemoveTracksFromLibrary(playing.Item.ID)
			if err != nil {
				log.Println("Error Removing track", err)
			}
		}
	})

	needPremiumLabel := widget.NewLabelWithStyle("I'm sorry but you can't change playback \n state without spotify premium", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	needPremiumSubLabel := widget.NewLabelWithStyle("Close and reopen the app if you upgraded to spotify premium", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	stop = Schedule(func() {
		app.UpdatePlayerLabels()
	}, 2*time.Second)

	playerControls := widget.NewHBox(
		playButton,
		pauseButton,
		nextButton,
		backButton,
		app.buttons.likeButton,
	)

	if app.user.Product != "premium" {
		playerControls = widget.NewVBox(
			needPremiumLabel,
			needPremiumSubLabel,
		)
	}

	app.windows.mainWindow.SetContent(widget.NewVBox(
		currentPlayingLabel,
		currentArtistLabel,
		playerControls,
	))

	return stop
}

// ShowPlayerView - Show the player view
func (app *App) ShowPlayerView() {
	if !app.authenticated {
		return
	}
	app.windows.mainWindow.Show()
}

// DrawConfigScreen - Draw the configuration screen to allow user to set prefs and spotify client settings
func (app *App) DrawConfigScreen() {
	app.windows.configWindow = app.appInstance.NewWindow("Configuration")
	clientID := app.appInstance.Preferences().StringWithFallback("Client ID", "")

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
		app.appInstance.Preferences().SetString("Client ID", clientIDEntry.Text)
		app.SaveScopes(scopes...)
		initiateOAuthFlow(clientIDEntry.Text, codeChallenge)
	})

	app.windows.configWindow.SetContent(
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

	app.windows.configWindow.Show()
}

// ShowConfigScreen - Show the config screen
func (app *App) ShowConfigScreen() {

}

// SyncScopes - sync current scope requests with existing prefs
func (app *App) SyncScopes(askingScopes ...string) {
	savedScopes := app.appInstance.Preferences().StringWithFallback("Scopes", "")
	askingScopesString := strings.Join(askingScopes[:], ",")
	log.Println("savedScopes", savedScopes)
	log.Println("askingScopesString", askingScopesString)

	if len(askingScopesString) != len(savedScopes) {
		app.appInstance.Preferences().RemoveValue("Scopes")
		app.appInstance.Preferences().RemoveValue("Access Token")
		app.appInstance.Preferences().RemoveValue("Refresh Token")
		app.authenticated = false
		app.SaveScopes(askingScopes...)
	}
}

// SaveScopes - Save new scopes
func (app *App) SaveScopes(toSave ...string) {
	toSaveString := strings.Join(toSave[:], ",")
	app.appInstance.Preferences().SetString("Scopes", toSaveString)
}

// UpdatePlayerLabels - Update the player labels
func (app *App) UpdatePlayerLabels() {
	playing, err := app.client.PlayerCurrentlyPlaying()
	if err != nil {
		log.Print("Label Update Fail:", err)
	}
	if playing == nil {
		return
	}
	if !playing.Playing {
		app.labels.trackLabel.SetText("Not Playing anything...")
		app.labels.artistLabel.SetText("-")
	} else {
		app.labels.trackLabel.SetText(playing.Item.Name)
		app.labels.artistLabel.SetText(playing.Item.Artists[0].Name)
	}

	likeButtonText := ""

	if playing.Playing && app.checkIfUserHasTracks(playing.Item.ID) {
		likeButtonText = "Remove From Library"
	} else {
		likeButtonText = "Add to Library"
	}

	if app.buttons.likeButton != nil {
		app.buttons.likeButton.SetText(likeButtonText)
	}
	return
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

func (app *App) checkIfUserHasTracks(trackID spotify.ID) bool {
	hasTrack := false

	userHasTrack, err := app.client.UserHasTracks(trackID)
	if err != nil {
		log.Println("Error checking user tracks", err)
	}

	hasTrack = (len(userHasTrack) > 0 && userHasTrack[0]) || false

	return hasTrack
}
