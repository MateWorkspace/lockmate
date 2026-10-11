package presentationhttpresponse

import (
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
)

type ClientSecretResponse struct {
	Client ClientResponse `json:"client"`
	Secret string         `json:"secret" example:"returned-only-once"`
}

func ClientSecret(value domainusecasesmanagement.ClientSecretResult) ClientSecretResponse {
	return ClientSecretResponse{Client: Client(value.Client), Secret: value.Secret}
}
