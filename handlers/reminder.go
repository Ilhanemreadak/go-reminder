package handlers

import (
	"email-reminder-system/models"
	"email-reminder-system/services"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

// ReminderHandler handles reminder-related HTTP requests
type ReminderHandler struct {
	reminderService services.ReminderService
	templates       *template.Template
}

// NewReminderHandler creates a new reminder handler
func NewReminderHandler(reminderService services.ReminderService, templates *template.Template) *ReminderHandler {
	return &ReminderHandler{
		reminderService: reminderService,
		templates:       templates,
	}
}

// ListReminders displays the reminder list page (GET /)
func (h *ReminderHandler) ListReminders(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all reminders for the user
	reminders, err := h.reminderService.ListReminders(user.ID)
	if err != nil {
		http.Error(w, "Failed to load reminders", http.StatusInternalServerError)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "reminders",
		"Reminders":       reminders,
		"Success":         r.URL.Query().Get("success"),
	}

	// Render template
	err = h.templates.ExecuteTemplate(w, "reminders.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// NewReminderForm displays the reminder creation form (GET /reminders/new)
func (h *ReminderHandler) NewReminderForm(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "reminders",
		"Reminder":        &models.Reminder{}, // Empty reminder for new form
		"IsEdit":          false,
	}

	// Render template
	err := h.templates.ExecuteTemplate(w, "reminder_form.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// EditReminderForm displays the reminder edit form (GET /reminders/:id/edit)
func (h *ReminderHandler) EditReminderForm(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract reminder ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	// Get reminder
	reminder, err := h.reminderService.GetReminder(id, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Reminder not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to load reminder", http.StatusInternalServerError)
		}
		return
	}

	// Prepare template data
	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "reminders",
		"Reminder":        reminder,
		"IsEdit":          true,
	}

	// Render template
	err = h.templates.ExecuteTemplate(w, "reminder_form.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// CreateReminder handles reminder creation (POST /reminders)
func (h *ReminderHandler) CreateReminder(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Parse recipients (comma-separated emails)
	recipientsStr := r.FormValue("recipients")
	recipients := []string{}
	for _, email := range strings.Split(recipientsStr, ",") {
		trimmed := strings.TrimSpace(email)
		if trimmed != "" {
			recipients = append(recipients, trimmed)
		}
	}

	// Parse schedule-specific fields
	dayOfWeek, _ := strconv.Atoi(r.FormValue("day_of_week"))
	dayOfMonth, _ := strconv.Atoi(r.FormValue("day_of_month"))
	intervalDays, _ := strconv.Atoi(r.FormValue("interval_days"))

	// Create reminder object
	reminder := &models.Reminder{
		Title:        r.FormValue("title"),
		Recipients:   recipients,
		EmailContent: r.FormValue("email_content"),
		ScheduleType: r.FormValue("schedule_type"),
		ScheduleDate: r.FormValue("schedule_date"),
		IntervalDays: intervalDays,
		DayOfWeek:    dayOfWeek,
		DayOfMonth:   dayOfMonth,
		TimeOfDay:    r.FormValue("time_of_day"),
	}

	// Create reminder
	err = h.reminderService.CreateReminder(user.ID, reminder)
	if err != nil {
		// Display validation errors on form
		data := map[string]interface{}{
			"User":            user,
			"IsAuthenticated": true,
			"ActivePage":      "reminders",
			"Reminder":        reminder,
			"IsEdit":          false,
			"Error":           err.Error(),
		}
		h.templates.ExecuteTemplate(w, "reminder_form.html", data)
		return
	}

	// Redirect to list with success message
	http.Redirect(w, r, "/?success=created", http.StatusSeeOther)
}

// UpdateReminder handles reminder updates (POST /reminders/:id)
func (h *ReminderHandler) UpdateReminder(w http.ResponseWriter, r *http.Request) {
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

	// Extract reminder ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	// Parse form data
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Parse recipients (comma-separated emails)
	recipientsStr := r.FormValue("recipients")
	recipients := []string{}
	for _, email := range strings.Split(recipientsStr, ",") {
		trimmed := strings.TrimSpace(email)
		if trimmed != "" {
			recipients = append(recipients, trimmed)
		}
	}

	// Parse schedule-specific fields
	dayOfWeek, _ := strconv.Atoi(r.FormValue("day_of_week"))
	dayOfMonth, _ := strconv.Atoi(r.FormValue("day_of_month"))
	intervalDays, _ := strconv.Atoi(r.FormValue("interval_days"))

	// Create reminder object with ID
	reminder := &models.Reminder{
		ID:           id,
		Title:        r.FormValue("title"),
		Recipients:   recipients,
		EmailContent: r.FormValue("email_content"),
		ScheduleType: r.FormValue("schedule_type"),
		ScheduleDate: r.FormValue("schedule_date"),
		IntervalDays: intervalDays,
		DayOfWeek:    dayOfWeek,
		DayOfMonth:   dayOfMonth,
		TimeOfDay:    r.FormValue("time_of_day"),
	}

	// Update reminder
	err = h.reminderService.UpdateReminder(reminder, user.ID)
	if err != nil {
		// Display validation errors on form
		data := map[string]interface{}{
			"User":            user,
			"IsAuthenticated": true,
			"ActivePage":      "reminders",
			"Reminder":        reminder,
			"IsEdit":          true,
			"Error":           err.Error(),
		}
		h.templates.ExecuteTemplate(w, "reminder_form.html", data)
		return
	}

	// Redirect to list with success message
	http.Redirect(w, r, "/?success=updated", http.StatusSeeOther)
}

// DeleteReminder handles reminder deletion (POST /reminders/:id/delete)
func (h *ReminderHandler) DeleteReminder(w http.ResponseWriter, r *http.Request) {
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

	// Extract reminder ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	// Delete reminder
	err = h.reminderService.DeleteReminder(id, user.ID)
	if err != nil {
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusForbidden)
		} else if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Reminder not found", http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete reminder: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Redirect to list with success message
	http.Redirect(w, r, "/?success=deleted", http.StatusSeeOther)
}

// ToggleReminder toggles reminder active status (POST /reminders/:id/toggle)
func (h *ReminderHandler) ToggleReminder(w http.ResponseWriter, r *http.Request) {
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

	// Extract reminder ID from URL path
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/reminders/")
	idStr = strings.TrimSuffix(idStr, "/toggle")

	var id int64
	_, err := fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		http.Error(w, "Invalid reminder ID", http.StatusBadRequest)
		return
	}

	// Get reminder
	reminder, err := h.reminderService.GetReminder(id, user.ID)
	if err != nil {
		http.Error(w, "Reminder not found", http.StatusNotFound)
		return
	}

	// Toggle active status
	reminder.IsActive = !reminder.IsActive

	// Update reminder
	err = h.reminderService.UpdateReminder(reminder, user.ID)
	if err != nil {
		http.Error(w, "Failed to update reminder", http.StatusInternalServerError)
		return
	}

	// Redirect with success message
	if reminder.IsActive {
		http.Redirect(w, r, "/?success=activated", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/?success=deactivated", http.StatusSeeOther)
	}
}
