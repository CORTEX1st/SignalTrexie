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
	Info("XAUUSD PRO SIGNAL BOT STARTED")
	Info("Mode      : " + MODE)
	Info("Session   : " + SESSION)
	Info("Timezone  : " + YOUR_TIMEZONE)
	Info(fmt.Sprintf("Polling   : %ds", POLLING_SECONDS))
	Info("Status    : ONLINE")

	// Get current session info
	currentSession := GetCurrentSession()

	// Get session emoji
	sessionEmoji := "🌐"
	sessionMode := SESSION
	if SESSION == "AUTO" {
		sessionMode = "AUTO (Smart Detection)"
		if currentSession.Name == "ASIA" || currentSession.Name == "ASIA_DEAD_HOURS" {
			sessionEmoji = "🌏"
		} else if currentSession.Name == "LONDON" || currentSession.Name == "NEW_YORK" || currentSession.Name == "LONDON_NY_OVERLAP" {
			sessionEmoji = "🌍"
		}
	} else if SESSION == "ASIA" {
		sessionEmoji = "🌏"
		sessionMode = "Asia (Tokyo/Sydney)"
	} else if SESSION == "LONDON_NY" {
		sessionEmoji = "🌍"
		sessionMode = "London-New York"
	} else if SESSION == "ALL" {
		sessionEmoji = "🌐"
		sessionMode = "24/7 Global"
	}

	// ===== TELEGRAM START NOTIF =====
	startMsg := fmt.Sprintf(
		"🟢 XAUUSD PRO SIGNAL BOT ONLINE\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"📊 Mode     : %s\n"+
			"%s Session  : %s\n"+
			"⏱ Polling  : %ds\n"+
			"🌍 Timezone : %s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"%s\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"✨ Multi-Session Strategy Active\n"+
			"📈 EMA • RSI • MACD • BB • ADX\n"+
			"🎯 Dynamic SL/TP Based on ATR\n"+
			"🤖 Smart Session Detection\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"💡 Asia: Range-focused, tighter stops\n"+
			"💡 London-NY: Trend-focused, wider stops\n"+
			"💡 Auto: Best strategy per session\n"+
			"━━━━━━━━━━━━━━━━━━━━━━\n"+
			"%s",
		MODE,
		sessionEmoji,
		sessionMode,
		POLLING_SECONDS,
		YOUR_TIMEZONE,
		GetCurrentStatusInfo(),
		GetSessionScheduleInfo(),
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
				currentSession := GetCurrentSession()
				Info(fmt.Sprintf("Outside trading hours - %s (Volatility: %s)",
					currentSession.Description, currentSession.Volatility))

				// Send periodic status update every hour during off-hours
				// (optional - comment out if you don't want hourly updates)
				/*
					if time.Now().Minute() == 0 {
						statusMsg := fmt.Sprintf(
							"⏸ WAITING FOR TRADING SESSION\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"%s",
							GetCurrentStatusInfo(),
						)
						wg.Add(1)
						go func() {
							defer wg.Done()
							SendTelegram(statusMsg)
						}()
					}
				*/
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

					// Session emoji
					sessionIcon := "🌍"
					if IsAsiaSession() {
						sessionIcon = "🌏"
					}

					reasonsText := strings.Join(sig.Reasons, "\n")

					// Calculate pip distance for SL and TPs
					slPips := (sig.Entry - sig.StopLoss)
					if sig.Action == "SELL" {
						slPips = (sig.StopLoss - sig.Entry)
					}

					message := fmt.Sprintf(
						"%s XAUUSD %s SIGNAL\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"💰 Entry Price : %.2f\n"+
							"🛑 Stop Loss   : %.2f (%.2f pips)\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"🎯 TP1         : %.2f\n"+
							"🎯 TP2         : %.2f\n"+
							"🎯 TP3         : %.2f\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"📊 Risk/Reward : 1:%.2f\n"+
							"💪 Confidence  : %d%%\n"+
							"⚙️ Mode        : %s\n"+
							"%s Session     : %s\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"📌 CONFIRMATIONS:\n%s\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"💡 Trade Management:\n"+
							"• Close 50%% at TP1\n"+
							"• Close 30%% at TP2\n"+
							"• Close 20%% at TP3\n"+
							"• Always use SL!\n"+
							"━━━━━━━━━━━━━━━━━━━━━━\n"+
							"🕐 Your Time: %s\n"+
							"🌍 UTC Time : %s",
						emoji,
						sig.Action,
						sig.Entry,
						sig.StopLoss,
						slPips,
						sig.TakeProfit1,
						sig.TakeProfit2,
						sig.TakeProfit3,
						sig.RiskReward,
						sig.Confidence,
						MODE,
						sessionIcon,
						sig.Session,
						reasonsText,
						GetUserLocalTime(),
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
