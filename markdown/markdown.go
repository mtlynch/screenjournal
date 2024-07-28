package markdown

import (
	"strings"

	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	gomarkdown_parser "github.com/gomarkdown/markdown/parser"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func RenderBlurb(blurb screenjournal.Blurb) string {
	return renderUntrusted(blurb.String())
}

func RenderComment(comment screenjournal.CommentText) string {
	return renderUntrusted(comment.String())
}

func renderUntrusted(s string) string {
	parser := gomarkdown_parser.New()

	renderer := html.NewRenderer(html.RendererOptions{Flags: html.SkipHTML | html.SkipImages | html.SkipLinks})

	html := string(gomarkdown.ToHTML([]byte(s), parser, renderer))

	return strings.TrimSpace(html)
}

func RenderEmail(email string) string {
	parser := gomarkdown_parser.New()

	renderer := html.NewRenderer(html.RendererOptions{Flags: html.SkipHTML | html.SkipImages})

	html := string(gomarkdown.ToHTML([]byte(email), parser, renderer))

	return strings.TrimSpace(html)
}
