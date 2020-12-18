package lib

import "fyne.io/fyne"

var (
	app fyne.App
)

// SetApp - set app instance for the lib package
func SetApp(appInstance fyne.App) {
	app = appInstance
}
