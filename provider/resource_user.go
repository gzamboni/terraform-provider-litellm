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

func UserSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"user_id": {
			Type:        schema.TypeString,
			Required:    false,
			Optional:    true,
			Description: "Specify a user id. If not set, a unique id will be generated.",
		},
		"user_alias": {
			Type:        schema.TypeString,
			Required:    false,
			Optional:    true,
			Description: "A descriptive name for you to know who this user id refers to.",
		},
		"user_email": {
			Type:        schema.TypeString,
			Required:    false,
			Optional:    true,
			Description: "Specify a user email.",
		},
		"send_invite_mail": {
			Type:        schema.TypeBool,
			Optional:    true,
			Required:    false,
			Description: "Specify if an invite email should be sent.",
		},
		"user_role": {
			Type:        schema.TypeString,
			Optional:    true,
			Required:    false,
			Description: "Specify a user role - \"proxy_admin\", \"proxy_admin_viewer\", \"internal_user\", \"internal_user_viewer\", \"team\", \"customer\".",
		},
		"max_budget": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Required:    false,
			Description: "Specify max budget for a given user.",
		},
		"budget_duration": {
			Type:        schema.TypeString,
			Optional:    true,
			Required:    false,
			Description: "Budget is reset at the end of specified duration. If not set, budget is never reset. You can set duration as seconds (\"30s\"), minutes (\"30m\"), hours (\"30h\"), days (\"30d\").",
		},
		"models": {
			Type:        schema.TypeSet,
			Optional:    true,
			Required:    false,
			Description: "Model_name's a user is allowed to call. (if empty, key is allowed to call all models)",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"tpm_limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Required:    false,
			Description: "Specify tpm limit for a given user (Tokens per minute)",
		},
		"rpm_limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Required:    false,
			Description: "Specify rpm limit for a given user (Requests per minute)",
		},
		"auto_create_key": {
			Type:        schema.TypeBool,
			Optional:    true,
			Required:    false,
			Default:     true,
			Description: "Flag used for returning a key as part of the /user/new response",
		},
		"metadata": {
			Type:        schema.TypeMap,
			Optional:    true,
			Required:    false,
			Description: "Metadata to add to the user",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func resourceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Schema:        UserSchema(),
		Description:   `The resource user allow you to create/delete/update/import a user in your litellm instance.`,
	}
}

func fromResourceDataGetUser(d *schema.ResourceData) litellm.User {
	rawUserRole := d.Get("user_role").(string)
	schemaModelsSet := d.Get("models").(*schema.Set)

	// Convert schema set string to []litellm.Role
	role, isValidated := litellm.ValidateRole(rawUserRole)
	if !isValidated {
		panic(rawUserRole + " is not a valid role")
	}

	// Convert schema set string to []string
	models := make([]string, schemaModelsSet.Len())
	for i, v := range schemaModelsSet.List() {
		models[i] = fmt.Sprint(v)
	}

	// Convert schema map string to map[string]string
	schemaMetadata := d.Get("metadata").(map[string]interface{})
	metadata := make(map[string]string)
	for k, v := range schemaMetadata {
		metadata[k] = fmt.Sprint(v)
	}

	user := litellm.User{
		UserId:         d.Get("user_id").(string),
		UserAlias:      d.Get("user_alias").(string),
		UserEmail:      d.Get("user_email").(string),
		UserRole:       role,
		MaxBudget:      d.Get("max_budget").(float64),
		BudgetDuration: d.Get("budget_duration").(string),
		Models:         models,
		TpmLimit:       d.Get("tpm_limit").(int),
		RpmLimit:       d.Get("rpm_limit").(int),
		AutoCreateKey:  d.Get("auto_create_key").(bool),
		Metadata:       metadata,
	}

	return user
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*LitellmClient)

	user := fromResourceDataGetUser(d)

	jsonPayload, err := json.Marshal(user)
	if err != nil {
		diag.FromErr(err)
	}

	apiUrl := fmt.Sprintf("%s/user/new", client.ApiBaseURL)
	req, err := client.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API Request to create a user has returned with status code %d", resp.StatusCode)
	}

	d.SetId(user.UserId)

	return diags
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*LitellmClient)

	userIds := map[string]interface{}{
		"user_ids": []string{d.Get("user_id").(string)},
	}

	apiUrl := fmt.Sprintf("%s/user/delete", client.ApiBaseURL)

	jsonPayload, err := json.Marshal(userIds)
	if err != nil {
		diag.FromErr(err)
	}

	req, err := client.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(jsonPayload))
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API Request to delete users has returned with status code %d", resp.StatusCode)
	}

	d.SetId("")

	return diags
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*LitellmClient)

	userId := d.Get("user_id").(string)

	apiUrl := fmt.Sprintf("%s/user/info?user_id=%s", client.ApiBaseURL, userId)

	req, err := client.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API Request to read user info has returned with status code %d", resp.StatusCode)
	}

	var jsonBody litellm.UserInfo
	err = json.NewDecoder(resp.Body).Decode(&jsonBody)
	if err != nil {
		diag.FromErr(err)
	}

	d.SetId(jsonBody.UserInfo.UserId)
	d.Set("user_id", jsonBody.UserInfo.UserId)
	d.Set("user_email", jsonBody.UserInfo.UserAlias)
	d.Set("user_alias", jsonBody.UserInfo.UserAlias)
	d.Set("send_invite_mail", jsonBody.UserInfo.SendInviteEmail)
	d.Set("user_role", jsonBody.UserInfo.UserRole)
	d.Set("max_budget", jsonBody.UserInfo.MaxBudget)
	d.Set("tpm_limit", jsonBody.UserInfo.TpmLimit)
	d.Set("rpm_limit", jsonBody.UserInfo.RpmLimit)
	d.Set("models", jsonBody.UserInfo.Models)
	d.Set("auto_create_key", jsonBody.UserInfo.AutoCreateKey)

	return diags
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := m.(*LitellmClient)

	user := fromResourceDataGetUser(d)
	// Field not used in /user/update
	user.UserAlias = ""

	apiUrl := fmt.Sprintf("%s/user/update", client.ApiBaseURL)
	jsonData, err := json.Marshal(user)
	if err != nil {
		diag.FromErr(err)
	}

	req, err := client.NewRequest(http.MethodPost, apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		diag.FromErr(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		diag.Errorf("API Request to update user has returned with status code %d", resp.StatusCode)
	}

	return diags
}
