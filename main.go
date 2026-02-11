package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"email-reminder-system/config"
	"email-reminder-system/database"
	"email-reminder-system/handlers"
	"email-reminder-system/services"
)

func main() {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			logger := services.GetLogger()
			logger.Error("Application panic recovered", map[string]interface{}{
				"panic": fmt.Sprintf("%v", r),
			})
			os.Exit(1)
		}
	}()

	// Initialize logger
	logger := services.GetLogger()
	defer logger.Close()

	// Load configuration from environment variables
	cfg := config.Load()

	// Set encryption key if provided in config
	if cfg.EncryptionKey != "" {
		os.Setenv("ENCRYPTION_KEY", cfg.EncryptionKey)
	} else {
		logger.Warn("ENCRYPTION_KEY not set, using default key (not secure for production)")
		// Generate a base64-encoded 32-byte key for development
		// This is exactly 32 bytes when decoded: "12345678901234567890123456789012"
		defaultKey := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI="
		os.Setenv("ENCRYPTION_KEY", defaultKey)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.DatabasePath), 0755); err != nil {
		logger.Error("Failed to create data directory", map[string]interface{}{
			"path":  filepath.Dir(cfg.DatabasePath),
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Initialize database connection
	db, err := database.InitDB(cfg.DatabasePath)
	if err != nil {
		logger.Error("Failed to initialize database", map[string]interface{}{
			"path":  cfg.DatabasePath,
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer db.Close()

	logger.Info("Database initialized successfully", map[string]interface{}{
		"path": cfg.DatabasePath,
	})

	// Initialize repository
	repo := database.NewSQLiteRepository(db)

	// Initialize services
	authService := services.NewAuthService(repo, cfg.SessionDuration)
	reminderService := services.NewReminderService(repo)
	encryptionService, err := services.NewEncryptionService()
	if err != nil {
		logger.Error("Failed to initialize encryption service", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	settingsService := services.NewSettingsService(repo, encryptionService)
	emailHistoryService := services.NewEmailHistoryService(repo, logger)
	emailService := services.NewEmailService(emailHistoryService)

	// Initialize scheduler service
	schedulerService := services.NewSchedulerService(repo, emailService, reminderService, encryptionService)

	// Load templates with custom functions
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
	}
	templates, err := template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		logger.Error("Failed to load templates", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	// Initialize handlers with dependency injection
	authHandler := handlers.NewAuthHandler(authService, templates)
	reminderHandler := handlers.NewReminderHandler(reminderService, templates)
	settingsHandler := handlers.NewSettingsHandler(settingsService, emailService, templates)
	emailHistoryHandler := handlers.NewEmailHistoryHandler(emailHistoryService, templates)
	groupHandler := handlers.NewRecipientGroupHandler(repo, templates)
	authMiddleware := handlers.NewAuthMiddleware(authService)

	// Set up HTTP routes with authentication middleware
	// Public routes
	http.HandleFunc("/login", handlers.RecoverMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			authHandler.LoginPage(w, r)
		} else if r.Method == http.MethodPost {
			authHandler.Login(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))
	http.HandleFunc("/logout", handlers.RecoverMiddleware(authHandler.Logout))
	http.HandleFunc("/health", handlers.RecoverMiddleware(healthHandler))

	// Protected routes - Reminder list
	http.HandleFunc("/", handlers.RecoverMiddleware(authMiddleware.RequireAuth(reminderHandler.ListReminders)))

	// Protected routes - Reminder forms
	http.HandleFunc("/reminders/new", handlers.RecoverMiddleware(authMiddleware.RequireAuth(reminderHandler.NewReminderForm)))

	// Protected routes - Reminder CRUD operations
	http.HandleFunc("/reminders", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			reminderHandler.CreateReminder(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Handle reminder edit, update, and delete with path-based routing
	http.HandleFunc("/reminders/", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/reminders/") {
			// Check if it's an edit form request
			if len(path) > 5 && path[len(path)-5:] == "/edit" {
				reminderHandler.EditReminderForm(w, r)
			} else if len(path) > 7 && path[len(path)-7:] == "/delete" {
				reminderHandler.DeleteReminder(w, r)
			} else if len(path) > 7 && path[len(path)-7:] == "/toggle" {
				reminderHandler.ToggleReminder(w, r)
			} else {
				// It's an update request (POST /reminders/:id)
				if r.Method == http.MethodPost {
					reminderHandler.UpdateReminder(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})))

	// Protected routes - Settings
	http.HandleFunc("/settings", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			settingsHandler.SettingsPage(w, r)
		} else if r.Method == http.MethodPost {
			settingsHandler.UpdateSettings(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Protected routes - Test SMTP settings
	http.HandleFunc("/settings/test", handlers.RecoverMiddleware(authMiddleware.RequireAuth(settingsHandler.TestSMTPSettings)))

	// Protected routes - Email History
	http.HandleFunc("/email-history", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			emailHistoryHandler.ListEmailHistory(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Protected routes - Email History Detail
	http.HandleFunc("/email-history/", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			emailHistoryHandler.GetEmailLogDetail(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Protected routes - Recipient Groups
	http.HandleFunc("/groups", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			groupHandler.ListGroups(w, r)
		} else if r.Method == http.MethodPost {
			groupHandler.CreateGroup(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.HandleFunc("/groups/new", handlers.RecoverMiddleware(authMiddleware.RequireAuth(groupHandler.NewGroupForm)))
	http.HandleFunc("/groups/export", handlers.RecoverMiddleware(authMiddleware.RequireAuth(groupHandler.ExportCSV)))
	http.HandleFunc("/groups/import", handlers.RecoverMiddleware(authMiddleware.RequireAuth(groupHandler.ImportCSV)))

	http.HandleFunc("/groups/", handlers.RecoverMiddleware(authMiddleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if len(path) > len("/groups/") {
			if len(path) > 5 && path[len(path)-5:] == "/edit" {
				groupHandler.EditGroupForm(w, r)
			} else if len(path) > 7 && path[len(path)-7:] == "/delete" {
				groupHandler.DeleteGroup(w, r)
			} else {
				if r.Method == http.MethodPost {
					groupHandler.UpdateGroup(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})))

	// Protected routes - API for AJAX
	http.HandleFunc("/api/groups", handlers.RecoverMiddleware(authMiddleware.RequireAuth(groupHandler.GetGroupsJSON)))

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start scheduler in background goroutine
	if err := schedulerService.Start(); err != nil {
		logger.Error("Failed to start scheduler service", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	defer schedulerService.Stop()

	logger.Info("Scheduler service started successfully", nil)

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start HTTP server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Starting Email Reminder System", map[string]interface{}{
			"port":             cfg.ServerPort,
			"database":         cfg.DatabasePath,
			"session_duration": cfg.SessionDuration.String(),
		})

		serverErrors <- http.ListenAndServe(":"+cfg.ServerPort, nil)
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		logger.Error("Failed to start server", map[string]interface{}{
			"port":  cfg.ServerPort,
			"error": err.Error(),
		})
		os.Exit(1)
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", map[string]interface{}{
			"signal": sig.String(),
		})

		// Stop scheduler gracefully
		if err := schedulerService.Stop(); err != nil {
			logger.Error("Error stopping scheduler", map[string]interface{}{
				"error": err.Error(),
			})
		}

		logger.Info("Application shutdown complete", nil)
	}
}

// healthHandler provides a health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
