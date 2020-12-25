package lib

import (
	"strings"

	"fyne.io/fyne"
)

// SyncScopes - synchorise config asked scopes and currently required scopes in-case they change over time
func SyncScopes(appInstance fyne.App, askingScopes ...string) {
	savedScopes := appInstance.Preferences().StringWithFallback("Scopes", "")
	askingScopesString := strings.Join(askingScopes[:], ",")
	if len(askingScopesString) != len(savedScopes) {
		appInstance.Preferences().RemoveValue("Scopes")
		appInstance.Preferences().RemoveValue("Access Token")
		appInstance.Preferences().RemoveValue("Refresh Token")
	}
}

// SaveScopes - Save scopes to app prefs
func SaveScopes(appInstance fyne.App, toSave ...string) {
	toSaveString := strings.Join(toSave[:], ",")
	appInstance.Preferences().SetString("Scopes", toSaveString)
}
