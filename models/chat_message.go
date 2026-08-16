package models

import "time"

const (
	ChatRoleUser = "user"
	ChatRoleMilo = "milo"
)

type ChatMessage struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	Role      string    `gorm:"size:10;not null" json:"role"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
