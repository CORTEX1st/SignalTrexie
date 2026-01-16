package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			Error(fmt.Sprintf("PANIC RECOVERED: %v", r))
			os.Exit(1)
		}
	}()

	var wg sync.WaitGroup

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		Info("Shutdown signal received")
		wg.Wait()
		os.Exit(0)
	}()
	// ===== START LOG =====
	Info("XAUUSD SIGNAL BOT STARTED")
	Info("Mode      : " + MODE)
	Info(fmt.Sprintf("Polling   : %ds", POLLING_SECONDS))
	Info("Session   : London—New York")
	Info("Status    : ONLINE")

	// ===== TELEGRAM START NOTIF =====
	startMsg := fmt.Sprintf(
		"🟢 XAUUSD PRO SIGNAL BOT ONLINE\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📊 Mode    : %s\n"+
			"⏱ Polling : %ds\n"+
			"🌍 Session : London—New York\n"+
			"🕐 Time    : %s UTC\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"✨ Multi-Indicator Strategy Active\n"+
			"📈 EMA • RSI • MACD • BB • ADX\n"+
			"🎯 Dynamic SL/TP Based on ATR",
		MODE,
		POLLING_SECONDS,
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	)

	Info("Sending start notification...")
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			if r := recover(); r != nil {
				Error(fmt.Sprintf("SendTelegram panic: %v", r))
			}
		}()
		SendTelegram(startMsg)
	}()

	// Give telegram time to send
	time.Sleep(2 * time.Second)

	// ===== ENGINE =====
	ticker := time.NewTicker(time.Duration(POLLING_SECONDS) * time.Second)
	defer ticker.Stop()

	var prices []float64
	lastSignal := ""

	maxBuffer := 120
	if MODE == "LONG" {
		maxBuffer = 300
	}

	for range ticker.C {
		func() {
			defer func() {
				if r := recover(); r != nil {
					Error(fmt.Sprintf("Loop panic: %v", r))
				}
			}()

			if !IsTradingSession() {
				Info("Outside trading session, waiting...")
				return
			}

			price, err := FetchXAUUSD()
			if err != nil {
				Error("Fetch price failed: " + err.Error())
				return
			}

			prices = append(prices, price)
			if len(prices) > maxBuffer {
				prices = prices[len(prices)-maxBuffer:]
			}

			signal := GenerateSignalAdvanced(prices)
			
			if signal.Action != "WAIT" && signal.Action != lastSignal {
				wg.Add(1)
				go func(sig SignalData) {
					defer func() {
						wg.Done()
						if r := recover(); r != nil {
							Error(fmt.Sprintf("SendTelegram signal panic: %v", r))
						}
					}()

					// Format professional signal message
					var emoji string
					if sig.Action == "BUY" {
						emoji = "🟢📈"
					} else {
						emoji = "🔴📉"
					}

					reasonsText := strings.Join(sig.Reasons, "\n")

					message := fmt.Sprintf(
						"%s XAUUSD %s SIGNAL\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"💰 Entry Price : %.2f\n"+
							"🛑 Stop Loss   : %.2f\n"+
							"🎯 TP1         : %.2f\n"+
							"🎯 TP2         : %.2f\n"+
							"🎯 TP3         : %.2f\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"📊 Risk/Reward : 1:%.2f\n"+
							"💪 Confidence  : %d%%\n"+
							"⚙️ Mode        : %s\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"📌 CONFIRMATIONS:\n%s\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"⏰ %s UTC",
						emoji,
						sig.Action,
						sig.Entry,
						sig.StopLoss,
						sig.TakeProfit1,
						sig.TakeProfit2,
						sig.TakeProfit3,
						sig.RiskReward,
						sig.Confidence,
						MODE,
						reasonsText,
						time.Now().UTC().Format("15:04:05"),
					)

					SendTelegram(message)
				}(signal)
				lastSignal = signal.Action
			}
		}()
	}

	Info("Engine started, waiting for signals...")
	select {}
}