package api

import (
	"context"
	"fmt"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/uuid"
)

// ListSettings returns all settings for the authenticated user.
func (h *Handler) ListSettings(ctx context.Context, input *struct{}) (*ListSettingsOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not found in context")
	}

	rows, err := h.queries.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}

	out := make([]SettingResponse, len(rows))
	for i, r := range rows {
		out[i] = toSettingResponse(r)
	}

	return &ListSettingsOutput{Body: out}, nil
}

// UpdateSettings bulk-upserts settings for the authenticated user.
func (h *Handler) UpdateSettings(ctx context.Context, input *UpdateSettingsInput) (*ListSettingsOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not found in context")
	}

	for _, s := range input.Body.Settings {
		_, err := h.queries.UpsertUserSetting(ctx, sqlc.UpsertUserSettingParams{
			ID:     uuid.New(),
			UserID: userID,
			Key:    s.Key,
			Value:  s.Value,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to upsert setting %q: %w", s.Key, err)
		}
	}

	rows, err := h.queries.GetUserSettings(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list settings: %w", err)
	}

	out := make([]SettingResponse, len(rows))
	for i, r := range rows {
		out[i] = toSettingResponse(r)
	}

	return &ListSettingsOutput{Body: out}, nil
}

func toSettingResponse(s sqlc.UserSetting) SettingResponse {
	return SettingResponse{
		Key:       s.Key,
		Value:     s.Value,
		UpdatedAt: s.UpdatedAt,
	}
}
