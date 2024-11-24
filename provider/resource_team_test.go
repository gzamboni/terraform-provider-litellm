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

func TestResourceTeamCreateUpdateDelete(t *testing.T) {
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	valueRead := map[string]interface{}{
		"team_id": "mock-team-read",
		"team_info": map[string]interface{}{
			"team_alias": "mock-team",
			"metadata": map[string]interface{}{
				"test-metadata": "mock-value",
			},
			"tpm_limit":       20,
			"rpm_limit":       10,
			"max_budget":      250.25,
			"budget_duration": "30d",
			"models":          []interface{}{"azure/gpt-4o"},
			"blocked":         false,
		},
	}

	mux.HandleFunc("/team/info", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		jsonOutput, err := json.Marshal(valueRead)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonOutput)
	})

	mux.HandleFunc("/team/update", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/team/delete", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/team/new", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	p := NewProvider()
	providerConfig := schema.TestResourceDataRaw(t, p.Schema, map[string]interface{}{
		"api_token":    apiToken,
		"api_base_url": server.URL,
	})

	meta, diags := p.ConfigureContextFunc(context.Background(), providerConfig)
	if diags.HasError() {
		t.Fatalf("Failed to configure provider: %s", diags[0].Summary)
	}

	resourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_team"].Schema, map[string]interface{}{
		"team_id":    "mock-team",
		"team_alias": "mock-team",
		"metadata": map[string]interface{}{
			"test-metadata": "mock-value",
		},
		"tpm_limit":       20,
		"rpm_limit":       10,
		"max_budget":      250.25,
		"budget_duration": "30d",
		"models":          []interface{}{"azure/gpt-4o"},
		"blocked":         false,
	})

	updateResourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_team"].Schema, map[string]interface{}{
		"team_id":    "mock-team",
		"team_alias": "mock-team-updated",
		"metadata": map[string]interface{}{
			"test-metadata": "mock-value",
		},
		"tpm_limit":       20,
		"rpm_limit":       10,
		"max_budget":      250.25,
		"budget_duration": "30d",
		"models":          []interface{}{"azure/gpt-4o"},
		"blocked":         false,
	})

	readResourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_team"].Schema, valueRead)

	// Test Create
	diags = p.ResourcesMap["litellm_team"].CreateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "mock-team", resourceData.Get("team_id").(string))

	// Test Update
	diags = p.ResourcesMap["litellm_team"].UpdateContext(context.Background(), updateResourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "mock-team-updated", updateResourceData.Get("team_alias").(string))
	assert.Equal(t, "mock-value", updateResourceData.Get("metadata").(map[string]interface{})["test-metadata"].(string))

	// Test Delete
	diags = p.ResourcesMap["litellm_team"].DeleteContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", resourceData.Id())

	// Test Read
	diags = p.ResourcesMap["litellm_team"].ReadContext(context.Background(), readResourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "mock-team-read", readResourceData.Get("team_id").(string))
	assert.Equal(t, "mock-team-read", readResourceData.Id())
}
