package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

type CreateCollectionRequest struct {
	Name        string   `json:"name" validate:"required,min=2"`
	Slug        string   `json:"slug" validate:"omitempty"`
	Description *string  `json:"description"`
	ImageURL    *string  `json:"image_url" validate:"omitempty,url"`
	SortOrder   int      `json:"sort_order"`
	CategoryIDs []string `json:"category_ids"`
}

type UpdateCollectionRequest struct {
	Name        string   `json:"name" validate:"omitempty,min=2"`
	Description *string  `json:"description"`
	ImageURL    *string  `json:"image_url" validate:"omitempty,url"`
	SortOrder   int      `json:"sort_order"`
	CategoryIDs []string `json:"category_ids"`
}

func (r CreateCollectionRequest) ToInput() usecase.CollectionInput {
	return usecase.CollectionInput{
		Name: r.Name, Slug: r.Slug, Description: r.Description,
		ImageURL: r.ImageURL, SortOrder: r.SortOrder, CategoryIDs: r.CategoryIDs,
	}
}

func (r UpdateCollectionRequest) ToInput() usecase.CollectionInput {
	return usecase.CollectionInput{
		Name: r.Name, Description: r.Description,
		ImageURL: r.ImageURL, SortOrder: r.SortOrder, CategoryIDs: r.CategoryIDs,
	}
}

type CollectionResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	ImageURL    string             `json:"image_url,omitempty"`
	SortOrder   int                `json:"sort_order"`
	Categories  []CategoryResponse `json:"categories"`
}

func ToCollectionResponse(c *entity.Collection) CollectionResponse {
	resp := CollectionResponse{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		SortOrder: c.SortOrder,
	}
	if c.Description != nil {
		resp.Description = *c.Description
	}
	if c.ImageURL != nil {
		resp.ImageURL = *c.ImageURL
	}
	resp.Categories = ToCategoryResponses(c.Categories)
	return resp
}

func ToCollectionResponses(collections []entity.Collection) []CollectionResponse {
	out := make([]CollectionResponse, 0, len(collections))
	for i := range collections {
		out = append(out, ToCollectionResponse(&collections[i]))
	}
	return out
}
