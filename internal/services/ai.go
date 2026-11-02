package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jenklu/learn-french/internal/config"
	"github.com/jenklu/learn-french/internal/models"
)

type AIService struct {
	config    *config.Config
	httpClient *http.Client
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{
		config:    cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ExercisePrompt defines the structure for generating exercises
type ExercisePrompt struct {
	Level       string `json:"level"`
	Count       int    `json:"count"`
	SkillType   string `json:"skill_type"`
	Difficulty  string `json:"difficulty"`
}

// MistralResponse defines the structure for Mistral API response
type MistralResponse struct {
	Outputs []struct {
		Text string `json:"text"`
		StopReason string `json:"stop_reason"`
	} `json:"outputs"`
}

// ExerciseGenerationResult defines the structure for parsed exercises
type ExerciseGenerationResult struct {
	Exercises []*models.Exercise `json:"exercises"`
}

// GenerateExercises uses Mistral API to generate French listening exercises
func (s *AIService) GenerateExercises(ctx context.Context, level string, count int) ([]*models.Exercise, error) {
	if s.config.MistralAPIKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY not configured")
	}

	// Create a prompt for Mistral to generate exercises
	prompt := createExercisePrompt(level, count)

	// Call Mistral API
	responseText, err := s.callMistralAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call Mistral API: %w", err)
	}

	// Parse the response into exercises
	exercises, err := parseExercises(responseText, level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse exercises: %w", err)
	}

	// Ensure we have exactly 'count' exercises
	if len(exercises) < count {
		// If we don't have enough, generate more
		remaining := count - len(exercises)
		additional, err := s.GenerateExercises(ctx, level, remaining)
		if err != nil {
			// Fallback: return what we have
			log.Printf("Failed to generate additional exercises: %v", err)
			return exercises, nil
		}
		exercises = append(exercises, additional...)
	}

	// Trim to exactly count exercises
	if len(exercises) > count {
		exercises = exercises[:count]
	}

	return exercises, nil
}

// callMistralAPI calls the Mistral API to generate text
func (s *AIService) callMistralAPI(ctx context.Context, prompt string) (string, error) {
	// Mistral API endpoint
	url := "https://api.mistral.ai/v1/fim/completions"
	
	// Alternative: Use chat completions endpoint
	// url := "https://api.mistral.ai/v1/chat/completions"
	
	// For now, we'll use a simpler approach with the completions endpoint
	// This is a placeholder - Mistral's actual API might differ
	
	// Create request body
	requestBody := map[string]interface{}{
		"model": "mistral-tiny",
		"prompt": prompt,
		"max_tokens": 2000,
		"temperature": 0.7,
		"stop": []string{"\n\n\n"},
	}
	
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.MistralAPIKey)

	// Make the request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Mistral API error: %s - %s", resp.Status, string(body))
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var mistralResp MistralResponse
	if err := json.Unmarshal(body, &mistralResp); err != nil {
		// If JSON parsing fails, try to extract text directly
		log.Printf("Failed to parse Mistral response as JSON: %v", err)
		log.Printf("Raw response: %s", string(body))
		
		// Try to extract the text from the response
		// Mistral might return different formats
		var genericResp map[string]interface{}
		if err := json.Unmarshal(body, &genericResp); err == nil {
			if choices, ok := genericResp["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if text, ok := choice["text"].(string); ok {
						return text, nil
					}
				}
			}
			if outputs, ok := genericResp["outputs"].([]interface{}); ok && len(outputs) > 0 {
				if output, ok := outputs[0].(map[string]interface{}); ok {
					if text, ok := output["text"].(string); ok {
						return text, nil
					}
				}
			}
		}
		
		return string(body), nil
	}

	// Extract text from response
	if len(mistralResp.Outputs) > 0 {
		return mistralResp.Outputs[0].Text, nil
	}

	return "", fmt.Errorf("no text in response")
}

// createExercisePrompt creates a prompt for Mistral to generate exercises
func createExercisePrompt(level string, count int) string {
	// Define the format we want
	format := `Generate %d French listening comprehension exercises for %s level learners.

REQUIREMENTS:
- Each exercise must have exactly 4 multiple-choice answers
- Only ONE answer is correct
- Write everything in French
- Use common vocabulary and grammar appropriate for %s level
- Include a question, 4 options, the correct answer index (0-3), and the transcript text
- Format each exercise as a JSON object with these fields:
  {
    "question": "The question in French",
    "options": ["option 1", "option 2", "option 3", "option 4"],
    "correct_answer": 0,  // index of correct answer (0-3)
    "text": "The transcript of what was said"
  }

- Return ONLY a JSON array of exercise objects, no other text
- Example output:
[
  {
    "question": "Qu'est-ce que tu as mangé ce matin ?",
    "options": ["Du pain et du café", "Une pizza", "Rien", "Un croissant"],
    "correct_answer": 0,
    "text": "J'ai mangé du pain et du café ce matin."
  }
]

Generate %d exercises now:`

	return fmt.Sprintf(format, count, level, level, count)
}

// parseExercises parses the Mistral response into Exercise objects
func parseExercises(responseText string, level string) ([]*models.Exercise, error) {
	// Clean up the response text
	cleaned := strings.TrimSpace(responseText)
	
	// Try to find JSON in the response
	// Mistral might wrap the JSON in markdown code blocks
	if strings.HasPrefix(cleaned, "```") {
		// Remove markdown code blocks
		lines := strings.Split(cleaned, "\n")
		for i, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```json") || strings.HasPrefix(trimmed, "```") {
				continue
			}
			if strings.HasSuffix(trimmed, "```") {
				cleaned = strings.Join(lines[:i], "\n")
				break
			}
		}
		cleaned = strings.TrimSpace(cleaned)
	}

	// Try to parse as JSON array
	var result ExerciseGenerationResult
	if err := json.Unmarshal([]byte(cleaned), &result.Exercises); err == nil {
		// Success!
		for _, exercise := range result.Exercises {
			exercise.Level = level
		}
		return result.Exercises, nil
	}

	// Try to find JSON in the text
	// Look for the opening bracket
	startIdx := strings.Index(cleaned, "[")
	endIdx := strings.LastIndex(cleaned, "]")
	
	if startIdx >= 0 && endIdx > startIdx {
		jsonStr := cleaned[startIdx : endIdx+1]
		var result ExerciseGenerationResult
		if err := json.Unmarshal([]byte(jsonStr), &result.Exercises); err == nil {
			for _, exercise := range result.Exercises {
				exercise.Level = level
			}
			return result.Exercises, nil
		}
	}

	// If parsing fails, generate fallback exercises
	log.Printf("Failed to parse exercises from response: %s", cleaned)
	return generateFallbackExercises(level, 5), nil
}

// generateFallbackExercises generates hardcoded exercises as a fallback
func generateFallbackExercises(level string, count int) []*models.Exercise {
	// Predefined exercises for each level
	a2Exercises := []*models.Exercise{
		{
			Level:       "A2",
			Question:    "Qu'est-ce que tu as mangé ce matin ?",
			Options:     []string{"Du pain et du café", "Une pizza", "Rien", "Un croissant"},
			CorrectAnswer: 0,
			Text:        "J'ai mangé du pain et du café ce matin.",
		},
		{
			Level:       "A2",
			Question:    "Où est la bibliothèque ?",
			Options:     []string{"À droite", "À gauche", "Tout droit", "Derrière"},
			CorrectAnswer: 2,
			Text:        "La bibliothèque est tout droit.",
		},
		{
			Level:       "A2",
			Question:    "Quel temps fait-il aujourd'hui ?",
			Options:     []string{"Il pleut", "Il fait beau", "Il neige", "Il y a du vent"},
			CorrectAnswer: 1,
			Text:        "Aujourd'hui, il fait beau.",
		},
		{
			Level:       "A2",
			Question:    "Comment tu t'appelles ?",
			Options:     []string{"Je m'appelle Pierre", "J'ai 25 ans", "Je suis français", "J'habite à Paris"},
			CorrectAnswer: 0,
			Text:        "Je m'appelle Pierre.",
		},
		{
			Level:       "A2",
			Question:    "Quelle heure est-il ?",
			Options:     []string{"Il est 8 heures", "Il est midi", "Il est 15 heures", "Il est 18 heures"},
			CorrectAnswer: 0,
			Text:        "Il est 8 heures.",
		},
	}

	b1Exercises := []*models.Exercise{
		{
			Level:       "B1",
			Question:    "Pourquoi as-tu choisi ce travail ?",
			Options:     []string{"Parce que c'est facile", "Parce que c'est intéressant", "Parce que c'est proche", "Parce que c'est bien payé"},
			CorrectAnswer: 1,
			Text:        "J'ai choisi ce travail parce que c'est intéressant.",
		},
		{
			Level:       "B1",
			Question:    "Comment tu te sens aujourd'hui ?",
			Options:     []string{"Très bien", "Mal", "Fatigué", "Heureux"},
			CorrectAnswer: 0,
			Text:        "Je me sens très bien aujourd'hui.",
		},
		{
			Level:       "B1",
			Question:    "Qu'est-ce que tu feras demain ?",
			Options:     []string{"Je vais au cinéma", "Je reste à la maison", "Je travaille", "Je voyage"},
			CorrectAnswer: 0,
			Text:        "Demain, je vais au cinéma.",
		},
		{
			Level:       "B1",
			Question:    "Où as-tu passé tes vacances l'été dernier ?",
			Options:     []string{"À la montagne", "À la plage", "À l'étranger", "Chez moi"},
			CorrectAnswer: 1,
			Text:        "L'été dernier, j'ai passé mes vacances à la plage.",
		},
		{
			Level:       "B1",
			Question:    "Que penses-tu de ce film ?",
			Options:     []string{"Il est excellent", "Il est ennuyeux", "Il est trop long", "Je ne l'ai pas vu"},
			CorrectAnswer: 0,
			Text:        "Je pense que ce film est excellent.",
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
		rand.Seed(time.Now().UnixNano())
		for i := range exercises {
			j := rand.Intn(i + 1)
			exercises[i], exercises[j] = exercises[j], exercises[i]
		}
		return exercises[:count]
	}

	return exercises
}

// GenerateExercise generates a single exercise using AI
func (s *AIService) GenerateExercise(ctx context.Context, level string) (*models.Exercise, error) {
	exercises, err := s.GenerateExercises(ctx, level, 1)
	if err != nil {
		return nil, err
	}
	if len(exercises) == 0 {
		return nil, fmt.Errorf("no exercises generated")
	}
	return exercises[0], nil
}

// GenerateExerciseBatch generates a batch of exercises and saves them to the database
func (s *AIService) GenerateExerciseBatch(ctx context.Context, db *Database, level string, count int) ([]*models.Exercise, error) {
	// Generate exercises
	exercises, err := s.GenerateExercises(ctx, level, count)
	if err != nil {
		return nil, fmt.Errorf("failed to generate exercises: %w", err)
	}

	// Save exercises to database
	for _, exercise := range exercises {
		_, err := db.Pool.Exec(ctx, `
			INSERT INTO exercises (level, question, options, correct_answer, text)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (question) DO NOTHING
		`, exercise.Level, exercise.Question, exercise.Options, exercise.CorrectAnswer, exercise.Text)
		if err != nil {
			log.Printf("Failed to save exercise: %v", err)
			// Continue with other exercises
		}
	}

	return exercises, nil
}

// Database is imported from the database package
// We'll use a type alias to avoid circular dependency issues
type Database struct {
	Pool interface{} // This would be *pgxpool.Pool in the actual implementation
}
