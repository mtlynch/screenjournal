package markdown

import (
	"strings"

	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	gomarkdown_parser "github.com/gomarkdown/markdown/parser"
)

// TODO: Change this to FromBlurb so that it's specific to the typed Blurb, then
// add a different one for trusted strings.
func Render(markdown string) string {
	parser := gomarkdown_parser.New()

	renderer := html.NewRenderer(html.RendererOptions{Flags: html.SkipHTML | html.SkipImages | html.SkipLinks})

	html := string(gomarkdown.ToHTML([]byte(markdown), parser, renderer))

	return strings.TrimSpace(html)
}
