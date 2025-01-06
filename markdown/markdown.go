package markdown

import (
	"html"
	"strings"

	gomarkdown "github.com/gomarkdown/markdown"
	gomarkdown_html "github.com/gomarkdown/markdown/html"
	gomarkdown_parser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

var (
	untrustedParser   *gomarkdown_parser.Parser
	trustedParser     *gomarkdown_parser.Parser
	untrustedRenderer *gomarkdown_html.Renderer
	trustedRenderer   *gomarkdown_html.Renderer
)

func init() {
	untrustedParser = gomarkdown_parser.NewWithExtensions(gomarkdown_parser.NoExtensions)
	trustedParser = gomarkdown_parser.NewWithExtensions(gomarkdown_parser.Autolink)
	untrustedRenderer = gomarkdown_html.NewRenderer(gomarkdown_html.RendererOptions{Flags: gomarkdown_html.SkipHTML | gomarkdown_html.SkipImages})
	trustedRenderer = gomarkdown_html.NewRenderer(gomarkdown_html.RendererOptions{Flags: gomarkdown_html.SkipHTML | gomarkdown_html.SkipImages})
}

func RenderBlurb(blurb screenjournal.Blurb) string {
	return renderUntrusted(trimSpacesFromEachLine(blurb.String()))
}

func RenderBlurbAsPlaintext(blurb screenjournal.Blurb) string {
	asHtml := renderUntrusted(blurb.String())
	plaintext := bluemonday.StrictPolicy().Sanitize(asHtml)

	// Decode HTML entities like ' back to characters
	plaintext = html.UnescapeString(plaintext)

	return strings.TrimSpace(plaintext)
}

func RenderComment(comment screenjournal.CommentText) string {
	return renderUntrusted(trimSpacesFromEachLine(comment.String()))
}

func renderUntrusted(s string) string {
	asHtml := string(gomarkdown.ToHTML([]byte(s), untrustedParser, untrustedRenderer))
	return strings.TrimSpace(asHtml)
}

func RenderEmail(body screenjournal.EmailBodyMarkdown) string {
	asHtml := string(gomarkdown.ToHTML([]byte(body.String()), trustedParser, trustedRenderer))
	return strings.TrimSpace(asHtml)
}

func trimSpacesFromEachLine(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}
