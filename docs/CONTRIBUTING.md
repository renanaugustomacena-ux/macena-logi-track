# Contributing to LogiTrack

## Dev setup
```bash
# Go 1.26, Node 22, Docker, make.
cd backend  && make tidy && make test
cd frontend && npm ci && npm run typecheck && npm run test
```

## Branching
- `main` is protected. All changes land via pull request.
- Feature branches: `feature/<short-slug>`.
- Hotfixes: `hotfix/<incident-id>`.
- Never push directly to `main`.

## PR template

Each PR must include:
- Summary of the change (one paragraph).
- Test plan (bullet list of unit/integration/E2E tests added or
  touched).
- Security review checkbox if the PR touches auth, data-residency,
  or outbound HTTP.
- Compliance review checkbox if the PR touches
  `docs/ITALIAN-COMPLIANCE.md` or the entities referenced there
  (Shipment, Vehicle, Driver, ChainOfCustody).

## Testing requirements
- New code ≥ 70% line coverage on the new module.
- Unit tests pass on `go test ./... -race`.
- `golangci-lint run --timeout 5m` passes.
- `vue-tsc --noEmit` passes on the frontend.

## Commit messages
- Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`, `test:`,
  `refactor:`, `build:`, `security:`).
- Reference the issue id in the footer (`Refs #123`).

## Code of conduct
Italian labour culture: respect, punctuality, chiarezza. No slurs,
no harassment; a DEI-incident channel is published separately.

## Maintainers
- Release steward: 1 rotating role, weekly.
- Security contact: `security@logitrack.it`.
