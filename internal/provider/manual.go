package provider

import (
	"context"
	"errors"
	"strings"
)

// Manual is the degenerate intake provider for hand-entered events. No network, no
// filtering.
type Manual struct{}

func (Manual) Name() string { return "manual" }

// Intake normalizes a user-supplied event: forces Source="manual"/SourceID="", trims, and
// defaults Kind to "other". Name and StartDate are required. Free text (name, venue, lineup)
// is preserved verbatim as UTF-8 - only surrounding whitespace is trimmed.
func (Manual) Intake(_ context.Context, in *ProviderEvent) (*ProviderEvent, error) {
	if in == nil {
		return nil, errors.New("manual intake: nil event")
	}
	in.Source = "manual"
	in.SourceID = ""
	in.Name = strings.TrimSpace(in.Name)
	in.StartDate = strings.TrimSpace(in.StartDate)
	if in.Kind == "" {
		in.Kind = "other"
	}
	if in.Name == "" {
		return nil, errors.New("manual intake: name is required")
	}
	if in.StartDate == "" {
		return nil, errors.New("manual intake: start_date is required")
	}
	if in.Venue != nil {
		in.Venue.Source = "manual"
		in.Venue.SourceID = ""
	}
	return in, nil
}
