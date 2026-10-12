package infrastructureutilitytoken

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"time"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/golang-jwt/jwt/v5"
)

type jwtImpl struct {
	refreshSecret   []byte
	accessDuration  time.Duration
	refreshDuration time.Duration
	now             func() time.Time
}

type accessClaims struct {
	domainmodelsentity.AccessClaims
	jwt.RegisteredClaims
	TokenType string `json:"token_type"`
}

type refreshClaims struct {
	domainmodelsentity.RefreshClaims
	jwt.RegisteredClaims
	TokenType string `json:"token_type"`
}

func NewJwtImpl(refreshSecret string, accessDuration, refreshDuration time.Duration) domaincontractsutility.Token {
	return &jwtImpl{
		refreshSecret:   []byte(refreshSecret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		now:             time.Now,
	}
}

func (j *jwtImpl) GenerateAccess(claims *domainmodelsentity.AccessClaims, kid string, privateKeyPKCS8 []byte) (string, error) {
	if claims == nil {
		return "", domainmodelserror.NewError("access claims are required", domainmodelserror.BadArgs, nil)
	}
	if kid == "" || len(privateKeyPKCS8) == 0 {
		return "", domainmodelserror.NewError("an access signing key is required", domainmodelserror.TokenFailed, nil)
	}

	key, err := parseECPrivateKey(privateKeyPKCS8)
	if err != nil {
		return "", err
	}

	registered, err := j.registered(claims.UserId, claims.SpaceId, claims.SpaceMemberId, j.accessDuration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, accessClaims{
		AccessClaims:     *claims,
		RegisteredClaims: registered,
		TokenType:        "access",
	})
	token.Header["kid"] = kid

	signed, err := token.SignedString(key)
	if err != nil {
		return "", domainmodelserror.NewError("failed to generate token", domainmodelserror.TokenFailed, err)
	}

	return signed, nil
}

func (j *jwtImpl) GenerateRefresh(claims *domainmodelsentity.RefreshClaims) (string, error) {
	if claims == nil {
		return "", domainmodelserror.NewError("refresh claims are required", domainmodelserror.BadArgs, nil)
	}
	registered, err := j.registered(claims.UserId, claims.SpaceId, claims.SpaceMemberId, j.refreshDuration)
	if err != nil {
		return "", err
	}
	return j.sign(
		refreshClaims{
			RefreshClaims:    *claims,
			RegisteredClaims: registered,
			TokenType:        "refresh",
		},
		j.refreshSecret,
	)
}

func (j *jwtImpl) ValidateAccess(keys []domainmodelsentity.SigningKey, token string) (*domainmodelsentity.AccessClaims, error) {
	if len(keys) == 0 || token == "" {
		return nil, domainmodelserror.NewError("token and signing keys are required", domainmodelserror.BadArgs, nil)
	}

	publicKeys := make(map[string]*ecdsa.PublicKey, len(keys))
	for i := range keys {
		public, err := parseJWKPublicKey(keys[i].PublicJWK)
		if err != nil {
			continue
		}

		publicKeys[keys[i].Kid] = public
	}
	if len(publicKeys) == 0 {
		return nil, domainmodelserror.NewError("no verifiable signing keys are available", domainmodelserror.TokenFailed, nil)
	}

	claims := &accessClaims{}
	keyFunc := func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, domainmodelserror.NewError("unexpected token signing method", domainmodelserror.TokenInvalid, nil)
		}

		kid, _ := token.Header["kid"].(string)
		public, ok := publicKeys[kid]
		if !ok {
			return nil, domainmodelserror.NewError("token signing key is unknown", domainmodelserror.TokenInvalid, nil)
		}

		return public, nil
	}

	parsed, err := jwt.ParseWithClaims(token, claims, keyFunc,
		jwt.WithValidMethods([]string{"ES256"}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(j.now),
	)
	if err != nil {
		return nil, classify(err)
	}
	if !parsed.Valid {
		return nil, domainmodelserror.NewError("token is invalid", domainmodelserror.TokenInvalid, nil)
	}

	if err := j.identity(claims.UserId, claims.SpaceId, claims.SpaceMemberId, claims.TokenType, "access", claims.RegisteredClaims); err != nil {
		return nil, err
	}

	return &claims.AccessClaims, nil
}

func (j *jwtImpl) ValidateRefresh(spaceId int64, token string) (*domainmodelsentity.RefreshClaims, error) {
	claims := &refreshClaims{}
	if err := j.parse(spaceId, token, claims, j.refreshSecret, "refresh"); err != nil {
		return nil, err
	}
	if err := j.identity(claims.UserId, claims.SpaceId, claims.SpaceMemberId, claims.TokenType, "refresh", claims.RegisteredClaims); err != nil {
		return nil, err
	}
	if claims.SpaceId != spaceId {
		return nil, domainmodelserror.NewError("token identity or scope is invalid", domainmodelserror.TokenInvalid, nil)
	}
	return &claims.RefreshClaims, nil
}

func (j *jwtImpl) registered(userId, spaceId, memberId int64, duration time.Duration) (jwt.RegisteredClaims, error) {
	if userId <= 0 || spaceId <= 0 || memberId <= 0 {
		return jwt.RegisteredClaims{}, domainmodelserror.NewError("positive identity and scope ids are required", domainmodelserror.BadArgs, nil)
	}
	if duration < time.Second {
		return jwt.RegisteredClaims{}, domainmodelserror.NewError("token lifetime must be at least one second", domainmodelserror.TokenFailed, nil)
	}
	now := j.now().UTC()
	return jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userId, 10),
		Audience:  jwt.ClaimStrings{audience(spaceId)},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
	}, nil
}

func (j *jwtImpl) sign(claims jwt.Claims, secret []byte) (string, error) {
	if len(secret) == 0 {
		return "", domainmodelserror.NewError("token secret is not configured", domainmodelserror.TokenFailed, nil)
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		return "", domainmodelserror.NewError("failed to generate token", domainmodelserror.TokenFailed, err)
	}
	return raw, nil
}

func (j *jwtImpl) parse(spaceId int64, raw string, claims jwt.Claims, secret []byte, purpose string) error {
	if spaceId <= 0 || raw == "" {
		return domainmodelserror.NewError("token and positive space id are required", domainmodelserror.BadArgs, nil)
	}
	if len(secret) == 0 {
		return domainmodelserror.NewError("token secret is not configured", domainmodelserror.TokenFailed, nil)
	}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithAudience(audience(spaceId)), jwt.WithIssuedAt(), jwt.WithTimeFunc(j.now))
	if err != nil {
		return classify(err)
	}
	if !token.Valid {
		return domainmodelserror.NewError("token is invalid", domainmodelserror.TokenInvalid, nil)
	}
	return nil
}

func (j *jwtImpl) identity(userId, spaceId, memberId int64, purpose, expectedPurpose string, registered jwt.RegisteredClaims) error {
	if userId <= 0 || spaceId <= 0 || memberId <= 0 || purpose != expectedPurpose || registered.Subject != strconv.FormatInt(userId, 10) || registered.IssuedAt == nil || registered.NotBefore == nil || registered.ExpiresAt == nil || len(registered.Audience) != 1 || registered.Audience[0] != audience(spaceId) {
		return domainmodelserror.NewError("token identity or scope is invalid", domainmodelserror.TokenInvalid, nil)
	}
	if registered.IssuedAt.After(registered.NotBefore.Time) || !registered.NotBefore.Before(registered.ExpiresAt.Time) {
		return domainmodelserror.NewError("token timestamps are invalid", domainmodelserror.TokenInvalid, nil)
	}
	return nil
}

func classify(err error) error {
	kind := domainmodelserror.TokenInvalid
	if errors.Is(err, jwt.ErrTokenExpired) && !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		kind = domainmodelserror.TokenExpired
	}
	return domainmodelserror.NewError("failed to validate token", kind, err)
}

func audience(spaceId int64) string { return "lockmate:space:" + strconv.FormatInt(spaceId, 10) }

func parseECPrivateKey(pkcs8 []byte) (*ecdsa.PrivateKey, error) {
	parsed, err := x509.ParsePKCS8PrivateKey(pkcs8)
	if err != nil {
		return nil, domainmodelserror.NewError("failed to parse access signing key", domainmodelserror.TokenFailed, err)
	}

	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, domainmodelserror.NewError("access signing key is not an ECDSA key", domainmodelserror.TokenFailed, nil)
	}

	return key, nil
}

func parseJWKPublicKey(raw json.RawMessage) (*ecdsa.PublicKey, error) {
	var value struct {
		Kty string `json:"kty"`
		Crv string `json:"crv"`
		X   string `json:"x"`
		Y   string `json:"y"`
	}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	if value.Kty != "EC" || value.Crv != "P-256" {
		return nil, domainmodelserror.NewError("unsupported public JWK", domainmodelserror.TokenFailed, nil)
	}

	x, err := base64.RawURLEncoding.DecodeString(value.X)
	if err != nil {
		return nil, err
	}
	y, err := base64.RawURLEncoding.DecodeString(value.Y)
	if err != nil {
		return nil, err
	}

	public := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(x),
		Y:     new(big.Int).SetBytes(y),
	}
	if !public.Curve.IsOnCurve(public.X, public.Y) {
		return nil, domainmodelserror.NewError("public JWK point is not on the curve", domainmodelserror.TokenFailed, nil)
	}

	return public, nil
}
