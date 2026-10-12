package presentationgrpcmapper

import "encoding/json"

// Preferences serializes a raw JSON preference value, normalizing empty to `{}`.
func Preferences(value json.RawMessage) string {
	if len(value) == 0 {
		return "{}"
	}

	return string(value)
}

// OptionalPreferences carries a raw JSON preference replacement when supplied.
func OptionalPreferences(value *string) *json.RawMessage {
	if value == nil {
		return nil
	}

	raw := json.RawMessage(*value)
	return &raw
}

// RequiredPreferences carries a raw JSON preference replacement.
func RequiredPreferences(value string) json.RawMessage {
	return json.RawMessage(value)
}
