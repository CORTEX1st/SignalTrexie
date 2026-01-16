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

// MACD with Signal Line and Histogram
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

func ATR(prices []float64, period int) float64 {
	if len(prices) < period+1 {
		return 0
	}

	var sum float64
	for i := len(prices) - period; i < len(prices); i++ {
		sum += math.Abs(prices[i] - prices[i-1])
	}
	return sum / float64(period)
}

// Bollinger Bands
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

	// Calculate standard deviation
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

// Stochastic Oscillator
type Stochastic struct {
	K float64
	D float64
}

func StochasticOscillator(prices []float64, period int) Stochastic {
	stoch := Stochastic{}
	if len(prices) < period {
		return stoch
	}

	// Find highest high and lowest low in period
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

	// Simplified %D as 3-period SMA of %K (would need to store previous K values for accuracy)
	stoch.D = stoch.K

	return stoch
}

// ADX - Average Directional Index (Trend Strength)
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

// Support and Resistance Levels
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
