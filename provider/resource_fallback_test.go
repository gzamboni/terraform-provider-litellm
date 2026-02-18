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

func TestResourceFallbackCreateUpdateDelete(t *testing.T) {
	// Mock LiteLLM API server
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Mock responses
	mux.HandleFunc("/fallback", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"model": "gpt-3.5-turbo",
			"fallback_models": ["gpt-4", "claude-3-haiku"],
			"fallback_type": "general",
			"message": "Fallback configuration created successfully"
		}`))
	})

	mux.HandleFunc("/fallback/gpt-3.5-turbo", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"model": "gpt-3.5-turbo",
				"fallback_models": ["gpt-4", "claude-3-haiku"],
				"fallback_type": "general"
			}`))
		} else if r.Method == "DELETE" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"model": "gpt-3.5-turbo",
				"fallback_type": "general",
				"message": "Fallback configuration deleted successfully"
			}`))
		}
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

	resourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_fallback"].Schema, map[string]interface{}{
		"model":           "gpt-3.5-turbo",
		"fallback_models": []interface{}{"gpt-4", "claude-3-haiku"},
		"fallback_type":   "general",
	})

	// Test Create
	diags = p.ResourcesMap["litellm_fallback"].CreateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "gpt-3.5-turbo:general", resourceData.Id())

	// Test Read
	diags = p.ResourcesMap["litellm_fallback"].ReadContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "gpt-3.5-turbo", resourceData.Get("model"))
	assert.Equal(t, "general", resourceData.Get("fallback_type"))

	// Test Update
	diags = p.ResourcesMap["litellm_fallback"].UpdateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())

	// Test Delete
	diags = p.ResourcesMap["litellm_fallback"].DeleteContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", resourceData.Id())
}

func TestResourceFallbackContextWindow(t *testing.T) {
	// Mock LiteLLM API server
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	// Mock responses
	mux.HandleFunc("/fallback", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"model": "gpt-3.5-turbo",
			"fallback_models": ["gpt-4-32k", "claude-3-opus"],
			"fallback_type": "context_window",
			"message": "Fallback configuration created successfully"
		}`))
	})

	mux.HandleFunc("/fallback/gpt-3.5-turbo", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		fallbackType := r.URL.Query().Get("fallback_type")
		if fallbackType == "context_window" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"model": "gpt-3.5-turbo",
				"fallback_models": ["gpt-4-32k", "claude-3-opus"],
				"fallback_type": "context_window"
			}`))
		}
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

	resourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_fallback"].Schema, map[string]interface{}{
		"model":           "gpt-3.5-turbo",
		"fallback_models": []interface{}{"gpt-4-32k", "claude-3-opus"},
		"fallback_type":   "context_window",
	})

	// Test Create
	diags = p.ResourcesMap["litellm_fallback"].CreateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "gpt-3.5-turbo:context_window", resourceData.Id())

	// Test Read
	diags = p.ResourcesMap["litellm_fallback"].ReadContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "gpt-3.5-turbo", resourceData.Get("model"))
	assert.Equal(t, "context_window", resourceData.Get("fallback_type"))
}
