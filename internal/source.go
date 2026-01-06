package internal

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

type DiscSource struct {
	OrigDirName   string `json:"origDirName"`
	Path          string `json:"path"`
	AllFilesAdded bool   `json:"allFilesAdded"`
}
