package web

import (
	"embed"
	"net/http"
)

//go:embed index.html styles.css app.js
var content embed.FS

// Handler serves the capture UI.
func Handler() http.Handler {
	return http.FileServer(http.FS(content))
}

// FS exposes the embedded files.
func FS() embed.FS {
	return content
}
