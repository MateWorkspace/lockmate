package infrastructureutilitysecret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

const (
	keySize         = 32
	algorithm       = "ES256"
	curveName       = "P-256"
	fingerprintSize = 8
)

type sealImpl struct {
	currentKey    []byte
	previousKey   []byte
	currentPrint  string
	previousPrint string
}

func NewSealImpl(current, previous []byte) (domaincontractsutility.Secret, error) {
	if len(current) != keySize {
		return nil, domainmodelserror.NewError("the current key-encryption key must be 32 bytes", domainmodelserror.SecretKeyInvalid, nil)
	}

	if len(previous) != 0 && len(previous) != keySize {
		return nil, domainmodelserror.NewError("the previous key-encryption key must be 32 bytes", domainmodelserror.SecretKeyInvalid, nil)
	}

	impl := &sealImpl{
		currentKey:   current,
		previousKey:  previous,
		currentPrint: fingerprint(current),
	}
	if len(previous) != 0 {
		impl.previousPrint = fingerprint(previous)
	}

	return impl, nil
}

func (s *sealImpl) Fingerprint() string { return s.currentPrint }

func (s *sealImpl) PreviousFingerprint() string { return s.previousPrint }

func (s *sealImpl) Seal(plaintext []byte) (cipher []byte, nonce []byte, err error) {
	gcm, err := newGCM(s.currentKey)
	if err != nil {
		return nil, nil, domainmodelserror.NewError("failed to initialize key encryption", domainmodelserror.SecretEncryptFailed, err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, domainmodelserror.NewError("failed to generate encryption nonce", domainmodelserror.SecretEncryptFailed, err)
	}

	return gcm.Seal(nil, nonce, plaintext, nil), nonce, nil
}

func (s *sealImpl) Open(cipher []byte, nonce []byte, print string) ([]byte, error) {
	key := s.keyFor(print)
	if key == nil {
		return nil, domainmodelserror.NewError("the matching key-encryption key is not configured", domainmodelserror.SecretKeyInvalid, nil)
	}

	gcm, err := newGCM(key)
	if err != nil {
		return nil, domainmodelserror.NewError("failed to initialize key decryption", domainmodelserror.SecretDecryptFailed, err)
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, domainmodelserror.NewError("encrypted value has an invalid nonce", domainmodelserror.SecretDecryptFailed, nil)
	}

	plaintext, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return nil, domainmodelserror.NewError("failed to decrypt value", domainmodelserror.SecretDecryptFailed, err)
	}

	return plaintext, nil
}

func (s *sealImpl) GenerateKeypair() (privateKeyPKCS8 []byte, publicJWK []byte, kid string, alg string, err error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, "", "", domainmodelserror.NewError("failed to generate signing keypair", domainmodelserror.KeygenFailed, err)
	}

	privateKeyPKCS8, err = x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return nil, nil, "", "", domainmodelserror.NewError("failed to encode signing keypair", domainmodelserror.KeygenFailed, err)
	}

	kid = keyID(&key.PublicKey)
	publicJWK, err = marshalPublicJWK(&key.PublicKey, kid)
	if err != nil {
		return nil, nil, "", "", err
	}

	return privateKeyPKCS8, publicJWK, kid, algorithm, nil
}

func (s *sealImpl) keyFor(print string) []byte {
	switch {
	case print != "" && print == s.currentPrint:
		return s.currentKey
	case print != "" && print == s.previousPrint:
		return s.previousKey
	default:
		return nil
	}
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func fingerprint(key []byte) string {
	sum := sha256.Sum256(key)
	return hex.EncodeToString(sum[:fingerprintSize])
}

func keyID(public *ecdsa.PublicKey) string {
	sum := sha256.Sum256(publicBytes(public))
	return hex.EncodeToString(sum[:fingerprintSize])
}

func marshalPublicJWK(public *ecdsa.PublicKey, kid string) ([]byte, error) {
	size := (public.Curve.Params().BitSize + 7) / 8
	value := map[string]string{
		"kty": "EC",
		"crv": curveName,
		"alg": algorithm,
		"use": "sig",
		"kid": kid,
		"x":   base64.RawURLEncoding.EncodeToString(public.X.FillBytes(make([]byte, size))),
		"y":   base64.RawURLEncoding.EncodeToString(public.Y.FillBytes(make([]byte, size))),
	}

	raw, err := json.Marshal(value)
	if err != nil {
		return nil, domainmodelserror.NewError("failed to encode public JWK", domainmodelserror.KeygenFailed, err)
	}

	return raw, nil
}

func publicBytes(public *ecdsa.PublicKey) []byte {
	size := (public.Curve.Params().BitSize + 7) / 8
	result := make([]byte, 0, 2*size+1)
	result = append(result, public.X.FillBytes(make([]byte, size))...)
	result = append(result, public.Y.FillBytes(make([]byte, size))...)

	return result
}
