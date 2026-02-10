package handlers

import (
	"email-reminder-system/services"
	"html/template"
	"net/http"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService services.AuthService
	templates   *template.Template
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService services.AuthService, templates *template.Template) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		templates:   templates,
	}
}

// LoginPage displays the login page (GET /login)
func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to home
	cookie, err := r.Cookie("session_id")
	if err == nil {
		_, err := h.authService.ValidateSession(cookie.Value)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	// Render login template
	data := map[string]interface{}{
		"Error": r.URL.Query().Get("error"),
	}
	
	err = h.templates.ExecuteTemplate(w, "login.html", data)
	if err != nil {
		http.Error(w, "Failed to render login page", http.StatusInternalServerError)
		return
	}
}

// Login processes login credentials (POST /login)
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate credentials
	session, err := h.authService.Login(username, password)
	if err != nil {
		// Redirect back to login with error
		http.Redirect(w, r, "/login?error=invalid_credentials", http.StatusSeeOther)
		return
	}

	// Set secure session cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   true,  // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout invalidates the session (POST /logout)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session cookie
	cookie, err := r.Cookie("session_id")
	if err == nil {
		// Invalidate session in database
		h.authService.Logout(cookie.Value)
	}

	// Clear session cookie
	clearCookie := &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, clearCookie)

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
