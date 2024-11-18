package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestResourceModelCreateUpdateDeleteRead(t *testing.T) {
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

	mux.HandleFunc("/model/info", func(w http.ResponseWriter, r *http.Request) {
		readOutput := map[string]interface{}{
			"data": map[string]interface{}{"model_name": "test-model",
				"model_info": map[string]string{
					"id":         "unique-model-id",
					"base_model": "gpt-3.5-turbo",
					"tier":       "paid",
				},
				"litellm_params": map[string]string{
					"custom_llm_provider": "openai",
					"model":               "gpt-3.5-turbo",
					"api_key":             "underlying-api-key",
				},
			},
		}

		jsonOutput, err := json.Marshal(readOutput)
		if err != nil {
			panic(err)
		}

		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)
		assert.Equal(t, expectedToken, token)

		litellm_model_id := r.URL.Query().Get("litellm_model_id")
		assert.Equal(t, "unique-model-id", litellm_model_id)

		w.WriteHeader(http.StatusOK)
		w.Write(jsonOutput)
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
		"model_name":                         "test-model",
		"model_info_id":                      "unique-model-id",
		"model_info_base_model":              "gpt-3.5-turbo",
		"model_info_tier":                    "paid",
		"litellm_params_custom_llm_provider": "openai",
		"litellm_params_model":               "gpt-3.5-turbo",
		"litellm_params_api_key":             "underlying-api-key",
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

	// Test Read
	diags = p.ResourcesMap["litellm_model"].ReadContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
}
