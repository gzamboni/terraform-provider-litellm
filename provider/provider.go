// provider.go
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func NewProvider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LITELLM_API_TOKEN", nil),
				Description: "The API token (bearer token) for accessing the LiteLLM API.",
			},
			"api_base_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LITELLM_API_BASE_URL", nil),
				Description: "The base URL for the LiteLLM API (e.g., 'https://your-litellm-instance.com').",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"litellm_model":           resourceModel(),
			"litellm_user":            resourceUser(),
			"litellm_team":            resourceTeam(),
			"litellm_team_membership": resourceTeamMembership(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	apiToken := d.Get("api_token").(string)
	apiBaseURL := d.Get("api_base_url").(string)

	if apiToken == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API token is missing",
			Detail:   "An API token is required to authenticate with the LiteLLM API.",
		})
		return nil, diags
	}
	if apiBaseURL == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API base URL is missing",
			Detail:   "An API base URL is required to access the LiteLLM API.",
		})
		return nil, diags
	}

	client := &LitellmClient{}
	client.ApiBaseURL = apiBaseURL
	client.ApiToken = apiToken

	return client, diags
}
