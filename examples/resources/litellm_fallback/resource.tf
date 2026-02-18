# General fallback configuration
resource "litellm_fallback" "general" {
  model           = "gpt-3.5-turbo"
  fallback_models = ["gpt-4", "claude-3-haiku"]
  fallback_type   = "general"
}

# Context window fallback configuration
resource "litellm_fallback" "context_window" {
  model           = "gpt-3.5-turbo"
  fallback_models = ["gpt-4-32k", "claude-3-opus"]
  fallback_type   = "context_window"
}

# Content policy fallback configuration
resource "litellm_fallback" "content_policy" {
  model           = "gpt-4"
  fallback_models = ["claude-3-haiku"]
  fallback_type   = "content_policy"
}

# Multiple fallback configurations for different models
resource "litellm_fallback" "gpt4_fallback" {
  model           = "gpt-4o"
  fallback_models = ["gpt-4o-azure", "gpt-4-turbo", "claude-3-sonnet"]
  fallback_type   = "general"
}

resource "litellm_fallback" "embedding_fallback" {
  model           = "text-embedding-ada-002"
  fallback_models = ["text-embedding-ada-002-azure", "text-embedding-3-small"]
  fallback_type   = "general"
}
