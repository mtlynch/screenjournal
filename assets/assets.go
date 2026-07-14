package assets

import (
	"embed"
	"io/fs"
)

//go:embed html
var files embed.FS

var HTMLFiles = sub(files, "html")

func sub(f embed.FS, dir string) fs.FS {
	sub, err := fs.Sub(f, dir)
	if err != nil {
		panic(err)
	}
	return sub
}
