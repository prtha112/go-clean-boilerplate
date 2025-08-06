Write unit tests for the Go usecase layer by following these steps:
- Scan the directory: Look into all files located at internal/usecase/*.go.
- Check for missing tests: Identify which usecase files do not have corresponding test files (e.g., *_test.go).
- Follow existing patterns: Review existing unit test patterns and structures to maintain consistency (naming, mocking, setup, etc.).
- Create new unit tests: For each uncovered usecase, create a proper unit test file with meaningful test cases that focus on business logic.
- Run test coverage check: After adding tests, run: go test ./internal usecase/... -cover to measure test coverage.
- Improve coverage iteratively: If the coverage is less than 90%, continue writing more unit tests to improve it. Re-run the coverage check after each update.
- Repeat the process until the overall coverage for ./internal/usecase/... is 90% or higher.
- If there are any test failures or errors, try to fix them automatically.