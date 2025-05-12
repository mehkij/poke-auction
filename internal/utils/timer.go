package utils

import (
	"time"
)

/*
Timer takes several parameters:

A duration (in seconds);
A stop signal;
A callback function to be called every tick;
A callback function that is called when the timer ends;
And a final callback function that is called whenever the stop signal recieves a "true" signal.
*/
func Timer(duration int, stop <-chan bool, callback func(duration int), end func(), interrupt func()) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeLeft := duration

	for {
		select {
		case <-stop:
			// ticker.Stop()
			interrupt()
			return
		case <-ticker.C:
			timeLeft--
			callback(timeLeft)

			if timeLeft <= 0 {
				end()
				return
			}
		}
	}
}
