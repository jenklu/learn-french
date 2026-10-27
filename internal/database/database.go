package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jenklu/learn-french/internal/models"
)

type Database struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Database, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{Pool: pool}, nil
}

func (db *Database) Close() {
	db.Pool.Close()
}

func (db *Database) Init(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			timezone VARCHAR(50) DEFAULT 'UTC',
			is_admin BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			date DATE NOT NULL,
			duration_minutes INTEGER NOT NULL,
			listening_score INTEGER NOT NULL,
			completed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS exercise_results (
			id SERIAL PRIMARY KEY,
			session_id INTEGER REFERENCES sessions(id) ON DELETE CASCADE,
			exercise_index INTEGER NOT NULL,
			is_correct BOOLEAN NOT NULL,
			score INTEGER NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS progress (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			skill_type VARCHAR(50) NOT NULL,
			level VARCHAR(50) NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS reminders (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			enabled BOOLEAN DEFAULT false
		)`,
		`CREATE TABLE IF NOT EXISTS push_subscriptions (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			endpoint TEXT NOT NULL,
			auth TEXT NOT NULL,
			p256dh TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS exercises (
			id SERIAL PRIMARY KEY,
			level VARCHAR(10) NOT NULL,
			question TEXT NOT NULL,
			options TEXT[] NOT NULL,
			correct_answer INTEGER NOT NULL,
			audio_url TEXT,
			text TEXT
		)`,
	}

	for _, query := range queries {
		if _, err := db.Pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	// Seed exercises if empty
	var count int
	if err := db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM exercises").Scan(&count); err != nil {
		return fmt.Errorf("failed to count exercises: %w", err)
	}

	if count == 0 {
		if err := seedExercises(ctx, db.Pool); err != nil {
			return fmt.Errorf("failed to seed exercises: %w", err)
		}
	}

	return nil
}

func seedExercises(ctx context.Context, pool *pgxpool.Pool) error {
	a2Exercises := []struct {
		level        string
		question     string
		options      []string
		correctAnswer int
		text         string
	}{
		{
			level:        "A2",
			question:     "Qu'est-ce que tu as mangé ce matin ?",
			options:      []string{"Du pain et du café", "Une pizza", "Rien", "Un croissant"},
			correctAnswer: 0,
			text:         "J'ai mangé du pain et du café ce matin.",
		},
		{
			level:        "A2",
			question:     "Où est la bibliothèque ?",
			options:      []string{"À droite", "À gauche", "Tout droit", "Derrière"},
			correctAnswer: 2,
			text:         "La bibliothèque est tout droit.",
		},
		{
			level:        "A2",
			question:     "Quel temps fait-il aujourd'hui ?",
			options:      []string{"Il pleut", "Il fait beau", "Il neige", "Il y a du vent"},
			correctAnswer: 1,
			text:         "Aujourd'hui, il fait beau.",
		},
		{
			level:        "B1",
			question:     "Pourquoi as-tu choisi ce travail ?",
			options:      []string{"Parce que c'est facile", "Parce que c'est intéressant", "Parce que c'est proche", "Parce que c'est bien payé"},
			correctAnswer: 1,
			text:         "J'ai choisi ce travail parce que c'est intéressant.",
		},
		{
			level:        "B1",
			question:     "Comment tu te sens aujourd'hui ?",
			options:      []string{"Très bien", "Mal", "Fatigué", "Heureux"},
			correctAnswer: 0,
			text:         "Je me sens très bien aujourd'hui.",
		},
	}

	for _, e := range a2Exercises {
		_, err := pool.Exec(ctx, `
			INSERT INTO exercises (level, question, options, correct_answer, text)
			VALUES ($1, $2, $3, $4, $5)
		`, e.level, e.question, e.options, e.correctAnswer, e.text)
		if err != nil {
			return err
		}
	}

	return nil
}

// User operations

func (db *Database) CreateUser(ctx context.Context, user *models.User) error {
	return db.Pool.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, timezone, is_admin)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, timezone, is_admin, created_at
	`, user.Username, user.Email, user.PasswordHash, user.Timezone, user.IsAdmin).Scan(
		&user.ID, &user.Username, &user.Email, &user.Timezone, &user.IsAdmin, &user.CreatedAt)
}

func (db *Database) GetUserByID(ctx context.Context, id int) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, username, email, timezone, is_admin, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Email, &user.Timezone, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *Database) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, timezone, is_admin, created_at
		FROM users WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Timezone, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *Database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, username, email, password_hash, timezone, is_admin, created_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Timezone, &user.IsAdmin, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (db *Database) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, username, email, timezone, is_admin, created_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Timezone, &user.IsAdmin, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (db *Database) UpdateUser(ctx context.Context, id int, username, email, timezone string, isAdmin bool) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE users SET username = $1, email = $2, timezone = $3, is_admin = $4
		WHERE id = $5
	`, username, email, timezone, isAdmin, id)
	return err
}

func (db *Database) UpdateUserPassword(ctx context.Context, id int, passwordHash string) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE users SET password_hash = $1 WHERE id = $2
	`, passwordHash, id)
	return err
}

func (db *Database) DeleteUser(ctx context.Context, id int) error {
	_, err := db.Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}

// Session operations

func (db *Database) CreateSession(ctx context.Context, session *models.Session) error {
	return db.Pool.QueryRow(ctx, `
		INSERT INTO sessions (user_id, date, duration_minutes, listening_score, completed_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, date, duration_minutes, listening_score, completed_at
	`, session.UserID, session.Date, session.DurationMinutes, session.ListeningScore, session.CompletedAt).Scan(
		&session.ID, &session.UserID, &session.Date, &session.DurationMinutes, &session.ListeningScore, &session.CompletedAt)
}

func (db *Database) GetSessionsByUser(ctx context.Context, userID int) ([]*models.Session, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, date, duration_minutes, listening_score, completed_at
		FROM sessions WHERE user_id = $1 ORDER BY completed_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		if err := rows.Scan(&session.ID, &session.UserID, &session.Date, &session.DurationMinutes, &session.ListeningScore, &session.CompletedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (db *Database) GetSessionByID(ctx context.Context, id int) (*models.Session, error) {
	session := &models.Session{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, date, duration_minutes, listening_score, completed_at
		FROM sessions WHERE id = $1
	`, id).Scan(&session.ID, &session.UserID, &session.Date, &session.DurationMinutes, &session.ListeningScore, &session.CompletedAt)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// ExerciseResult operations

func (db *Database) CreateExerciseResult(ctx context.Context, result *models.ExerciseResult) error {
	return db.Pool.QueryRow(ctx, `
		INSERT INTO exercise_results (session_id, exercise_index, is_correct, score, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, session_id, exercise_index, is_correct, score, created_at
	`, result.SessionID, result.ExerciseIndex, result.IsCorrect, result.Score, result.CreatedAt).Scan(
		&result.ID, &result.SessionID, &result.ExerciseIndex, &result.IsCorrect, &result.Score, &result.CreatedAt)
}

func (db *Database) GetExerciseResultsBySession(ctx context.Context, sessionID int) ([]*models.ExerciseResult, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, session_id, exercise_index, is_correct, score, created_at
		FROM exercise_results WHERE session_id = $1 ORDER BY created_at
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.ExerciseResult
	for rows.Next() {
		result := &models.ExerciseResult{}
		if err := rows.Scan(&result.ID, &result.SessionID, &result.ExerciseIndex, &result.IsCorrect, &result.Score, &result.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Progress operations

func (db *Database) GetProgress(ctx context.Context, userID int) ([]*models.Progress, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, skill_type, level, updated_at
		FROM progress WHERE user_id = $1 ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var progress []*models.Progress
	for rows.Next() {
		p := &models.Progress{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.SkillType, &p.Level, &p.UpdatedAt); err != nil {
			return nil, err
		}
		progress = append(progress, p)
	}
	return progress, nil
}

func (db *Database) UpdateProgress(ctx context.Context, userID int, skillType, level string) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO progress (user_id, skill_type, level, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, skill_type) DO UPDATE
		SET level = EXCLUDED.level, updated_at = NOW()
	`, userID, skillType, level)
	return err
}

// Reminder operations

func (db *Database) GetReminder(ctx context.Context, userID int) (*models.Reminder, error) {
	reminder := &models.Reminder{}
	err := db.Pool.QueryRow(ctx, `
		SELECT id, user_id, enabled
		FROM reminders WHERE user_id = $1
	`, userID).Scan(&reminder.ID, &reminder.UserID, &reminder.Enabled)
	if err != nil {
		// If no reminder exists, create one
		if err := db.Pool.QueryRow(ctx, `
			INSERT INTO reminders (user_id, enabled)
			VALUES ($1, false)
			RETURNING id, user_id, enabled
		`, userID).Scan(&reminder.ID, &reminder.UserID, &reminder.Enabled); err != nil {
			return nil, err
		}
	}
	return reminder, nil
}

func (db *Database) UpdateReminder(ctx context.Context, userID int, enabled bool) error {
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO reminders (user_id, enabled)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE
		SET enabled = EXCLUDED.enabled
	`, userID, enabled)
	return err
}

func (db *Database) GetUsersWithRemindersEnabled(ctx context.Context) ([]*models.User, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT u.id, u.username, u.email, u.timezone, u.is_admin, u.created_at
		FROM users u
		JOIN reminders r ON u.id = r.user_id
		WHERE r.enabled = true
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.Timezone, &user.IsAdmin, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// PushSubscription operations

func (db *Database) CreatePushSubscription(ctx context.Context, sub *models.PushSubscription) error {
	return db.Pool.QueryRow(ctx, `
		INSERT INTO push_subscriptions (user_id, endpoint, auth, p256dh)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, endpoint, auth, p256dh
	`, sub.UserID, sub.Endpoint, sub.Auth, sub.P256DH).Scan(
		&sub.ID, &sub.UserID, &sub.Endpoint, &sub.Auth, &sub.P256DH)
}

func (db *Database) GetPushSubscriptionsByUser(ctx context.Context, userID int) ([]*models.PushSubscription, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, endpoint, auth, p256dh
		FROM push_subscriptions WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.PushSubscription
	for rows.Next() {
		sub := &models.PushSubscription{}
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.Auth, &sub.P256DH); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

func (db *Database) DeletePushSubscription(ctx context.Context, id int) error {
	_, err := db.Pool.Exec(ctx, "DELETE FROM push_subscriptions WHERE id = $1", id)
	return err
}

func (db *Database) GetAllPushSubscriptions(ctx context.Context) ([]*models.PushSubscription, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, user_id, endpoint, auth, p256dh
		FROM push_subscriptions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*models.PushSubscription
	for rows.Next() {
		sub := &models.PushSubscription{}
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.Endpoint, &sub.Auth, &sub.P256DH); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, nil
}

// Exercise operations

func (db *Database) GetExercisesByLevel(ctx context.Context, level string) ([]*models.Exercise, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, level, question, options, correct_answer, audio_url, text
		FROM exercises WHERE level = $1 ORDER BY id
	`, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []*models.Exercise
	for rows.Next() {
		exercise := &models.Exercise{}
		var options []string
		if err := rows.Scan(&exercise.ID, &exercise.Level, &exercise.Question, &options, &exercise.CorrectAnswer, &exercise.AudioURL, &exercise.Text); err != nil {
			return nil, err
		}
		exercise.Options = options
		exercises = append(exercises, exercise)
	}
	return exercises, nil
}

func (db *Database) GetExerciseByID(ctx context.Context, id int) (*models.Exercise, error) {
	exercise := &models.Exercise{}
	var options []string
	err := db.Pool.QueryRow(ctx, `
		SELECT id, level, question, options, correct_answer, audio_url, text
		FROM exercises WHERE id = $1
	`, id).Scan(&exercise.ID, &exercise.Level, &exercise.Question, &options, &exercise.CorrectAnswer, &exercise.AudioURL, &exercise.Text)
	if err != nil {
		return nil, err
	}
	exercise.Options = options
	return exercise, nil
}

func (db *Database) GetRandomExercises(ctx context.Context, level string, count int) ([]*models.Exercise, error) {
	rows, err := db.Pool.Query(ctx, `
		SELECT id, level, question, options, correct_answer, audio_url, text
		FROM exercises WHERE level = $1 ORDER BY RANDOM() LIMIT $2
	`, level, count)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []*models.Exercise
	for rows.Next() {
		exercise := &models.Exercise{}
		var options []string
		if err := rows.Scan(&exercise.ID, &exercise.Level, &exercise.Question, &options, &exercise.CorrectAnswer, &exercise.AudioURL, &exercise.Text); err != nil {
			return nil, err
		}
		exercise.Options = options
		exercises = append(exercises, exercise)
	}
	return exercises, nil
}
