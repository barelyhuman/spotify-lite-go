package lib

import (
	"time"

	"fyne.io/fyne"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// GetToken - Handle renew and creation of token
func GetToken(appInstance fyne.App, auth spotify.Authenticator, client *spotify.Client) bool {

	accessToken := appInstance.Preferences().StringWithFallback("Access Token", "")
	refreshToken := appInstance.Preferences().StringWithFallback("Refresh Token", "")
	clientID := appInstance.Preferences().StringWithFallback("Client ID", "")
	clientSecret := appInstance.Preferences().StringWithFallback("Client Secret", "")

	auth.SetAuthInfo(clientID, clientSecret)

	token := &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken}
	*client = auth.NewClient(token)

	if m, _ := time.ParseDuration("5m30s"); time.Until(token.Expiry) < m {
		newToken, _ := client.Token()
		appInstance.Preferences().SetString("Access Token", newToken.AccessToken)
		*client = auth.NewClient(newToken)
	}

	_, err := client.CurrentUser()

	if err != nil {
		return false
	}

	return true
}
