package domainmodelsapp

// ClientIdentity is produced by client authentication, never by request mapping.
// Usecases recheck the live client and space before trusting this identity.
// User tokens remain scoped to SpaceId rather than an individual client.
type ClientIdentity struct {
	Id      int64  `json:"id"`
	SpaceId int64  `json:"space_id"`
	Slug    string `json:"slug"`
}
