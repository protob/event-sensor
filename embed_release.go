//go:build release

package main

import (
	"embed"
	"io/fs"
)

// A missing frontend/dist is a compile error here, and that is the point: `just build`
// runs the frontend first, so a release binary cannot ship without the SPA.
//
//go:embed frontend/dist/*
var frontendEmbed embed.FS

// frontendFS returns the SPA built by `just build-fe`.
func frontendFS() (fs.FS, error) { return fs.Sub(frontendEmbed, "frontend/dist") }
