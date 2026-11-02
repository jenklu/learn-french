package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
)

type ExerciseService struct {
	db         *database.Database
	aiService  *AIService
}

func NewExerciseService(db *database.Database, aiService *AIService) *ExerciseService {
	return &ExerciseService{
		db:        db,
		aiService: aiService,
	}
}

func (s *ExerciseService) GetExercises(ctx context.Context, level string, count int) ([]*models.Exercise, error) {
	if count <= 0 {
		count = 5
	}

	// First, try to get exercises from the database
	exercises, err := s.db.GetRandomExercises(ctx, level, count)
	if err != nil {
		log.Printf("Failed to get exercises from database: %v", err)
	}

	// If we have enough exercises, return them
	if len(exercises) >= count {
		return exercises[:count], nil
	}

	// If we don't have enough, generate new ones using AI
	if s.aiService != nil {
		generated, err := s.aiService.GenerateExercises(ctx, level, count)
		if err != nil {
			log.Printf("Failed to generate exercises with AI: %v", err)
			// Fallback to hardcoded exercises
			return s.getFallbackExercises(level, count), nil
		}

		// Save generated exercises to database for future use
		for _, exercise := range generated {
			_, err := s.db.Pool.Exec(ctx, `
				INSERT INTO exercises (level, question, options, correct_answer, text)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (question) DO NOTHING
			`, exercise.Level, exercise.Question, exercise.Options, exercise.CorrectAnswer, exercise.Text)
			if err != nil {
				log.Printf("Failed to save generated exercise: %v", err)
			}
		}

		// Combine with existing exercises
		allExercises := append(exercises, generated...)
		
		// Return exactly count exercises
		if len(allExercises) > count {
			return allExercises[:count], nil
		}
		return allExercises, nil
	}

	// Fallback to hardcoded exercises
	return s.getFallbackExercises(level, count), nil
}

func (s *ExerciseService) getFallbackExercises(level string, count int) []*models.Exercise {
	// Predefined exercises for each level
	a2Exercises := []*models.Exercise{
		{
			Level:         "A2",
			Question:      "Qu'est-ce que tu as mangé ce matin ?",
			Options:       []string{"Du pain et du café", "Une pizza", "Rien", "Un croissant"},
			CorrectAnswer: 0,
			Text:          "J'ai mangé du pain et du café ce matin.",
		},
		{
			Level:         "A2",
			Question:      "Où est la bibliothèque ?",
			Options:       []string{"À droite", "À gauche", "Tout droit", "Derrière"},
			CorrectAnswer: 2,
			Text:          "La bibliothèque est tout droit.",
		},
		{
			Level:         "A2",
			Question:      "Quel temps fait-il aujourd'hui ?",
			Options:       []string{"Il pleut", "Il fait beau", "Il neige", "Il y a du vent"},
			CorrectAnswer: 1,
			Text:          "Aujourd'hui, il fait beau.",
		},
		{
			Level:         "A2",
			Question:      "Comment tu t'appelles ?",
			Options:       []string{"Je m'appelle Pierre", "J'ai 25 ans", "Je suis français", "J'habite à Paris"},
			CorrectAnswer: 0,
			Text:          "Je m'appelle Pierre.",
		},
		{
			Level:         "A2",
			Question:      "Quelle heure est-il ?",
			Options:       []string{"Il est 8 heures", "Il est midi", "Il est 15 heures", "Il est 18 heures"},
			CorrectAnswer: 0,
			Text:          "Il est 8 heures.",
		},
	}

	b1Exercises := []*models.Exercise{
		{
			Level:         "B1",
			Question:      "Pourquoi as-tu choisi ce travail ?",
			Options:       []string{"Parce que c'est facile", "Parce que c'est intéressant", "Parce que c'est proche", "Parce que c'est bien payé"},
			CorrectAnswer: 1,
			Text:          "J'ai choisi ce travail parce que c'est intéressant.",
		},
		{
			Level:         "B1",
			Question:      "Comment tu te sens aujourd'hui ?",
			Options:       []string{"Très bien", "Mal", "Fatigué", "Heureux"},
			CorrectAnswer: 0,
			Text:          "Je me sens très bien aujourd'hui.",
		},
		{
			Level:         "B1",
			Question:      "Qu'est-ce que tu feras demain ?",
			Options:       []string{"Je vais au cinéma", "Je reste à la maison", "Je travaille", "Je voyage"},
			CorrectAnswer: 0,
			Text:          "Demain, je vais au cinéma.",
		},
		{
			Level:         "B1",
			Question:      "Où as-tu passé tes vacances l'été dernier ?",
			Options:       []string{"À la montagne", "À la plage", "À l'étranger", "Chez moi"},
			CorrectAnswer: 1,
			Text:          "L'été dernier, j'ai passé mes vacances à la plage.",
		},
		{
			Level:         "B1",
			Question:      "Que penses-tu de ce film ?",
			Options:       []string{"Il est excellent", "Il est ennuyeux", "Il est trop long", "Je ne l'ai pas vu"},
			CorrectAnswer: 0,
			Text:          "Je pense que ce film est excellent.",
		},
	}

	// Select exercises based on level
	var exercises []*models.Exercise
	if level == "A2" {
		exercises = a2Exercises
	} else if level == "B1" {
		exercises = b1Exercises
	} else {
		// Mix of both
		exercises = append(a2Exercises, b1Exercises...)
	}

	// Return up to count exercises
	if count < len(exercises) {
		// Shuffle and return count exercises
		// Simple shuffle for fallback
		for i := range exercises {
			j := i
			if i > 0 {
				j = (i * 1664525 + 1013904223) % i
			}
			exercises[i], exercises[j] = exercises[j], exercises[i]
		}
		return exercises[:count]
	}

	return exercises
}

func (s *ExerciseService) SubmitExercise(ctx context.Context, userID int, submission *models.ExerciseSubmission) (*models.ExerciseResult, error) {
	// Get the current session or create a new one
	session, err := s.getOrCreateSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Create exercise result
	result := &models.ExerciseResult{
		SessionID:    session.ID,
		ExerciseIndex: submission.ExerciseIndex,
		IsCorrect:    submission.IsCorrect,
		Score:        submission.Score,
		CreatedAt:    time.Now(),
	}

	if err := s.db.CreateExerciseResult(ctx, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ExerciseService) CompleteSession(ctx context.Context, userID int, completion *models.SessionCompletion) (*models.Session, error) {
	// Get the current session
	session, err := s.getCurrentSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update session
	session.DurationMinutes = completion.DurationMinutes
	session.ListeningScore = completion.ListeningScore
	session.CompletedAt = time.Now()

	if err := s.db.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	// Update progress
	level := "A2"
	if completion.ListeningScore >= 80 {
		level = "B1"
	}
	if err := s.db.UpdateProgress(ctx, userID, "listening", level); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *ExerciseService) getOrCreateSession(ctx context.Context, userID int) (*models.Session, error) {
	// Check for an incomplete session today
	sessions, err := s.db.GetSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, session := range sessions {
		if session.Date.Truncate(24*time.Hour).Equal(today) && session.CompletedAt.IsZero() {
			return session, nil
		}
	}

	// Create new session
	session := &models.Session{
		UserID:    userID,
		Date:      today,
		CreatedAt: time.Now(),
	}

	// We don't create it yet, just return a placeholder
	// The actual creation happens when we have results
	return session, nil
}

func (s *ExerciseService) getCurrentSession(ctx context.Context, userID int) (*models.Session, error) {
	sessions, err := s.db.GetSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, session := range sessions {
		if session.Date.Truncate(24*time.Hour).Equal(today) && session.CompletedAt.IsZero() {
			return session, nil
		}
	}

	return nil, fmt.Errorf("no active session found")
}

func (s *ExerciseService) GetSessionResults(ctx context.Context, sessionID int) ([]*models.ExerciseResult, error) {
	return s.db.GetExerciseResultsBySession(ctx, sessionID)
}

func (s *ExerciseService) GetUserSessions(ctx context.Context, userID int) ([]*models.Session, error) {
	return s.db.GetSessionsByUser(ctx, userID)
}

func (s *ExerciseService) GetUserProgress(ctx context.Context, userID int) ([]*models.Progress, error) {
	return s.db.GetProgress(ctx, userID)
}

// GenerateNewExercises generates new exercises using AI and saves them to the database
func (s *ExerciseService) GenerateNewExercises(ctx context.Context, level string, count int) ([]*models.Exercise, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("AI service not configured")
	}

	return s.aiService.GenerateExercises(ctx, level, count)
}

// RegenerateExercises clears existing exercises and generates new ones
func (s *ExerciseService) RegenerateExercises(ctx context.Context, level string, count int) ([]*models.Exercise, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("AI service not configured")
	}

	// Delete existing exercises for this level
	_, err := s.db.Pool.Exec(ctx, "DELETE FROM exercises WHERE level = $1", level)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing exercises: %w", err)
	}

	// Generate new exercises
	exercises, err := s.aiService.GenerateExercises(ctx, level, count)
	if err != nil {
		return nil, fmt.Errorf("failed to generate exercises: %w", err)
	}

	// Save exercises to database
	for _, exercise := range exercises {
		_, err := s.db.Pool.Exec(ctx, `
			INSERT INTO exercises (level, question, options, correct_answer, text)
			VALUES ($1, $2, $3, $4, $5)
		`, exercise.Level, exercise.Question, exercise.Options, exercise.CorrectAnswer, exercise.Text)
		if err != nil {
			log.Printf("Failed to save exercise: %v", err)
			// Continue with other exercises
		}
	}

	return exercises, nil
}
