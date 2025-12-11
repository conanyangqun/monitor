package logger

import (
	"log"
	"time"
)

// CleanExpiredEntries removes entries older than maxDays from the history entry list.
func (h History) CleanExpiredEntries(maxDays int) History {
	if maxDays <= 0 {
		log.Println("Max days for cleaning history is not set or invalid, skipping cleaning.")
		return h // No cleaning needed if maxDays is not set
	}

	cutoffTime := time.Now().AddDate(0, 0, -maxDays)
	var cleanedHistory History

	for _, entry := range h {
		entryTime, err := time.Parse(time.RFC3339, entry.Time)
		if err != nil {
			log.Printf("Error parsing time %s: %v", entry.Time, err)
			continue // Skip entries with invalid time format
		}
		if entryTime.After(cutoffTime) {
			cleanedHistory = append(cleanedHistory, entry)
		}
	}

	return cleanedHistory
}

// AddEntry adds a new entry to the history entry list.
func (h History) AddEntry(entry HistoryEntry) History {
	newHistory := append(h, entry)
	return newHistory
}
