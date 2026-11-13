package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceModel() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceModelCreate,
		ReadContext:   resourceModelRead,
		UpdateContext: resourceModelUpdate,
		DeleteContext: resourceModelDelete,
		CustomizeDiff: resourceModelCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the model.",
			},
			"model_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the model to be managed.",
			},
			"litellm_params": {
				Type:        schema.TypeMap,
				Required:    true,
				Description: "Parameters for the model as per LiteLLM API.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"model_info": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional model information.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"litellm_params_json": {
				Type:             schema.TypeString,
				Optional:         true,
				Description:      "JSON object of LiteLLM parameters whose values are objects, arrays, numbers or booleans, such as `complexity_router_config` or `cache_control_injection_points`. Merged into `litellm_params`; a key may be set in only one of the two. Removing a key replaces the model, because LiteLLM keeps the parameters an update omits.",
				ValidateDiagFunc: validation.ToDiagFunc(validateJSONObject),
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
		},
	}
}

// parseJSONObject decodes a JSON object, keeping every value as written.
func parseJSONObject(raw string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		return nil, fmt.Errorf("must be a JSON object: %w", err)
	}
	// The literal null decodes without error and leaves the map nil.
	if object == nil {
		return nil, fmt.Errorf("must be a JSON object, got null")
	}
	return object, nil
}

func validateJSONObject(value interface{}, key string) ([]string, []error) {
	if _, err := parseJSONObject(value.(string)); err != nil {
		return nil, []error{fmt.Errorf("%s %w", key, err)}
	}
	return nil, nil
}

// structuredParams decodes litellm_params_json. Unset holds no parameters.
func structuredParams(raw string) (map[string]json.RawMessage, error) {
	if raw == "" {
		return nil, nil
	}
	return parseJSONObject(raw)
}

func duplicateParam(params map[string]interface{}, structured map[string]json.RawMessage) error {
	for key := range structured {
		if _, duplicate := params[key]; duplicate {
			return fmt.Errorf("%q is set in both litellm_params and litellm_params_json", key)
		}
	}
	return nil
}

// resourceModelCustomizeDiff reports at plan time what an apply cannot do.
//
// A key set in both attributes is an error. A key dropped from
// litellm_params_json replaces the model: LiteLLM's /model/update keeps every
// stored parameter the request omits, and treats an explicit null the same
// way, so an update cannot remove one.
func resourceModelCustomizeDiff(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	// Values unknown until apply cannot be compared; create and update repeat
	// the duplicate check once they are known.
	if !d.NewValueKnown("litellm_params") || !d.NewValueKnown("litellm_params_json") {
		return nil
	}

	previousJSON, plannedJSON := d.GetChange("litellm_params_json")
	params := d.Get("litellm_params").(map[string]interface{})
	planned, err := structuredParams(plannedJSON.(string))
	if err != nil {
		return fmt.Errorf("litellm_params_json %w", err)
	}
	if err := duplicateParam(params, planned); err != nil {
		return err
	}

	if d.Id() == "" {
		return nil
	}
	previous, err := structuredParams(previousJSON.(string))
	if err != nil {
		return fmt.Errorf("litellm_params_json in state %w", err)
	}
	for key := range previous {
		_, kept := planned[key]
		_, moved := params[key]
		if !kept && !moved {
			return d.ForceNew("litellm_params_json")
		}
	}
	return nil
}

// modelLitellmParams builds the litellm_params object sent to LiteLLM.
//
// params is sent as strings, except extra_headers, which is a JSON-encoded
// object of strings. paramsJSON carries the values LiteLLM does not accept as
// strings (objects, arrays) and is sent verbatim. Decoding is opt-in per value
// rather than applied to every string that happens to hold JSON: LiteLLM
// encrypts string params at rest and stores structured ones as they are, so a
// JSON-shaped secret such as vertex_credentials must stay a string.
func modelLitellmParams(params map[string]interface{}, paramsJSON string) (map[string]interface{}, error) {
	// A new map: params belongs to the resource data and must keep its strings.
	merged := make(map[string]interface{}, len(params))
	for key, value := range params {
		merged[key] = value
	}

	if raw, ok := params["extra_headers"]; ok {
		var extraHeaders map[string]string
		if err := json.Unmarshal([]byte(raw.(string)), &extraHeaders); err != nil {
			return nil, fmt.Errorf("litellm_params.extra_headers must be a JSON object of strings: %w", err)
		}
		merged["extra_headers"] = extraHeaders
	}

	structured, err := structuredParams(paramsJSON)
	if err != nil {
		return nil, fmt.Errorf("litellm_params_json %w", err)
	}
	if err := duplicateParam(params, structured); err != nil {
		return nil, err
	}
	for key, value := range structured {
		merged[key] = value
	}
	return merged, nil
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	modelName := d.Get("model_name").(string)
	litellmParams := d.Get("litellm_params").(map[string]interface{})
	modelInfo, _ := d.Get("model_info").(map[string]interface{})

	if modelInfo == nil || modelInfo["id"] == nil {
		return diag.Errorf("model_info.id is required")
	}

	litellmParams, err := modelLitellmParams(litellmParams, d.Get("litellm_params_json").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	requestBody := map[string]interface{}{
		"model_name":     modelName,
		"litellm_params": litellmParams,
		"model_info":     modelInfo,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/model/new", client.ApiBaseURL)
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
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	// Set the ID of the resource
	d.SetId(modelInfo["id"].(string))

	return diags
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)
	var diags diag.Diagnostics

	modelInfo := d.Get("model_info").(map[string]interface{})
	if modelInfo == nil || modelInfo["id"] == nil {
		return diag.Errorf("model_info.id is required")
	}

	url := fmt.Sprintf("%s/model/info?litellm_model_id=%s", client.ApiBaseURL, modelInfo["id"].(string))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	modelName := d.Get("model_name").(string)
	litellmParams := d.Get("litellm_params").(map[string]interface{})
	modelInfo, _ := d.Get("model_info").(map[string]interface{})

	if modelInfo == nil || modelInfo["id"] == nil {
		return diag.Errorf("model_info.id is required")
	}

	litellmParams, err := modelLitellmParams(litellmParams, d.Get("litellm_params_json").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	requestBody := map[string]interface{}{
		"model_name":     modelName,
		"litellm_params": litellmParams,
		"model_info":     modelInfo,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/model/update", client.ApiBaseURL)
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
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	return diags
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)

	var diags diag.Diagnostics

	modelInfo := d.Get("model_info").(map[string]interface{})
	if modelInfo == nil || modelInfo["id"] == nil {
		return diag.Errorf("model_info.id is required")
	}

	requestBody := map[string]interface{}{
		"id": modelInfo["id"].(string),
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/model/delete", client.ApiBaseURL)
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
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	d.SetId("")

	return diags
}
