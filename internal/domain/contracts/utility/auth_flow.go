package domaincontractsutility

// AuthFlow generates opaque URL-safe secrets and their stable SHA-256 digests.
type AuthFlow interface {
	Generate() (string, error)
	Hash(value string) string
}
