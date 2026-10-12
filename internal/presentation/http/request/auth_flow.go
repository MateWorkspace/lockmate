package presentationhttprequest

type AuthGoogleStartRequest struct {
	State string `json:"state" example:"f8d5b65dc01947658b04e768c98771ae"`
}

type AuthExchangeRequest struct {
	Code string `json:"code"`
}
