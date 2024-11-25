package litellm

type Team struct {
	TeamAlias        string                     `json:"team_alias,omitempty"`
	TeamId           string                     `json:"team_id,omitempty"`
	Metadata         map[string]string          `json:"metadata,omitempty"`
	TpmLimit         int                        `json:"tpm_limit,omitempty"`
	RpmLimit         int                        `json:"rpm_limit,omitempty"`
	MaxBudget        float64                    `json:"max_budget,omitempty"`
	BudgetDuration   string                     `json:"budget_duration,omitempty"`
	Models           []string                   `json:"models,omitempty"`
	Blocked          bool                       `json:"blocked,omitempty"`
	MembersWithRoles []TeamMembershipMemberRole `json:"members_with_roles,omitempty"`
}

type TeamInfoResponse struct {
	TeamId   string `json:"team_id"`
	TeamInfo Team   `json:"team_info"`
}
