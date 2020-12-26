package lib

import (
	"log"

	"fyne.io/fyne"
)

// ChangedSubscription - checks if subscription state changed
func ChangedSubscription(appInstance fyne.App, current string) bool {
	savedSubState := appInstance.Preferences().StringWithFallback("SubState", "")
	log.Println("Saved Sub State", savedSubState)
	log.Println("Current Sub State", current)
	if savedSubState != current {
		return true
	}
	return false
}

// SaveSubscriptionState - update state in config
func SaveSubscriptionState(appInstance fyne.App, state string) {
	appInstance.Preferences().SetString("SubState", state)
}
