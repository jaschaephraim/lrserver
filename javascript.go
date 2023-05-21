package lrserver

import "embed"

//go:embed assets
var assets embed.FS

func getLivereloadJS() ([]byte, error) {
	return assets.ReadFile("assets/livereload_4.0.1.min.js")
}
