// Package studioassets embeds the static HTML, CSS, and ES2022 JavaScript assets
// for the Abeta Layout Studio SPA served at /studio.
package studioassets

import "embed"

// Files contains the embedded Visual Layout Studio frontend assets.
//
//go:embed index.html app.css app.js
var Files embed.FS
