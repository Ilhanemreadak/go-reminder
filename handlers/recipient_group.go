package handlers

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"email-reminder-system/database"
	"email-reminder-system/models"
)

// RecipientGroupHandler handles HTTP requests for recipient group operations
type RecipientGroupHandler struct {
	repo      database.Repository
	templates *template.Template
}

// NewRecipientGroupHandler creates a new recipient group handler
func NewRecipientGroupHandler(repo database.Repository, templates *template.Template) *RecipientGroupHandler {
	return &RecipientGroupHandler{
		repo:      repo,
		templates: templates,
	}
}

// ListGroups displays all recipient groups for the user (GET /groups)
func (h *RecipientGroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := h.repo.GetRecipientGroupsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load groups", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "groups",
		"Groups":          groups,
		"Success":         r.URL.Query().Get("success"),
		"Error":           r.URL.Query().Get("error"),
	}

	err = h.templates.ExecuteTemplate(w, "groups.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// NewGroupForm displays the group creation form (GET /groups/new)
func (h *RecipientGroupHandler) NewGroupForm(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "groups",
		"Group":           &models.RecipientGroup{},
		"IsEdit":          false,
	}

	err := h.templates.ExecuteTemplate(w, "group_form.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// CreateGroup handles group creation (POST /groups)
func (h *RecipientGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormWithError(w, user, &models.RecipientGroup{}, false, "Grup adı zorunludur.")
		return
	}

	// Parse emails
	emailsStr := r.FormValue("emails")
	emails := parseEmails(emailsStr)
	if len(emails) == 0 {
		h.renderFormWithError(w, user, &models.RecipientGroup{Name: name}, false, "En az bir e-posta adresi zorunludur.")
		return
	}

	group := &models.RecipientGroup{
		UserID: user.ID,
		Name:   name,
		Emails: emails,
	}

	if err := h.repo.CreateRecipientGroup(group); err != nil {
		h.renderFormWithError(w, user, group, false, "Grup oluşturulurken hata: "+err.Error())
		return
	}

	http.Redirect(w, r, "/groups?success=created", http.StatusSeeOther)
}

// EditGroupForm displays the group edit form (GET /groups/:id/edit)
func (h *RecipientGroupHandler) EditGroupForm(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	group, err := h.repo.GetRecipientGroup(id)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	if group.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "groups",
		"Group":           group,
		"IsEdit":          true,
	}

	err = h.templates.ExecuteTemplate(w, "group_form.html", data)
	if err != nil {
		http.Error(w, "Failed to render page", http.StatusInternalServerError)
	}
}

// UpdateGroup handles group updates (POST /groups/:id)
func (h *RecipientGroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Check authorization
	existing, err := h.repo.GetRecipientGroup(id)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	if existing.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		h.renderFormWithError(w, user, existing, true, "Grup adı zorunludur.")
		return
	}

	emails := parseEmails(r.FormValue("emails"))
	if len(emails) == 0 {
		existing.Name = name
		h.renderFormWithError(w, user, existing, true, "En az bir e-posta adresi zorunludur.")
		return
	}

	existing.Name = name
	existing.Emails = emails

	if err := h.repo.UpdateRecipientGroup(existing); err != nil {
		h.renderFormWithError(w, user, existing, true, "Grup güncellenirken hata: "+err.Error())
		return
	}

	http.Redirect(w, r, "/groups?success=updated", http.StatusSeeOther)
}

// DeleteGroup handles group deletion (POST /groups/:id/delete)
func (h *RecipientGroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(pathParts[2], 10, 64)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	group, err := h.repo.GetRecipientGroup(id)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	if group.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if err := h.repo.DeleteRecipientGroup(id); err != nil {
		http.Redirect(w, r, "/groups?error=Silinemedi", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/groups?success=deleted", http.StatusSeeOther)
}

// ExportCSV exports groups as CSV (GET /groups/export)
func (h *RecipientGroupHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get all groups for user
	allGroups, err := h.repo.GetRecipientGroupsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load groups", http.StatusInternalServerError)
		return
	}

	// Filter by IDs if provided
	var groupsToExport []*models.RecipientGroup

	// Parse form to get query parameters (method GET supported)
	r.ParseForm()
	selectedIDs := r.Form["ids"]

	if len(selectedIDs) > 0 {
		// Creaet a map for faster lookup
		idMap := make(map[int64]bool)
		for _, idStr := range selectedIDs {
			if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				idMap[id] = true
			}
		}

		for _, group := range allGroups {
			if idMap[group.ID] {
				groupsToExport = append(groupsToExport, group)
			}
		}
	} else {
		// If no IDs selected, export ALL groups (or handle as error/empty if preferred,
		// but exporting all is a good default for "Export List" button)
		groupsToExport = allGroups
	}

	if len(groupsToExport) == 0 {
		// If filtering resulted in no groups (e.g. invalid IDs), redirect with error
		http.Redirect(w, r, "/groups?error=Dışa+aktarılacak+grup+bulunamadı", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=recipient_groups.csv")

	// Write BOM for Excel UTF-8 compatibility
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header row
	writer.Write([]string{"Grup Adı", "E-posta Adresleri"})

	// Data rows
	for _, group := range groupsToExport {
		writer.Write([]string{group.Name, strings.Join(group.Emails, ", ")})
	}
}

// ImportCSV imports groups from CSV (POST /groups/import)
func (h *RecipientGroupHandler) ImportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Redirect(w, r, "/groups?error=Dosya+yüklenemedi", http.StatusSeeOther)
		return
	}

	file, _, err := r.FormFile("csv_file")
	if err != nil {
		http.Redirect(w, r, "/groups?error=Dosya+seçilmedi", http.StatusSeeOther)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		http.Redirect(w, r, "/groups?error=CSV+dosyası+okunamadı", http.StatusSeeOther)
		return
	}

	if len(records) < 2 {
		http.Redirect(w, r, "/groups?error=CSV+dosyası+boş+veya+yalnızca+başlık+satırı+içeriyor", http.StatusSeeOther)
		return
	}

	imported := 0
	for i, record := range records {
		// Skip header row
		if i == 0 {
			continue
		}

		if len(record) < 2 {
			continue
		}

		name := strings.TrimSpace(record[0])
		if name == "" {
			continue
		}

		emails := parseEmails(record[1])
		if len(emails) == 0 {
			continue
		}

		group := &models.RecipientGroup{
			UserID: user.ID,
			Name:   name,
			Emails: emails,
		}

		if err := h.repo.CreateRecipientGroup(group); err != nil {
			continue
		}
		imported++
	}

	http.Redirect(w, r, fmt.Sprintf("/groups?success=imported&count=%d", imported), http.StatusSeeOther)
}

// GetGroupsJSON returns groups as JSON for AJAX (GET /api/groups)
func (h *RecipientGroupHandler) GetGroupsJSON(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groups, err := h.repo.GetRecipientGroupsByUserID(user.ID)
	if err != nil {
		http.Error(w, "Failed to load groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("["))
	for i, group := range groups {
		if i > 0 {
			w.Write([]byte(","))
		}
		emailsJSON := "["
		for j, email := range group.Emails {
			if j > 0 {
				emailsJSON += ","
			}
			emailsJSON += fmt.Sprintf(`"%s"`, strings.ReplaceAll(email, `"`, `\"`))
		}
		emailsJSON += "]"
		w.Write([]byte(fmt.Sprintf(`{"id":%d,"name":"%s","emails":%s}`,
			group.ID, strings.ReplaceAll(group.Name, `"`, `\"`), emailsJSON)))
	}
	w.Write([]byte("]"))
}

// Helper functions

func (h *RecipientGroupHandler) renderFormWithError(w http.ResponseWriter, user *models.User, group *models.RecipientGroup, isEdit bool, errorMsg string) {
	data := map[string]interface{}{
		"User":            user,
		"IsAuthenticated": true,
		"ActivePage":      "groups",
		"Group":           group,
		"IsEdit":          isEdit,
		"Error":           errorMsg,
	}
	h.templates.ExecuteTemplate(w, "group_form.html", data)
}

func parseEmails(emailsStr string) []string {
	var emails []string
	for _, email := range strings.Split(emailsStr, ",") {
		trimmed := strings.TrimSpace(email)
		if trimmed != "" {
			emails = append(emails, trimmed)
		}
	}
	return emails
}
