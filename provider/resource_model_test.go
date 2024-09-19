package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "test-model", resourceData.Id())

	// Test Update
	diags = p.ResourcesMap["litellm_model"].UpdateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())

	// Test Delete
	diags = p.ResourcesMap["litellm_model"].DeleteContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", resourceData.Id())
}
