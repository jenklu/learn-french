package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jenklu/learn-french/internal/models"
	"github.com/jenklu/learn-french/internal/services"
)

type AuthHandler struct {
	service     *services.AuthService
	store       *sessions.CookieStore
	sessionName string
}

func NewAuthHandler(service *services.AuthService, store *sessions.CookieStore, sessionName string) *AuthHandler {
	return &AuthHandler{
		service:     service,
		store:       store,
		sessionName: sessionName,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), &req)
	if err != nil {
		if err == services.ErrUsernameExists {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		if err == services.ErrEmailExists {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create session
	session, _ := h.store.Get(r, h.sessionName)
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["is_admin"] = user.IsAdmin
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":     user,
		"message": "Registration successful",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.UserLogin
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	user, err := h.service.Login(r.Context(), &req)
	if err != nil {
		if err == services.ErrUserNotFound || err == services.ErrInvalidPassword {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update timezone from request header
	timezone := r.Header.Get("X-Timezone")
	if timezone != "" {
		go h.service.UpdateTimezone(r.Context(), user.ID, timezone)
	}

	// Create session
	session, _ := h.store.Get(r, h.sessionName)
	session.Values["user_id"] = user.ID
	session.Values["username"] = user.Username
	session.Values["is_admin"] = user.IsAdmin
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":     user,
		"message": "Login successful",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Values["is_admin"] = nil
	session.Options.MaxAge = -1
	session.Save(r, w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logout successful"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// In a real implementation, we'd fetch the user from the database
	// For now, we'll just return the session data
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":   userID,
		"username":  session.Values["username"],
		"is_admin":  session.Values["is_admin"],
	})
}
