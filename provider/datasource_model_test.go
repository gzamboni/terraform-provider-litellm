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

func TestDatasourceModelRead(t *testing.T) {
	// Mock LiteLLM API server
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

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

	readResosurceData := schema.TestResourceDataRaw(t, p.DataSourcesMap["litellm_model"].Schema, map[string]interface{}{
		"model_info_id": "unique-model-id",
	})

	// Test Read
	diags = p.DataSourcesMap["litellm_model"].ReadContext(context.Background(), readResosurceData, meta)
	assert.False(t, diags.HasError())
}
