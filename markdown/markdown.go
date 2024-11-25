package markdown

import (
	"strings"

	gomarkdown "github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	gomarkdown_parser "github.com/gomarkdown/markdown/parser"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	untrustedRenderer *html.Renderer
	trustedRenderer   *html.Renderer
)

func init() {

	untrustedRenderer = html.NewRenderer(html.RendererOptions{Flags: html.SkipHTML | html.SkipImages})
	trustedRenderer = html.NewRenderer(html.RendererOptions{Flags: html.SkipHTML | html.SkipImages})
}

func RenderBlurb(blurb screenjournal.Blurb) string {
	return renderUntrusted(blurb.String())
}

func RenderComment(comment screenjournal.CommentText) string {
	return renderUntrusted(comment.String())
}

func renderUntrusted(s string) string {
	parser := gomarkdown_parser.NewWithExtensions(gomarkdown_parser.NoExtensions)
	html := string(gomarkdown.ToHTML([]byte(s), parser, untrustedRenderer))

	return strings.TrimSpace(html)
}

func RenderEmail(body screenjournal.EmailBodyMarkdown) string {
	parser := gomarkdown_parser.NewWithExtensions(gomarkdown_parser.Autolink)
	html := string(gomarkdown.ToHTML([]byte(body.String()), parser, trustedRenderer))

	return strings.TrimSpace(html)
}
