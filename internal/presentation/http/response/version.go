package presentationhttpresponse

type VersionResponse struct {
	Name    string `json:"name" example:"Alex Doe"`
	Version string `json:"version" example:"v1.0.0-dev.1"`
}
