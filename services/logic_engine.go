package services

import (
	"strings"

	"pet-wellness-backend/models"
)

type ActivityInput struct {
	WaterGlasses int
	SleepHours   float64
	JournalText  string
}

const (
	maxScore = 100
	minScore = 0

	waterTargetGlasses = 4
	sleepTargetHours   = 7.0

	waterHealthBonus   = 15
	sleepEnergyBonus   = 25
	journalHealthBonus = 10
	journalEnergyBonus = 10

	happyThreshold = 75
	tiredThreshold = 35
	sadThreshold   = 40
)

// EvaluateActivity applies wellness scoring rules to the pet based on the
// daily activity input and returns a new pet state.
func EvaluateActivity(pet *models.Pet, input ActivityInput) *models.Pet {
	health := pet.HealthScore
	energy := pet.EnergyScore

	if input.WaterGlasses >= waterTargetGlasses {
		health += waterHealthBonus
	}

	if input.SleepHours >= sleepTargetHours {
		energy += sleepEnergyBonus
	}

	if strings.TrimSpace(input.JournalText) != "" {
		health += journalHealthBonus
		energy += journalEnergyBonus
	}

	health = clamp(health)
	energy = clamp(energy)

	pet.HealthScore = health
	pet.EnergyScore = energy
	pet.CurrentState = determineMoodState(health, energy)

	return pet
}

func clamp(score int) int {
	if score < minScore {
		return minScore
	}
	if score > maxScore {
		return maxScore
	}
	return score
}

func determineMoodState(health, energy int) string {
	average := (health + energy) / 2

	switch {
	case average >= happyThreshold:
		return models.MoodHappy
	case energy < tiredThreshold:
		return models.MoodTired
	case health < sadThreshold:
		return models.MoodSad
	default:
		return models.MoodNeutral
	}
}
