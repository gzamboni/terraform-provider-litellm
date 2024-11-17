package litellm

type User struct {
	UserId          string            `json:"user_id,omitempty"`
	UserAlias       string            `json:"user_alias,omitempty"`
	Teams           []string          `json:"teams,omitempty"`
	UserEmail       string            `json:"user_email,omitempty"`
	SendInviteEmail bool              `json:"send_invite_email,omitempty"`
	UserRole        Role              `json:"user_role,omitempty"`
	MaxBudget       float64           `json:"max_budget,omitempty"`
	BudgetDuration  string            `json:"budget_duration,omitempty"`
	Models          []string          `json:"models,omitempty"`
	TpmLimit        int               `json:"tpm_limit,omitempty"`
	RpmLimit        int               `json:"rpm_limit,omitempty"`
	AutoCreateKey   bool              `json:"auto_create_key,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

type UserInfo struct {
	UserId   string `json:"user_id"`
	UserInfo User   `json:"user_info"`
}
