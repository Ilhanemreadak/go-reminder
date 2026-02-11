package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"email-reminder-system/config"
	"email-reminder-system/database"
	"email-reminder-system/models"
	"email-reminder-system/services"

	"golang.org/x/term"
)

func main() {
	// Define command-line flags
	username := flag.String("username", "", "Username for the new user")
	password := flag.String("password", "", "Password for the new user (not recommended, use interactive mode)")
	dbPath := flag.String("db", "", "Path to database file (default: from config)")
	interactive := flag.Bool("i", false, "Interactive mode (prompts for username and password)")

	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Use custom database path if provided
	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	}

	// Initialize database
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize repository and auth service
	repo := database.NewSQLiteRepository(db)
	authService := services.NewAuthService(repo, cfg.SessionDuration)

	// Get username and password
	var user, pass string

	if *interactive || (*username == "" && *password == "") {
		// Interactive mode
		user, pass, err = promptForCredentials()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Command-line mode
		if *username == "" {
			fmt.Fprintln(os.Stderr, "Error: Username is required. Use -username flag or -i for interactive mode.")
			flag.Usage()
			os.Exit(1)
		}

		user = *username

		if *password == "" {
			// Prompt for password if not provided
			fmt.Print("Enter password: ")
			passBytes, err := term.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading password: %v\n", err)
				os.Exit(1)
			}
			pass = string(passBytes)
		} else {
			pass = *password
			fmt.Println("Warning: Passing password via command-line flag is insecure. Consider using interactive mode.")
		}
	}

	// Validate inputs
	if user == "" {
		fmt.Fprintln(os.Stderr, "Error: Username cannot be empty")
		os.Exit(1)
	}

	if pass == "" {
		fmt.Fprintln(os.Stderr, "Error: Password cannot be empty")
		os.Exit(1)
	}

	// Check if user already exists
	existingUser, _ := repo.GetUserByUsername(user)
	if existingUser != nil {
		fmt.Fprintf(os.Stderr, "Error: User '%s' already exists\n", user)
		os.Exit(1)
	}

	// Hash password
	passwordHash, err := authService.HashPassword(pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to hash password: %v\n", err)
		os.Exit(1)
	}

	// Create user
	newUser := &models.User{
		Username:     user,
		PasswordHash: passwordHash,
	}

	err = repo.CreateUser(newUser)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to create user: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ User '%s' created successfully (ID: %d)\n", newUser.Username, newUser.ID)
	fmt.Println("\nYou can now log in to RemindMe with these credentials.")
}

// promptForCredentials prompts the user for username and password interactively
func promptForCredentials() (string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Prompt for username
	fmt.Print("Enter username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read username: %w", err)
	}
	username = strings.TrimSpace(username)

	// Prompt for password
	fmt.Print("Enter password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", "", fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)

	// Confirm password
	fmt.Print("Confirm password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", "", fmt.Errorf("failed to read password confirmation: %w", err)
	}
	confirm := string(confirmBytes)

	if password != confirm {
		return "", "", fmt.Errorf("passwords do not match")
	}

	return username, password, nil
}
