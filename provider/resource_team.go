package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gzamboni/terraform-provider-litellm/provider/litellm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ResourceTeamSchema = map[string]*schema.Schema{
	"team_alias": {
		Type:        schema.TypeString,
		Optional:    true,
		Required:    false,
		Description: "User defined team alias",
	},
	"team_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Required:    false,
		Description: "The team id of the user. if none passed, we'll generate it",
	},
	"metadata": {
		Type:        schema.TypeMap,
		Required:    false,
		Optional:    true,
		Description: "Metadata for team, store information for team",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"tpm_limit": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "The TPM (Tokens Per Minute) limit for this team - all keys with this team_id will have at max this TPM limit",
	},
	"rpm_limit": {
		Type:        schema.TypeInt,
		Required:    false,
		Optional:    true,
		Description: "The RPM (Requests Per Minute) limit for this team - all keys associated with this team_id will have at max this RPM limit",
	},
	"max_budget": {
		Type:        schema.TypeFloat,
		Required:    false,
		Optional:    true,
		Description: "The maximum budget allocated to the team - all keys for this team_id will have at max this max_budget",
	},
	"budget_duration": {
		Type:        schema.TypeString,
		Required:    false,
		Optional:    true,
		Description: "The duration of the budget for the team",
	},
	"models": {
		Type:        schema.TypeSet,
		Required:    false,
		Optional:    true,
		Description: "A list of models associated with the team - all keys for this team_id will have at most, these models. If empty, assumes all models are allowed",
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"blocked": {
		Type:        schema.TypeBool,
		Required:    false,
		Optional:    true,
		Description: "Flag indicating if the team is blocked or not - will stop all calls from keys with this team_id.",
		Default:     false,
	},
}

func resourceTeam() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamCreate,
		ReadContext:   resourceTeamRead,
		UpdateContext: resourceTeamUpdate,
		DeleteContext: resourceTeamDelete,
		Schema:        ResourceTeamSchema,
	}
}

func getTeamFromResourceData(d *schema.ResourceData) litellm.Team {
	m := d.Get("metadata").(map[string]interface{})
	metadata := make(map[string]string)
	for k, v := range m {
		metadata[k] = fmt.Sprintf("%v", v)
	}

	mo := d.Get("models").(*schema.Set)
	var models []string
	for _, value := range mo.List() {
		models = append(models, fmt.Sprintf("%v", value))
	}

	team := litellm.Team{
		TeamAlias:      d.Get("team_alias").(string),
		TeamId:         d.Get("team_id").(string),
		Metadata:       metadata,
		TpmLimit:       d.Get("tpm_limit").(int),
		RpmLimit:       d.Get("rpm_limit").(int),
		MaxBudget:      d.Get("max_budget").(float64),
		BudgetDuration: d.Get("budget_duration").(string),
		Models:         models,
		Blocked:        d.Get("blocked").(bool),
	}

	return team
}

func resourceTeamCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	team := getTeamFromResourceData(d)
	jsonData, err := json.Marshal(team)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/team/new", client.ApiBaseURL)
	req, err := client.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API Request to create the team has failed with status code %d", resp.StatusCode)
	}

	d.SetId(team.TeamId)

	return diags
}

func resourceTeamRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	teamId := d.Get("team_id").(string)
	apiUrl := fmt.Sprintf("%s/team/info?team_id=%s", client.ApiBaseURL, teamId)

	req, err := client.NewRequest("GET", apiUrl, nil)
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	var teamInfo litellm.TeamInfoResponse
	// body := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&teamInfo)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(teamInfo.TeamId)
	d.Set("team_id", teamInfo.TeamId)
	d.Set("team_alias", teamInfo.TeamInfo.TeamAlias)
	d.Set("metadata", teamInfo.TeamInfo.Metadata)
	d.Set("tpm_limit", teamInfo.TeamInfo.TpmLimit)
	d.Set("rpm_limit", teamInfo.TeamInfo.RpmLimit)
	d.Set("blocked", teamInfo.TeamInfo.Blocked)
	d.Set("models", teamInfo.TeamInfo.Models)
	d.Set("budget_duration", teamInfo.TeamInfo.BudgetDuration)
	d.Set("max_budget", teamInfo.TeamInfo.MaxBudget)

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API Request to read the team has failed with status code %d", resp.StatusCode)
	}

	return diags
}

func resourceTeamUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	team := getTeamFromResourceData(d)

	apiUrlUpdateTeam := fmt.Sprintf("%s/team/update", client.ApiBaseURL)
	teamUpdateBody, err := json.Marshal(team)
	if err != nil {
		diag.FromErr(err)
	}
	req, err := client.NewRequest("POST", apiUrlUpdateTeam, bytes.NewBuffer(teamUpdateBody))
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API Request to update the team has failed with status code %d", resp.StatusCode)
	}

	return diags
}

func resourceTeamDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	teamId := d.Get("team_id").(string)
	teamsToDelete := []string{}
	teamsToDelete = append(teamsToDelete, teamId)
	jsonPayload := map[string]interface{}{
		"team_ids": teamsToDelete,
	}
	body, err := json.Marshal(jsonPayload)
	if err != nil {
		diag.FromErr(err)
	}

	apiUrl := fmt.Sprintf("%s/team/delete", client.ApiBaseURL)
	req, err := client.NewRequest("POST", apiUrl, bytes.NewBuffer(body))
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API Request to delete the team has failed with status code %d", resp.StatusCode)
	}

	d.SetId("")

	return diags
}
