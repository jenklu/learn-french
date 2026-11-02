package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/jenklu/learn-french/internal/models"
	"github.com/jenklu/learn-french/internal/services"
)

type AIHandler struct {
	service     *services.ExerciseService
	store       *sessions.CookieStore
	sessionName string
}

func NewAIHandler(service *services.ExerciseService, store *sessions.CookieStore, sessionName string) *AIHandler {
	return &AIHandler{
		service:     service,
		store:       store,
		sessionName: sessionName,
	}
}

func (h *AIHandler) isAdmin(w http.ResponseWriter, r *http.Request) bool {
	session, _ := h.store.Get(r, h.sessionName)
	isAdmin, ok := session.Values["is_admin"].(bool)
	return ok && isAdmin
}

// GenerateExercises generates new exercises using AI
func (h *AIHandler) GenerateExercises(w http.ResponseWriter, r *http.Request) {
	// Only admins can generate exercises manually
	if !h.isAdmin(w, r) {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Level string `json:"level"`
		Count int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Level == "" {
		req.Level = "A2"
	}

	if req.Count <= 0 {
		req.Count = 5
	}

	exercises, err := h.service.GenerateNewExercises(r.Context(), req.Level, req.Count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exercises": exercises,
		"count":    len(exercises),
		"level":    req.Level,
	})
}

// RegenerateExercises clears and regenerates exercises for a specific level
func (h *AIHandler) RegenerateExercises(w http.ResponseWriter, r *http.Request) {
	// Only admins can regenerate exercises
	if !h.isAdmin(w, r) {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	level := r.URL.Query().Get("level")
	if level == "" {
		level = "A2"
	}

	countStr := r.URL.Query().Get("count")
	count := 10
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			count = 10
		}
	}

	exercises, err := h.service.RegenerateExercises(r.Context(), level, count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":   "Exercises regenerated successfully",
		"exercises": exercises,
		"count":     len(exercises),
		"level":     level,
	})
}

// GetAIStatus checks if AI is configured and working
func (h *AIHandler) GetAIStatus(w http.ResponseWriter, r *http.Request) {
	// This endpoint can be public or admin-only
	// For now, we'll make it admin-only
	if !h.isAdmin(w, r) {
		http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
		return
	}

	// Try to generate a single exercise to test the AI service
	exercise, err := h.service.GenerateNewExercises(r.Context(), "A2", 1)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "error",
			"message":   "AI service not configured or not working",
			"error":     err.Error(),
			"ai_ready":  false,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"message":   "AI service is working",
		"ai_ready":  true,
		"test_exercise": exercise[0],
	})
}

// SubmitAIExercise allows users to submit answers to AI-generated exercises
// This is the same as the regular exercise submission but ensures compatibility
func (h *AIHandler) SubmitAIExercise(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, _ := h.store.Get(r, h.sessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var submission models.ExerciseSubmission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.service.SubmitExercise(r.Context(), userID, &submission)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
