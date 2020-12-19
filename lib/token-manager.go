package lib

import (
	"strings"
	"time"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

// GetToken - Handle renew and creation of token
func GetToken(client *spotify.Client) bool {

	auth := GetAuthenticator()
	accessToken := app.Preferences().StringWithFallback("Access Token", "")
	refreshToken := app.Preferences().StringWithFallback("Refresh Token", "")
	clientID := app.Preferences().StringWithFallback("Client ID", "")
	clientSecret := app.Preferences().StringWithFallback("Client Secret", "")

	auth.SetAuthInfo(clientID, clientSecret)

	token := &oauth2.Token{AccessToken: accessToken, RefreshToken: refreshToken}
	*client = auth.NewClient(token)

	if m, _ := time.ParseDuration("5m30s"); time.Until(token.Expiry) < m {
		newToken, err := client.Token()
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				return false
			}
		}
		app.Preferences().SetString("Access Token", newToken.AccessToken)
		*client = auth.NewClient(newToken)
	}

	_, err := client.CurrentUser()

	if err != nil {
		return false
	}

	return true
}
