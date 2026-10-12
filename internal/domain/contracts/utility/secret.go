package domaincontractsutility

// Secret seals and opens private signing material with the configured
// key-encryption keys and generates asymmetric keypairs with public JWKs.
type Secret interface {
	// Fingerprint identifies the current key-encryption key.
	Fingerprint() string
	// PreviousFingerprint identifies the optional previous key-encryption key,
	// or an empty string when none is configured.
	PreviousFingerprint() string
	// Seal encrypts plaintext with the current key-encryption key.
	Seal(plaintext []byte) (cipher []byte, nonce []byte, err error)
	// Open decrypts ciphertext using the key matching the supplied fingerprint.
	Open(cipher []byte, nonce []byte, fingerprint string) (plaintext []byte, err error)
	// GenerateKeypair returns a PKCS#8 private key, its public JWK, a key ID,
	// and the signing algorithm.
	GenerateKeypair() (privateKeyPKCS8 []byte, publicJWK []byte, kid string, algorithm string, err error)
}
