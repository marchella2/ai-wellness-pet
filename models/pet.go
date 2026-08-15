package models

import "time"

const (
	MoodHappy   = "Happy"
	MoodTired   = "Tired"
	MoodSad     = "Sad"
	MoodNeutral = "Neutral"

	DefaultHealthScore = 50
	DefaultEnergyScore = 50
	DefaultState       = MoodNeutral

	NeglectedHealthScore = 20
	NeglectedEnergyScore = 20
	NeglectedState       = MoodSad
)

type Pet struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"type:uuid;not null;unique" json:"user_id"`
	PetName      string    `gorm:"size:50;not null;default:'Milo'" json:"pet_name"`
	HealthScore  int       `gorm:"not null;default:50" json:"health_score"`
	EnergyScore  int       `gorm:"not null;default:50" json:"energy_score"`
	CurrentState string    `gorm:"size:20;not null;default:'Neutral'" json:"current_state"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
