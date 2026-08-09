package dto

import (
	"encoding/json"

	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type UpsertSiteContentRequest struct {
	Key   string          `json:"key" validate:"required"`
	Value json.RawMessage `json:"value" validate:"required"`
}

func (r UpsertSiteContentRequest) ToInput() usecase.SiteContentInput {
	return usecase.SiteContentInput{
		Key:   r.Key,
		Value: string(r.Value),
	}
}

type SiteContentResponse struct {
	ID        string          `json:"id"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func ToSiteContentResponse(c *entity.SiteContent) SiteContentResponse {
	return SiteContentResponse{
		ID:        c.ID,
		Key:       c.Key,
		Value:     json.RawMessage(c.Value),
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToSiteContentResponses(contents []entity.SiteContent) []SiteContentResponse {
	out := make([]SiteContentResponse, 0, len(contents))
	for i := range contents {
		out = append(out, ToSiteContentResponse(&contents[i]))
	}
	return out
}
