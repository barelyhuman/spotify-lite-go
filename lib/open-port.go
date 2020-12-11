package lib

import (
	"fmt"
	"net"
	"sync"
)

type openPort string

var connectedPort openPort
var getOpenPortOnce sync.Once

// CheckOpenPort - look for available ports
func CheckOpenPort() string {
	ports := []string{"1821", "12314", "12312", "1293"}

	getOpenPortOnce.Do(func() {
		for _, port := range ports {
			l, err := net.Listen("tcp", ":"+port)

			if err != nil {
				// Log or report the error here
				fmt.Printf("Error: %s\n", err)
			}
			defer l.Close()

			connectedPort = openPort(port)
		}
	})

	return string(connectedPort)
}

// GetRedirectURL - Get redirect url for spotify app
func GetRedirectURL() string {
	return "http://localhost:" + string(connectedPort) + "/callback"

}
