package markdown_test

import (
	"testing"

	"github.com/mtlynch/screenjournal/v2/markdown"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

func TestRenderBlurbAndComment(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		out         string
	}{
		{
			"renders unformatted text",
			"hello, world!",
			"<p>hello, world!</p>",
		},
		{
			"formats italics",
			"hello, _world_!",
			"<p>hello, <em>world</em>!</p>",
		},
		{
			"formats bold text",
			"hello, **world**!",
			"<p>hello, <strong>world</strong>!</p>",
		},
		{
			"formats multiline text",
			`Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.

If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.

Daniel Radcliffe is fantastic, and it's a great film role for Rainn Wilson. There are a million great cameos.

You'll like it if you enjoy things like Children's Hospital, Comedy Bang Bang, or Popstar.`,
			`<p>Instant movie of the year for me. It's such a delightful and creative way to play with the genre of musical biopics.</p>

<p>If you think of Weird Al as just a parody music guy, give it a chance. I was never that excited about his parody music, but I always enjoy seeing him in TV and movies.</p>

<p>Daniel Radcliffe is fantastic, and it's a great film role for Rainn Wilson. There are a million great cameos.</p>

<p>You'll like it if you enjoy things like Children's Hospital, Comedy Bang Bang, or Popstar.</p>`,
		},
		{
			// We don't really want this behavior, but it doesn't hurt anything right
			// now, so keep the test to show the behavior.
			"renders backticks",
			"hello, `world`!",
			"<p>hello, <code>world</code>!</p>",
		},
		{
			// We don't really want this behavior, but it doesn't hurt anything right
			// now, so keep the test to show the behavior.
			"renders triple backticks",
			"hello, ```world```!",
			"<p>hello, <code>world</code>!</p>",
		},
		{
			"does not render script tags",
			"hello, <script>alert(1)</script>",
			"<p>hello, alert(1)</p>",
		},
		{
			"does not render images",
			"check out my cat! ![photo of a cat](https://example.com/cat.jpg \"My Cat Milo\")",
			"<p>check out my cat! </p>",
		},
		{
			"does not render HTML images",
			`check out my cat! <img src="http://example.com/cat.jpg">`,
			"<p>check out my cat! </p>",
		},
		{
			"does not render links",
			"you can see it [on my blog](http://example.com/blog)",
			"<p>you can see it <tt>on my blog</tt></p>",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			if got, want := markdown.RenderBlurb(screenjournal.Blurb(tt.in)), tt.out; got != want {
				t.Errorf("rendered blurb=%s, want=%s", got, want)
			}
			if got, want := markdown.RenderComment(screenjournal.CommentText(tt.in)), tt.out; got != want {
				t.Errorf("rendered comment=%s, want=%s", got, want)
			}
		})
	}
}

func TestRenderEmail(t *testing.T) {
	for _, tt := range []struct {
		description string
		in          string
		out         string
	}{
		{
			"renders unformatted text",
			"hello, world!",
			"<p>hello, world!</p>",
		},
		{
			"formats italics",
			"hello, _world_!",
			"<p>hello, <em>world</em>!</p>",
		},
		{
			"formats bold text",
			"hello, **world**!",
			"<p>hello, <strong>world</strong>!</p>",
		},
		{
			"renders links",
			"Check out their [latest review](http://example.com/review)",
			`<p>Check out their <a href="http://example.com/review">latest review</a></p>`,
		},
		{
			// We don't really want this behavior, but it doesn't hurt anything right
			// now, so keep the test to show the behavior.
			"renders backticks",
			"hello, `world`!",
			"<p>hello, <code>world</code>!</p>",
		},
		{
			// We don't really want this behavior, but it doesn't hurt anything right
			// now, so keep the test to show the behavior.
			"renders triple backticks",
			"hello, ```world```!",
			"<p>hello, <code>world</code>!</p>",
		},
		{
			"does not render script tags",
			"hello, <script>alert(1)</script>",
			"<p>hello, alert(1)</p>",
		},
		{
			"does not render images",
			"check out my cat! ![photo of a cat](https://example.com/cat.jpg \"My Cat Milo\")",
			"<p>check out my cat! </p>",
		},
		{
			"does not render HTML images",
			`check out my cat! <img src="http://example.com/cat.jpg">`,
			"<p>check out my cat! </p>",
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			if got, want := markdown.RenderEmail(screenjournal.EmailBodyMarkdown(tt.in)), tt.out; got != want {
				t.Errorf("rendered=%s, want=%s", got, want)
			}
		})
	}
}
