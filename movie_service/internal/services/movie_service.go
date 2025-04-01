package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iamsuteerth/skyfox-helper/tree/main/movie_service/internal/models"
)

type MovieService struct {
	movies   []models.Movie
	dataPath string
}

func NewMovieService(dataPath string) (*MovieService, error) {
	service := &MovieService{
		dataPath: dataPath,
	}

	if err := service.loadMovies(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *MovieService) loadMovies() error {
	data, err := os.ReadFile(s.dataPath)
	if err != nil {
		possiblePaths := []string{
			"data/movies.json",
			"../data/movies.json",
			"../../data/movies.json",
			filepath.Join(filepath.Dir(os.Args[0]), "data/movies.json"),
		}
		var readErr error
		for _, path := range possiblePaths {
			data, readErr = os.ReadFile(path)
			if readErr == nil {
				s.dataPath = path
				break
			}
		}
		if readErr != nil {
			return fmt.Errorf("failed to read movies data from any location: %w", err)
		}
	}
	if err := json.Unmarshal(data, &s.movies); err != nil {
		return fmt.Errorf("failed to parse movies JSON: %w", err)
	}
	return nil
}

func (s *MovieService) GetAllMovies() []models.Movie {
	return s.movies
}

func (s *MovieService) GetMovieByID(id string) (models.Movie, bool) {
	for _, movie := range s.movies {
		if movie.ImdbID == id {
			return movie, true
		}
	}
	return models.Movie{}, false
}
