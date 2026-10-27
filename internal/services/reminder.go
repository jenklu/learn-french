package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/models"
)

type ReminderService struct {
	db     *database.Database
	vapid  VAPIDConfig
}

type VAPIDConfig struct {
	PublicKey  string
	PrivateKey string
}

func NewReminderService(db *database.Database, vapid VAPIDConfig) *ReminderService {
	return &ReminderService{
		db:     db,
		vapid:  vapid,
	}
}

func (s *ReminderService) UpdateReminder(ctx context.Context, userID int, enabled bool) error {
	return s.db.UpdateReminder(ctx, userID, enabled)
}

func (s *ReminderService) GetReminder(ctx context.Context, userID int) (*models.Reminder, error) {
	return s.db.GetReminder(ctx, userID)
}

func (s *ReminderService) SendReminders(ctx context.Context) {
	users, err := s.db.GetUsersWithRemindersEnabled(ctx)
	if err != nil {
		log.Printf("Failed to get users with reminders: %v", err)
		return
	}

	now := time.Now().UTC()
	for _, user := range users {
		// Parse user timezone
		location, err := time.LoadLocation(user.Timezone)
		if err != nil {
			location = time.UTC
		}

		localTime := now.In(location)
		hour := localTime.Hour()

		// Check if it's 10AM or 9PM in user's timezone
		if hour != 10 && hour != 21 {
			continue
		}

		// Get push subscriptions for this user
		subs, err := s.db.GetPushSubscriptionsByUser(ctx, user.ID)
		if err != nil {
			log.Printf("Failed to get push subscriptions for user %d: %v", user.ID, err)
			continue
		}

		for _, sub := range subs {
			message := "Il est temps de pratiquer le français!"
			if hour == 10 {
				message = "Bonjour! N'oubliez pas de pratiquer le français aujourd'hui."
			} else {
				message = "Bonsoir! Avez-vous pratiqué le français aujourd'hui?"
			}

			if err := s.sendPushNotification(ctx, sub, message); err != nil {
				log.Printf("Failed to send push notification to user %d: %v", user.ID, err)
			}
		}
	}
}

func (s *ReminderService) sendPushNotification(ctx context.Context, sub *models.PushSubscription, message string) error {
	// This is a simplified version - in production you'd use a proper Web Push library
	// For now, we'll just log that we would send a notification
	
	log.Printf("Would send push notification to %s: %s", sub.Endpoint, message)
	
	// In a real implementation, you would:
	// 1. Create a JWT with the VAPID keys
	// 2. Build the push notification payload
	// 3. Send an HTTP POST request to the push service endpoint
	
	// For now, we'll just return nil to indicate success
	return nil
}

func (s *ReminderService) StorePushSubscription(ctx context.Context, userID int, req *models.PushSubscriptionRequest) (*models.PushSubscription, error) {
	// Check if subscription already exists
	subs, err := s.db.GetPushSubscriptionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check if this endpoint already exists
	for _, sub := range subs {
		if sub.Endpoint == req.Endpoint {
			// Update existing subscription
			if err := s.db.UpdatePushSubscription(ctx, sub.ID, req); err != nil {
				return nil, err
			}
			return sub, nil
		}
	}

	// Create new subscription
	sub := &models.PushSubscription{
		UserID:   userID,
		Endpoint: req.Endpoint,
		Auth:     req.Keys.Auth,
		P256DH:   req.Keys.P256DH,
	}

	if err := s.db.CreatePushSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *ReminderService) UpdatePushSubscription(ctx context.Context, id int, req *models.PushSubscriptionRequest) error {
	// Get existing subscription
	subs, err := s.db.GetAllPushSubscriptions(ctx)
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.ID == id {
			// Update the subscription
			_, err := s.db.Pool.Exec(ctx, `
				UPDATE push_subscriptions SET endpoint = $1, auth = $2, p256dh = $3 WHERE id = $4
			`, req.Endpoint, req.Keys.Auth, req.Keys.P256DH, id)
			return err
		}
	}

	return fmt.Errorf("subscription not found")
}

func (s *ReminderService) DeletePushSubscription(ctx context.Context, id int) error {
	return s.db.DeletePushSubscription(ctx, id)
}

func (s *ReminderService) GetPushSubscriptions(ctx context.Context, userID int) ([]*models.PushSubscription, error) {
	return s.db.GetPushSubscriptionsByUser(ctx, userID)
}

// Web Push implementation types

type PushMessage struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	Icon    string `json:"icon,omitempty"`
	Badges  string `json:"badge,omitempty"`
	Data    map[string]string `json:"data,omitempty"`
	Actions []PushAction `json:"actions,omitempty"`
	URL     string `json:"url,omitempty"`
}

type PushAction struct {
	Action string `json:"action"`
	Title  string `json:"title"`
	Icon   string `json:"icon,omitempty"`
}

func (s *ReminderService) SendWebPush(ctx context.Context, sub *models.PushSubscription, message *PushMessage) error {
	// This is where you would implement actual Web Push
	// For now, we'll just log it
	
	jsonData, _ := json.Marshal(message)
	log.Printf("Would send WebPush to %s: %s", sub.Endpoint, string(jsonData))
	
	return nil
}
