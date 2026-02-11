package handlers

import (
	"email-reminder-system/models"
	"email-reminder-system/services"
	"encoding/json"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// EmailHistoryHandler handles email history-related HTTP requests
type EmailHistoryHandler struct {
	historyService *services.EmailHistoryService
	templates      *template.Template
}

// NewEmailHistoryHandler creates a new email history handler
func NewEmailHistoryHandler(historyService *services.EmailHistoryService, templates *template.Template) *EmailHistoryHandler {
	return &EmailHistoryHandler{
		historyService: historyService,
		templates:      templates,
	}
}

// ListEmailHistory displays the email history page with filters and pagination (GET /email-history)
func (h *EmailHistoryHandler) ListEmailHistory(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	query := r.URL.Query()
	
	// Parse page number (default 1)
	page := 1
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse status filter
	status := query.Get("status")
	
	// Parse reminder_id filter
	var reminderID *int64
	if reminderIDStr := query.Get("reminder_id"); reminderIDStr != "" {
		if rid, err := strconv.ParseInt(reminderIDStr, 10, 64); err == nil {
			reminderID = &rid
		}
	}

	// Parse date range filters
	var startDate, endDate *time.Time
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		if sd, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &sd
		}
	}
	if endDateStr := query.Get("end_date"); endDateStr != "" {
		if ed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			// Set to end of day
			endOfDay := ed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &endOfDay
		}
	}

	// Build filter
	filter := &models.EmailHistoryFilter{
		UserID:     user.ID,
		ReminderID: reminderID,
		Status:     status,
		StartDate:  startDate,
		EndDate:    endDate,
		Page:       page,
		PageSize:   20,
	}

	// Get email history
	result, err := h.historyService.GetEmailHistory(filter)
	if err != nil {
		http.Error(w, "Failed to load email history", http.StatusInternalServerError)
		return
	}

	// Get user's reminders for filter dropdown
	reminders, err := h.historyService.GetUserReminders(user.ID)
	if err != nil {
		http.Error(w, "Failed to load reminders", http.StatusInternalServerError)
		return
	}

	// Build filter query string for pagination links
	filterQuery := ""
	if status != "" {
		filterQuery += "&status=" + status
	}
	if reminderID != nil {
		filterQuery += "&reminder_id=" + strconv.FormatInt(*reminderID, 10)
	}
	if startDate != nil {
		filterQuery += "&start_date=" + query.Get("start_date")
	}
	if endDate != nil {
		filterQuery += "&end_date=" + query.Get("end_date")
	}

	// Check if any filters are applied
	filtersApplied := status != "" || reminderID != nil || startDate != nil || endDate != nil

	// Prepare template data
	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "email-history",
		"Logs":            result.Logs,
		"TotalCount":      result.TotalCount,
		"Page":            result.Page,
		"PageSize":        result.PageSize,
		"TotalPages":      result.TotalPages,
		"Reminders":       reminders,
		"Status":          status,
		"ReminderID":      reminderID,
		"StartDate":       query.Get("start_date"),
		"EndDate":         query.Get("end_date"),
		"FilterQuery":     filterQuery,
		"FiltersApplied":  filtersApplied,
	}

	// Render template
	err = h.templates.ExecuteTemplate(w, "email_history.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
		return
	}
}

// GetEmailLogDetail returns detailed information about a specific email log (GET /email-history/:id)
func (h *EmailHistoryHandler) GetEmailLogDetail(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user from context
	user, ok := GetUserFromContext(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Unauthorized",
		})
		return
	}

	// Extract log ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid log ID",
		})
		return
	}

	logID, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid log ID",
		})
		return
	}

	// Get email log detail
	log, err := h.historyService.GetEmailLogDetail(logID, user.ID)
	if err != nil {
		// Check error type for appropriate status code
		if strings.Contains(err.Error(), "not found") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Email log not found",
			})
		} else if strings.Contains(err.Error(), "unauthorized") {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Access denied",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Failed to load email log",
			})
		}
		return
	}

	// Escape HTML in error message for XSS prevention
	var escapedErrorMessage *string
	if log.ErrorMessage != nil {
		escaped := html.EscapeString(*log.ErrorMessage)
		escapedErrorMessage = &escaped
	}

	// Prepare JSON response
	response := map[string]interface{}{
		"ID":            log.ID,
		"ReminderID":    log.ReminderID,
		"ReminderTitle": log.ReminderTitle,
		"Recipients":    log.Recipients,
		"Subject":       log.Subject,
		"EmailContent":  log.EmailContent,
		"Status":        log.Status,
		"ErrorMessage":  escapedErrorMessage,
		"SentAt":        log.SentAt.Format(time.RFC3339),
		"CreatedAt":     log.CreatedAt.Format(time.RFC3339),
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
