package management_test

import (
	"context"
	"errors"
	"fmt"
	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
	"strings"
	"testing"
)

func TestClientAuthRedirectUpdates(t *testing.T) {
	for _, mode := range []string{"replace", "clear", "conflict", "invalid", "commit failure", "cosmetic"} {
		t.Run(mode, func(t *testing.T) {
			f := testshared.NewFixture()
			calls := 0
			uri := "https://app.example.com/callback"
			request := domainusecasesmanagement.UpdateClientRequest{AuthRedirectUri: &uri}
			switch mode {
			case "clear":
				request.AuthRedirectUri = nil
				request.SetAuthRedirectUriNull = true
			case "conflict":
				request.SetAuthRedirectUriNull = true
			case "invalid":
				uri = "https://app.example.com/callback?next=evil"
			case "commit failure":
				f.CommitErr = errors.New("commit failed")
			case "cosmetic":
				request.AuthRedirectUri = nil
			}
			f.Client.UpdateByIdFn = func(ctx context.Context, spaceId, id int64, args domaincontractsrepository.UpdateClientArgs) error {
				testshared.RequireTx(t, ctx)
				calls++
				if spaceId != 10 || id != 30 || args.AuthRedirectUri != request.AuthRedirectUri || args.SetAuthRedirectUriNull != request.SetAuthRedirectUriNull {
					t.Fatal("redirect update mapping lost")
				}
				return nil
			}
			err := client(f).UpdateById(context.Background(), f.Caller, 10, 30, request)
			switch mode {
			case "conflict":
				if !errors.Is(err, domainmodelserror.BadArgs) || calls != 0 {
					t.Fatal(err)
				}
			case "invalid":
				if !errors.Is(err, domainmodelserror.ClientAuthRedirectUriInvalid) || calls != 0 {
					t.Fatal(err)
				}
			case "commit failure":
				if err != f.CommitErr {
					t.Fatal(err)
				}
			default:
				if err != nil || calls != 1 {
					t.Fatal(err)
				}
			}
			info := 0
			for _, entry := range f.Logger.Entries {
				if entry.Level == "info" {
					info++
					if entry.Message != "updated client authentication redirect" {
						t.Fatal("generic redirect log")
					}
				}
			}
			expected := 0
			if mode == "replace" || mode == "clear" {
				expected = 1
			}
			if info != expected {
				t.Fatalf("commit success logging: %d", info)
			}
		})
	}
}

func TestClientCreationNeverLogsOneTimeSecret(t *testing.T) {
	f := testshared.NewFixture()
	f.Client.CreateFn = func(context.Context, int64, domaincontractsrepository.CreateClientArgs) (int64, error) {
		return 30, nil
	}
	result, err := client(f).Create(context.Background(), f.Caller, 10, domainusecasesmanagement.CreateClientRequest{Slug: "web", Name: "Web"})
	if err != nil || result.Secret != "raw-secret" {
		t.Fatal(err)
	}
	if len(f.Logger.Entries) != 1 || f.Logger.Entries[0].Meta["id"] != int64(30) || strings.Contains(fmt.Sprint(f.Logger.Entries), "raw-secret") {
		t.Fatal("client secret logged")
	}
}
