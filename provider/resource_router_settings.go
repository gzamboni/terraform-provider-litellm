package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceRouterSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRouterSettingsCreate,
		ReadContext:   resourceRouterSettingsRead,
		UpdateContext: resourceRouterSettingsUpdate,
		DeleteContext: resourceRouterSettingsDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the router settings resource.",
			},
			"routing_strategy": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "simple-shuffle",
				Description: "The routing strategy to use for load balancing. Options: simple-shuffle, least-busy, usage-based-routing, latency-based-routing.",
			},
			"routing_strategy_args": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional arguments for the routing strategy.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"allowed_fails": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     3,
				Description: "Number of allowed failures before marking a model as unhealthy.",
			},
			"cooldown_time": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     5,
				Description: "Cooldown time in seconds before retrying a failed model.",
			},
			"num_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     4,
				Description: "Number of retries for failed requests.",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     6000,
				Description: "Request timeout in milliseconds.",
			},
			"retry_after": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Time to wait before retrying in seconds.",
			},
			"fallbacks": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Model fallback configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"primary_model": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The primary model name.",
						},
						"fallback_models": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "List of fallback models.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"context_window_fallbacks": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Context window fallback configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"primary_model": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The primary model name.",
						},
						"fallback_models": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "List of fallback models.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func resourceRouterSettingsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceRouterSettingsUpdate(ctx, d, m)
}

func resourceRouterSettingsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// For router settings, we'll assume the resource always exists since it's a configuration
	// In a real implementation, you might want to fetch the current config from the API
	return diags
}

func resourceRouterSettingsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*LitellmClient)
	var diags diag.Diagnostics

	// Build the router settings payload
	routerSettings := map[string]interface{}{
		"routing_strategy":      d.Get("routing_strategy").(string),
		"routing_strategy_args": d.Get("routing_strategy_args").(map[string]interface{}),
		"allowed_fails":         d.Get("allowed_fails").(int),
		"cooldown_time":         d.Get("cooldown_time").(int),
		"num_retries":           d.Get("num_retries").(int),
		"timeout":               d.Get("timeout").(int),
		"retry_after":           d.Get("retry_after").(int),
	}

	// Handle fallbacks
	if fallbacksRaw, ok := d.GetOk("fallbacks"); ok {
		fallbacksList := fallbacksRaw.([]interface{})
		fallbacks := make([]map[string]interface{}, 0)

		for _, fallbackRaw := range fallbacksList {
			fallbackMap := fallbackRaw.(map[string]interface{})
			primaryModel := fallbackMap["primary_model"].(string)
			fallbackModelsRaw := fallbackMap["fallback_models"].([]interface{})

			fallbackModels := make([]string, len(fallbackModelsRaw))
			for i, model := range fallbackModelsRaw {
				fallbackModels[i] = model.(string)
			}

			fallbackConfig := map[string]interface{}{
				primaryModel: fallbackModels,
			}
			fallbacks = append(fallbacks, fallbackConfig)
		}
		routerSettings["fallbacks"] = fallbacks
	} else {
		routerSettings["fallbacks"] = nil
	}

	// Handle context window fallbacks
	if contextFallbacksRaw, ok := d.GetOk("context_window_fallbacks"); ok {
		contextFallbacksList := contextFallbacksRaw.([]interface{})
		contextFallbacks := make([]map[string]interface{}, 0)

		for _, fallbackRaw := range contextFallbacksList {
			fallbackMap := fallbackRaw.(map[string]interface{})
			primaryModel := fallbackMap["primary_model"].(string)
			fallbackModelsRaw := fallbackMap["fallback_models"].([]interface{})

			fallbackModels := make([]string, len(fallbackModelsRaw))
			for i, model := range fallbackModelsRaw {
				fallbackModels[i] = model.(string)
			}

			fallbackConfig := map[string]interface{}{
				primaryModel: fallbackModels,
			}
			contextFallbacks = append(contextFallbacks, fallbackConfig)
		}
		routerSettings["context_window_fallbacks"] = contextFallbacks
	} else {
		routerSettings["context_window_fallbacks"] = nil
	}

	// Create the full payload
	requestBody := map[string]interface{}{
		"router_settings": routerSettings,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return diag.FromErr(err)
	}

	url := fmt.Sprintf("%s/config/update", client.ApiBaseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return diag.FromErr(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.ApiToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return diag.Errorf("API request failed with status code %d", resp.StatusCode)
	}

	// Set the ID for the resource (using a fixed ID since this is a singleton configuration)
	d.SetId("router_settings")

	return diags
}

func resourceRouterSettingsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// For router settings, we don't actually delete the configuration
	// Instead, we could reset to default values or just remove the resource from state
	// For now, we'll just remove it from state
	d.SetId("")

	return diags
}