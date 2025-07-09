package authz

// AllowList maintains an allow list for IP addresses
type AllowList interface {
	GetPermission(ipAddress string) Permission
}
