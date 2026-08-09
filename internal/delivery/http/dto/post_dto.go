package dto

import (
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/entity"
	"github.com/pisaupediaprojek/pisau-pedia-backend/internal/usecase"
)

// ---------- PostCategory ----------

type CreatePostCategoryRequest struct {
	Slug   string  `json:"slug" validate:"required"`
	NameID string  `json:"name_id" validate:"required"`
	NameEN string  `json:"name_en" validate:"required"`
	DescID *string `json:"desc_id"`
	DescEN *string `json:"desc_en"`
}

func (r CreatePostCategoryRequest) ToInput() usecase.PostCategoryInput {
	return usecase.PostCategoryInput{
		Slug:   r.Slug,
		NameID: r.NameID,
		NameEN: r.NameEN,
		DescID: r.DescID,
		DescEN: r.DescEN,
	}
}

type UpdatePostCategoryRequest struct {
	Slug   string  `json:"slug" validate:"required"`
	NameID string  `json:"name_id" validate:"required"`
	NameEN string  `json:"name_en" validate:"required"`
	DescID *string `json:"desc_id"`
	DescEN *string `json:"desc_en"`
}

func (r UpdatePostCategoryRequest) ToInput() usecase.PostCategoryInput {
	return usecase.PostCategoryInput{
		Slug:   r.Slug,
		NameID: r.NameID,
		NameEN: r.NameEN,
		DescID: r.DescID,
		DescEN: r.DescEN,
	}
}

type LocalizedString struct {
	ID string `json:"id"`
	EN string `json:"en"`
}

type LocalizedNullableString struct {
	ID *string `json:"id"`
	EN *string `json:"en"`
}

type PostCategoryResponse struct {
	ID          string                  `json:"id"`
	Slug        string                  `json:"slug"`
	Name        LocalizedString         `json:"name"`
	Description LocalizedNullableString `json:"description"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

func ToPostCategoryResponse(c *entity.PostCategory) PostCategoryResponse {
	return PostCategoryResponse{
		ID:   c.ID,
		Slug: c.Slug,
		Name: LocalizedString{ID: c.NameID, EN: c.NameEN},
		Description: LocalizedNullableString{ID: c.DescID, EN: c.DescEN},
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToPostCategoryListResponse(cats []entity.PostCategory) []PostCategoryResponse {
	out := make([]PostCategoryResponse, 0, len(cats))
	for i := range cats {
		out = append(out, ToPostCategoryResponse(&cats[i]))
	}
	return out
}

// ---------- Post ----------

type CreatePostRequest struct {
	Slug           string  `json:"slug" validate:"required"`
	CategoryID     string  `json:"category_id" validate:"required"`
	TitleID        string  `json:"title_id" validate:"required"`
	TitleEN        string  `json:"title_en" validate:"required"`
	ExcerptID      *string `json:"excerpt_id"`
	ExcerptEN      *string `json:"excerpt_en"`
	ContentID      *string `json:"content_id"`
	ContentEN      *string `json:"content_en"`
	Image          *string `json:"image"`
	ReadingMinutes int     `json:"reading_minutes"`
	Status         string  `json:"status"`
	PublishedAt    *string `json:"published_at"`
}

func (r CreatePostRequest) ToInput() usecase.PostInput {
	return usecase.PostInput{
		Slug:           r.Slug,
		CategoryID:     r.CategoryID,
		TitleID:        r.TitleID,
		TitleEN:        r.TitleEN,
		ExcerptID:      r.ExcerptID,
		ExcerptEN:      r.ExcerptEN,
		ContentID:      r.ContentID,
		ContentEN:      r.ContentEN,
		Image:          r.Image,
		ReadingMinutes: r.ReadingMinutes,
		Status:         r.Status,
		PublishedAt:    r.PublishedAt,
	}
}

type UpdatePostRequest struct {
	Slug           string  `json:"slug" validate:"required"`
	CategoryID     string  `json:"category_id" validate:"required"`
	TitleID        string  `json:"title_id" validate:"required"`
	TitleEN        string  `json:"title_en" validate:"required"`
	ExcerptID      *string `json:"excerpt_id"`
	ExcerptEN      *string `json:"excerpt_en"`
	ContentID      *string `json:"content_id"`
	ContentEN      *string `json:"content_en"`
	Image          *string `json:"image"`
	ReadingMinutes int     `json:"reading_minutes"`
	Status         string  `json:"status"`
	PublishedAt    *string `json:"published_at"`
}

func (r UpdatePostRequest) ToInput() usecase.PostInput {
	return usecase.PostInput{
		Slug:           r.Slug,
		CategoryID:     r.CategoryID,
		TitleID:        r.TitleID,
		TitleEN:        r.TitleEN,
		ExcerptID:      r.ExcerptID,
		ExcerptEN:      r.ExcerptEN,
		ContentID:      r.ContentID,
		ContentEN:      r.ContentEN,
		Image:          r.Image,
		ReadingMinutes: r.ReadingMinutes,
		Status:         r.Status,
		PublishedAt:    r.PublishedAt,
	}
}

type PostResponse struct {
	ID             string                  `json:"id"`
	Slug           string                  `json:"slug"`
	CategorySlug   string                  `json:"category_slug"`
	CategoryName   string                  `json:"category_name"`
	Title          LocalizedString         `json:"title"`
	Excerpt        LocalizedNullableString `json:"excerpt"`
	Content        LocalizedNullableString `json:"content"`
	Image          *string                 `json:"image,omitempty"`
	ReadingMinutes int                     `json:"reading_minutes"`
	Status         string                  `json:"status"`
	PublishedAt    *string                 `json:"published_at,omitempty"`
	CreatedAt      string                  `json:"created_at"`
}

func ToPostResponse(p *entity.Post) PostResponse {
	var publishedAt *string
	if p.PublishedAt != nil {
		s := p.PublishedAt.Format("2006-01-02T15:04:05Z")
		publishedAt = &s
	}
	return PostResponse{
		ID:           p.ID,
		Slug:         p.Slug,
		CategorySlug: p.CategorySlug,
		CategoryName: p.CategoryName,
		Title:        LocalizedString{ID: p.TitleID, EN: p.TitleEN},
		Excerpt:      LocalizedNullableString{ID: p.ExcerptID, EN: p.ExcerptEN},
		Content:      LocalizedNullableString{ID: p.ContentID, EN: p.ContentEN},
		Image:          p.Image,
		ReadingMinutes: p.ReadingMinutes,
		Status:         p.Status,
		PublishedAt:    publishedAt,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToPostListResponse(posts []entity.Post) []PostResponse {
	out := make([]PostResponse, 0, len(posts))
	for i := range posts {
		out = append(out, ToPostResponse(&posts[i]))
	}
	return out
}
