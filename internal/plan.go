package internal

import "github.com/google/uuid"

type PlanKind string

const (
	PlanKindDirect       PlanKind = "direct"
	PlanKindChapterRange PlanKind = "chapter_range"
)

func (k PlanKind) IsValid() bool {
	switch k {
	case PlanKindDirect, PlanKindChapterRange:
		return true
	default:
		return false
	}
}

type DirectPlan struct {
	SourceUUID uuid.UUID `json:"sourceUuid"`
	WorkUUID   uuid.UUID `json:"workUuid"`
}

type ChapterRangePlan struct {
	SourceUUID   uuid.UUID `json:"sourceUuid"`
	WorkUUID     uuid.UUID `json:"workUuid"`
	StartChapter *int      `json:"startChapter,omitempty"`
	EndChapter   *int      `json:"endChapter,omitempty"`
}
