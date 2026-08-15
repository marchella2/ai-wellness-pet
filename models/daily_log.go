package models

import "time"

type DailyLog struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID       string    `gorm:"type:uuid;not null;index" json:"user_id"`
	WaterGlasses int       `gorm:"not null;default:0" json:"water_glasses"`
	SleepHours   float64   `gorm:"type:numeric(3,1);not null;default:0" json:"sleep_hours"`
	JournalText  string    `gorm:"type:text" json:"journal_text"`
	CreatedAt    time.Time `json:"created_at"`
}
