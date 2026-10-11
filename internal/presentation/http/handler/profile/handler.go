package presentationhttphandlerprofile

import (
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
)

type handler struct {
	accountUseCase  domainusecasesprofile.Account
	securityUseCase domainusecasesprofile.Security
}

func NewHandler(accountUseCase domainusecasesprofile.Account, securityUseCase domainusecasesprofile.Security) *handler {
	return &handler{
		accountUseCase:  accountUseCase,
		securityUseCase: securityUseCase,
	}
}
