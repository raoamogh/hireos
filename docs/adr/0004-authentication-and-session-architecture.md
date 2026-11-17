# ADR 0004: Authentication and Session Architecture

- Status: Accepted
- Date: 2026-11-17

## Context

HIREOS requires authentication for recruiters, organization members
and candidates.

The PRD requires:

- Email and password registration and authentication.
- Google login.
- LinkedIn login.
- Argon2id password hashing.
- Secure session management.
- Short-lived access tokens where JWT is used.
- Refresh credential rotation.
- Session revocation.
- Password reset protection.
- Email verification.
- Organization-scoped authorization.
- Candidate authentication and isolation.

A single user may belong to multiple organizations. Candidate
accounts are separate from organization memberships.

## Decisions

### 1. Canonical user identity

HIREOS maintains one global user identity per person.

Authentication methods are associated with this canonical identity.
Recruiter organization memberships and candidate accounts remain
separate authorization contexts.

A user may have both an organization membership and a candidate
account.

### 2. Supported authentication methods

HIREOS supports:

- Email and password.
- Google OpenID Connect.
- LinkedIn OpenID Connect.

All three methods are available to recruiter and candidate users.

### 3. Password authentication

Passwords are hashed using Argon2id.

Password credentials are stored separately from the core users
record so that users authenticated exclusively through external
identity providers do not require an artificial password.

Password hashes must include the algorithm parameters and salt
required for verification and future rehashing.

Plaintext passwords must never be persisted or logged.

### 4. External identity providers

Google and LinkedIn are integrated using OpenID Connect.

The authorization-code flow uses:

- Transaction-specific state.
- Transaction-specific nonce.
- PKCE using S256.
- Server-side authorization-code exchange.
- ID token signature and claim validation.
- Exact issuer and audience validation.
- Expiration and nonce validation.

External identities are identified using the combination of
issuer and subject.

Provider access tokens are not used as HIREOS application sessions.
Provider tokens are not persisted unless a separately approved
integration requirement needs them.

### 5. Email verification

HIREOS requires verified email addresses for account association
and email-based account recovery.

Provider email claims are not assumed to be verified merely because
an email address is present.

If a provider does not supply a verified email, HIREOS requires its
own email verification process.

An unverified matching email must never trigger automatic account
linking.

### 6. Account linking

Users may link additional authentication methods to an existing
HIREOS account only after authenticating into that account.

External identities must be unique across HIREOS accounts.

Matching email addresses alone are insufficient to link accounts.

### 7. Session architecture

HIREOS uses stateful opaque sessions backed by PostgreSQL.

The browser receives a cryptographically random session credential
in a secure, HttpOnly cookie.

Only a verifier/hash of the credential is persisted.

The session record contains the user association, creation time,
expiry, last activity and revocation state.

### 8. Session rotation and replay detection

Session credentials are rotated through the session renewal
endpoint.

Rotation must be atomic.

Previously consumed credential hashes are retained for the
appropriate session lifetime to detect credential reuse.

Detected credential reuse invalidates the associated session.

Session expiration and revocation are enforced server-side.

### 9. Browser security

Production session cookies use Secure, HttpOnly and an appropriate
SameSite policy.

CSRF protection includes synchronizer tokens and origin validation.

OAuth callback transactions use state, nonce and PKCE validation.

Local development cookie configuration must not weaken production
cookie settings.

### 10. Authorization boundaries

Authentication establishes the HIREOS user identity only.

Organization access requires an active organization membership
and appropriate organization-scoped permissions.

Candidate access requires candidate-specific authorization.

PostgreSQL RLS remains a defense-in-depth control.

The application database role must not be a superuser or have
BYPASSRLS privileges.

## Consequences

- One HIREOS user can authenticate through multiple methods.
- Recruiter and candidate authorization remain independent.
- Sessions can be revoked centrally.
- JWT signing infrastructure is not required for the initial
  first-party browser authentication architecture.
- PostgreSQL stores session state and credential verifiers.
- Authentication provider integrations require secure callback
  handling and provider-specific configuration.
- Email verification is required where provider verification
  cannot be established.

## Deferred implementation details

The following will be specified and tested during implementation:

- Session absolute and idle expiration durations.
- Argon2id parameters after benchmarking.
- Password policy and maximum input sizes.
- Email verification token lifecycle.
- Password reset token lifecycle.
- OAuth transaction storage and expiration.
- Provider client configuration and redirect URIs.
- CSRF token lifecycle.
- Rate limits and abuse controls.
- Account recovery and security notifications.

## Official references

- Google OpenID Connect:
  https://developers.google.com/identity/openid-connect/openid-connect

- LinkedIn OpenID Connect:
  https://learn.microsoft.com/linkedin/consumer/integrations/self-serve/sign-in-with-linkedin-v2

- OAuth 2.0 Security Best Current Practice:
  https://www.rfc-editor.org/rfc/rfc9700.html

- OWASP Password Storage:
  https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html

- OWASP Session Management:
  https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
