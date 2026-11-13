package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

// modelAPI is a LiteLLM stand-in that records the litellm_params of each
// /model/new and /model/update request.
type modelAPI struct {
	meta interface{}

	mu   sync.Mutex
	sent map[string]map[string]interface{}
}

func newModelAPI(t *testing.T) *modelAPI {
	t.Helper()

	api := &modelAPI{sent: map[string]map[string]interface{}{}}
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	for _, path := range []string{"/model/new", "/model/update"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			var body struct {
				LitellmParams map[string]interface{} `json:"litellm_params"`
			}
			// The handler runs outside the test goroutine, so it must not FailNow.
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decoding %s request: %v", r.URL.Path, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			api.mu.Lock()
			api.sent[r.URL.Path] = body.LitellmParams
			api.mu.Unlock()
			w.Write([]byte(`{}`))
		})
	}

	p := NewProvider()
	meta, diags := p.ConfigureContextFunc(context.Background(), schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"api_token":    "test-token",
		"api_base_url": server.URL,
	}))
	require.False(t, diags.HasError())
	api.meta = meta
	return api
}

func (api *modelAPI) params(path string) map[string]interface{} {
	api.mu.Lock()
	defer api.mu.Unlock()
	return api.sent[path]
}

func TestResourceModelSendsStructuredParamsAsJSON(t *testing.T) {
	api := newModelAPI(t)
	model := resourceModel()
	data := schema.TestResourceDataRaw(t, model.Schema, map[string]interface{}{
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
	require.False(t, model.CreateContext(context.Background(), data, api.meta).HasError())
	require.False(t, model.UpdateContext(context.Background(), data, api.meta).HasError())

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
	assert.Equal(t, expected, api.params("/model/new"))
	assert.Equal(t, expected, api.params("/model/update"))
}

// Apply repeats the plan-time duplicate check, for values unknown until then.
func TestResourceModelRejectsParamSetInBothAttributes(t *testing.T) {
	const duplicate = `"rpm" is set in both litellm_params and litellm_params_json`
	conflicting := map[string]interface{}{
		"model_name":          "test-model",
		"litellm_params":      map[string]interface{}{"model": "gpt-4o", "rpm": "10"},
		"litellm_params_json": `{"rpm": 60}`,
		"model_info":          map[string]interface{}{"id": "test-model-id"},
	}
	model := resourceModel()

	t.Run("create", func(t *testing.T) {
		api := newModelAPI(t)
		diags := model.CreateContext(context.Background(), schema.TestResourceDataRaw(t, model.Schema, conflicting), api.meta)
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, duplicate)
		assert.Nil(t, api.params("/model/new"))
	})

	t.Run("update", func(t *testing.T) {
		api := newModelAPI(t)
		data := schema.TestResourceDataRaw(t, model.Schema, conflicting)
		data.SetId("test-model-id")
		diags := model.UpdateContext(context.Background(), data, api.meta)
		require.True(t, diags.HasError())
		assert.Contains(t, diags[0].Summary, duplicate)
		assert.Nil(t, api.params("/model/update"))
	})
}

func TestResourceModelValidatesLitellmParamsJSON(t *testing.T) {
	validate := resourceModel().Schema["litellm_params_json"].ValidateDiagFunc
	path := cty.GetAttrPath("litellm_params_json")

	assert.False(t, validate(`{"tags": ["a"]}`, path).HasError())
	for _, invalid := range []string{`null`, `["a"]`, `"a"`, `{"tags": `} {
		assert.True(t, validate(invalid, path).HasError(), invalid)
	}
}

// planModel plans a litellm_model whose state holds priorParams and priorJSON
// against a configuration holding params and paramsJSON. An empty priorParams
// plans a create.
func planModel(t *testing.T, priorParams map[string]string, priorJSON string, params map[string]interface{}, paramsJSON string) (*terraform.InstanceDiff, error) {
	t.Helper()

	var state *terraform.InstanceState
	if len(priorParams) > 0 {
		attributes := map[string]string{
			"id":               "test-model-id",
			"model_name":       "test-model",
			"model_info.%":     "1",
			"model_info.id":    "test-model-id",
			"litellm_params.%": fmt.Sprint(len(priorParams)),
		}
		for key, value := range priorParams {
			attributes["litellm_params."+key] = value
		}
		if priorJSON != "" {
			attributes["litellm_params_json"] = priorJSON
		}
		state = &terraform.InstanceState{ID: "test-model-id", Attributes: attributes}
	}

	config := map[string]interface{}{
		"model_name":     "test-model",
		"litellm_params": params,
		"model_info":     map[string]interface{}{"id": "test-model-id"},
	}
	if paramsJSON != "" {
		config["litellm_params_json"] = paramsJSON
	}
	return resourceModel().Diff(context.Background(), state, terraform.NewResourceConfigRaw(config), nil)
}

func TestResourceModelPlan(t *testing.T) {
	const duplicate = `"rpm" is set in both litellm_params and litellm_params_json`
	stored := map[string]string{"model": "gpt-4o"}
	params := map[string]interface{}{"model": "gpt-4o"}
	const storedJSON = `{"rpm":60,"tags":["a","b"]}`

	t.Run("rejects a duplicate key on create", func(t *testing.T) {
		_, err := planModel(t, nil, "", map[string]interface{}{"model": "gpt-4o", "rpm": "10"}, `{"rpm": 60}`)
		assert.ErrorContains(t, err, duplicate)
	})

	t.Run("rejects a duplicate key on update", func(t *testing.T) {
		_, err := planModel(t, stored, storedJSON, map[string]interface{}{"model": "gpt-4o", "rpm": "10"}, storedJSON)
		assert.ErrorContains(t, err, duplicate)
	})

	t.Run("ignores formatting and key order", func(t *testing.T) {
		diff, err := planModel(t, stored, storedJSON, params, "{\n  \"tags\": [\"a\", \"b\"],\n  \"rpm\": 60\n}\n")
		require.NoError(t, err)
		assert.True(t, diff == nil || diff.Empty(), "%v", diff)
	})

	t.Run("keeps array order significant", func(t *testing.T) {
		diff, err := planModel(t, stored, storedJSON, params, `{"rpm":60,"tags":["b","a"]}`)
		require.NoError(t, err)
		require.NotNil(t, diff)
		assert.Contains(t, diff.Attributes, "litellm_params_json")
		assert.False(t, diff.RequiresNew())
	})

	t.Run("updates in place when a value changes or a key is added", func(t *testing.T) {
		diff, err := planModel(t, stored, storedJSON, params, `{"rpm":90,"tags":["a","b"],"tpm":1000}`)
		require.NoError(t, err)
		require.NotNil(t, diff)
		assert.False(t, diff.RequiresNew())
	})

	t.Run("replaces the model when a key is removed", func(t *testing.T) {
		diff, err := planModel(t, stored, storedJSON, params, `{"rpm":60}`)
		require.NoError(t, err)
		require.NotNil(t, diff)
		assert.True(t, diff.RequiresNew())
	})

	t.Run("replaces the model when the attribute is unset", func(t *testing.T) {
		diff, err := planModel(t, stored, storedJSON, params, "")
		require.NoError(t, err)
		require.NotNil(t, diff)
		assert.True(t, diff.RequiresNew())
	})

	t.Run("updates in place when a key moves to litellm_params", func(t *testing.T) {
		diff, err := planModel(t, stored, storedJSON, map[string]interface{}{"model": "gpt-4o", "rpm": "60"}, `{"tags":["a","b"]}`)
		require.NoError(t, err)
		require.NotNil(t, diff)
		assert.False(t, diff.RequiresNew())
	})
}
