package internal

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
	ReleaseYear *int   `json:"releaseYear,omitempty"`
	TmdbId      *int   `json:"tmdbId,omitempty"`
}

type MovieEditionWork struct {
	EditionType string `json:"editionType"`
}
