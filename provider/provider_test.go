package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestProvider(t *testing.T) {
	litellm_provider := NewProvider()

	assert.NotNil(t, litellm_provider)
	assert.IsType(t, &schema.Provider{}, litellm_provider)
}

func TestProviderSchema(t *testing.T) {
	litellm_provider := NewProvider()

	schema := litellm_provider.Schema

	assert.Contains(t, schema, "api_token")
	assert.Contains(t, schema, "api_base_url")
}
