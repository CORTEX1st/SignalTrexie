package main

import (
	"fmt"
	"time"
)

type SessionInfo struct {
	Name       string
	IsActive   bool
	Volatility string // HIGH, MODERATE, LOW
	Description string
}

// Get location based on YOUR_TIMEZONE config
func GetUserLocation() *time.Location {
	loc, err := time.LoadLocation(YOUR_TIMEZONE)
	if err != nil {
		Error(fmt.Sprintf("Invalid timezone '%s', falling back to UTC", YOUR_TIMEZONE))
		return time.UTC
	}
	return loc
}

// Get current session based on UTC time
func GetCurrentSessionByUTC() SessionInfo {
	hour := time.Now().UTC().Hour()
	
	// Asia Session: 22:00 - 08:00 UTC (Tokyo: 00:00-09:00, Sydney: 22:00-07:00)
	if (hour >= 22 || hour < 8) {
		// Avoid dead hours: 02:00 - 04:00 UTC (11:00 - 13:00 WIB)
		if hour >= 2 && hour < 4 {
			return SessionInfo{
				Name:        "ASIA_DEAD_HOURS",
				IsActive:    false,
				Volatility:  "VERY_LOW",
				Description: "Asia dead hours (low liquidity)",
			}
		}
		return SessionInfo{
			Name:        "ASIA",
			IsActive:    true,
			Volatility:  "MODERATE",
			Description: "Tokyo/Sydney session",
		}
	}
	
	// London-NY Overlap: 13:00 - 17:00 UTC (20:00 - 00:00 WIB) - HIGHEST VOLATILITY
	if hour >= 13 && hour < 17 {
		return SessionInfo{
			Name:        "LONDON_NY_OVERLAP",
			IsActive:    true,
			Volatility:  "VERY_HIGH",
			Description: "London-NY overlap (best liquidity)",
		}
	}
	
	// London Session: 08:00 - 17:00 UTC (15:00 - 00:00 WIB)
	if hour >= 8 && hour < 17 {
		return SessionInfo{
			Name:        "LONDON",
			IsActive:    true,
			Volatility:  "HIGH",
			Description: "London session",
		}
	}
	
	// NY Session: 13:00 - 22:00 UTC (20:00 - 05:00 WIB)
	if hour >= 13 && hour < 22 {
		return SessionInfo{
			Name:        "NEW_YORK",
			IsActive:    true,
			Volatility:  "HIGH",
			Description: "New York session",
		}
	}
	
	return SessionInfo{
		Name:        "OFF_HOURS",
		IsActive:    false,
		Volatility:  "LOW",
		Description: "Market closed/low liquidity",
	}
}

// Main session detection - considers both user's SESSION config and current time
func GetCurrentSession() SessionInfo {
	utcSession := GetCurrentSessionByUTC()
	
	// If user forces a specific session, respect that
	if SESSION == "LONDON_NY" {
		if utcSession.Name == "LONDON" || utcSession.Name == "NEW_YORK" || utcSession.Name == "LONDON_NY_OVERLAP" {
			return utcSession
		}
		return SessionInfo{
			Name:        "WAITING_LONDON_NY",
			IsActive:    false,
			Volatility:  "LOW",
			Description: "Waiting for London-NY session",
		}
	}
	
	if SESSION == "ASIA" {
		if utcSession.Name == "ASIA" {
			return utcSession
		}
		if utcSession.Name == "ASIA_DEAD_HOURS" {
			return utcSession
		}
		return SessionInfo{
			Name:        "WAITING_ASIA",
			IsActive:    false,
			Volatility:  "LOW",
			Description: "Waiting for Asia session",
		}
	}
	
	if SESSION == "ALL" {
		// Trade all sessions except dead hours
		if utcSession.Name == "ASIA_DEAD_HOURS" {
			return utcSession
		}
		if utcSession.IsActive {
			return utcSession
		}
	}
	
	// AUTO mode (default) - automatically pick best session
	if SESSION == "AUTO" {
		return utcSession
	}
	
	return utcSession
}

// Check if we should trade now
func IsTradingSession() bool {
	session := GetCurrentSession()
	return session.IsActive
}

// Check if current session has high volatility
func IsHighVolatilityPeriod() bool {
	session := GetCurrentSession()
	return session.Volatility == "HIGH" || session.Volatility == "VERY_HIGH"
}

// Check if currently in Asia session
func IsAsiaSession() bool {
	session := GetCurrentSession()
	return session.Name == "ASIA"
}

// Get current session name
func GetSessionName() string {
	return GetCurrentSession().Name
}

// Get user's local time for display
func GetUserLocalTime() string {
	loc := GetUserLocation()
	return time.Now().In(loc).Format("2006-01-02 15:04:05 MST")
}

// Get session info in user's timezone for display
func GetSessionScheduleInfo() string {
	loc := GetUserLocation()
	
	// Calculate session times in user's timezone
	// Asia: 22:00-08:00 UTC
	asiaStart := time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC).In(loc)
	asiaEnd := time.Date(2024, 1, 2, 8, 0, 0, 0, time.UTC).In(loc)
	
	// London-NY: 08:00-22:00 UTC  
	londonStart := time.Date(2024, 1, 1, 8, 0, 0, 0, time.UTC).In(loc)
	londonEnd := time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC).In(loc)
	
	// Overlap: 13:00-17:00 UTC
	overlapStart := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC).In(loc)
	overlapEnd := time.Date(2024, 1, 1, 17, 0, 0, 0, time.UTC).In(loc)
	
	return fmt.Sprintf(
		"📅 SESSION SCHEDULE (Your Time - %s):\n"+
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
		"🌏 ASIA Session:\n"+
		"   %s - %s\n"+
		"   Strategy: Range trading, BB bounces\n"+
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
		"🌍 LONDON-NY Session:\n"+
		"   %s - %s\n"+
		"   Strategy: Trend following, breakouts\n"+
		"━━━━━━━━━━━━━━━━━━━━━━\n"+
		"🔥 BEST TIME (Overlap):\n"+
		"   %s - %s\n"+
		"   Highest volatility & liquidity!",
		loc.String(),
		asiaStart.Format("15:04"),
		asiaEnd.Format("15:04"),
		londonStart.Format("15:04"),
		londonEnd.Format("15:04"),
		overlapStart.Format("15:04"),
		overlapEnd.Format("15:04"),
	)
}

// Get detailed current status
func GetCurrentStatusInfo() string {
	session := GetCurrentSession()
	localTime := GetUserLocalTime()
	utcTime := time.Now().UTC().Format("2006-01-02 15:04:05 MST")
	
	statusEmoji := "🟢"
	if !session.IsActive {
		statusEmoji = "🔴"
	}
	
	sessionEmoji := "🌐"
	if session.Name == "ASIA" || session.Name == "ASIA_DEAD_HOURS" {
		sessionEmoji = "🌏"
	} else if session.Name == "LONDON" || session.Name == "NEW_YORK" || session.Name == "LONDON_NY_OVERLAP" {
		sessionEmoji = "🌍"
	}
	
	return fmt.Sprintf(
		"%s Status: %s\n"+
		"%s Session: %s\n"+
		"📊 Volatility: %s\n"+
		"🕐 Your Time: %s\n"+
		"🌍 UTC Time: %s",
		statusEmoji,
		map[bool]string{true: "ACTIVE", false: "WAITING"}[session.IsActive],
		sessionEmoji,
		session.Description,
		session.Volatility,
		localTime,
		utcTime,
	)
}