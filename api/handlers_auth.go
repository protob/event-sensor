package api

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/protob/event-sensor/db/sqlc"
	"github.com/protob/event-sensor/internal/auth"
)

const tokenExpiry = 7 * 24 * time.Hour

// Login authenticates a user and returns a JWT token.
func (h *Handler) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	user, err := h.queries.GetUserByUsername(ctx, input.Body.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma400("invalid username or password")
		}
		return nil, fmt.Errorf("failed to look up user: %w", err)
	}

	if err := auth.CheckPassword(user.Password, input.Body.Password); err != nil {
		return nil, huma400("invalid username or password")
	}

	token, err := auth.GenerateToken(user.ID, user.Username, h.config.JWTSecret, tokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginOutput{
		Body: AuthResponseBody{
			Token: token,
			User:  toUserResponse(user),
		},
	}, nil
}

// Me returns the current authenticated user's info.
func (h *Handler) Me(ctx context.Context, input *struct{}) (*MeOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not found in context")
	}

	user, err := h.queries.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &MeOutput{
		Body: toUserResponse(user),
	}, nil
}

// ChangePassword changes the authenticated user's password.
func (h *Handler) ChangePassword(ctx context.Context, input *ChangePasswordInput) (*ChangePasswordOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not found in context")
	}

	user, err := h.queries.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := auth.CheckPassword(user.Password, input.Body.CurrentPassword); err != nil {
		return nil, huma400("current password is incorrect")
	}

	hash, err := auth.HashPassword(input.Body.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := h.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		Password: hash,
		ID:       userID,
	}); err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	return &ChangePasswordOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "password changed successfully"},
	}, nil
}

// UpdateProfile updates the authenticated user's profile (username and/or email).
func (h *Handler) UpdateProfile(ctx context.Context, input *UpdateProfileInput) (*UpdateProfileOutput, error) {
	userID := UserIDFromContext(ctx)
	if userID == "" {
		return nil, huma400("user not found in context")
	}

	user, err := h.queries.GetUser(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma404("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	newUsername := input.Body.Username
	if newUsername == "" {
		newUsername = user.Username
	}
	newEmail := input.Body.Email
	if newEmail == "" {
		newEmail = user.Email
	}

	updated, err := h.queries.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		Username: newUsername,
		Email:    newEmail,
		ID:       userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &UpdateProfileOutput{Body: toUserResponse(updated)}, nil
}

// ResetPassword resets a user's password by verifying username + email. The pair is
// only a real secret once the operator has set their own credentials; while the
// seeded account still authenticates with "password", both values are public
// constants in the repo, so off-loopback the flow refuses rather than hand over
// the account to whoever read the seed.
func (h *Handler) ResetPassword(ctx context.Context, input *ResetPasswordInput) (*ResetPasswordOutput, error) {
	user, err := h.queries.GetUserByUsername(ctx, input.Body.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, huma400("no account found with that username and email")
		}
		return nil, fmt.Errorf("failed to look up user: %w", err)
	}

	if user.Email != input.Body.Email {
		return nil, huma400("no account found with that username and email")
	}

	if !h.config.IsLoopbackBind() && auth.CheckPassword(user.Password, "password") == nil {
		return nil, huma403("password reset is disabled until the default seeded password is changed")
	}

	hash, err := auth.HashPassword(input.Body.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := h.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		Password: hash,
		ID:       user.ID,
	}); err != nil {
		return nil, fmt.Errorf("failed to reset password: %w", err)
	}

	return &ResetPasswordOutput{
		Body: struct {
			Message string `json:"message"`
		}{Message: "password reset successfully"},
	}, nil
}

func toUserResponse(u sqlc.User) UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		Active:    u.Active == 1,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
