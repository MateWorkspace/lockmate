package utility_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilityauthflow "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/auth_flow"
	infrastructureutilitygoogleoauth "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/google_oauth"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type attemptMemory struct {
	mu     sync.Mutex
	values map[string]domainmodelsentity.GoogleOAuthAttempt
}

func (s *attemptMemory) Create(_ context.Context, value domainmodelsentity.GoogleOAuthAttempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[value.StateHash] = value
	return nil
}
func (s *attemptMemory) Consume(_ context.Context, state, binding string) (*domainmodelsentity.GoogleOAuthAttempt, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.values[state]
	if !ok || v.BindingHash != binding || !time.Now().Before(v.ExpiresAt) {
		return nil, domainmodelserror.NewError("invalid attempt", domainmodelserror.GoogleOAuthStateInvalid, nil)
	}
	delete(s.values, state)
	return &v, nil
}
func (s *attemptMemory) PurgeExpired(context.Context, int64) (int64, error) { return 0, nil }

func TestGoogleOIDCProvider(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	flow := infrastructureutilityauthflow.NewRandomImpl()
	for _, mode := range []string{"valid", "wrong issuer", "wrong audience", "expired", "wrong nonce", "unverified email", "missing subject", "missing id token", "bad signature", "provider failure", "malformed claims", "denied", "timeout", "key fetch failure", "key rotation"} {
		t.Run(mode, func(t *testing.T) {
			store := &attemptMemory{values: map[string]domainmodelsentity.GoogleOAuthAttempt{}}
			var token string
			var exchangeCalls, keyCalls atomic.Int64
			var expectedChallenge string
			kid := "one"
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/keys" {
					keyCalls.Add(1)
					if mode == "key fetch failure" {
						w.WriteHeader(503)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(map[string]any{"keys": []any{map[string]any{"kty": "RSA", "kid": kid, "use": "sig", "alg": "RS256", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes())}}})
					return
				}
				exchangeCalls.Add(1)
				if err := r.ParseForm(); err != nil {
					t.Error(err)
				}
				challenge := base64.RawURLEncoding.EncodeToString(mustDecodeHex(t, flow.Hash(r.Form.Get("code_verifier"))))
				if challenge != expectedChallenge || r.Form.Get("redirect_uri") != "https://lockmate-dev.dhonan.my.id/api/v1/auth/google/callback" {
					t.Error("PKCE or callback URI lost")
				}
				if mode == "timeout" {
					time.Sleep(150 * time.Millisecond)
				}
				w.Header().Set("Content-Type", "application/json")
				if mode == "provider failure" {
					w.WriteHeader(400)
					_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"PRIVATE_PROVIDER_TOKEN private-code private-client-secret"}`))
					return
				}
				if mode == "missing id token" {
					_, _ = w.Write([]byte(`{"access_token":"private-provider-access","token_type":"Bearer"}`))
					return
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "private-provider-access", "token_type": "Bearer", "id_token": token})
			}))
			defer server.Close()
			client := &http.Client{Timeout: time.Second}
			if mode == "timeout" {
				client.Timeout = 30 * time.Millisecond
			}
			adapter := infrastructureutilitygoogleoauth.NewOIDCImpl(store, flow, client, infrastructureutilitygoogleoauth.Config{Enabled: true, ClientId: "client", ClientSecret: "private-client-secret", CallbackUri: "https://lockmate-dev.dhonan.my.id/api/v1/auth/google/callback", AttemptDuration: time.Minute, Issuer: server.URL, KeysUrl: server.URL + "/keys", Endpoint: oauth2.Endpoint{AuthURL: server.URL + "/authorize", TokenURL: server.URL + "/token", AuthStyle: oauth2.AuthStyleInParams}})
			run := func() (*domaincontractsutility.CompleteGoogleOAuthResult, error) {
				started, err := adapter.Begin(context.Background(), domaincontractsutility.BeginGoogleOAuthArgs{ClientId: 1, SpaceId: 2, Binding: "private-binding", RedirectUri: "https://app.example.com/callback", ClientState: "project-state"})
				if err != nil {
					t.Fatal(err)
				}
				auth, err := url.Parse(started.AuthorizationUrl)
				if err != nil {
					t.Fatal(err)
				}
				if auth.Query().Get("scope") != "openid email profile" || auth.Query().Get("code_challenge_method") != "S256" || auth.Query().Get("access_type") == "offline" {
					t.Fatal("incorrect authorization request")
				}
				expectedChallenge = auth.Query().Get("code_challenge")
				claims := jwt.MapClaims{"iss": server.URL, "aud": "client", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix(), "sub": "stable-sub", "nonce": auth.Query().Get("nonce"), "email": "user@example.com", "email_verified": true, "name": "Provider Name"}
				switch mode {
				case "wrong issuer":
					claims["iss"] = "https://wrong.example.com"
				case "wrong audience":
					claims["aud"] = "another-client"
				case "expired":
					claims["exp"] = time.Now().Add(-time.Hour).Unix()
				case "wrong nonce":
					claims["nonce"] = "wrong"
				case "unverified email":
					claims["email_verified"] = false
				case "missing subject":
					claims["sub"] = ""
				case "malformed claims":
					claims["email_verified"] = "wrong-type"
				}
				signed := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
				signed.Header["kid"] = kid
				token, err = signed.SignedString(key)
				if err != nil {
					t.Fatal(err)
				}
				if mode == "bad signature" {
					token = token[:strings.LastIndex(token, ".")+1] + base64.RawURLEncoding.EncodeToString(make([]byte, 256))
				}
				args := domaincontractsutility.CompleteGoogleOAuthArgs{State: started.State, Binding: "private-binding", Code: "private-code"}
				if mode == "denied" {
					args.Code = ""
					args.ProviderError = "access_denied"
				}
				// Wrong binding must neither exchange nor consume a legitimate attempt.
				wrong := args
				wrong.Binding = "wrong-binding"
				if _, err := adapter.Complete(context.Background(), wrong); !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
					t.Fatalf("binding accepted: %v", err)
				}
				result, err := adapter.Complete(context.Background(), args)
				if _, replay := adapter.Complete(context.Background(), args); !errors.Is(replay, domainmodelserror.GoogleOAuthStateInvalid) {
					t.Fatalf("replay accepted: %v", replay)
				}
				return result, err
			}
			result, err := run()
			if mode == "valid" || mode == "key rotation" {
				if err != nil || result == nil || result.ClientId != 1 || result.SpaceId != 2 || result.Identity.Subject != "stable-sub" || result.RedirectUri != "https://app.example.com/callback" {
					t.Fatalf("verified flow: %v", err)
				}
				if mode == "valid" {
					if _, err := run(); err != nil {
						t.Fatal(err)
					}
					if keyCalls.Load() != 1 {
						t.Fatal("signing keys were not cached")
					}
				}
				if mode == "key rotation" {
					kid = "two"
					rotation, err := rsa.GenerateKey(rand.Reader, 2048)
					if err != nil {
						t.Fatal(err)
					}
					key = rotation
					if _, err := run(); err != nil {
						t.Fatal(err)
					}
					if keyCalls.Load() < 2 {
						t.Fatal("rotated signing key was not fetched")
					}
				}
			} else if mode == "denied" {
				if err != nil || !result.Denied || exchangeCalls.Load() != 0 {
					t.Fatalf("denial exchanged code: %v", err)
				}
			} else {
				if err == nil || result != nil {
					t.Fatal("invalid provider response accepted")
				}
				for _, secret := range []string{"PRIVATE_PROVIDER_TOKEN", "private-code", "private-client-secret", "private-provider-access"} {
					if strings.Contains(err.Error(), secret) {
						t.Fatal("provider diagnostics leaked secrets")
					}
				}
				if mode == "timeout" && !errors.Is(err, domainmodelserror.Timeout) {
					t.Fatalf("timeout classification: %v", err)
				}
			}
		})
	}
}

func mustDecodeHex(t *testing.T, value string) []byte {
	t.Helper()
	decoded, err := hex.DecodeString(value)
	if err != nil {
		t.Fatal(err)
	}
	return decoded
}

func TestGoogleDisabledAndExpiredAttempt(t *testing.T) {
	flow := infrastructureutilityauthflow.NewRandomImpl()
	store := &attemptMemory{values: map[string]domainmodelsentity.GoogleOAuthAttempt{}}
	disabled := infrastructureutilitygoogleoauth.NewOIDCImpl(store, flow, http.DefaultClient, infrastructureutilitygoogleoauth.Config{})
	if _, err := disabled.Begin(context.Background(), domaincontractsutility.BeginGoogleOAuthArgs{}); !errors.Is(err, domainmodelserror.Unimplemented) {
		t.Fatal(err)
	}
	store.values[flow.Hash("state")] = domainmodelsentity.GoogleOAuthAttempt{BindingHash: flow.Hash("binding"), ExpiresAt: time.Now().Add(-time.Second)}
	enabled := infrastructureutilitygoogleoauth.NewOIDCImpl(store, flow, http.DefaultClient, infrastructureutilitygoogleoauth.Config{Enabled: true})
	if _, err := enabled.Complete(context.Background(), domaincontractsutility.CompleteGoogleOAuthArgs{State: "state", Binding: "binding", Code: "code"}); !errors.Is(err, domainmodelserror.GoogleOAuthStateInvalid) {
		t.Fatal(err)
	}
}
