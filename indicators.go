package main

import "math"

func EMA(data []float64, period int) []float64 {
	k := 2.0 / float64(period+1)
	ema := make([]float64, len(data))
	ema[0] = data[0]

	for i := 1; i < len(data); i++ {
		ema[i] = data[i]*k + ema[i-1]*(1-k)
	}
	return ema
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

func MACD(data []float64) float64 {
	if len(data) < 26 {
		return 0
	}
	ema12 := EMA(data, 12)
	ema26 := EMA(data, 26)
	return ema12[len(data)-1] - ema26[len(data)-1]
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
