# Assistant guidelines

## Planning

- When in "Plan" mode, assume that a junior developer will perform the implementation. Mention details that may not be obvious to them.
- When in "Plan" mode, ask for clarification about anything that is ambiguous.

## Documenting lessons learned

After successfully completing a task where the user had to provide corrections or guidance, consider adding the lessons to `.clinerules`. This helps build institutional knowledge and prevents repeating mistakes.

### When to add new guidelines

- The user corrected a misunderstanding about the codebase
- You learned a new pattern or best practice specific to this project
- The user revealed a preference or requirement not previously documented

### How to add guidelines

1. Identify the key principle or pattern learned
2. Determine which existing file fits best:
   - `api-design.md` - API structure and design patterns
   - `planning.md` - Task planning and scoping
   - `security.md` - Security considerations
   - `style.md` - Code formatting and conventions
   - `testing.md` - Testing approaches and requirements
3. Add a concise, actionable guideline
4. Keep entries brief but clear for future LLM conversations

### Example

If you learned that methods shouldn't be exported just for testing, add to `api-design.md`:

"Don't export methods just for testing - test through public APIs instead."
