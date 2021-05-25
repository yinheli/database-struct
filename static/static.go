package static

import "embed"

var (
	//go:embed assets template
	FS embed.FS
)
