Goal: Modify handlers/templates/pages/reviews-for-single-media-entry.html so that above the "Comment" button there's a set of emoji reactions the user can leave.

👍 👀 😯 🤔 🥞

Clicking an emoji adds a line below the review but above the comments like

🥞 reacted mike, 5 days ago
👀 reacted jeff, 20 days ago

The owner of the reaction or ScreenJournal admin can delete their reaction with a small delete button to the right of the reaction line

---

## Revised Implementation Plan for Emoji Reactions

### 1. Create domain types (`screenjournal/reactions.go`)

```go
type ReactionID uint64

type ReactionEmoji struct {
    value string
}

func (e ReactionEmoji) String() string {
    return e.value
}

type ReviewReaction struct {
    ID      ReactionID
    Owner   Username
    Emoji   ReactionEmoji
    Created time.Time
    Review  Review
}
```

The `ReactionEmoji` type contains a string field (not aliasing) with a `String()` method.

### 2. Create parsing function (`handlers/parse/reaction.go`)

```go
var allowedEmojis = map[string]bool{
    "👍": true,
    "👀": true,
    "😯": true,
    "🤔": true,
    "🥞": true,
}

func ReactionEmoji(raw string) (screenjournal.ReactionEmoji, error) {
    // Validate emoji is in allowed list, return parsed type
}

func ReactionID(raw string) (screenjournal.ReactionID, error) {
    // Parse uint64 ID
}
```

This is the **only code path** that creates `ReactionEmoji` values.

### 3. Add reactions to `Review` struct (`screenjournal/review.go`)

- Add `Reactions []ReviewReaction` field

### 4. Create database migration (`store/sqlite/migrations/014-reactions-create.sql`)

```sql
CREATE TABLE review_reactions (
    id INTEGER PRIMARY KEY,
    review_id INTEGER NOT NULL,
    reaction_owner TEXT NOT NULL,
    emoji TEXT NOT NULL,
    created_time TEXT NOT NULL,
    FOREIGN KEY (review_id) REFERENCES reviews (id),
    FOREIGN KEY (reaction_owner) REFERENCES users (username),
    UNIQUE(review_id, reaction_owner, emoji)
);
```

### 5. Create store layer (`store/sqlite/reactions.go`)

- `ReadReactions(reviewID ReviewID) ([]ReviewReaction, error)` - ordered by `created_time ASC` (matches comments)
- `InsertReaction(ReviewReaction) (ReactionID, error)`
- `DeleteReaction(ReactionID) error`

No `ReadReaction` or `ErrReactionNotFound` as per your feedback.

### 6. Update store interface (`handlers/store.go`)

- Add the three reaction methods to the Store interface

### 7. Create HTTP handlers (`handlers/reactions.go`)

Using the new htmx pattern (return HTML directly, not JSON):

**`reactionsPost()`** - POST /reactions

- Parse review ID and emoji from form
- Insert reaction into database
- On success: Return HTML fragment with updated reaction section (the user's new reaction + hide emoji picker since user has reacted)
- On error: Return error via `http.Error()`

**`reactionsDelete()`** - DELETE /reactions/{reactionID}

- Parse reaction ID from path
- Look up the reaction in the database to get the owner
- Permission check: Allow if `owner == loggedInUser` OR `isAdmin(ctx)`
- Delete reaction
- On success: Return `http.StatusNoContent` (or empty HTML to remove element)
- On error: Return error via `http.Error()`

### 8. Register routes (`handlers/routes.go`)

Add under `authenticatedRoutes` (the "new" subrouter pattern):

```go
authenticatedRoutes.HandleFunc("/reactions", s.reactionsPost()).Methods(http.MethodPost)
authenticatedRoutes.HandleFunc("/reactions/{reactionID}", s.reactionsDelete()).Methods(http.MethodDelete)
```

### 9. Update views (`handlers/views.go`)

In `moviesReadGet()` and `tvShowsReadGet()`:

- Load reactions for each review (similar to how comments are loaded)
- Check if the logged-in user has already reacted to each review (to hide emoji picker)

### 10. Update template (`handlers/templates/pages/reviews-for-single-media-entry.html`)

Structure within each review:

```html
<!-- Review content -->
<div class="review-content">...</div>

<!-- Emoji picker (only shown if user hasn't reacted) -->
{{ if not .UserHasReacted }}
<div class="emoji-picker" id="reactions-picker-{{.ID}}">
  {{ range $emoji := .AvailableEmojis }}
  <button
    hx-post="/reactions"
    hx-vals='{"review-id": "{{$.ID}}", "emoji": "{{$emoji}}"}'
    hx-target="#reactions-section-{{$.ID}}"
    hx-swap="outerHTML"
  >
    {{$emoji}}
  </button>
  {{ end }}
</div>
{{ end }}

<!-- Reactions section -->
<div id="reactions-section-{{.ID}}">
  {{ range .Reactions }} {{ template "reaction" dict "Reaction" .
  "LoggedInUsername" $loggedInUsername "IsAdmin" $.IsAdmin }} {{ end }}
</div>

<!-- Comments section -->
<div class="comments">...</div>

<!-- Comment button -->
{{ template "add-comment-button" . }}
```

**`{{ define "reaction" }}`** template block:

```html
<div id="reaction{{.Reaction.ID}}" class="reaction">
  <span
    >{{.Reaction.Emoji}} reacted
    <a href="/reviews/by/{{.Reaction.Owner}}">{{.Reaction.Owner}}</a>, {{
    relativeCommentDate .Reaction.Created }}
  </span>
  {{ if or (eq .Reaction.Owner .LoggedInUsername) .IsAdmin }}
  <button
    hx-delete="/reactions/{{.Reaction.ID}}"
    hx-confirm="Delete this reaction?"
    hx-target="#reaction{{.Reaction.ID}}"
    hx-swap="outerHTML"
    class="btn btn-sm"
  >
    ×
  </button>
  {{ end }}
</div>
```

### 11. Add CSS styling (`handlers/static/css/reviews.css`)

- Style `.emoji-picker` buttons
- Style `.reaction` display rows (small text, delete button alignment)

### Summary of key design decisions:

- `ReactionEmoji` is a struct containing a string (not aliasing), created only via `parse.ReactionEmoji()`
- Reactions ordered chronologically (oldest first) like comments
- User can only react once per emoji per review (DB unique constraint)
- Once a user has reacted, the emoji picker is hidden for that review
- Owner or admin can delete reactions
- Using htmx pattern with direct HTML responses

## Unit Tests

### 1. Parsing Tests (`handlers/parse/reaction_test.go`)

Following the pattern from `comment_test.go`:

```go
func TestReactionID(t *testing.T) {
    for _, tt := range []struct {
        description string
        in          string
        id          screenjournal.ReactionID
        err         error
    }{
        {"ID of 1 is valid", "1", screenjournal.ReactionID(1), nil},
        {"ID of 0 is invalid", "0", screenjournal.ReactionID(0), parse.ErrInvalidReactionID},
        {"non-numeric ID is invalid", "banana", screenjournal.ReactionID(0), parse.ErrInvalidReactionID},
    } {
        t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
            id, err := parse.ReactionID(tt.in)
            if got, want := err, tt.err; got != want {
                t.Fatalf("err=%v, want=%v", got, want)
            }
            if got, want := id.UInt64(), tt.id.UInt64(); got != want {
                t.Errorf("id=%d, want=%d", got, want)
            }
        })
    }
}

func TestReactionEmoji(t *testing.T) {
    for _, tt := range []struct {
        description string
        in          string
        emoji       string  // expected String() output
        err         error
    }{
        {"thumbs up emoji is valid", "👍", "👍", nil},
        {"eyes emoji is valid", "👀", "👀", nil},
        {"surprised emoji is valid", "😯", "😯", nil},
        {"thinking emoji is valid", "🤔", "🤔", nil},
        {"pancakes emoji is valid", "🥞", "🥞", nil},
        {"heart emoji is not in allowed list", "❤️", "", parse.ErrInvalidReactionEmoji},
        {"empty string is invalid", "", "", parse.ErrInvalidReactionEmoji},
        {"text is invalid", "hello", "", parse.ErrInvalidReactionEmoji},
    } {
        t.Run(fmt.Sprintf("%s [%s]", tt.description, tt.in), func(t *testing.T) {
            emoji, err := parse.ReactionEmoji(tt.in)
            if got, want := err, tt.err; got != want {
                t.Fatalf("err=%v, want=%v", got, want)
            }
            if got, want := emoji.String(), tt.emoji; got != want {
                t.Errorf("emoji=%v, want=%v", got, want)
            }
        })
    }
}
```

### 2. Handler Tests (`handlers/reactions_test.go`)

Following the pattern from `comments_test.go`:

**Test cases for `TestReactionsPost`:**

- Allows user to add a reaction to an existing review
- Rejects request with invalid emoji
- Rejects request with invalid review ID
- Returns 404 if user attempts to react to non-existent review
- Rejects request if user is not authenticated
- (Implicitly via DB constraint) Rejects duplicate reaction (same user + same emoji + same review)

**Test cases for `TestReactionsDelete`:**

- Allows a user to delete their own reaction
- Allows an admin to delete another user's reaction
- Prevents a non-admin user from deleting another user's reaction (403 Forbidden)
- Prevents deleting a non-existent reaction (404)
- Prevents deleting with invalid reaction ID (400)
- Prevents unauthenticated user from deleting any reaction (401)

Example structure:

```go
func TestReactionsPost(t *testing.T) {
    for _, tt := range []struct {
        description       string
        payload           string
        sessionToken      string
        sessions          []mockSessionEntry
        movies            []screenjournal.Movie
        reviews           []screenjournal.Review
        status            int
        expectedReactions []screenjournal.ReviewReaction
    }{
        {
            description:  "allows user to add a reaction to an existing review",
            payload:      "review-id=1&emoji=%F0%9F%91%8D", // 👍 URL-encoded
            sessionToken: makeReactionsTestData().sessions.userA.token,
            // ... setup data ...
            status: http.StatusOK,
            expectedReactions: []screenjournal.ReviewReaction{
                {
                    ID:    screenjournal.ReactionID(1),
                    Owner: makeReactionsTestData().sessions.userA.session.Username,
                    // ... emoji and review ...
                },
            },
        },
        // ... more test cases ...
    } {
        t.Run(tt.description, func(t *testing.T) {
            // Setup datastore, insert users/movies/reviews
            // Make HTTP request
            // Assert status code
            // If success, verify reactions in datastore
        })
    }
}

func TestReactionsDelete(t *testing.T) {
    for _, tt := range []struct {
        description       string
        route             string
        sessionToken      string
        sessions          []mockSessionEntry
        reactions         []screenjournal.ReviewReaction
        status            int
        expectedReactions []screenjournal.ReviewReaction
    }{
        {
            description:  "allows a user to delete their own reaction",
            route:        "/reactions/1",
            // ... test data with owner == logged in user ...
            status: http.StatusNoContent,
        },
        {
            description:  "allows an admin to delete another user's reaction",
            route:        "/reactions/1",
            // ... test data with admin session, reaction owned by different user ...
            status: http.StatusNoContent,
        },
        {
            description:  "prevents a non-admin user from deleting another user's reaction",
            route:        "/reactions/1",
            // ... test data with non-admin session, reaction owned by different user ...
            status: http.StatusForbidden,
        },
        // ... more test cases ...
    } {
        // ... test implementation ...
    }
}
```

---

## Playwright E2E Tests (`e2e/reactions.spec.ts`)

Following patterns from `e2e/reviews.spec.ts` for testing comments:

```typescript
import { test, expect } from "@playwright/test";
import { populateDummyData, readDbTokenCookie } from "./helpers/db";
import { loginAsUserA, loginAsUserB, loginAsAdmin } from "./helpers/login";

test.beforeEach(async ({ page }) => {
  await populateDummyData(page);
  await loginAsUserA(page);
});

test("adds an emoji reaction to a review", async ({ page }) => {
  // Navigate to a movie page with an existing review
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();
  await expect(page).toHaveURL("/movies/1#review1");

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });

  // Click the thumbs up emoji button
  await reviewDiv.locator(".emoji-picker button", { hasText: "👍" }).click();

  // Verify reaction appears
  const reactionDiv = await reviewDiv.locator(".reaction", {
    hasText: /👍 reacted userA/,
  });
  await expect(reactionDiv).toBeVisible();
  await expect(reactionDiv.getByTestId("relative-time")).toHaveText("just now");

  // Verify emoji picker is no longer visible (user already reacted)
  await expect(reviewDiv.locator(".emoji-picker")).not.toBeVisible();
});

test("emoji picker is hidden after user reacts", async ({ page }) => {
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });

  // Initially emoji picker is visible
  await expect(reviewDiv.locator(".emoji-picker")).toBeVisible();

  // Add a reaction
  await reviewDiv.locator(".emoji-picker button", { hasText: "🥞" }).click();

  // Emoji picker should be hidden
  await expect(reviewDiv.locator(".emoji-picker")).not.toBeVisible();
});

test("user can delete their own reaction", async ({ page }) => {
  // First add a reaction
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });

  await reviewDiv.locator(".emoji-picker button", { hasText: "👀" }).click();

  // Verify reaction exists
  const reactionDiv = await reviewDiv.locator(".reaction", {
    hasText: /👀 reacted userA/,
  });
  await expect(reactionDiv).toBeVisible();

  // Delete the reaction (accept confirmation dialog)
  page.on("dialog", (dialog) => dialog.accept());
  await reactionDiv.locator("button[data-sj-purpose='delete']").click();

  // Verify reaction is removed
  await expect(reactionDiv).not.toBeVisible();

  // Verify emoji picker reappears
  await expect(reviewDiv.locator(".emoji-picker")).toBeVisible();
});

test("user cannot delete another user's reaction", async ({
  page,
  browser,
}) => {
  // UserA adds a reaction
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });
  await reviewDiv.locator(".emoji-picker button", { hasText: "🤔" }).click();

  // Switch to userB
  const userBContext = await browser.newContext();
  await userBContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const userBPage = await userBContext.newPage();
  await loginAsUserB(userBPage);

  await userBPage.goto("/movies/1#review1");

  // UserB should see userA's reaction but NOT see delete button
  const userBReviewDiv = await userBPage.locator(".review", {
    has: userBPage.getByText("I love water!"),
  });
  const reactionDiv = await userBReviewDiv.locator(".reaction", {
    hasText: /🤔 reacted userA/,
  });
  await expect(reactionDiv).toBeVisible();
  await expect(
    reactionDiv.locator("button[data-sj-purpose='delete']")
  ).not.toBeVisible();

  await userBContext.close();
});

test("admin can delete another user's reaction", async ({ page, browser }) => {
  // UserA adds a reaction
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });
  await reviewDiv.locator(".emoji-picker button", { hasText: "😯" }).click();

  // Switch to admin user
  const adminContext = await browser.newContext();
  await adminContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const adminPage = await adminContext.newPage();
  await loginAsAdmin(adminPage); // Need to add this helper if it doesn't exist

  await adminPage.goto("/movies/1#review1");

  // Admin should see delete button on userA's reaction
  const adminReviewDiv = await adminPage.locator(".review", {
    has: adminPage.getByText("I love water!"),
  });
  const reactionDiv = await adminReviewDiv.locator(".reaction", {
    hasText: /😯 reacted userA/,
  });
  await expect(
    reactionDiv.locator("button[data-sj-purpose='delete']")
  ).toBeVisible();

  // Admin deletes userA's reaction
  adminPage.on("dialog", (dialog) => dialog.accept());
  await reactionDiv.locator("button[data-sj-purpose='delete']").click();

  await expect(reactionDiv).not.toBeVisible();

  await adminContext.close();
});

test("reactions are displayed in chronological order", async ({
  page,
  browser,
}) => {
  // UserA adds first reaction
  await page
    .getByRole("heading", { name: "The Waterboy" })
    .getByRole("link")
    .click();

  const reviewDiv = await page.locator(".review", {
    has: page.getByText("I love water!"),
  });
  await reviewDiv.locator(".emoji-picker button", { hasText: "👍" }).click();

  // UserB adds second reaction
  const userBContext = await browser.newContext();
  await userBContext.addCookies([
    readDbTokenCookie(await page.context().cookies()),
  ]);
  const userBPage = await userBContext.newPage();
  await loginAsUserB(userBPage);
  await userBPage.goto("/movies/1#review1");

  const userBReviewDiv = await userBPage.locator(".review", {
    has: userBPage.getByText("I love water!"),
  });
  await userBReviewDiv
    .locator(".emoji-picker button", { hasText: "🥞" })
    .click();

  // Verify reactions appear in chronological order (userA's first, then userB's)
  const reactions = await userBReviewDiv.locator(".reaction").all();
  await expect(reactions[0]).toContainText("userA");
  await expect(reactions[1]).toContainText("userB");

  await userBContext.close();
});
```

---

## Summary of Test Files to Create

| File                              | Purpose                                             |
| --------------------------------- | --------------------------------------------------- |
| `handlers/parse/reaction_test.go` | Unit tests for parsing ReactionID and ReactionEmoji |
| `handlers/reactions_test.go`      | HTTP handler tests for POST and DELETE endpoints    |
| `e2e/reactions.spec.ts`           | End-to-end Playwright tests for UI interactions     |

The tests follow the existing patterns:

- Table-driven tests with `if got, want` pattern for Go
- Mock session manager for simulating authenticated users
- `test_sqlite.New()` for in-memory database
- Playwright tests using `populateDummyData()` helper and multi-browser contexts for testing different users
