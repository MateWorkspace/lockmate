package presentationhttpresponse

import "encoding/json"

// JwksResponse is the standard JSON Web Key Set document.
type JwksResponse struct {
	Keys []json.RawMessage `json:"keys"`
}

func Jwks(keys []json.RawMessage) JwksResponse {
	if keys == nil {
		keys = []json.RawMessage{}
	}

	return JwksResponse{Keys: keys}
}
