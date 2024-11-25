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

var SchemaTeamMembership = map[string]*schema.Schema{
	"team_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Team ID",
		ForceNew:    true,
	},
	"user_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "User ID",
		ForceNew:    true,
	},
	"role": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Team role for the user in the team. Only roles `admin` and role `user` are valid in this field.",
		ForceNew:    true,
		ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
			value := val.(string)
			if value != "admin" && value != "user" {
				errs = append(errs, fmt.Errorf("only roles `admin` and `user` are authorized in this field."))
			}
			return warns, errs
		},
	},
	// Max budget in theam doesnt anything for now (2024-11-15) so its better not to be set. When getting a team user are returned through the member_with_roles attribute, that doesn't include any mention of a budget so it doesnt work
	// Users with a budget should be returned through the team_memberships but that's not the case
	// "max_budget_in_team": {
	// 	Type:        schema.TypeFloat,
	// 	Required:    false,
	// 	Optional:    true,
	// 	Description: "Allowed budget in the team",
	// },
}

func teamMembershipCalculatedResourceId(d *schema.ResourceData) string {
	teamId := d.Get("team_id").(string)
	userId := d.Get("user_id").(string)
	return fmt.Sprintf("%s_%s", teamId, userId)
}

func resourceTeamMembership() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTeamMembershipCreate,
		ReadContext:   resourceTeamMembershipRead,
		// UpdateContext: resourceTeamMembershipUpdate,
		DeleteContext: resourceTeamMembershipDelete,
		Schema:        SchemaTeamMembership,
		Description: `This resource allow your to Create/Delete/Import a team membership. 
		This resource doesn't support max_budget_in_team because it doesnt seems to work right now (2024-11-24). 
		
		Also take into account, the underlying LITELLM API used by this resource is in preview, bugs can happen.`,
	}
}

func resourceTeamMembershipCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	role := d.Get("role").(string)
	teamId := d.Get("team_id").(string)
	userId := d.Get("user_id").(string)

	apiUrl := fmt.Sprintf("%s/team/member_add", client.ApiBaseURL)

	addMemberData := litellm.TeamMembershipMemberAddPayload{
		Members: []litellm.TeamMembershipMemberRole{
			{
				UserId: userId,
				Role:   role,
			},
		},
		TeamId: teamId,
	}

	body, err := json.Marshal(addMemberData)

	if err != nil {
		diag.FromErr(err)
	}
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
		diag.Errorf("API request to create team membership has failed with status code %d", resp.StatusCode)
	}

	d.SetId(teamMembershipCalculatedResourceId(d))

	return diags
}

func resourceTeamMembershipRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	teamId := d.Get("team_id").(string)
	userId := d.Get("user_id").(string)

	apiUrl := fmt.Sprintf("%s/team/info?team_id=%s", client.ApiBaseURL, teamId)
	req, err := client.NewRequest("GET", apiUrl, nil)
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	var teamInfo litellm.TeamInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&teamInfo)
	if err != nil {
		diag.FromErr(err)
	}

	var myTeamAssociation litellm.TeamMembershipMemberRole
	for _, v := range teamInfo.TeamInfo.MembersWithRoles {
		if v.UserId == userId {
			myTeamAssociation = v
			break
		}
	}

	d.Set("team_id", teamId)
	d.Set("user_id", userId)
	d.Set("role", myTeamAssociation.Role)

	return diags
}

// func resourceTeamMembershipUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
// 	// client := m.(*LitellmClient)

// 	var diags diag.Diagnostics

// 	// Team membership update is to update the attribute budget in team, not to update team/user association
// 	// This part should be implemented when max_budget_in_team is working

// 	return diags
// }

func resourceTeamMembershipDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	userId := d.Get("user_id").(string)
	teamId := d.Get("team_id").(string)

	apiUrl := fmt.Sprintf("%s/team/member_delete", client.ApiBaseURL)

	jsonPayload, err := json.Marshal(map[string]interface{}{
		"user_id": userId,
		"team_id": teamId,
	})

	if err != nil {
		diag.FromErr(err)
	}
	req, err := client.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API request to create team membership has failed with status code %d", resp.StatusCode)
	}

	d.SetId("")

	return diags
}
