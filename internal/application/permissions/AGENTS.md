# Application permission requirements

permissions.go is the source-based permission catalog for protected application
operations. Group immutable constants by usecase group and capability; name each
constant after its group, capability, and method. Keep the existing permission
slug vocabulary consistent with domain contracts and future seed definitions.

Each protected application method calls its injected PermissionGuard directly,
using catalog constants immediately below its tag declaration. The check must
precede identity reads, validation, and transaction creation. Live identity and
space-boundary checks remain afterward in the feature. Authentication and trace
injection use their own identity/credential checks rather than permissions.

Do not add mutable global policies, runtime dispatch, role bypasses, or concrete
infrastructure dependencies. Constructors and domain contracts stay independent
of the catalog. Keep independent permission expectations in application tests so
an accidental policy change cannot silently change both implementation and tests.
