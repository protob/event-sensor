//go:build release

package spa

import (
	"embed"
	"io/fs"
)

// A missing dist is a compile error here, and that is the point: `just build` runs the
// frontend first, so a release binary cannot ship without the SPA. Vite writes here
// (frontend/vite.config.ts) because go:embed cannot reach outside its own package.
//
//go:embed dist/*
var frontendEmbed embed.FS

// frontendFS returns the SPA built by `just build-fe`.
func frontendFS() (fs.FS, error) { return fs.Sub(frontendEmbed, "dist") }
