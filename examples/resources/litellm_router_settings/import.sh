#!/bin/bash

# Import existing router settings configuration
# Since router settings is a singleton resource, it uses a fixed ID
terraform import litellm_router_settings.example router_settings