# Import a fallback configuration using the model name and fallback type
# Format: terraform import litellm_fallback.<resource_name> <model>:<fallback_type>

# Import general fallback for gpt-3.5-turbo
terraform import litellm_fallback.general gpt-3.5-turbo:general

# Import context_window fallback for gpt-3.5-turbo
terraform import litellm_fallback.context_window gpt-3.5-turbo:context_window

# Import content_policy fallback for gpt-4
terraform import litellm_fallback.content_policy gpt-4:content_policy
