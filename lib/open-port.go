package lib

import (
	"fmt"
	"net"
)

// CheckOpenPort - check for open ports on localhost
func CheckOpenPort(host string, ports []string) string {
	var openPort string
	for _, port := range ports {
		l, err := net.Listen("tcp", ":"+port)
		defer l.Close()

		if err != nil {
			// Log or report the error here
			fmt.Printf("Error: %s\n", err)
		}

		openPort = port
	}
	return openPort
}
