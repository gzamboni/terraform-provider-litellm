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

func TestResourceTeamMembershipCreateUpdateDelete(t *testing.T) {
	apiToken := "test-token"

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mockUserId := "mock-user"
	mockTeamId := "mock-team"
	mockRole := "internal_user"

	valueRead := map[string]interface{}{
		"team_id": mockTeamId,
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
			"members_with_roles": []map[string]interface{}{
				{
					"role":    "internal_user",
					"user_id": "mock-user",
				},
			},
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

	mux.HandleFunc("/team/member_delete", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		expectedToken := fmt.Sprintf("Bearer %s", apiToken)

		assert.Equal(t, expectedToken, token)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	mux.HandleFunc("/team/member_add", func(w http.ResponseWriter, r *http.Request) {
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

	resourceData := schema.TestResourceDataRaw(t, p.ResourcesMap["litellm_team_membership"].Schema, map[string]interface{}{
		"user_id": mockUserId,
		"role":    mockRole,
		"team_id": mockTeamId,
	})

	// Test Create
	diags = p.ResourcesMap["litellm_team_membership"].CreateContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, teamMembershipCalculatedResourceId(resourceData), resourceData.Id())

	// Test Delete
	diags = p.ResourcesMap["litellm_team_membership"].DeleteContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, "", resourceData.Id())

	// Test Read
	diags = p.ResourcesMap["litellm_team_membership"].ReadContext(context.Background(), resourceData, meta)
	assert.False(t, diags.HasError())
	assert.Equal(t, mockUserId, resourceData.Get("user_id").(string))
	assert.Equal(t, mockTeamId, resourceData.Get("team_id").(string))
	assert.Equal(t, mockRole, resourceData.Get("role").(string))
}
