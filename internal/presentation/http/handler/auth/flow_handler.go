package presentationhttphandlerauth

import domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"

type flowHandler struct {
	google  domainusecasesauth.Google
	handoff domainusecasesauth.Handoff
}

func NewFlowHandler(google domainusecasesauth.Google, handoff domainusecasesauth.Handoff) *flowHandler {
	return &flowHandler{
		google:  google,
		handoff: handoff,
	}
}
