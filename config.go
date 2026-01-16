package main

const (
	API_KEY = "a790dde799e34bce82e320563c98fc94"

	WS_URL = "wss://api.twelvedata.com/v1/quotes/price?symbol=XAU/USD&apikey=" + API_KEY

	TELEGRAM_TOKEN = "8281709039:AAH9-RsVgawaFI21vOQBHalgSg5C7K1lGJI"
	TELEGRAM_CHAT  = "8245959182"

	MODE = "SCALPING" // SCALPING or LONG
	// ===== SESSION SELECTION =====
	// "AUTO" = Automatic session detection based on your timezone
	// "LONDON_NY" = Force London-NY session only
	// "ASIA" = Force Asia session only
	// "ALL" = Trade all sessions (24/7)
	SESSION = "AUTO" // 🔥 RECOMMENDED: Auto-detect best session!

	// ===== YOUR TIMEZONE =====
	// Set your local timezone for accurate session detection
	// Examples: "Asia/Jakarta", "Asia/Singapore", "Asia/Tokyo"
	// Full list: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
	YOUR_TIMEZONE = "Asia/Jakarta" // WIB (UTC+7)

	// ===== POLLING =====
	POLLING_SECONDS = 15

	// ===== SESSION CHARACTERISTICS =====
	// London-NY: High volatility, strong trends, tighter spreads
	// Asia: Moderate-low volatility, more ranging, wider spreads early session

	// ===== AUTO MODE LOGIC =====
	// Bot will automatically choose the best session based on your local time:
	//
	// WIB Time (Jakarta/Cikarang):
	// 03:00-15:00 WIB → ASIA Session (Tokyo/Sydney active)
	// 15:00-05:00 WIB → LONDON-NY Session (Europe/US active)
	//
	// Dead hours (09:00-11:00 WIB) will be skipped automatically
)
