// Package desktop embeds the desktop frontend production build so the Wails
// backend can serve it. The frontend lives beside the backend (apps/desktop/
// frontend), so the embed must be declared from apps/desktop/ where the
// frontend/dist pattern stays within the package directory.
package desktop

import "embed"

//go:embed all:frontend/dist
var FrontendDist embed.FS
