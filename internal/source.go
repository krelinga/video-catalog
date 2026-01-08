package internal

import (
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
)

type SourceKind string

const (
	SourceKindFile SourceKind = "file"
	SourceKindDisc SourceKind = "disc"
)

func (k SourceKind) IsValid() bool {
	switch k {
	case SourceKindFile, SourceKindDisc:
		return true
	default:
		return false
	}
}

type FileSource struct {
	Path string `json:"path"`
}

// ToAPI converts the FileSource to its API representation.
func (s *FileSource) ToAPI() *vcrest.File {
	return &vcrest.File{
		Path: nullable.NewNullableWithValue(s.Path),
	}
}

type DiscSource struct {
	OrigDirName   string `json:"origDirName"`
	Path          string `json:"path"`
	AllFilesAdded bool   `json:"allFilesAdded"`
}

// ToAPI converts the DiscSource to its API representation.
func (s *DiscSource) ToAPI() *vcrest.Disc {
	return &vcrest.Disc{
		OrigDirName:   nullable.NewNullableWithValue(s.OrigDirName),
		Path:          nullable.NewNullableWithValue(s.Path),
		AllFilesAdded: nullable.NewNullableWithValue(s.AllFilesAdded),
	}
}
