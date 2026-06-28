# Security Policy

## Supported versions

whook is pre-1.0 and moving quickly. Security fixes land on `main` and in the
latest tagged release. Please run a recent version before reporting an issue.

## Reporting a vulnerability

Please do not open a public issue for security problems.

Report privately through GitHub's [private vulnerability reporting](https://github.com/edaywalid/whook/security/advisories/new),
or email becharwalid9@gmail.com. Include the version or commit, a description of
the issue, and steps to reproduce if you have them.

You can expect an initial response within a few days. Once a fix is ready we will
release it and credit you in the advisory unless you prefer to stay anonymous.

## Scope

The most sensitive parts of whook are signature verification, the encryption of
signing secrets at rest, and the admin bearer token that guards the management
API and dashboard. Reports touching those areas are especially welcome.
