package internal

import (
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
)

type WorkKind string

const (
	WorkKindMovie        WorkKind = "movie"
	WorkKindMovieEdition WorkKind = "movieEdition"
)

func (k WorkKind) IsValid() bool {
	switch k {
	case WorkKindMovie, WorkKindMovieEdition:
		return true
	default:
		return false
	}
}

type MovieWork struct {
	Title       string `json:"title"`
	ReleaseYear *int32   `json:"releaseYear,omitempty"`
	TmdbId      *int32   `json:"tmdbId,omitempty"`
}

// ToAPI converts the MovieWork to its API representation.
func (w *MovieWork) ToAPI() *vcrest.Movie {
	result := &vcrest.Movie{
		Title: nullable.NewNullableWithValue(w.Title),
	}
	if w.ReleaseYear != nil {
		result.ReleaseYear = nullable.NewNullableWithValue(*w.ReleaseYear)
	}
	if w.TmdbId != nil {
		result.TmdbId = nullable.NewNullableWithValue(*w.TmdbId)
	}
	return result
}

type MovieEditionWork struct {
	EditionType string `json:"editionType"`
}

// ToAPI converts the MovieEditionWork to its API representation.
func (w *MovieEditionWork) ToAPI() *vcrest.MovieEdition {
	return &vcrest.MovieEdition{
		EditionType: nullable.NewNullableWithValue(w.EditionType),
	}
}
