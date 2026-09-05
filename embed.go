package projectroot

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:ui/build
var frontendFS embed.FS

// GetFrontendFS returnerer SvelteKit-filerne klar til brug
func GetFrontendFS() http.FileSystem {
	strippedFS, err := fs.Sub(frontendFS, "ui/build")
	if err != nil {
		panic(err)
	}
	return http.FS(strippedFS)
}
