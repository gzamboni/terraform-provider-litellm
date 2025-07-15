terraform {
  required_providers {
    litellm = {
      source  = "registry.terraform.io/gzamboni/litellm"
      version = "0.2.0"
    }
  }
}

provider "litellm" {
  api_base_url = "https://your-litellm-instance.com" # or omit to use LITELLM_API_BASE_URL

  jwt_token_endpoint = "https://my-identity-provider/token" # or omit to use LITELLM_JWT_TOKEN_ENDPOINT
  jwt_request_header = {
    "Content-Type": "application/json" # Only application/json and application/x-www-form-urlencoded are supported for now
  }
  jwt_request_payload = {
    client_id = "my-idp-client-id"
    client_secret = "my-idp-client-secret"
    scope = "my-litellm-admin-scope"
    grant_type = "client_credentials"
  }
}
