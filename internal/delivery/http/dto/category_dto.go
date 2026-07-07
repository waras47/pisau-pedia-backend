package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=2"`
	Slug        string  `json:"slug" validate:"omitempty"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url"`
}

type UpdateCategoryRequest struct {
	Name        string  `json:"name" validate:"omitempty,min=2"`
	Description *string `json:"description"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url"`
}

func (r CreateCategoryRequest) ToInput() usecase.CategoryInput {
	return usecase.CategoryInput{Name: r.Name, Slug: r.Slug, Description: r.Description, ImageURL: r.ImageURL}
}

func (r UpdateCategoryRequest) ToInput() usecase.CategoryInput {
	return usecase.CategoryInput{Name: r.Name, Description: r.Description, ImageURL: r.ImageURL}
}

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
}

func ToCategoryResponse(c *entity.Category) CategoryResponse {
	resp := CategoryResponse{ID: c.ID, Name: c.Name, Slug: c.Slug}
	if c.Description != nil {
		resp.Description = *c.Description
	}
	if c.ImageURL != nil {
		resp.ImageURL = *c.ImageURL
	}
	return resp
}

func ToCategoryResponses(categories []entity.Category) []CategoryResponse {
	out := make([]CategoryResponse, 0, len(categories))
	for i := range categories {
		out = append(out, ToCategoryResponse(&categories[i]))
	}
	return out
}
