package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceModelCreateUpdateDelete(t *testing.T) {
	// Mock LiteLLM API server
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Mock responses
	mux.HandleFunc("/model/new", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/model/update", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/model/delete", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	// Configure the provider with the mock server URL
	p := NewProvider()

	// Create provider configuration data
	providerConfig := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"api_token":    apiToken,
		"api_base_url": server.URL,
	})

	// Call ConfigureContextFunc and get the meta (client)
	meta, diags := p.ConfigureContextFunc(context.Background(), providerConfig)
	if diags.HasError() {
		t.Fatalf("Failed to configure provider: %s", diags[0].Summary)
	}

	resourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_model"].Schema, map[string]interface{}{
		"model_name": "test-model",
		"litellm_params": map[string]interface{}{
			"custom_llm_provider": "openai",
			"model":               "gpt-3.5-turbo",
			"api_key":             "underlying-api-key",
		},
		"model_info": map[string]interface{}{
			"id":         "unique-model-id",
			"base_model": "gpt-3.5-turbo",
			"tier":       "paid",
		},
	})

	// Test Create
	diags = p.ResourcesMap["litellm_model"].CreateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "unique-model-id", resourceData.Id())

	// Test Update
	diags = p.ResourcesMap["litellm_model"].UpdateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())

	// Test Delete
	diags = p.ResourcesMap["litellm_model"].DeleteContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", resourceData.Id())
}

// modelRequestBodies runs create and update for a litellm_model resource and
// returns the litellm_params each call sent to LiteLLM.
func modelRequestBodies(t *testing.T, resource map[string]interface{}) (created, updated map[string]interface{}, diags diag.Diagnostics) {
	t.Helper()

	sent := map[string]map[string]interface{}{}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()
	for _, path := range []string{"/model/new", "/model/update"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				LitellmParams map[string]interface{} `json:"litellm_params"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			sent[r.URL.Path] = body.LitellmParams
			w.Write([]byte(`{}`))
		})
	}

	p := NewProvider()
	meta, diags := p.ConfigureContextFunc(context.Background(), schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"api_token":    "test-token",
		"api_base_url": server.URL,
	}))
	require.False(t, diags.HasError())

	model := p.ResourcesMap["litellm_model"]
	data := schema.TestResourceDataRaw(t, model.Schema, resource)
	if diags = model.CreateContext(context.Background(), data, meta); diags.HasError() {
		return nil, nil, diags
	}
	diags = model.UpdateContext(context.Background(), data, meta)
	return sent["/model/new"], sent["/model/update"], diags
}

func TestResourceModelSendsStructuredParamsAsJSON(t *testing.T) {
	created, updated, diags := modelRequestBodies(t, map[string]interface{}{
		"model_name": "smart-router",
		"litellm_params": map[string]interface{}{
			"model":                           "auto_router/complexity_router",
			"complexity_router_default_model": "claude-sonnet",
			"extra_headers":                   `{"x-api-key":"secret"}`,
			// A JSON-shaped string that LiteLLM must still receive as a string.
			"vertex_credentials": `{"type":"service_account"}`,
		},
		"litellm_params_json": `{
			"complexity_router_config": {"tiers": {"SIMPLE": "claude-haiku", "REASONING": ["claude-opus", "claude-sonnet"]}, "session_affinity": true},
			"cache_control_injection_points": [{"location": "message", "role": "system"}],
			"rpm": 60
		}`,
		"model_info": map[string]interface{}{"id": "smart-router-id"},
	})
	require.False(t, diags.HasError())

	expected := map[string]interface{}{
		"model":                           "auto_router/complexity_router",
		"complexity_router_default_model": "claude-sonnet",
		"extra_headers":                   map[string]interface{}{"x-api-key": "secret"},
		"vertex_credentials":              `{"type":"service_account"}`,
		"complexity_router_config": map[string]interface{}{
			"tiers": map[string]interface{}{
				"SIMPLE":    "claude-haiku",
				"REASONING": []interface{}{"claude-opus", "claude-sonnet"},
			},
			"session_affinity": true,
		},
		"cache_control_injection_points": []interface{}{
			map[string]interface{}{"location": "message", "role": "system"},
		},
		"rpm": float64(60),
	}
	assert.Equal(t, expected, created)
	assert.Equal(t, expected, updated)
}

func TestResourceModelRejectsParamSetInBothAttributes(t *testing.T) {
	_, _, diags := modelRequestBodies(t, map[string]interface{}{
		"model_name":          "test-model",
		"litellm_params":      map[string]interface{}{"model": "gpt-4o", "rpm": "10"},
		"litellm_params_json": `{"rpm": 60}`,
		"model_info":          map[string]interface{}{"id": "test-model-id"},
	})
	require.True(t, diags.HasError())
	assert.Contains(t, diags[0].Summary, `"rpm" is set in both litellm_params and litellm_params_json`)
}

func TestResourceModelValidatesLitellmParamsJSON(t *testing.T) {
	validate := resourceModel().Schema["litellm_params_json"].ValidateDiagFunc
	path := cty.GetAttrPath("litellm_params_json")

	assert.False(t, validate(`{"tags": ["a"]}`, path).HasError())
	for _, invalid := range []string{`["a"]`, `"a"`, `{"tags": `} {
		assert.True(t, validate(invalid, path).HasError(), invalid)
	}
}
