package main

import (
	"github.com/gzamboni/terraform-provider-litellm/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// provider documentation generation
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name adguard

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.NewProvider,
		ProviderAddr: "registry.terraform.io/gzamboni/litellm",
	})
}
