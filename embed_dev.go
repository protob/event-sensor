//go:build !release

package main

import (
	"io/fs"
	"testing/fstest"
)

const devNotice = `<!doctype html>
<meta charset="utf-8">
<title>Event Sensor - development build</title>
<h1>Frontend not embedded</h1>
<p>This binary was built without <code>-tags release</code>.</p>
<p>In development Vite serves the SPA on
<a href="http://localhost:5173/">http://localhost:5173/</a> and proxies /api to this
port. Build the embedded one with <code>just build</code>.</p>
`

// The default build embeds nothing, so a fresh clone compiles and `go test ./...` runs
// before any frontend build. In development Vite owns the SPA, so this page only answers
// someone pointing a browser at the backend port directly.
//
// fstest.MapFS is the standard library's only ready-made in-memory fs.FS and it drags in
// no test machinery: `go list -deps testing/fstest` does not include package testing.
var devFS = fstest.MapFS{
	"index.html": &fstest.MapFile{Data: []byte(devNotice)},
}

// frontendFS returns a one-page filesystem explaining that this is a development build.
func frontendFS() (fs.FS, error) { return devFS, nil }
