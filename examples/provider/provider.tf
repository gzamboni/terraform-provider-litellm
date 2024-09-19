terraform {
  required_providers {
    litellm = {
      source  = "registry.terraform.io/gzamboni/litellm"
      version = "0.1.0"
    }
  }
}

provider "litellm" {
  api_token    = "your_api_token_here"               # or omit to use LITELLM_API_TOKEN
  api_base_url = "https://your-litellm-instance.com" # or omit to use LITELLM_API_BASE_URL
}
