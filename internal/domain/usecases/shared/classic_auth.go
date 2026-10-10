package domainusecasesshared

import domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"

// ClassicAuthView reuses the repository projection that excludes password hashes.
type ClassicAuthView = domaincontractsrepository.ClassicAuthSummary
