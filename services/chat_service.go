package services

import (
	"context"

	"gorm.io/gorm"

	"pet-wellness-backend/config"
	"pet-wellness-backend/models"
)

// chatHistoryLimit caps how many past messages are sent to Gemini as
// context for each turn, keeping prompts small and responses fast.
const chatHistoryLimit = 20

// ChatService holds the free-form conversation business logic (persist +
// Gemini call, aware of the pet's current mood for a consistent persona).
type ChatService struct {
	db     *gorm.DB
	env    *config.Env
	petSvc *PetService
}

func NewChatService(db *gorm.DB, env *config.Env, petService *PetService) *ChatService {
	return &ChatService{db: db, env: env, petSvc: petService}
}

// SendMessage stores the owner's message, asks Gemini for Milo's reply given
// the recent conversation and the pet's mood, stores that reply, and returns
// it. The pet must already exist (set up first).
func (s *ChatService) SendMessage(ctx context.Context, userID, message string) (string, error) {
	pet, err := s.petSvc.GetByUserID(userID)
	if err != nil {
		return "", err
	}

	var history []models.ChatMessage
	if err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(chatHistoryLimit).
		Find(&history).Error; err != nil {
		return "", err
	}
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	userMsg := models.ChatMessage{UserID: userID, Role: models.ChatRoleUser, Content: message}
	if err := s.db.Create(&userMsg).Error; err != nil {
		return "", err
	}

	reply, err := GenerateChatReply(ctx, s.env.GeminiAPIKey, pet.CurrentState, pet.HealthScore, pet.EnergyScore, history, message)
	if err != nil {
		return "", err
	}

	miloMsg := models.ChatMessage{UserID: userID, Role: models.ChatRoleMilo, Content: reply}
	if err := s.db.Create(&miloMsg).Error; err != nil {
		return "", err
	}

	return reply, nil
}

// GetHistory returns the full conversation for a user, oldest first.
func (s *ChatService) GetHistory(userID string) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := s.db.Where("user_id = ?", userID).
		Order("created_at ASC").
		Find(&messages).Error
	return messages, err
}
