package main

const (
	API_KEY = "a790dde799e34bce82e320563c98fc94"

	WS_URL = "wss://api.twelvedata.com/v1/quotes/price?symbol=XAU/USD&apikey=" + API_KEY

	TELEGRAM_TOKEN = "8281709039:AAH9-RsVgawaFI21vOQBHalgSg5C7K1lGJI"
	TELEGRAM_CHAT  = "8245959182"

	MODE = "SCALPING" // SCALPING or LONG

	// ===== POLLING =====
	POLLING_SECONDS = 15

	// ===== STRATEGY PARAMETERS =====
	// Minimum number of confirmations needed for signal
	// SCALPING: 3 confirmations (more aggressive)
	// LONG: 4 confirmations (more conservative)
)
