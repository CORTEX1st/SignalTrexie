package main

import "time"

func IsTradingSession() bool {
	hour := time.Now().UTC().Hour()

	// London + New York
	return hour >= 7 && hour <= 21
}
