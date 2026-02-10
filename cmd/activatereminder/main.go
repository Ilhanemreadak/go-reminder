package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"email-reminder-system/config"
	"email-reminder-system/database"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/activatereminder/main.go <reminder_id>")
		os.Exit(1)
	}

	reminderID, err := strconv.ParseInt(os.Args[1], 10, 64)
	if err != nil {
		log.Fatal("Invalid reminder ID:", err)
	}

	cfg := config.Load()
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Activate the reminder
	_, err = db.Exec("UPDATE reminders SET is_active = 1 WHERE id = ?", reminderID)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Reminder %d activated successfully!\n", reminderID)
}
