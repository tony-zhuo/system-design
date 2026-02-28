# CLAUDE.md

## Git Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/) format. All commit messages must be written in **English**.

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type

| Type       | Description                                      |
|------------|--------------------------------------------------|
| `feat`     | A new feature                                    |
| `fix`      | A bug fix                                        |
| `docs`     | Documentation only changes                       |
| `style`    | Code style changes (formatting, no logic change) |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf`     | Performance improvement                          |
| `test`     | Adding or updating tests                         |
| `chore`    | Build process, tooling, or auxiliary changes      |
| `revert`   | Revert a previous commit                         |

### Subject

- Max 50 characters
- Use imperative mood (e.g. "add" not "added")
- No period at the end
- One topic per commit

### Body (optional)

- Explain **what** and **why**, not just how
- Wrap each line at 72 characters
- Separate from subject with a blank line

### Footer (optional)

- Reference issues: `Closes #123`, `Fixes #456`
- Note breaking changes: `BREAKING CHANGE: <description>`

### Examples

```
feat(auth): add OAuth2 login support

Integrate Google OAuth2 to allow users to sign in with their
Google account. This replaces the legacy session-based auth.

Closes #42
```

```
fix(api): handle null response from payment gateway
```

```
docs: update README with setup instructions
```
