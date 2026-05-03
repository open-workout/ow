package storage

import (
	"context"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type MockMediaStorage struct {
	Called     bool
	UploadFunc func(_ context.Context, file *domain.ExerciseMediaUpload) (string, error)
}

func (m MockMediaStorage) Upload(ctx context.Context, file *domain.ExerciseMediaUpload) (string, error) {
	m.Called = true
	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, file)
	}
	return "", nil
}
