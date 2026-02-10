package main

import (
	"fmt"
	"log"
	"time"

	"email-reminder-system/config"
	"email-reminder-system/database"
)

func main() {
	cfg := config.Load()
	
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	repo := database.NewSQLiteRepository(db)
	
	now := time.Now()
	fmt.Println("=== Reminders Debug ===")
	fmt.Printf("Current time: %s\n\n", now.Format("2006-01-02 15:04:05"))

	// Get all reminders
	rows, err := db.Query(`
		SELECT id, title, next_send_at, is_active, schedule_type, time_of_day 
		FROM reminders
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var title string
		var nextSendAt time.Time
		var isActive bool
		var scheduleType string
		var timeOfDay string

		err := rows.Scan(&id, &title, &nextSendAt, &isActive, &scheduleType, &timeOfDay)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("ID: %d\n", id)
		fmt.Printf("Title: %s\n", title)
		fmt.Printf("Schedule Type: %s\n", scheduleType)
		fmt.Printf("Time of Day: %s\n", timeOfDay)
		fmt.Printf("Next Send At: %s\n", nextSendAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Is Active: %v\n", isActive)
		
		// Check if it's due
		if nextSendAt.Before(now) || nextSendAt.Equal(now) {
			fmt.Printf("STATUS: ✓ DUE (should be sent)\n")
		} else {
			diff := nextSendAt.Sub(now)
			fmt.Printf("STATUS: ✗ NOT DUE (in %v)\n", diff.Round(time.Second))
		}
		fmt.Println("---")
	}
	
	// Test GetDueReminders
	fmt.Println("\n=== Testing GetDueReminders ===")
	dueReminders, err := repo.GetDueReminders(now)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %d due reminders\n", len(dueReminders))
	for _, r := range dueReminders {
		fmt.Printf("- ID %d: %s (next: %s)\n", r.ID, r.Title, r.NextSendAt.Format("2006-01-02 15:04:05"))
	}
}
