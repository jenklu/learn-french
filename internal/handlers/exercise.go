package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/jenklu/learn-french/internal/models"
	"github.com/jenklu/learn-french/internal/services"
)

type ExerciseHandler struct {
	service     *services.ExerciseService
	store       *sessions.CookieStore
	sessionName string
}

func NewExerciseHandler(service *services.ExerciseService, store *sessions.CookieStore, sessionName string) *ExerciseHandler {
	return &ExerciseHandler{
		service:     service,
		store:       store,
		sessionName: sessionName,
	}
}

func (h *ExerciseHandler) GetExercises(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		level = "A2"
	}

	countStr := r.URL.Query().Get("count")
	count := 5
	if countStr != "" {
		var err error
		count, err = strconv.Atoi(countStr)
		if err != nil || count <= 0 {
			count = 5
		}
	}

	exercises, err := h.service.GetExercises(r.Context(), level, count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exercises)
}

func (h *ExerciseHandler) SubmitExercise(w http.ResponseWriter, r *http.Request) {
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

func (h *ExerciseHandler) CompleteSession(w http.ResponseWriter, r *http.Request) {
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

	var completion models.SessionCompletion
	if err := json.NewDecoder(r.Body).Decode(&completion); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set default values
	if completion.DurationMinutes <= 0 {
		completion.DurationMinutes = 1
	}

	sessionResult, err := h.service.CompleteSession(r.Context(), userID, &completion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessionResult)
}

func (h *ExerciseHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessions, err := h.service.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func (h *ExerciseHandler) GetSessionResults(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	_, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	sessionIDStr := r.URL.Query().Get("session_id")
	if sessionIDStr == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	sessionID, err := strconv.Atoi(sessionIDStr)
	if err != nil {
		http.Error(w, "Invalid session_id", http.StatusBadRequest)
		return
	}

	results, err := h.service.GetSessionResults(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *ExerciseHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	progress, err := h.service.GetUserProgress(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

func (h *ExerciseHandler) GetDailyStats(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, h.sessionName)
	userID, ok := session.Values["user_id"].(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get today's session
	sessions, err := h.service.GetUserSessions(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	var todaySession *models.Session
	for _, s := range sessions {
		if s.Date.Truncate(24*time.Hour).Equal(today) {
			todaySession = s
			break
		}
	}

	if todaySession == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"has_session": false,
			"message":    "No practice session today yet",
		})
		return
	}

	// Get results for today's session
	results, err := h.service.GetSessionResults(r.Context(), todaySession.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate stats
	totalExercises := len(results)
	correctCount := 0
	for _, result := range results {
		if result.IsCorrect {
			correctCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"session":         todaySession,
		"total_exercises": totalExercises,
		"correct_count":   correctCount,
		"score":           todaySession.ListeningScore,
		"results":         results,
	})
}
