package markdown

import (
	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	gomarkdown_parser "github.com/gomarkdown/markdown/parser"
)

func Render(markdown string) string {
	parser := gomarkdown_parser.New()

	renderer := html.NewRenderer(html.RendererOptions{Flags: html.FlagsNone})

	return string(gomarkdown.ToHTML([]byte(markdown), parser, renderer))
}
