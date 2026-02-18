package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFallback() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFallbackCreate,
		ReadContext:   resourceFallbackRead,
		UpdateContext: resourceFallbackUpdate,
		DeleteContext: resourceFallbackDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the fallback configuration (combination of model and fallback_type).",
			},
			"model": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The primary model name to configure fallbacks for.",
			},
			"fallback_models": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of fallback model names in priority order.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"fallback_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "general",
				ForceNew:    true,
				Description: "Type of fallback. Options: general (default), context_window, content_policy.",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					validTypes := []string{"general", "context_window", "content_policy"}
					isValid := false
					for _, t := range validTypes {
						if v == t {
							isValid = true
							break
						}
					}
					if !isValid {
						errs = append(errs, fmt.Errorf("%s must be one of: general, context_window, content_policy", key))
					}
					return
				},
			},
		},
	}
}

func resourceFallbackCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)
	var diags diag.Diagnostics

	model := d.Get("model").(string)
	fallbackType := d.Get("fallback_type").(string)
	fallbackModelsRaw := d.Get("fallback_models").([]interface{})

	fallbackModels := make([]string, len(fallbackModelsRaw))
	for i, v := range fallbackModelsRaw {
		fallbackModels[i] = v.(string)
	}

	requestBody := map[string]interface{}{
		"model":           model,
		"fallback_models": fallbackModels,
		"fallback_type":   fallbackType,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/fallback", client.ApiBaseURL)
	req, err := client.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	httpClient := &http.Client{
		Timeout: 30 * 1000 * 1000 * 1000, // 30 seconds
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respBody := buf.String()
		return diag.Errorf("Failed to create fallback configuration: status code %d, response: %s", resp.StatusCode, respBody)
	}

	// Set the ID as a combination of model and fallback_type
	d.SetId(fmt.Sprintf("%s:%s", model, fallbackType))

	return diags
}

func resourceFallbackRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)
	var diags diag.Diagnostics

	// Parse the ID to get model and fallback_type
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.Errorf("Invalid ID format: %s (expected model:fallback_type)", d.Id())
	}
	model := idParts[0]
	fallbackType := idParts[1]

	url := fmt.Sprintf("%s/fallback/%s?fallback_type=%s", client.ApiBaseURL, model, fallbackType)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiToken))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: 30 * 1000 * 1000 * 1000, // 30 seconds
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Resource doesn't exist anymore
		d.SetId("")
		return diags
	}

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respBody := buf.String()
		return diag.Errorf("Failed to read fallback configuration: status code %d, response: %s", resp.StatusCode, respBody)
	}

	var response struct {
		Model          string   `json:"model"`
		FallbackModels []string `json:"fallback_models"`
		FallbackType   string   `json:"fallback_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return diag.FromErr(err)
	}

	// Set the values in state
	if err := d.Set("model", response.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fallback_type", response.FallbackType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("fallback_models", response.FallbackModels); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceFallbackUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)
	var diags diag.Diagnostics

	model := d.Get("model").(string)
	fallbackType := d.Get("fallback_type").(string)
	fallbackModelsRaw := d.Get("fallback_models").([]interface{})

	fallbackModels := make([]string, len(fallbackModelsRaw))
	for i, v := range fallbackModelsRaw {
		fallbackModels[i] = v.(string)
	}

	requestBody := map[string]interface{}{
		"model":           model,
		"fallback_models": fallbackModels,
		"fallback_type":   fallbackType,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/fallback", client.ApiBaseURL)
	req, err := client.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	httpClient := &http.Client{
		Timeout: 30 * 1000 * 1000 * 1000, // 30 seconds
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respBody := buf.String()
		return diag.Errorf("Failed to update fallback configuration: status code %d, response: %s", resp.StatusCode, respBody)
	}

	return diags
}

func resourceFallbackDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)
	var diags diag.Diagnostics

	// Parse the ID to get model and fallback_type
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 {
		return diag.Errorf("Invalid ID format: %s (expected model:fallback_type)", d.Id())
	}
	model := idParts[0]
	fallbackType := idParts[1]

	url := fmt.Sprintf("%s/fallback/%s?fallback_type=%s", client.ApiBaseURL, model, fallbackType)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiToken))
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{
		Timeout: 30 * 1000 * 1000 * 1000, // 30 seconds
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		respBody := buf.String()
		return diag.Errorf("Failed to delete fallback configuration: status code %d, response: %s", resp.StatusCode, respBody)
	}

	d.SetId("")

	return diags
}
