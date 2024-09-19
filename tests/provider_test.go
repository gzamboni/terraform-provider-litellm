package tests

import (
	"testing"

	"github.com/gzamboni/terraform-provider-litellm/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	provider := provider.NewProvider()

	assert.NotNil(t, provider)
	assert.IsType(t, &schema.Provider{}, provider)
}

func TestProviderSchema(t *testing.T) {
	provider := provider.NewProvider()

	schema := provider.Schema

	assert.Contains(t, schema, "api_token")
	assert.Contains(t, schema, "api_base_url")
}
