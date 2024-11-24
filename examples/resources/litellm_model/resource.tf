resource "litellm_model" "example" {
  model_name = "example-model"

  litellm_params_custom_llm_provider = "openai"
  litellm_params_model               = "gpt-3.5-turbo"
  litellm_params_api_key             = "your_underlying_model_api_key"
  litellm_params_api_base            = "https://api.openai.com/v1"

  model_info_id         = "unique-model-id"
  model_info_base_model = "gpt-3.5-turbo"
  model_info_tier       = "paid"
}
