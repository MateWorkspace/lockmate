package infrastructureutilitytoken

import (
	"errors"
	"strconv"
	"time"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/golang-jwt/jwt/v5"
)

type jwtImpl struct {
	accessSecret    []byte
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

func NewJwtImpl(accessSecret, refreshSecret string, accessDuration, refreshDuration time.Duration) domaincontractsutility.Token {
	return &jwtImpl{
		accessSecret:    []byte(accessSecret),
		refreshSecret:   []byte(refreshSecret),
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		now:             time.Now,
	}
}

func (j *jwtImpl) GenerateAccess(claims *domainmodelsentity.AccessClaims) (string, error) {
	if claims == nil {
		return "", domainmodelserror.NewError("access claims are required", domainmodelserror.BadArgs, nil)
	}
	registered, err := j.registered(claims.UserId, claims.SpaceId, claims.SpaceMemberId, j.accessDuration)
	if err != nil {
		return "", err
	}
	return j.sign(
		accessClaims{
			AccessClaims:     *claims,
			RegisteredClaims: registered,
			TokenType:        "access",
		},
		j.accessSecret,
	)
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

func (j *jwtImpl) ValidateAccess(spaceId int64, token string) (*domainmodelsentity.AccessClaims, error) {
	claims := &accessClaims{}
	if err := j.parse(spaceId, token, claims, j.accessSecret); err != nil {
		return nil, err
	}
	if err := j.identity(
		claims.UserId,
		claims.SpaceId,
		claims.SpaceMemberId,
		spaceId,
		claims.TokenType,
		"access",
		claims.RegisteredClaims,
	); err != nil {
		return nil, err
	}
	return &claims.AccessClaims, nil
}

func (j *jwtImpl) ValidateRefresh(spaceId int64, token string) (*domainmodelsentity.RefreshClaims, error) {
	claims := &refreshClaims{}
	if err := j.parse(spaceId, token, claims, j.refreshSecret); err != nil {
		return nil, err
	}
	if err := j.identity(
		claims.UserId,
		claims.SpaceId,
		claims.SpaceMemberId,
		spaceId,
		claims.TokenType,
		"refresh",
		claims.RegisteredClaims,
	); err != nil {
		return nil, err
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

func (j *jwtImpl) parse(spaceId int64, raw string, claims jwt.Claims, secret []byte) error {
	if spaceId <= 0 || raw == "" {
		return domainmodelserror.NewError("token and positive space id are required", domainmodelserror.BadArgs, nil)
	}
	if len(secret) == 0 {
		return domainmodelserror.NewError("token secret is not configured", domainmodelserror.TokenFailed, nil)
	}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) { return secret, nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithAudience(audience(spaceId)), jwt.WithIssuedAt(), jwt.WithTimeFunc(j.now))
	if err != nil {
		kind := domainmodelserror.TokenInvalid
		if errors.Is(err, jwt.ErrTokenExpired) && !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			kind = domainmodelserror.TokenExpired
		}
		return domainmodelserror.NewError("failed to validate token", kind, err)
	}
	if !token.Valid {
		return domainmodelserror.NewError("token is invalid", domainmodelserror.TokenInvalid, nil)
	}
	return nil
}

func (j *jwtImpl) identity(userId, spaceId, memberId, expectedSpace int64, purpose, expectedPurpose string, registered jwt.RegisteredClaims) error {
	if userId <= 0 || spaceId <= 0 || memberId <= 0 || spaceId != expectedSpace || purpose != expectedPurpose || registered.Subject != strconv.FormatInt(userId, 10) || registered.IssuedAt == nil || registered.NotBefore == nil || registered.ExpiresAt == nil || len(registered.Audience) != 1 || registered.Audience[0] != audience(spaceId) {
		return domainmodelserror.NewError("token identity or scope is invalid", domainmodelserror.TokenInvalid, nil)
	}
	if registered.IssuedAt.After(registered.NotBefore.Time) || !registered.NotBefore.Before(registered.ExpiresAt.Time) {
		return domainmodelserror.NewError("token timestamps are invalid", domainmodelserror.TokenInvalid, nil)
	}
	return nil
}

func audience(spaceId int64) string { return "lockmate:space:" + strconv.FormatInt(spaceId, 10) }
