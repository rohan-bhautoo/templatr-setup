package main

import "embed"

// WebAssets contains the built React web UI assets.
// The web/dist directory is populated by running `cd web && npm run build`.
// When the directory doesn't exist (development without web UI), the embedded
// filesystem will be empty and the server will return a fallback page.
//
//go:embed all:web/dist
var WebAssets embed.FS
