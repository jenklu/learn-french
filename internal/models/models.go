package models

import (
	"time"
)

type User struct {
	ID            int       `json:"id"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Timezone      string    `json:"timezone"`
	IsAdmin       bool      `json:"is_admin"`
	CreatedAt     time.Time `json:"created_at"`
}

type Session struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	Date           time.Time `json:"date"`
	DurationMinutes int       `json:"duration_minutes"`
	ListeningScore int       `json:"listening_score"`
	CompletedAt    time.Time `json:"completed_at"`
}

type ExerciseResult struct {
	ID           int       `json:"id"`
	SessionID    int       `json:"session_id"`
	ExerciseIndex int       `json:"exercise_index"`
	IsCorrect    bool      `json:"is_correct"`
	Score        int       `json:"score"`
	CreatedAt    time.Time `json:"created_at"`
}

type Progress struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	SkillType string    `json:"skill_type"`
	Level     string    `json:"level"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Reminder struct {
	ID      int  `json:"id"`
	UserID  int  `json:"user_id"`
	Enabled bool `json:"enabled"`
}

type PushSubscription struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Endpoint string `json:"endpoint"`
	Auth     string `json:"auth"`
	P256DH   string `json:"p256dh"`
}

type Exercise struct {
	ID          int    `json:"id"`
	Level       string `json:"level"`
	Question    string `json:"question"`
	Options     []string `json:"options"`
	CorrectAnswer int `json:"correct_answer"`
	AudioURL    string `json:"audio_url"`
	Text        string `json:"text"`
}

type UserRegistration struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Timezone string `json:"timezone"`
}

type UserLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ExerciseSubmission struct {
	SessionID    int    `json:"session_id"`
	ExerciseID   int    `json:"exercise_id"`
	ExerciseIndex int    `json:"exercise_index"`
	SelectedOption int `json:"selected_option"`
	IsCorrect    bool   `json:"is_correct"`
	Score        int    `json:"score"`
	Duration     int    `json:"duration"`
}

type SessionCompletion struct {
	DurationMinutes int `json:"duration_minutes"`
	ListeningScore  int `json:"listening_score"`
}

type PushSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		Auth   string `json:"auth"`
		P256DH string `json:"p256dh"`
	} `json:"keys"`
}

type ReminderRequest struct {
	Enabled bool `json:"enabled"`
}
