package domainusecasesshared

import domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"

// ClientView reuses the repository projection that excludes secret hashes.
type ClientView = domaincontractsrepository.ClientSummary
