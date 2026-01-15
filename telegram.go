package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func SendTelegram(msg string) {
	escaped := url.QueryEscape(msg)

	apiURL := fmt.Sprintf(
		"https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s",
		TELEGRAM_TOKEN,
		TELEGRAM_CHAT,
		escaped,
	)

	// Create HTTP client with timeout to prevent hanging
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var lastErr error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.Get(apiURL)
		if err != nil {
			lastErr = err
			Error(fmt.Sprintf("Telegram request failed (attempt %d/%d): %v", attempt, maxRetries, err))
			if attempt < maxRetries {
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			Error("Failed to read Telegram response: " + err.Error())
			return
		}

		if resp.StatusCode != 200 {
			Error(fmt.Sprintf("Telegram API error (status %d): %s", resp.StatusCode, string(body)))
			return
		}

		Info("Telegram sent successfully")
		return
	}

	Error(fmt.Sprintf("Telegram failed after %d attempts: %v", maxRetries, lastErr))
}
