package api

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/uuid"
)

// ListCategories returns all categories for the authenticated user with artist counts.
func (h *Handler) ListCategories(ctx context.Context, input *struct{}) (*ListCategoriesOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not found in context")
	}

	categories, err := h.queries.ListCategories(ctx, sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	resp := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		artists, err := h.queries.ListArtistsByCategory(ctx, c.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to list artists for category %s: %w", c.ID, err)
		}
		resp = append(resp, CategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			ArtistCount: len(artists),
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		})
	}

	return &ListCategoriesOutput{Body: resp}, nil
}

// CreateCategory creates a new category.
func (h *Handler) CreateCategory(ctx context.Context, input *CreateCategoryInput) (*CreateCategoryOutput, error) {
	id := uuid.New()
	cat, err := h.queries.CreateCategory(ctx, sqlc.CreateCategoryParams{
		ID:     id,
		Name:   input.Body.Name,
		UserID: sql.NullString{String: input.Body.UserID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &CreateCategoryOutput{
		Body: CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			ArtistCount: 0,
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		},
	}, nil
}

// UpdateCategory updates a category's name.
func (h *Handler) UpdateCategory(ctx context.Context, input *UpdateCategoryInput) (*UpdateCategoryOutput, error) {
	cat, err := h.queries.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		Name: input.Body.Name,
		ID:   input.ID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("category not found")
		}
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	artists, err := h.queries.ListArtistsByCategory(ctx, cat.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to count artists: %w", err)
	}

	return &UpdateCategoryOutput{
		Body: CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			ArtistCount: len(artists),
			CreatedAt:   cat.CreatedAt,
			UpdatedAt:   cat.UpdatedAt,
		},
	}, nil
}

// DeleteCategory deletes a category.
func (h *Handler) DeleteCategory(ctx context.Context, input *DeleteCategoryInput) (*DeleteCategoryOutput, error) {
	if err := h.queries.DeleteArtistCategoriesByCategory(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to clean up artist categories: %w", err)
	}

	if err := h.queries.DeleteCategory(ctx, input.ID); err != nil {
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}

	return &DeleteCategoryOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "category deleted"},
	}, nil
}
