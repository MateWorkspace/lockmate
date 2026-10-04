# JWT tokens

Access and refresh tokens are bound to one space membership. Both contain
`user_id`, `space_id`, `member_id`, `name`, and `username`. Access claims also
contain `roles` and `permissions`: slugs from that space, rather than display
names. Empty grant lists are valid. Refresh claims contain no grants.

Generation adds a canonical user-ID `sub`, audience `lockmate:space:{space_id}`,
`token_type` (`access` or `refresh`), and integer-second `iat`, `nbf`, and `exp`.
The adapter emits HS256 and accepts the established HS256/HS384/HS512 allowlist.
IDs must be positive. Invalid generation inputs return `BadArgs`; missing secrets,
zero lifetimes, or unusable expiration ranges return `Failure`.

Call `validate_access(context, expected_space_id, token)` or
`validate_refresh(context, expected_space_id, token)` with the caller's expected
space. Validation requires identity, scope, purpose, audience, subject, and time
claims. It checks positive IDs, exact space/audience and token type, canonical
subject consistency, expiration/not-before without clock skew, and
`iat <= nbf < exp` with no future issue time. Each token carries one string
audience. Old unscoped tokens and the previous subject fallback are rejected.

Empty tokens or nonpositive expected-space arguments return `BadArgs`. Expired
tokens return `Expired`; other validation failures return `Invalid`. Access and
refresh purposes are checked independently of signing keys, including when both
configured secrets are equal. Each failed public operation logs once with a safe
message and without tokens, secrets, or claim contents.

The utility validates token contents and does not query repositories. Application
authentication must check live membership state and current privileges when
refreshing or switching spaces. Token issuance, session management, and application
authorization remain separate work.
