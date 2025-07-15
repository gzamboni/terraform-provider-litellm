# Basic router settings configuration
resource "litellm_router_settings" "basic" {
  routing_strategy = "simple-shuffle"
  allowed_fails    = 3
  cooldown_time    = 5
  num_retries      = 4
  timeout          = 6000
  retry_after      = 0
}

# Advanced router settings with fallbacks
resource "litellm_router_settings" "advanced" {
  routing_strategy = "least-busy"
  allowed_fails    = 2
  cooldown_time    = 10
  num_retries      = 3
  timeout          = 8000
  retry_after      = 5

  routing_strategy_args = {
    "window_size" = "10"
    "ttl"         = "60"
  }

  # Model fallbacks for embedding models
  fallbacks {
    primary_model   = "text-embedding-ada-002"
    fallback_models = ["text-embedding-ada-002-azure"]
  }

  # Model fallbacks for chat models
  fallbacks {
    primary_model   = "gpt-4o"
    fallback_models = ["gpt-4o-azure", "gpt-4-turbo"]
  }

  # Context window fallbacks
  context_window_fallbacks {
    primary_model   = "gpt-4o"
    fallback_models = ["claude-3-sonnet"]
  }
}

# Production-ready configuration
resource "litellm_router_settings" "production" {
  routing_strategy = "latency-based-routing"
  allowed_fails    = 2
  cooldown_time    = 30
  num_retries      = 5
  timeout          = 10000
  retry_after      = 10

  routing_strategy_args = {
    "latency_window" = "100"
    "min_requests"   = "10"
  }

  # Multiple fallback chains for high availability
  fallbacks {
    primary_model   = "gpt-4o"
    fallback_models = ["gpt-4o-azure", "gpt-4-turbo", "claude-3-sonnet"]
  }

  fallbacks {
    primary_model   = "gpt-3.5-turbo"
    fallback_models = ["gpt-3.5-turbo-azure", "claude-3-haiku"]
  }

  fallbacks {
    primary_model   = "text-embedding-ada-002"
    fallback_models = ["text-embedding-ada-002-azure", "text-embedding-3-small"]
  }

  context_window_fallbacks {
    primary_model   = "gpt-4o"
    fallback_models = ["claude-3-sonnet", "claude-3-opus"]
  }

  context_window_fallbacks {
    primary_model   = "gpt-3.5-turbo"
    fallback_models = ["claude-3-haiku"]
  }
}
