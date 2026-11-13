resource "litellm_model" "example" {
  model_name = "example-model"

  litellm_params = {
    custom_llm_provider = "openai"
    model               = "gpt-3.5-turbo"
    api_key             = "your_underlying_model_api_key"
    api_base            = "https://api.openai.com/v1"
  }

  model_info = {
    id         = "unique-model-id"
    base_model = "gpt-3.5-turbo"
    tier       = "paid"
  }
}

# Parameters LiteLLM expects as objects or arrays go in `litellm_params_json`.
resource "litellm_model" "auto_router" {
  model_name = "smart-router"

  litellm_params = {
    model                           = "auto_router/complexity_router"
    complexity_router_default_model = "example-model"
  }

  litellm_params_json = jsonencode({
    complexity_router_config = {
      tiers = {
        SIMPLE    = "example-model"
        MEDIUM    = "example-model"
        COMPLEX   = "example-model"
        REASONING = "example-model"
      }
    }
  })

  model_info = {
    id = "smart-router-id"
  }
}
