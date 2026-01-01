package internal

type FileSource struct {
	Path string `json:"path"`
}

type DiscSource struct {
	OrigDirName   string `json:"origDirName"`
	Path          string `json:"path"`
	AllFilesAdded bool   `json:"allFilesAdded"`
}
