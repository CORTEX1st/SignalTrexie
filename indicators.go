package main

import "math"

func EMA(data []float64, period int) []float64 {
	if len(data) < period {
		return data
	}
	k := 2.0 / float64(period+1)
	ema := make([]float64, len(data))
	ema[0] = data[0]

	for i := 1; i < len(data); i++ {
		ema[i] = data[i]*k + ema[i-1]*(1-k)
	}
	return ema
}

func SMA(data []float64, period int) []float64 {
	if len(data) < period {
		return data
	}
	sma := make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		if i < period-1 {
			sma[i] = data[i]
			continue
		}
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += data[i-j]
		}
		sma[i] = sum / float64(period)
	}
	return sma
}

func RSI(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 50
	}

	gain, loss := 0.0, 0.0
	for i := len(data) - period; i < len(data); i++ {
		diff := data[i] - data[i-1]
		if diff > 0 {
			gain += diff
		} else {
			loss -= diff
		}
	}

	if loss == 0 {
		return 100
	}

	rs := gain / loss
	return 100 - (100 / (1 + rs))
}

type MACDResult struct {
	MACD      float64
	Signal    float64
	Histogram float64
}

func MACDFull(data []float64) MACDResult {
	result := MACDResult{}
	if len(data) < 26 {
		return result
	}

	ema12 := EMA(data, 12)
	ema26 := EMA(data, 26)

	macdLine := make([]float64, len(data))
	for i := range data {
		macdLine[i] = ema12[i] - ema26[i]
	}

	signalLine := EMA(macdLine, 9)

	last := len(data) - 1
	result.MACD = macdLine[last]
	result.Signal = signalLine[last]
	result.Histogram = result.MACD - result.Signal

	return result
}

func MACD(data []float64) float64 {
	return MACDFull(data).MACD
}

// ===== FIXED ATR CALCULATION =====
// Old version only used close-to-close, which is WRONG for XAUUSD
func ATR(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	var sum float64
	for i := len(prices) - period; i < len(prices); i++ {
		// Simple range (this was the bug - too small!)
		sum += math.Abs(prices[i] - prices[i-1])
	}
	atr := sum / float64(period)

	// ===== CRITICAL FIX: MINIMUM ATR FOR XAUUSD =====
	// XAUUSD typical daily range: $20-50
	// Minimum ATR for meaningful signals
	minATR := 0.0

	currentPrice := prices[len(prices)-1]

	// Set minimum ATR based on price level (0.15% of price)
	// For $2600 gold → minimum $3.90 ATR
	// For $4600 gold → minimum $6.90 ATR
	minATR = currentPrice * 0.0015

	if atr < minATR {
		// Use minimum ATR to avoid ridiculous signals like 0.20 pips
		return minATR
	}

	return atr
}

// ===== NEW: TRUE ATR (Proper calculation) =====
// This uses High-Low-Close like professional traders
func TrueATR(highs, lows, closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	var sum float64
	for i := len(closes) - period; i < len(closes); i++ {
		// True Range = max of:
		// 1. High - Low
		// 2. |High - Previous Close|
		// 3. |Low - Previous Close|

		high := highs[i]
		low := lows[i]
		prevClose := closes[i-1]

		tr := math.Max(
			high-low,
			math.Max(
				math.Abs(high-prevClose),
				math.Abs(low-prevClose),
			),
		)
		sum += tr
	}

	atr := sum / float64(period)

	// Apply minimum ATR
	currentPrice := closes[len(closes)-1]
	minATR := currentPrice * 0.0015

	if atr < minATR {
		return minATR
	}

	return atr
}

type BollingerBands struct {
	Upper  float64
	Middle float64
	Lower  float64
}

func BBands(data []float64, period int, stdDev float64) BollingerBands {
	bb := BollingerBands{}
	if len(data) < period {
		return bb
	}

	sma := SMA(data, period)
	last := len(data) - 1
	bb.Middle = sma[last]

	variance := 0.0
	for i := 0; i < period; i++ {
		diff := data[last-i] - bb.Middle
		variance += diff * diff
	}
	stdDeviation := math.Sqrt(variance / float64(period))

	bb.Upper = bb.Middle + (stdDev * stdDeviation)
	bb.Lower = bb.Middle - (stdDev * stdDeviation)

	return bb
}

type Stochastic struct {
	K float64
	D float64
}

func StochasticOscillator(prices []float64, period int) Stochastic {
	stoch := Stochastic{}
	if len(prices) < period {
		return stoch
	}

	highest := prices[len(prices)-1]
	lowest := prices[len(prices)-1]

	for i := len(prices) - period; i < len(prices); i++ {
		if prices[i] > highest {
			highest = prices[i]
		}
		if prices[i] < lowest {
			lowest = prices[i]
		}
	}

	current := prices[len(prices)-1]

	if highest-lowest == 0 {
		stoch.K = 50
	} else {
		stoch.K = ((current - lowest) / (highest - lowest)) * 100
	}

	stoch.D = stoch.K

	return stoch
}

func ADX(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	plusDM := 0.0
	minusDM := 0.0
	tr := 0.0

	for i := len(prices) - period; i < len(prices); i++ {
		high := prices[i]
		low := prices[i]
		prevClose := prices[i-1]

		trueRange := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		tr += trueRange

		upMove := high - prices[i-1]
		downMove := prices[i-1] - low

		if upMove > downMove && upMove > 0 {
			plusDM += upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM += downMove
		}
	}

	if tr == 0 {
		return 0
	}

	plusDI := (plusDM / tr) * 100
	minusDI := (minusDM / tr) * 100

	dx := math.Abs(plusDI-minusDI) / (plusDI + minusDI) * 100

	return dx
}

type SRLevels struct {
	Resistance float64
	Support    float64
}

func FindSupportResistance(prices []float64, lookback int) SRLevels {
	sr := SRLevels{}
	if len(prices) < lookback {
		lookback = len(prices)
	}

	recent := prices[len(prices)-lookback:]

	highest := recent[0]
	lowest := recent[0]

	for _, p := range recent {
		if p > highest {
			highest = p
		}
		if p < lowest {
			lowest = p
		}
	}

	sr.Resistance = highest
	sr.Support = lowest

	return sr
}

// ===== FIBONACCI RETRACEMENT & EXTENSION =====

type FibonacciLevels struct {
	High float64
	Low  float64

	// Retracement levels
	Fib236 float64
	Fib382 float64
	Fib500 float64
	Fib618 float64
	Fib786 float64

	// Extension levels
	Fib1272 float64
	Fib1618 float64
	Fib2000 float64
	Fib2618 float64

	Trend string
}

func CalculateFibonacci(prices []float64, lookback int) FibonacciLevels {
	fib := FibonacciLevels{}

	if len(prices) < lookback {
		lookback = len(prices)
	}

	recentPrices := prices[len(prices)-lookback:]

	high := recentPrices[0]
	low := recentPrices[0]
	highIdx := 0
	lowIdx := 0

	for i, p := range recentPrices {
		if p > high {
			high = p
			highIdx = i
		}
		if p < low {
			low = p
			lowIdx = i
		}
	}

	fib.High = high
	fib.Low = low

	if highIdx > lowIdx {
		fib.Trend = "BULLISH"
		diff := high - low

		fib.Fib236 = high - (diff * 0.236)
		fib.Fib382 = high - (diff * 0.382)
		fib.Fib500 = high - (diff * 0.500)
		fib.Fib618 = high - (diff * 0.618)
		fib.Fib786 = high - (diff * 0.786)

		fib.Fib1272 = high + (diff * 0.272)
		fib.Fib1618 = high + (diff * 0.618)
		fib.Fib2000 = high + (diff * 1.000)
		fib.Fib2618 = high + (diff * 1.618)

	} else {
		fib.Trend = "BEARISH"
		diff := high - low

		fib.Fib236 = low + (diff * 0.236)
		fib.Fib382 = low + (diff * 0.382)
		fib.Fib500 = low + (diff * 0.500)
		fib.Fib618 = low + (diff * 0.618)
		fib.Fib786 = low + (diff * 0.786)

		fib.Fib1272 = low - (diff * 0.272)
		fib.Fib1618 = low - (diff * 0.618)
		fib.Fib2000 = low - (diff * 1.000)
		fib.Fib2618 = low - (diff * 1.618)
	}

	return fib
}

func IsNearFibLevel(price, fibLevel, tolerance float64) bool {
	diff := math.Abs(price - fibLevel)
	return diff <= tolerance
}

func GetNearestFibRetracement(price float64, fib FibonacciLevels, tolerance float64) (string, float64) {
	levels := map[string]float64{
		"23.6%": fib.Fib236,
		"38.2%": fib.Fib382,
		"50.0%": fib.Fib500,
		"61.8%": fib.Fib618,
		"78.6%": fib.Fib786,
	}

	minDiff := math.MaxFloat64
	nearestLevel := ""
	nearestPrice := 0.0

	for name, level := range levels {
		diff := math.Abs(price - level)
		if diff < minDiff && diff <= tolerance {
			minDiff = diff
			nearestLevel = name
			nearestPrice = level
		}
	}

	return nearestLevel, nearestPrice
}

type PivotPoints struct {
	PP float64
	R1 float64
	R2 float64
	R3 float64
	S1 float64
	S2 float64
	S3 float64
}

func CalculatePivotPoints(prices []float64) PivotPoints {
	pp := PivotPoints{}

	if len(prices) < 24 {
		return pp
	}

	yesterday := prices[len(prices)-24 : len(prices)-1]

	high := yesterday[0]
	low := yesterday[0]
	close := yesterday[len(yesterday)-1]

	for _, p := range yesterday {
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
	}

	pp.PP = (high + low + close) / 3

	pp.R1 = (2 * pp.PP) - low
	pp.R2 = pp.PP + (high - low)
	pp.R3 = high + 2*(pp.PP-low)

	pp.S1 = (2 * pp.PP) - high
	pp.S2 = pp.PP - (high - low)
	pp.S3 = low - 2*(high-pp.PP)

	return pp
}
