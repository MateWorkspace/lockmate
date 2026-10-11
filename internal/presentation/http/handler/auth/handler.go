package presentationhttphandlerauth

import (
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
)

type handler struct {
	sessionUseCase domainusecasesauth.Session
}

func NewHandler(sessionUseCase domainusecasesauth.Session) *handler {
	return &handler{
		sessionUseCase: sessionUseCase,
	}
}
