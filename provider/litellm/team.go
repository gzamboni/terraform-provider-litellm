package litellm

type TeamInfoResponse struct {
	TeamId   string `json:"team_id"`
	TeamInfo struct {
		TeamAlias      string            `json:"team_alias"`
		TeamId         string            `json:"team_id"`
		OrganizationId string            `json:"organization_id"`
		Metadata       map[string]string `json:"metadata"`
		TpmLimit       int               `json:"tpm_limit"`
		RpmLimit       int               `json:"rpm_limit"`
		MaxBudget      float64           `json:"max_budget"`
		BudgetDuration string            `json:"budget_duration"`
		Models         []string          `json:"models"`
		Blocked        bool              `json:"blocked"`
	} `json:"team_info"`
}
