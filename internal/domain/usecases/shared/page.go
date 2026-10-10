package domainusecasesshared

// Page preserves repository pagination semantics without promising a snapshot.
type Page[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}
