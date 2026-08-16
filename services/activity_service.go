package services

import (
	"context"

	"gorm.io/gorm"

	"pet-wellness-backend/config"
	"pet-wellness-backend/models"
)

// ActivityService holds the activity business logic (daily log + logic engine + AI).
type ActivityService struct {
	db     *gorm.DB
	env    *config.Env
	petSvc *PetService
}

func NewActivityService(db *gorm.DB, env *config.Env, petService *PetService) *ActivityService {
	return &ActivityService{db: db, env: env, petSvc: petService}
}

// ActivityLogInput is the validated input of a daily activity submission.
type ActivityLogInput struct {
	UserID       string
	WaterGlasses int
	SleepHours   float64
	JournalText  string
}

// LogActivity stores the daily log, recalculates the pet scores and mood via
// the logic engine, persists the updated pet, and asks the AI companion for an
// empathetic reply. The pet must already exist (set up first).
func (s *ActivityService) LogActivity(ctx context.Context, input ActivityLogInput) (*models.Pet, string, error) {
	p, err := s.petSvc.GetByUserID(input.UserID)
	if err != nil {
		return nil, "", err
	}

	dailyLog := models.DailyLog{
		UserID:       input.UserID,
		WaterGlasses: input.WaterGlasses,
		SleepHours:   input.SleepHours,
		JournalText:  input.JournalText,
	}
	if err := s.db.Create(&dailyLog).Error; err != nil {
		return nil, "", err
	}

	p = EvaluateActivity(p, ActivityInput{
		WaterGlasses: input.WaterGlasses,
		SleepHours:   input.SleepHours,
		JournalText:  input.JournalText,
	})
	if err := s.db.Save(p).Error; err != nil {
		return nil, "", err
	}

	aiMessage, err := GenerateAIResponse(ctx, s.env.GeminiAPIKey, p.PetName, p.CurrentState, p.HealthScore, p.EnergyScore, input.JournalText)
	if err != nil {
		return nil, "", err
	}

	return p, aiMessage, nil
}

// GetHistory returns the 10 most recent daily logs of the given user.
func (s *ActivityService) GetHistory(userID string) ([]models.DailyLog, error) {
	var logs []models.DailyLog
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&logs).Error
	return logs, err
}
