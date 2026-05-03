package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type LocalMediaStorage struct {
	BaseDir string // e.g. "/var/www/uploads"
	BaseURL string // e.g. "https://yourserver.com/uploads"
}

func NewLocalMediaStorage(baseDir, baseURL string) (*LocalMediaStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating upload directory: %w", err)
	}
	return &LocalMediaStorage{BaseDir: baseDir, BaseURL: baseURL}, nil
}

func (s *LocalMediaStorage) Upload(ctx context.Context, upload *domain.ExerciseMediaUpload) (string, error) {
	filename := fmt.Sprintf("%d_%d_%s", upload.UserID, upload.ExerciseID, upload.Filename)
	destPath := filepath.Join(s.BaseDir, filename)

	out, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("creating file: %w", err)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}(out)

	if _, err := io.Copy(out, upload.File); err != nil {
		return "", fmt.Errorf("writing file: %w", err)
	}

	url := fmt.Sprintf("%s/%s", s.BaseURL, filename)
	return url, nil
}
