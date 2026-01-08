package internal

import (
	"github.com/google/uuid"
	"github.com/krelinga/video-catalog/vcrest"
	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

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

// ToAPI converts the DirectPlan to its API representation.
func (p *DirectPlan) ToAPI() *vcrest.DirectPlan {
	return &vcrest.DirectPlan{
		SourceUuid: nullable.NewNullableWithValue(openapi_types.UUID(p.SourceUUID)),
		WorkUuid:   nullable.NewNullableWithValue(openapi_types.UUID(p.WorkUUID)),
	}
}

type ChapterRangePlan struct {
	SourceUUID   uuid.UUID `json:"sourceUuid"`
	WorkUUID     uuid.UUID `json:"workUuid"`
	StartChapter *int      `json:"startChapter,omitempty"`
	EndChapter   *int      `json:"endChapter,omitempty"`
}

// ToAPI converts the ChapterRangePlan to its API representation.
func (p *ChapterRangePlan) ToAPI() *vcrest.ChapterRangePlan {
	result := &vcrest.ChapterRangePlan{
		SourceUuid: nullable.NewNullableWithValue(openapi_types.UUID(p.SourceUUID)),
		WorkUuid:   nullable.NewNullableWithValue(openapi_types.UUID(p.WorkUUID)),
	}
	if p.StartChapter != nil {
		result.StartChapter = nullable.NewNullableWithValue(int32(*p.StartChapter))
	}
	if p.EndChapter != nil {
		result.EndChapter = nullable.NewNullableWithValue(int32(*p.EndChapter))
	}
	return result
}
