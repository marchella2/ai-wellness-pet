package services

import (
	"testing"

	"pet-wellness-backend/models"
)

func TestEvaluateActivityAllBonuses(t *testing.T) {
	pet := &models.Pet{HealthScore: 50, EnergyScore: 50}

	result := EvaluateActivity(pet, ActivityInput{
		WaterGlasses: 4,
		SleepHours:   7.0,
		JournalText:  "hari ini menyenangkan",
	})

	if result.HealthScore != 75 {
		t.Errorf("expected health 75, got %d", result.HealthScore)
	}
	if result.EnergyScore != 85 {
		t.Errorf("expected energy 85, got %d", result.EnergyScore)
	}
	if result.CurrentState != models.MoodHappy {
		t.Errorf("expected mood Happy, got %s", result.CurrentState)
	}
}

func TestEvaluateActivityClampsScores(t *testing.T) {
	pet := &models.Pet{HealthScore: 95, EnergyScore: 90}

	result := EvaluateActivity(pet, ActivityInput{
		WaterGlasses: 4,
		SleepHours:   8.0,
		JournalText:  "sangat bahagia",
	})

	if result.HealthScore != 100 {
		t.Errorf("expected health clamped to 100, got %d", result.HealthScore)
	}
	if result.EnergyScore != 100 {
		t.Errorf("expected energy clamped to 100, got %d", result.EnergyScore)
	}
	if result.CurrentState != models.MoodHappy {
		t.Errorf("expected mood Happy, got %s", result.CurrentState)
	}
}

func TestEvaluateActivityNoBonusesLowState(t *testing.T) {
	pet := &models.Pet{HealthScore: 30, EnergyScore: 30}

	result := EvaluateActivity(pet, ActivityInput{
		WaterGlasses: 0,
		SleepHours:   4.0,
		JournalText:  "",
	})

	if result.HealthScore != 30 {
		t.Errorf("expected health 30, got %d", result.HealthScore)
	}
	if result.EnergyScore != 30 {
		t.Errorf("expected energy 30, got %d", result.EnergyScore)
	}
	if result.CurrentState != models.MoodTired {
		t.Errorf("expected mood Tired, got %s", result.CurrentState)
	}
}

func TestEvaluateActivitySadState(t *testing.T) {
	pet := &models.Pet{HealthScore: 20, EnergyScore: 80}

	result := EvaluateActivity(pet, ActivityInput{
		WaterGlasses: 0,
		SleepHours:   5.0,
		JournalText:  "",
	})

	if result.CurrentState != models.MoodSad {
		t.Errorf("expected mood Sad, got %s", result.CurrentState)
	}
}

func TestEvaluateActivityNeutralState(t *testing.T) {
	pet := &models.Pet{HealthScore: 40, EnergyScore: 50}

	result := EvaluateActivity(pet, ActivityInput{
		WaterGlasses: 0,
		SleepHours:   5.0,
		JournalText:  "",
	})

	if result.CurrentState != models.MoodNeutral {
		t.Errorf("expected mood Neutral, got %s", result.CurrentState)
	}
}

func TestDetermineMoodStateHappyTakesPriority(t *testing.T) {
	if mood := determineMoodState(90, 70); mood != models.MoodHappy {
		t.Errorf("expected Happy priority, got %s", mood)
	}
}
