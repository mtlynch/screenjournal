package markdown

import (
	"fmt"
	"html"
	"strings"

	gomarkdown "github.com/gomarkdown/markdown"
	gomarkdown_html "github.com/gomarkdown/markdown/html"
	gomarkdown_parser "github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

const SpoilersKeyword = "!spoilers"

var (
	untrustedRenderer *gomarkdown_html.Renderer
	trustedRenderer   *gomarkdown_html.Renderer
)

func init() {
	untrustedRenderer = gomarkdown_html.NewRenderer(gomarkdown_html.RendererOptions{Flags: gomarkdown_html.SkipHTML | gomarkdown_html.SkipImages})
	trustedRenderer = gomarkdown_html.NewRenderer(gomarkdown_html.RendererOptions{Flags: gomarkdown_html.SkipHTML | gomarkdown_html.SkipImages})
}

func RenderBlurb(blurb screenjournal.Blurb) string {
	return renderUntrusted(blurb.String())
}

func RenderBlurbAsPlaintext(blurb screenjournal.Blurb) string {
	unspoiled, _, _ := splitSpoilers(blurb.String())
	asHtml := renderUntrusted(unspoiled)
	plaintext := bluemonday.StrictPolicy().Sanitize(asHtml)

	// Decode HTML entities like ' back to characters
	plaintext = html.UnescapeString(plaintext)

	return strings.TrimSpace(plaintext)
}

func RenderComment(comment screenjournal.CommentText) string {
	return renderUntrusted(comment.String())
}

func renderUntrusted(s string) string {
	renderMarkdown := func(markdown string) string {
		parser := gomarkdown_parser.NewWithExtensions(gomarkdown_parser.NoExtensions)
		trimmed := trimSpacesFromEachLine(markdown)
		asHtml := string(gomarkdown.ToHTML([]byte(trimmed), parser, untrustedRenderer))
		return strings.TrimSpace(asHtml)
	}

	unspoiled, spoilers, hasSpoilers := strings.Cut(s, SpoilersKeyword)
	unspoiledRendered := renderMarkdown(unspoiled)
	if !hasSpoilers {
		return unspoiledRendered
	}

	return fmt.Sprintf("%s\n\n<div class=\"spoilers d-none\">\n\n%s\n\n</div>", unspoiledRendered, renderMarkdown(spoilers))
}

func RenderEmail(body screenjournal.EmailBodyMarkdown) string {
	parser := gomarkdown_parser.NewWithExtensions(gomarkdown_parser.Autolink)
	asHtml := string(gomarkdown.ToHTML([]byte(body.String()), parser, trustedRenderer))

	return strings.TrimSpace(asHtml)
}

func trimSpacesFromEachLine(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}

func splitSpoilers(s string) (string, string, bool) {
	return strings.Cut(s, SpoilersKeyword)
}
