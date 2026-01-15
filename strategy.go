package main

func GenerateSignal(prices []float64) string {
	if len(prices) < 60 {
		return "WAIT"
	}

	ema20 := EMA(prices, 20)
	ema50 := EMA(prices, 50)
	rsi := RSI(prices, 14)
	macd := MACD(prices)

	last := len(prices) - 1

	// BUY
	if ema20[last] > ema50[last] &&
		rsi > 50 && rsi < 70 &&
		macd > 0 {
		return "BUY"
	}

	// SELL
	if ema20[last] < ema50[last] &&
		rsi < 50 && rsi > 30 &&
		macd < 0 {
		return "SELL"
	}

	return "WAIT"
}
