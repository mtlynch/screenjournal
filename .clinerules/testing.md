# Testing

- To run unit tests, run ./dev-scripts/run-go-tests
- When writing tests to verify a bugfix, follow TDD conventions: write the test with the failing test first, verify the test fails, fix the bug, and verify that the test passes.
- When writing new test cases, avoid having t.Run have special-case behavior for particular inputs. Instead, use general purpose logic that doesn't assume particular inputs.
- Don't export methods just for testing - test through public APIs instead.

## if got, want

Use the `if got, want` pattern when writing or editing unit tests. See this snippet as an example:

```go
func TestParseTwitterHandle(t *testing.T) {
  for _, tt := range []struct {
    explanation    string
    input          string
    handleExpected social.TwitterHandle
    errExpected    error
  }{
    {
      "regular handle on its own is valid",
      "jerry",
      social.TwitterHandle("jerry"),
      nil,
    },
    {
      "regular handle in URL is valid",
      "https://twitter.com/jerry",
      social.TwitterHandle("jerry"),
      nil,
    },
    {
      "handle with exactly 15 characters is valid",
      "https://twitter.com/" + strings.Repeat("A", 15),
      social.TwitterHandle(strings.Repeat("A", 15)),
      nil,
    },
    {
      "handle with more than 15 characters is invalid",
      "https://twitter.com/" + strings.Repeat("A", 16),
      social.TwitterHandle(""),
      social.ErrInvalidTwitterHandle,
    },
  } {
    t.Run(fmt.Sprintf("%s [%s]", tt.explanation, tt.input), func(t *testing.T) {
      handle, err := social.ParseTwitterHandle(tt.input)
      if got, want := err, tt.errExpected; got != want {
        t.Fatalf("err=%v, want=%v", got, want)
      }
      if got, want := handle, tt.handleExpected; got != want {
        t.Errorf("handle=%v, want=%v", got, want)
      }
    })
  }
}
```
