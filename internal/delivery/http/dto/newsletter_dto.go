package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type SubscribeRequest struct {
	Email  string  `json:"email" validate:"required,email"`
	Name   *string `json:"name"`
	Source string  `json:"source" validate:"omitempty,oneof=checkout footer popup blog manual"`
}

func (r SubscribeRequest) ToInput() usecase.SubscribeInput {
	return usecase.SubscribeInput{Email: r.Email, Name: r.Name, Source: r.Source}
}

type SubscriberResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	Source    string `json:"source"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func ToSubscriberResponse(s *entity.NewsletterSubscriber) SubscriberResponse {
	resp := SubscriberResponse{
		ID:        s.ID,
		Email:     s.Email,
		Source:    string(s.Source),
		Status:    string(s.Status),
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if s.Name != nil {
		resp.Name = *s.Name
	}
	return resp
}

func ToSubscriberResponses(subs []entity.NewsletterSubscriber) []SubscriberResponse {
	out := make([]SubscriberResponse, 0, len(subs))
	for i := range subs {
		out = append(out, ToSubscriberResponse(&subs[i]))
	}
	return out
}
