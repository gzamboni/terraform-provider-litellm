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
