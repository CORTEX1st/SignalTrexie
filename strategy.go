package main

import "fmt"

type SignalData struct {
	Action      string // BUY, SELL, WAIT
	Entry       float64
	StopLoss    float64
	TakeProfit1 float64
	TakeProfit2 float64
	TakeProfit3 float64
	RiskReward  float64
	Confidence  int // 0-100
	Reasons     []string
}

func GenerateSignal(prices []float64) string {
	signal := GenerateSignalAdvanced(prices)
	return signal.Action
}

func GenerateSignalAdvanced(prices []float64) SignalData {
	signal := SignalData{
		Action:  "WAIT",
		Reasons: []string{},
	}

	if len(prices) < 60 {
		return signal
	}

	// Get current price
	currentPrice := prices[len(prices)-1]
	signal.Entry = currentPrice

	// Calculate all indicators
	ema9 := EMA(prices, 9)
	ema21 := EMA(prices, 21)
	ema50 := EMA(prices, 50)
	ema200 := EMA(prices, 200)

	rsi := RSI(prices, 14)
	macd := MACDFull(prices)
	bb := BBands(prices, 20, 2.0)
	stoch := StochasticOscillator(prices, 14)
	adx := ADX(prices, 14)
	atr := ATR(prices, 14)
	sr := FindSupportResistance(prices, 50)

	last := len(prices) - 1

	// Risk management based on ATR
	atrMultiplier := 1.5
	if MODE == "SCALPING" {
		atrMultiplier = 1.2
	}

	// Count confirmations for BUY
	buyConfirmations := 0
	buyReasons := []string{}

	// 1. EMA Trend (9 > 21 > 50)
	if ema9[last] > ema21[last] && ema21[last] > ema50[last] {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ EMA Uptrend (9>21>50)")
	}

	// 2. Price above EMA200 (long-term bullish)
	if len(ema200) > 0 && currentPrice > ema200[last] {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ Above EMA200")
	}

	// 3. RSI in buy zone (40-70)
	if rsi > 40 && rsi < 70 {
		buyConfirmations++
		buyReasons = append(buyReasons, fmt.Sprintf("✅ RSI Bullish (%.1f)", rsi))
	}

	// 4. MACD bullish (histogram positive and increasing)
	if macd.Histogram > 0 && macd.MACD > macd.Signal {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ MACD Bullish Cross")
	}

	// 5. Bollinger Bands (price near lower band = oversold bounce)
	if currentPrice <= bb.Middle && currentPrice > bb.Lower {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ BB Oversold Bounce")
	}

	// 6. Stochastic oversold recovery
	if stoch.K > 20 && stoch.K < 80 {
		buyConfirmations++
		buyReasons = append(buyReasons, fmt.Sprintf("✅ Stochastic (%.1f)", stoch.K))
	}

	// 7. Strong trend confirmation (ADX > 25)
	if adx > 25 {
		buyConfirmations++
		buyReasons = append(buyReasons, fmt.Sprintf("✅ Strong Trend (ADX %.1f)", adx))
	}

	// 8. Near support level
	if currentPrice <= sr.Support*1.01 {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ Near Support")
	}

	// Count confirmations for SELL
	sellConfirmations := 0
	sellReasons := []string{}

	// 1. EMA Trend (9 < 21 < 50)
	if ema9[last] < ema21[last] && ema21[last] < ema50[last] {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ EMA Downtrend (9<21<50)")
	}

	// 2. Price below EMA200 (long-term bearish)
	if len(ema200) > 0 && currentPrice < ema200[last] {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ Below EMA200")
	}

	// 3. RSI in sell zone (30-60)
	if rsi < 60 && rsi > 30 {
		sellConfirmations++
		sellReasons = append(sellReasons, fmt.Sprintf("✅ RSI Bearish (%.1f)", rsi))
	}

	// 4. MACD bearish (histogram negative and decreasing)
	if macd.Histogram < 0 && macd.MACD < macd.Signal {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ MACD Bearish Cross")
	}

	// 5. Bollinger Bands (price near upper band = overbought)
	if currentPrice >= bb.Middle && currentPrice < bb.Upper {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ BB Overbought")
	}

	// 6. Stochastic overbought
	if stoch.K < 80 && stoch.K > 20 {
		sellConfirmations++
		sellReasons = append(sellReasons, fmt.Sprintf("✅ Stochastic (%.1f)", stoch.K))
	}

	// 7. Strong trend confirmation (ADX > 25)
	if adx > 25 {
		sellConfirmations++
		sellReasons = append(sellReasons, fmt.Sprintf("✅ Strong Trend (ADX %.1f)", adx))
	}

	// 8. Near resistance level
	if currentPrice >= sr.Resistance*0.99 {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ Near Resistance")
	}

	// Decision: Need at least 4 confirmations for high-quality signal
	minConfirmations := 4
	if MODE == "SCALPING" {
		minConfirmations = 3 // More aggressive for scalping
	}

	// BUY SIGNAL
	if buyConfirmations >= minConfirmations {
		signal.Action = "BUY"
		signal.Reasons = buyReasons
		signal.Confidence = (buyConfirmations * 100) / 8

		// Calculate Stop Loss and Take Profit
		signal.StopLoss = currentPrice - (atr * atrMultiplier)

		if MODE == "SCALPING" {
			// Scalping: tighter targets
			signal.TakeProfit1 = currentPrice + (atr * 1.5)
			signal.TakeProfit2 = currentPrice + (atr * 2.5)
			signal.TakeProfit3 = currentPrice + (atr * 4.0)
		} else {
			// Long-term: wider targets
			signal.TakeProfit1 = currentPrice + (atr * 2.0)
			signal.TakeProfit2 = currentPrice + (atr * 4.0)
			signal.TakeProfit3 = currentPrice + (atr * 6.0)
		}

		signal.RiskReward = (signal.TakeProfit1 - currentPrice) / (currentPrice - signal.StopLoss)
	}

	// SELL SIGNAL
	if sellConfirmations >= minConfirmations {
		signal.Action = "SELL"
		signal.Reasons = sellReasons
		signal.Confidence = (sellConfirmations * 100) / 8

		// Calculate Stop Loss and Take Profit
		signal.StopLoss = currentPrice + (atr * atrMultiplier)

		if MODE == "SCALPING" {
			// Scalping: tighter targets
			signal.TakeProfit1 = currentPrice - (atr * 1.5)
			signal.TakeProfit2 = currentPrice - (atr * 2.5)
			signal.TakeProfit3 = currentPrice - (atr * 4.0)
		} else {
			// Long-term: wider targets
			signal.TakeProfit1 = currentPrice - (atr * 2.0)
			signal.TakeProfit2 = currentPrice - (atr * 4.0)
			signal.TakeProfit3 = currentPrice - (atr * 6.0)
		}

		signal.RiskReward = (currentPrice - signal.TakeProfit1) / (signal.StopLoss - currentPrice)
	}

	return signal
}
