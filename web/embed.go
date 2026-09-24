// Package web embeds the strike-fw room console build output.
package web

import "embed"

// Dist holds the Vite build under dist/ (placeholder until npm run build).
//
//go:embed all:dist
var Dist embed.FS
