# API Design

## Minimize Exported Surface Area

- Don't export methods just for testing - test through public APIs instead.
- Only export what external packages actually need to use.

## Avoid Platform Coupling

- Don't pass platform-specific types (e.g., AWS Lambda events) to business logic.
- Create simple structs with only the data needed, making code portable.

## Encapsulate Related Operations

- Group related operations (e.g., verification + processing) in a single method.
- This simplifies APIs and prevents steps from being accidentally skipped.

## Design for Testing

- Consider allowing bypass mechanisms for tests (e.g., empty secret = skip verification).
- Test private methods indirectly through public APIs.
- Structure code so unit tests don't need complex setup (e.g., generating valid signatures).

## Keep Interfaces Simple

- Group related parameters into structs rather than multiple arguments.
- Return single error types that can represent multiple failure modes.
- One method should have one clear responsibility from the caller's perspective.
