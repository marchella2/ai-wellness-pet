package services

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"pet-wellness-backend/models"
)

// ErrUserNotRegistered is returned when a pet cannot be created because the
// user does not exist in the users table (foreign key violation).
var ErrUserNotRegistered = errors.New("user not registered")

// PetService holds the pet business logic (read, setup, reset, simulate).
type PetService struct {
	db *gorm.DB
}

func NewPetService(db *gorm.DB) *PetService {
	return &PetService{db: db}
}

// GetByUserID returns the most recent pet belonging to the given user.
func (s *PetService) GetByUserID(userID string) (*models.Pet, error) {
	var pet models.Pet
	err := s.db.Where("user_id = ?", userID).Order("id DESC").First(&pet).Error
	return &pet, err
}

// Setup creates a default pet if none exists for the user, or renames the
// existing pet. The pet name is mandatory.
func (s *PetService) Setup(userID, petName string) (*models.Pet, error) {
	pet, err := s.GetByUserID(userID)
	switch {
	case err == nil:
		pet.PetName = petName
		if err := s.db.Model(pet).Update("pet_name", pet.PetName).Error; err != nil {
			return nil, err
		}
		return pet, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		pet = &models.Pet{
			UserID:       userID,
			PetName:      petName,
			HealthScore:  models.DefaultHealthScore,
			EnergyScore:  models.DefaultEnergyScore,
			CurrentState: models.DefaultState,
		}
		if err := s.db.Create(pet).Error; err != nil {
			if isForeignKeyViolation(err) {
				return nil, ErrUserNotRegistered
			}
			return nil, err
		}
		return pet, nil
	default:
		return nil, err
	}
}

// Reset restores the pet to its default state without deleting daily logs.
func (s *PetService) Reset(userID string) (*models.Pet, error) {
	pet, err := s.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	pet.HealthScore = models.DefaultHealthScore
	pet.EnergyScore = models.DefaultEnergyScore
	pet.CurrentState = models.DefaultState

	if err := s.db.Save(pet).Error; err != nil {
		return nil, err
	}
	return pet, nil
}

// SimulateNeglect forces the pet into a sad, neglected state for demos.
func (s *PetService) SimulateNeglect(userID string) (*models.Pet, error) {
	pet, err := s.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	pet.HealthScore = models.NeglectedHealthScore
	pet.EnergyScore = models.NeglectedEnergyScore
	pet.CurrentState = models.NeglectedState

	if err := s.db.Save(pet).Error; err != nil {
		return nil, err
	}
	return pet, nil
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
