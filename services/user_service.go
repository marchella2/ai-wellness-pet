package services

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"pet-wellness-backend/models"
)

// ErrEmailAlreadyRegistered is returned when the email is already taken
// (unique constraint violation on users.email).
var ErrEmailAlreadyRegistered = errors.New("email already registered")

// UserService holds the user business logic (registration).
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register creates a new user row so that its id can later be used as the
// user_id for POST /pet/setup (which enforces a foreign key against users).
func (s *UserService) Register(name, email string) (*models.User, error) {
	user := &models.User{
		Name:  name,
		Email: email,
	}
	if err := s.db.Create(user).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyRegistered
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail looks up an existing user by email, used to make registration
// idempotent (so re-registering with the same email doesn't error out).
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := s.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
