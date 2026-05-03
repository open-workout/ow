package service

import "github.com/open-workout/ow/services/exercise-service/internal/domain"

const primaryWeight = 1.2
const secondaryWeight = 1

func scoreExercise(ex *domain.ExerciseModel, state domain.MuscleState) float64 {
	score := 0.0
	count := 0.0

	// Primary muscle
	if val, ok := state.Muscles[ex.PrimaryMuscle]; ok {
		score += val * primaryWeight
		count += primaryWeight
	}

	for _, m := range ex.SecondaryMuscles {
		if val, ok := state.Muscles[m]; ok {
			score += val * secondaryWeight
			count += secondaryWeight
		}
	}

	if count == 0 {
		return 0
	}

	return score / count

}
