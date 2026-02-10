package handlers

import (
	"email-reminder-system/models"
	"email-reminder-system/services"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

// SettingsHandler handles SMTP settings-related HTTP requests
type SettingsHandler struct {
	settingsService *services.SettingsService
	emailService    *services.EmailService
	templates       *template.Template
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(settingsService *services.SettingsService, emailService *services.EmailService, templates *template.Template) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		emailService:    emailService,
		templates:       templates,
	}
}

// SettingsPage displays the SMTP settings page (GET /settings)
func (h *SettingsHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get current SMTP settings
	settings, err := h.settingsService.GetSMTPSettings(user.ID)
	
	// Prepare template data
	data := map[string]interface{}{
		"Success": r.URL.Query().Get("success"),
		"Error":   r.URL.Query().Get("error"),
	}

	// If settings exist, add them to template data with masked password
	if err == nil && settings != nil {
		// Mask the password for display
		maskedSettings := *settings
		if len(maskedSettings.Password) > 0 {
			maskedSettings.Password = "********"
		}
		data["Settings"] = &maskedSettings
	}

	// Render settings template
	err = h.templates.ExecuteTemplate(w, "settings.html", data)
	if err != nil {
		http.Error(w, "Failed to render settings page", http.StatusInternalServerError)
		return
	}
}

// UpdateSettings processes SMTP settings update (POST /settings)
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/settings?error=invalid_form", http.StatusSeeOther)
		return
	}

	// Extract form values
	host := strings.TrimSpace(r.FormValue("host"))
	port := 0
	_, err = fmt.Sscanf(r.FormValue("port"), "%d", &port)
	if err != nil {
		port = 0
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	fromEmail := strings.TrimSpace(r.FormValue("from_email"))
	fromName := strings.TrimSpace(r.FormValue("from_name"))
	useTLS := r.FormValue("use_tls") == "on"

	// If password is masked (all asterisks), retrieve existing password
	if password == "********" || strings.TrimSpace(password) == "" {
		existingSettings, err := h.settingsService.GetSMTPSettings(user.ID)
		if err == nil && existingSettings != nil {
			password = existingSettings.Password
		}
	}

	// Create settings object
	settings := &models.SMTPSettings{
		UserID:    user.ID,
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		FromEmail: fromEmail,
		FromName:  fromName,
		UseTLS:    useTLS,
	}

	// Validate and save settings
	err = h.settingsService.UpsertSMTPSettings(settings)
	if err != nil {
		// Validation or save failed, redirect with error
		http.Redirect(w, r, "/settings?error="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}

	// Success, redirect with success message
	http.Redirect(w, r, "/settings?success=updated", http.StatusSeeOther)
}

// TestSMTPSettings tests SMTP connection (POST /settings/test)
func (h *SettingsHandler) TestSMTPSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/settings?error=invalid_form", http.StatusSeeOther)
		return
	}

	// Extract form values
	host := strings.TrimSpace(r.FormValue("host"))
	port := 0
	_, err = fmt.Sscanf(r.FormValue("port"), "%d", &port)
	if err != nil {
		port = 0
	}
	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	fromEmail := strings.TrimSpace(r.FormValue("from_email"))
	fromName := strings.TrimSpace(r.FormValue("from_name"))
	useTLS := r.FormValue("use_tls") == "on"

	// Debug log
	logger := services.GetLogger()
	logger.Info("Testing SMTP settings - Form values", map[string]interface{}{
		"host":         host,
		"port":         port,
		"username":     username,
		"from_email":   fromEmail,
		"use_tls":      useTLS,
		"has_password": password != "",
	})

	// If form values are empty or password is masked, load from database
	if host == "" || password == "********" || strings.TrimSpace(password) == "" {
		logger.Info("Loading SMTP settings from database", nil)
		existingSettings, err := h.settingsService.GetSMTPSettings(user.ID)
		if err != nil {
			http.Redirect(w, r, "/settings?error="+url.QueryEscape("Ayarlar bulunamadı: "+err.Error()), http.StatusSeeOther)
			return
		}
		
		// Use database values if form values are empty
		if host == "" {
			host = existingSettings.Host
		}
		if port == 0 {
			port = existingSettings.Port
		}
		if username == "" {
			username = existingSettings.Username
		}
		if password == "********" || strings.TrimSpace(password) == "" {
			password = existingSettings.Password
		}
		if fromEmail == "" {
			fromEmail = existingSettings.FromEmail
		}
		if fromName == "" {
			fromName = existingSettings.FromName
		}
		// useTLS from form is fine
	}

	// Create settings object
	settings := &models.SMTPSettings{
		UserID:    user.ID,
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		FromEmail: fromEmail,
		FromName:  fromName,
		UseTLS:    useTLS,
	}

	// Test SMTP connection
	err = h.emailService.TestSMTPConnection(settings)
	if err != nil {
		logger.Error("SMTP test failed", map[string]interface{}{
			"error": err.Error(),
			"host":  host,
			"port":  port,
		})
		http.Redirect(w, r, "/settings?error="+url.QueryEscape("Test başarısız: "+err.Error()), http.StatusSeeOther)
		return
	}

	logger.Info("SMTP test successful", map[string]interface{}{
		"host": host,
		"port": port,
	})

	// Success
	http.Redirect(w, r, "/settings?success=test_ok", http.StatusSeeOther)
}
