package domaincontractsutility

type Keygen interface {
	// Generate returns a random secret, SHA-256 hash, and masked display value.
	// Random generation does not replace persistence-level uniqueness constraints.
	Generate() (raw string, hash string, redacted string, err error)
	Hash(raw string) string
	// Validate checks the key format; it does not establish ownership.
	Validate(raw string) error
	Compare(storedHash string, raw string) error
}
