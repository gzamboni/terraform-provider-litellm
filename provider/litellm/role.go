package litellm

type Role string

const (
	PROXY_ADMIN          = "proxy_admin"
	PROXY_ADMIN_VIEWER   = "proxy_admin_viewer"
	ORG_ADMIN            = "org_admin"
	INTERNAL_USER        = "internal_user"
	INTERNAL_USER_VIEWER = "internal_user_viewer"
)

var ROLE_LIST = map[Role]string{
	PROXY_ADMIN:          "proxy_admin",
	PROXY_ADMIN_VIEWER:   "proxy_admin_viewer",
	ORG_ADMIN:            "org_admin",
	INTERNAL_USER:        "internal_user",
	INTERNAL_USER_VIEWER: "internal_user_viewer",
}

func ValidateRole(inputRole string) (Role, bool) {
	//Role is empty no role to return also validation is correct, this field is normally optional and can is omitted if empty
	if len(inputRole) <= 0 {
		return "", true
	}

	for role, strValue := range ROLE_LIST {
		if strValue == inputRole {
			return role, true
		}
	}
	return "", false
}
