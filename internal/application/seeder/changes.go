package applicationseeder

// changes retains identifiers only; credential material never enters success logs.
type record struct {
	id         int64
	identifier string
}

type changes struct {
	spaceId         int64
	spaces          []record
	permissions     []record
	roles           []record
	rolePermissions []record
	users           []record
	classicAuths    []record
	members         []record
	memberRoles     []record
	clients         []record
	origins         []record
}
