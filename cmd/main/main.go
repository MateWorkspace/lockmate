package main

import (
	"os"

	compositionmain "github.com/MateWorkspace/lockmate/internal/composition/main"
)

// @title Lockmate
// @version 1.0.0-dev.1
// @description Multi-space authentication and access management. Client-authenticated operations require credentials and a nonzero execution UUID; Google browser navigation uses saved flow state and binding.
// @BasePath /api
// @tag.name 00. Version
// @tag.description Service information.
// @tag.name 01. Auth
// @tag.description Classic registration, sign-in, refresh, Google sign-in/registration, and single-use authentication handoff.
// @tag.name 02. Profile
// @tag.description The authenticated user's own profile and credentials.
// @tag.name 03. Spaces
// @tag.description Space administration.
// @tag.name 04. Permissions
// @tag.description Permission administration.
// @tag.name 05. Roles
// @tag.description Role administration.
// @tag.name 06. Users
// @tag.description User administration.
// @tag.name 07. Classic Auth
// @tag.description ClassicAuth administration.
// @tag.name 08. Google Auth
// @tag.description GoogleAuth administration.
// @tag.name 09. Clients
// @tag.description Client administration.
// @tag.name 10. Client Origins
// @tag.description Client origin allowlist administration.
// @tag.name 11. Space Members
// @tag.description SpaceMember administration.
// @tag.name 12. Role Permissions
// @tag.description RolePermission administration.
// @tag.name 13. Space Member Roles
// @tag.description SpaceMemberRole administration.
//
// @securityDefinitions.apikey ClientSlug
// @in header
// @name X-Client-Slug
// @description Registered client slug.
//
// @securityDefinitions.apikey ClientSecret
// @in header
// @name X-Client-Secret
// @description Client secret supplied by the consuming application.
//
// @securityDefinitions.apikey ClientExecId
// @in header
// @name X-Client-Exec-Id
// @description Nonzero UUID used for execution tracing, not authentication.
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer followed by a space-scoped access token.
func main() {
	os.Exit(compositionmain.Launch())
}
