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

	for timeLeft > 0 {
		select {
		case <-stop:
			interrupt()
			return
		case <-time.After(1 * time.Second):
			timeLeft--

			go func(t int) {
				callback(t)
			}(timeLeft)

			if timeLeft <= 0 {
				go end()
				return
			}
		}
	}
}
