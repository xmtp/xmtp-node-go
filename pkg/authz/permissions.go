package authz

type Permission int64

const (
	Unspecified Permission = 0
	AllowAll    Permission = 1
	Priority    Permission = 2
	Denied      Permission = 3
)

// Add a string function to use for logging
func (p Permission) String() string {
	switch p {
	case AllowAll:
		return "allow_all"
	case Priority:
		return "priority"
	case Denied:
		return "denied"
	case Unspecified:
		return "unspecified"
	}
	return "unknown"
}

func permissionFromString(permission string) Permission {
	switch permission {
	case "allow_all":
		return AllowAll
	case "priority":
		return Priority
	case "denied":
		return Denied
	default:
		return Unspecified
	}
}
