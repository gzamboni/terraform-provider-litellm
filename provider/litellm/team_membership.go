package litellm

import "errors"

type TeamMembershipMemberRole struct {
	Role      string `json:"role"`
	UserId    string `json:"user_id,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

type TeamMembershipMemberAddPayload struct {
	Members []TeamMembershipMemberRole `json:"member"`
	TeamId  string                     `json:"team_id"`
}

type TeamMembershipUserTeamAssociation struct {
	UserId    string `json:"user_id"`
	UserEmail string `json:"user_email"`
	TeamId    string `json:"team_id"`
}

func (t *TeamMembershipMemberRole) validateRole() error {
	if t.Role == "proxy_admin" || t.Role == "proxy_admin_viewer" {
		return errors.New("Team membership a user cannot have the role 'proxy_admin' or 'proxy_admin_viewer' in a team")
	}
	return nil
}
