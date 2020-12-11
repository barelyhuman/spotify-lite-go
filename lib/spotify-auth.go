package lib

import (
	"sync"

	"github.com/zmb3/spotify"
)

var auth spotify.Authenticator

var getAuthInstanceOnce sync.Once

// GetAuthenticator - Get Spotify Authenticator instance
func GetAuthenticator() spotify.Authenticator {
	getAuthInstanceOnce.Do(func() {
		SetAuthenticator()
	})
	return auth
}

// SetAuthenticator - set a new authenticator instance
func SetAuthenticator() {
	redirectURL := GetRedirectURL()
	auth = spotify.NewAuthenticator(redirectURL, spotify.ScopeUserReadCurrentlyPlaying, spotify.ScopeUserReadPlaybackState, spotify.ScopeUserModifyPlaybackState)
}
