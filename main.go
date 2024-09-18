package main

import (
	"github.com/gzamboni/terraform-provider-litellm/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.NewProvider,
	})
}
