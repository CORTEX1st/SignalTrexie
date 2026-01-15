package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func FetchXAUUSD() (float64, error) {
	url := fmt.Sprintf(
		"https://api.twelvedata.com/price?symbol=XAU/USD&apikey=%s",
		API_KEY,
	)

	// Add timeout to prevent hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		Error("Market API request failed: " + err.Error())
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Error("Failed to read market response: " + err.Error())
		return 0, err
	}

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf("http status not 200 (got %d): %s", resp.StatusCode, string(body))
		Error("Market API error: " + errMsg)
		return 0, errors.New(errMsg)
	}

	var data struct {
		Price   string `json:"price"`
		Status  string `json:"status"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		Error("Failed to parse market response: " + err.Error())
		return 0, err
	}

	if data.Price == "" {
		Error("Market API error: price empty - " + string(body))
		return 0, errors.New("price empty, api error")
	}

	price, err := strconv.ParseFloat(data.Price, 64)
	if err != nil {
		Error("Failed to parse price value: " + err.Error())
		return 0, err
	}

	Info(fmt.Sprintf("Price fetched: %.2f", price))
	return price, nil
}
