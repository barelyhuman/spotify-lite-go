package main

import "time"

// Schedule - Schedule a function to a dedicated time delay
func Schedule(toExec func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			toExec()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}
