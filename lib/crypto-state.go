package lib

import "sync"

// OpenPort - string type for singleton
type crptoState string

var state crptoState
var getStateOnce sync.Once

// GetState - get current crypto state
func GetState() string {

	getStateOnce.Do(func() {
		NewState()
	})

	return string(state)
}

// NewState - Create a new state
func NewState() {
	state = "abc123"
}
