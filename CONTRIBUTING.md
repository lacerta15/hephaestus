# Contributing to Hephaestus

Thank you for your interest in contributing! Hephaestus is an open proof-of-concept that aims to demonstrate how regulatory reporting in Indonesian banking could be re-imagined on a permissioned blockchain.

## How to contribute

1. **Open an issue first** — describe the bug, feature, or design question before writing code.
2. **Fork & branch** — branch from `main` using a descriptive name: `feat/cross-channel-query`, `fix/chaincode-validation-edge-case`.
3. **Follow the style guide** — `gofmt` for chaincode, `prettier` for JS/TS, `ansible-lint` for playbooks.
4. **Add tests** — every chaincode change requires a corresponding unit test in `chaincode/antasena/test/`.
5. **Update docs** — if you change behaviour, update both `docs/en/` and `docs/id/`.
6. **Sign your commits** — `git commit -s` (DCO).

## Local development

```bash
# Bring the network up
make up

# Run chaincode unit tests
make test-chaincode

# Lint everything
make lint
```

## Code review

All PRs require:
- ✅ At least one approval from a maintainer
- ✅ Green CI (lint + chaincode tests + docker-compose smoke test)
- ✅ Updated documentation if behaviour changes

## Responsible disclosure

If you discover a vulnerability, **do not open a public issue**. See [SECURITY.md](SECURITY.md).

## Code of Conduct

By participating, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).
