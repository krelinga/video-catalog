package internal

type MovieSource struct {
	Title       string `json:"title"`
	ReleaseYear *int   `json:"releaseYear,omitempty"`
	TmdbID      *int   `json:"tmdbId,omitempty"`
}

type MovieEditionSource struct {
	EditionType string `json:"editionType"`
}
