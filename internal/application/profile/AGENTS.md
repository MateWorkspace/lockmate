# Profile application

Implement Account and Security in `account/usecase.go` and `security/usecase.go`.
Keep private dependency structs, explicit NewUsecaseImpl constructors returning
domain interfaces, domain import aliases, and context-first early-return methods.
Use shared identity, validation, and projection helpers; log directly in methods.
application must not import concrete infrastructure or transport types.

Immediately below the tag declaration, call PermissionGuard using the operation
constant from application/permissions. Return permission failures before identity
reads, validation, or starting transactions. Then read the caller's live user,
active space/membership, and membership ownership.
Derive every target and audit actor from those authenticated IDs. Global profile
changes affect all spaces; membership preferences affect only this membership.

Writes own a strict transaction, including live authorization and persistence.
Pass the callback context throughout and log success only after commit. Preserve
pointer omission and null clearing; reject both together. Avatar paths are
metadata, and preference updates replace JSON. Classic contacts cannot override
verification flags; repository SQL resets changed contacts.

ChangePassword compares the current password before validating/hashing its
replacement. Google-only users cannot implicitly create classic credentials.
Authentication summaries expose optional safe records, without password hashes.
Missing methods are omitted; operational failures remain errors.

Tags follow profile/<capability>/<MethodName>. Write Error/Warn/Info directly at
handling points with specific messages and inline err/ID metadata. Do not use
logging defers, shared logging wrappers, or generic failure messages. A callback
failure logs both its specific step and the outer transaction failure. Native error
details are retained; never add passwords, tokens, or request bodies to metadata.
Committed classic credential/password changes emit info explicitly after commit.
Successful reads, profile edits, and membership preference updates stay quiet.
Keep blank lines between checks and operation steps.

Tests live in tests/application/profile. No linking, method removal, OTP,
self-deletion, object-storage I/O, or session revocation is implemented here.
