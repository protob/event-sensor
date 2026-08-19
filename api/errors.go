package api

import (
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func huma404(msg string) error {
	return huma.Error404NotFound(msg)
}

func huma400(msg string) error {
	return huma.Error400BadRequest(msg)
}

func huma409(msg string) error {
	return huma.Error409Conflict(msg)
}

func huma403(msg string) error {
	return huma.Error403Forbidden(msg)
}

// isForeignKeyErr matches the SQLite error an ON DELETE RESTRICT raises when deleting an
// event that still has a library claim.
func isForeignKeyErr(err error) bool {
	return err != nil && strings.Contains(strings.ToUpper(err.Error()), "FOREIGN KEY CONSTRAINT")
}
