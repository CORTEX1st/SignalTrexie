package main

import "fmt"

type SignalData struct {
	Action      string
	Entry       float64
	StopLoss    float64
	TakeProfit1 float64
	TakeProfit2 float64
	TakeProfit3 float64
	RiskReward  float64
	Confidence  int
	Reasons     []string
	Session     string
}

func GenerateSignal(prices []float64) string {
	signal := GenerateSignalAdvanced(prices)
	return signal.Action
}

func GenerateSignalAdvanced(prices []float64) SignalData {
	signal := SignalData{
		Action:  "WAIT",
		Reasons: []string{},
		Session: GetSessionName(),
	}

	if len(prices) < 60 {
		return signal
	}

	currentPrice := prices[len(prices)-1]
	signal.Entry = currentPrice

	// Calculate indicators
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

	fib := CalculateFibonacci(prices, 50)
	pivot := CalculatePivotPoints(prices)

	tolerance := currentPrice * 0.001

	last := len(prices) - 1

	isAsia := IsAsiaSession()
	isHighVol := IsHighVolatilityPeriod()

	// ===== CRITICAL: MINIMUM TP/SL FOR XAUUSD =====
	// Never allow TP/SL smaller than these values!
	var minSL, minTP1, minTP2, minTP3 float64

	if isAsia {
		// Asia session: Lower volatility, but still need meaningful targets
		if MODE == "SCALPING" {
			minSL = currentPrice * 0.0008  // 0.08% = $3.68 for $4600 gold
			minTP1 = currentPrice * 0.0015 // 0.15% = $6.90
			minTP2 = currentPrice * 0.0025 // 0.25% = $11.50
			minTP3 = currentPrice * 0.0040 // 0.40% = $18.40
		} else {
			minSL = currentPrice * 0.0012
			minTP1 = currentPrice * 0.0020
			minTP2 = currentPrice * 0.0035
			minTP3 = currentPrice * 0.0055
		}
	} else {
		// London-NY: Higher volatility, wider targets
		if MODE == "SCALPING" {
			minSL = currentPrice * 0.0010  // 0.10% = $4.60
			minTP1 = currentPrice * 0.0020 // 0.20% = $9.20
			minTP2 = currentPrice * 0.0035 // 0.35% = $16.10
			minTP3 = currentPrice * 0.0055 // 0.55% = $25.30
		} else {
			minSL = currentPrice * 0.0015
			minTP1 = currentPrice * 0.0025
			minTP2 = currentPrice * 0.0045
			minTP3 = currentPrice * 0.0070
		}
	}

	// ===== ATR MULTIPLIERS (as before) =====
	var atrMultiplierSL, atrMultiplierTP1, atrMultiplierTP2, atrMultiplierTP3 float64

	if isAsia {
		if MODE == "SCALPING" {
			atrMultiplierSL = 1.5
			atrMultiplierTP1 = 2.0
			atrMultiplierTP2 = 3.2
			atrMultiplierTP3 = 5.0
		} else {
			atrMultiplierSL = 1.8
			atrMultiplierTP1 = 2.5
			atrMultiplierTP2 = 4.0
			atrMultiplierTP3 = 6.0
		}
	} else {
		if MODE == "SCALPING" {
			atrMultiplierSL = 1.8
			atrMultiplierTP1 = 2.5
			atrMultiplierTP2 = 4.0
			atrMultiplierTP3 = 6.5
		} else {
			atrMultiplierSL = 2.0
			atrMultiplierTP1 = 3.0
			atrMultiplierTP2 = 5.0
			atrMultiplierTP3 = 7.5
		}

		if isHighVol {
			atrMultiplierSL = atrMultiplierSL * 1.2
			atrMultiplierTP1 = atrMultiplierTP1 * 1.15
			atrMultiplierTP2 = atrMultiplierTP2 * 1.15
			atrMultiplierTP3 = atrMultiplierTP3 * 1.15
		}
	}

	// ===== BUY CONFIRMATIONS =====
	buyConfirmations := 0
	buyReasons := []string{}

	if isAsia {
		if currentPrice > ema21[last] && ema21[last] > ema50[last] {
			buyConfirmations++
			buyReasons = append(buyReasons, "✅ Price Above EMA21>50")
		}
	} else {
		if ema9[last] > ema21[last] && ema21[last] > ema50[last] {
			buyConfirmations++
			buyReasons = append(buyReasons, "✅ EMA Uptrend (9>21>50)")
		}
	}

	if len(ema200) > 0 && currentPrice > ema200[last] {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ Above EMA200")
	}

	if isAsia {
		if rsi > 35 && rsi < 65 {
			buyConfirmations++
			buyReasons = append(buyReasons, fmt.Sprintf("✅ RSI Range (%.1f)", rsi))
		}
		if rsi < 40 {
			buyConfirmations++
			buyReasons = append(buyReasons, "✅ RSI Oversold Bounce")
		}
	} else {
		if rsi > 40 && rsi < 70 {
			buyConfirmations++
			buyReasons = append(buyReasons, fmt.Sprintf("✅ RSI Bullish (%.1f)", rsi))
		}
	}

	if macd.Histogram > 0 && macd.MACD > macd.Signal {
		buyConfirmations++
		buyReasons = append(buyReasons, "✅ MACD Bullish Cross")
	}

	if isAsia {
		if currentPrice <= bb.Middle*1.005 && currentPrice > bb.Lower {
			buyConfirmations += 2
			buyReasons = append(buyReasons, "✅✅ BB Oversold Bounce (Asia)")
		}
	} else {
		if currentPrice <= bb.Middle && currentPrice > bb.Lower {
			buyConfirmations++
			buyReasons = append(buyReasons, "✅ BB Support Zone")
		}
	}

	if stoch.K > 20 && stoch.K < 80 {
		buyConfirmations++
		buyReasons = append(buyReasons, fmt.Sprintf("✅ Stochastic (%.1f)", stoch.K))
	}

	if isAsia {
		if adx > 15 && adx < 30 {
			buyConfirmations++
			buyReasons = append(buyReasons, fmt.Sprintf("✅ Moderate Trend (ADX %.1f)", adx))
		}
	} else {
		minADX := 25.0
		if isHighVol {
			minADX = 30.0
		}
		if adx > minADX {
			buyConfirmations++
			buyReasons = append(buyReasons, fmt.Sprintf("✅ Strong Trend (ADX %.1f)", adx))
		}
	}

	if currentPrice <= sr.Support*1.01 {
		if isAsia {
			buyConfirmations += 2
			buyReasons = append(buyReasons, "✅✅ At Support Level (Asia)")
		} else {
			buyConfirmations++
			buyReasons = append(buyReasons, "✅ Near Support")
		}
	}

	// Fibonacci confirmations
	if fib.Trend == "BULLISH" {
		fibLevel, fibPrice := GetNearestFibRetracement(currentPrice, fib, tolerance)
		if fibLevel != "" {
			buyConfirmations += 2
			buyReasons = append(buyReasons, fmt.Sprintf("✅✅ Fib %s Pullback (%.2f)", fibLevel, fibPrice))
		}
	}

	if IsNearFibLevel(currentPrice, pivot.S1, tolerance) {
		buyConfirmations += 1
		buyReasons = append(buyReasons, fmt.Sprintf("✅ At Pivot S1 (%.2f)", pivot.S1))
	}

	// ===== SELL CONFIRMATIONS =====
	sellConfirmations := 0
	sellReasons := []string{}

	if isAsia {
		if currentPrice < ema21[last] && ema21[last] < ema50[last] {
			sellConfirmations++
			sellReasons = append(sellReasons, "✅ Price Below EMA21<50")
		}
	} else {
		if ema9[last] < ema21[last] && ema21[last] < ema50[last] {
			sellConfirmations++
			sellReasons = append(sellReasons, "✅ EMA Downtrend (9<21<50)")
		}
	}

	if len(ema200) > 0 && currentPrice < ema200[last] {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ Below EMA200")
	}

	if isAsia {
		if rsi < 65 && rsi > 35 {
			sellConfirmations++
			sellReasons = append(sellReasons, fmt.Sprintf("✅ RSI Range (%.1f)", rsi))
		}
		if rsi > 60 {
			sellConfirmations++
			sellReasons = append(sellReasons, "✅ RSI Overbought Drop")
		}
	} else {
		if rsi < 60 && rsi > 30 {
			sellConfirmations++
			sellReasons = append(sellReasons, fmt.Sprintf("✅ RSI Bearish (%.1f)", rsi))
		}
	}

	if macd.Histogram < 0 && macd.MACD < macd.Signal {
		sellConfirmations++
		sellReasons = append(sellReasons, "✅ MACD Bearish Cross")
	}

	if isAsia {
		if currentPrice >= bb.Middle*0.995 && currentPrice < bb.Upper {
			sellConfirmations += 2
			sellReasons = append(sellReasons, "✅✅ BB Overbought Drop (Asia)")
		}
	} else {
		if currentPrice >= bb.Middle && currentPrice < bb.Upper {
			sellConfirmations++
			sellReasons = append(sellReasons, "✅ BB Resistance Zone")
		}
	}

	if stoch.K < 80 && stoch.K > 20 {
		sellConfirmations++
		sellReasons = append(sellReasons, fmt.Sprintf("✅ Stochastic (%.1f)", stoch.K))
	}

	if isAsia {
		if adx > 15 && adx < 30 {
			sellConfirmations++
			sellReasons = append(sellReasons, fmt.Sprintf("✅ Moderate Trend (ADX %.1f)", adx))
		}
	} else {
		minADX := 25.0
		if isHighVol {
			minADX = 30.0
		}
		if adx > minADX {
			sellConfirmations++
			sellReasons = append(sellReasons, fmt.Sprintf("✅ Strong Trend (ADX %.1f)", adx))
		}
	}

	if currentPrice >= sr.Resistance*0.99 {
		if isAsia {
			sellConfirmations += 2
			sellReasons = append(sellReasons, "✅✅ At Resistance Level (Asia)")
		} else {
			sellConfirmations++
			sellReasons = append(sellReasons, "✅ Near Resistance")
		}
	}

	if fib.Trend == "BEARISH" {
		fibLevel, fibPrice := GetNearestFibRetracement(currentPrice, fib, tolerance)
		if fibLevel != "" {
			sellConfirmations += 2
			sellReasons = append(sellReasons, fmt.Sprintf("✅✅ Fib %s Pullback (%.2f)", fibLevel, fibPrice))
		}
	}

	if IsNearFibLevel(currentPrice, pivot.R1, tolerance) {
		sellConfirmations += 1
		sellReasons = append(sellReasons, fmt.Sprintf("✅ At Pivot R1 (%.2f)", pivot.R1))
	}

	// ===== MINIMUM CONFIRMATIONS =====
	minConfirmations := 5 // Raised from 4 to reduce false signals
	if MODE == "SCALPING" {
		minConfirmations = 4 // Raised from 3
	}
	if isAsia {
		minConfirmations = 6 // Raised from 5
		if MODE == "SCALPING" {
			minConfirmations = 5 // Raised from 4
		}
	}

	// ===== HELPER FUNCTION: Apply minimum thresholds =====
	applyMinimum := func(value, minimum float64) float64 {
		if value < minimum {
			return minimum
		}
		return value
	}

	// ===== BUY SIGNAL =====
	if buyConfirmations >= minConfirmations {
		signal.Action = "BUY"
		signal.Reasons = buyReasons
		signal.Confidence = (buyConfirmations * 100) / 12

		// Calculate ATR-based values
		atrSL := currentPrice - (atr * atrMultiplierSL)
		atrTP1 := currentPrice + (atr * atrMultiplierTP1)
		atrTP2 := currentPrice + (atr * atrMultiplierTP2)
		atrTP3 := currentPrice + (atr * atrMultiplierTP3)

		// Apply MINIMUM thresholds (CRITICAL FIX!)
		signal.StopLoss = applyMinimum(currentPrice-atrSL, minSL)
		signal.StopLoss = currentPrice - signal.StopLoss // Convert back to price

		signal.TakeProfit1 = applyMinimum(atrTP1-currentPrice, minTP1)
		signal.TakeProfit1 = currentPrice + signal.TakeProfit1

		signal.TakeProfit2 = applyMinimum(atrTP2-currentPrice, minTP2)
		signal.TakeProfit2 = currentPrice + signal.TakeProfit2

		signal.TakeProfit3 = applyMinimum(atrTP3-currentPrice, minTP3)
		signal.TakeProfit3 = currentPrice + signal.TakeProfit3

		// Use Fibonacci if available and better than ATR
		if fib.Trend == "BULLISH" && MODE == "SCALPING" {
			if fib.Fib1272 > currentPrice && fib.Fib1272 < signal.TakeProfit1*2 {
				signal.TakeProfit1 = fib.Fib1272
			}
			if fib.Fib1618 > currentPrice && fib.Fib1618 < signal.TakeProfit2*2 {
				signal.TakeProfit2 = fib.Fib1618
			}
			if fib.Fib2000 > currentPrice && fib.Fib2000 < signal.TakeProfit3*2 {
				signal.TakeProfit3 = fib.Fib2000
			}
		}

		signal.RiskReward = (signal.TakeProfit1 - currentPrice) / (currentPrice - signal.StopLoss)

		// FINAL SAFETY CHECK: Reject if R:R too low
		if signal.RiskReward < 1.3 {
			signal.Action = "WAIT"
			signal.Reasons = []string{"❌ Risk:Reward too low (< 1.3)"}
			return signal
		}
	}

	// ===== SELL SIGNAL =====
	if sellConfirmations >= minConfirmations {
		signal.Action = "SELL"
		signal.Reasons = sellReasons
		signal.Confidence = (sellConfirmations * 100) / 12

		atrSL := currentPrice + (atr * atrMultiplierSL)
		atrTP1 := currentPrice - (atr * atrMultiplierTP1)
		atrTP2 := currentPrice - (atr * atrMultiplierTP2)
		atrTP3 := currentPrice - (atr * atrMultiplierTP3)

		signal.StopLoss = applyMinimum(atrSL-currentPrice, minSL)
		signal.StopLoss = currentPrice + signal.StopLoss

		signal.TakeProfit1 = applyMinimum(currentPrice-atrTP1, minTP1)
		signal.TakeProfit1 = currentPrice - signal.TakeProfit1

		signal.TakeProfit2 = applyMinimum(currentPrice-atrTP2, minTP2)
		signal.TakeProfit2 = currentPrice - signal.TakeProfit2

		signal.TakeProfit3 = applyMinimum(currentPrice-atrTP3, minTP3)
		signal.TakeProfit3 = currentPrice - signal.TakeProfit3

		if fib.Trend == "BEARISH" && MODE == "SCALPING" {
			if fib.Fib1272 < currentPrice && fib.Fib1272 > signal.TakeProfit1*0.5 {
				signal.TakeProfit1 = fib.Fib1272
			}
			if fib.Fib1618 < currentPrice && fib.Fib1618 > signal.TakeProfit2*0.5 {
				signal.TakeProfit2 = fib.Fib1618
			}
			if fib.Fib2000 < currentPrice && fib.Fib2000 > signal.TakeProfit3*0.5 {
				signal.TakeProfit3 = fib.Fib2000
			}
		}

		signal.RiskReward = (currentPrice - signal.TakeProfit1) / (signal.StopLoss - currentPrice)

		if signal.RiskReward < 1.3 {
			signal.Action = "WAIT"
			signal.Reasons = []string{"❌ Risk:Reward too low (< 1.3)"}
			return signal
		}
	}

	return signal
}
