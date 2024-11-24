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

func TestResourceUserCreateUpdateDelete(t *testing.T) {
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	valueRead := map[string]interface{}{
		"user_id": "mock-user-read",
		"user_info": map[string]interface{}{
			"user_id":          "mock-user-read",
			"user_alias":       "mock-user",
			"user_email":       "user_email",
			"send_invite_mail": false,
			"metadata": map[string]interface{}{
				"test-metadata": "mock-value",
			},
			"tpm_limit":       20,
			"rpm_limit":       10,
			"max_budget":      250.25,
			"budget_duration": "30d",
			"models":          []interface{}{"azure/gpt-4o"},
		},
	}

	mux.HandleFunc("/user/info", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/user/update", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/user/delete", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/user/new", func(w http.ResponseWriter, r *http.Request) {
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

	resourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_user"].Schema, map[string]interface{}{
		"user_id":    "mock-user",
		"user_alias": "mock-user",
		"user_email": "mock-user@mock.com",
		"user_role":  "proxy_admin",
		"metadata": map[string]interface{}{
			"test-metadata": "mock-value",
		},
		"tpm_limit":       20,
		"rpm_limit":       10,
		"max_budget":      250.25,
		"budget_duration": "30d",
		"models":          []interface{}{"azure/gpt-4o"},
	})

	updateResourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_user"].Schema, map[string]interface{}{
		"user_id":    "mock-user",
		"user_alias": "mock-user-updated",
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

	readResourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_user"].Schema, valueRead)

	// Test Create
	diags = p.ResourcesMap["litellm_user"].CreateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "mock-user", resourceData.Get("user_id").(string))

	// Test Update
	diags = p.ResourcesMap["litellm_user"].UpdateContext(context.Background(), updateResourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "mock-user-updated", updateResourceData.Get("user_alias").(string))
	assert.Equal(t, "mock-value", updateResourceData.Get("metadata").(map[string]interface{})["test-metadata"].(string))

	// Test Delete
	diags = p.ResourcesMap["litellm_user"].DeleteContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", resourceData.Id())

	// Test Read
	diags = p.ResourcesMap["litellm_user"].ReadContext(context.Background(), readResourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "mock-user-read", readResourceData.Get("user_id").(string))
	assert.Equal(t, "mock-user-read", readResourceData.Id())
}
