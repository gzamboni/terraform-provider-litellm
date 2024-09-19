# Terraform Provider for LiteLLM

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Terraform provider to manage models in a LiteLLM instance. This provider allows you to automate the creation, updating, and deletion of models in LiteLLM using Terraform.

## Table of Contents

- [Terraform Provider for LiteLLM](#terraform-provider-for-litellm)
  - [Table of Contents](#table-of-contents)
  - [Requirements](#requirements)
  - [Installation](#installation)
    - [Using Pre-built Binaries](#using-pre-built-binaries)
    - [Building from Source](#building-from-source)
  - [Usage](#usage)
    - [Provider Configuration](#provider-configuration)
    - [Resource: `litellm_model`](#resource-litellm_model)
      - [Example Usage](#example-usage)
      - [Argument Reference](#argument-reference)
      - [Attributes Reference](#attributes-reference)
    - [Importing Models](#importing-models)
  - [Building the Provider](#building-the-provider)
  - [Contributing](#contributing)
    - [Running Tests](#running-tests)
    - [Reporting Issues](#reporting-issues)
    - [Coding Standards](#coding-standards)
  - [License](#license)

## Requirements

- **Go** (version 1.16 or higher)
- **Terraform** (version 0.12 or higher)
- A LiteLLM instance with API access
- An API token (bearer token) for authentication with the LiteLLM API

## Installation

### Using Pre-built Binaries

Pre-built binaries are not yet available. For now, you need to build the provider from source.

### Building from Source

1. **Clone the repository:**

   ```bash
   git clone https://github.com/gzamboni/terraform-provider-litellm.git
   cd terraform-provider-litellm
   ```

2. **Build the provider:**

   ```bash
   go build -o terraform-provider-litellm
   ```

3. **Install the provider:**

   Create the following directory structure in your Terraform plugins directory:

   ```bash
   mkdir -p ~/.terraform.d/plugins/registry.terraform.io/gzamboni/litellm/0.1.0/YOUR_OS_ARCH
   ```

   Replace `YOUR_OS_ARCH` with your operating system and architecture, for example:

   - `linux_amd64`
   - `darwin_amd64` (macOS)
   - `windows_amd64`

   Move the binary to the appropriate directory:

   ```bash
   mv terraform-provider-litellm ~/.terraform.d/plugins/registry.terraform.io/gzamboni/litellm/0.1.0/YOUR_OS_ARCH/
   ```

## Usage

### Provider Configuration

Configure the LiteLLM provider with the required `api_token` and `api_base_url`. You can set these values directly in your Terraform configuration or use environment variables:

- `LITELLM_API_TOKEN`
- `LITELLM_API_BASE_URL`

```hcl
terraform {
  required_providers {
    litellm = {
      source  = "registry.terraform.io/gzamboni/litellm"
      version = "0.1.0"
    }
  }
}

provider "litellm" {
  api_token    = "your_api_token_here"       # or omit to use LITELLM_API_TOKEN
  api_base_url = "https://your-litellm-instance.com"  # or omit to use LITELLM_API_BASE_URL
}
```

### Resource: `litellm_model`

Manage models in your LiteLLM instance.

#### Example Usage

```hcl
resource "litellm_model" "example" {
  model_name = "example-model"

  litellm_params = {
    custom_llm_provider = "openai"
    model               = "gpt-3.5-turbo"
    api_key             = "your_underlying_model_api_key"
    api_base            = "https://api.openai.com/v1"
  }

  model_info = {
    id        = "unique-model-id"
    base_model = "gpt-3.5-turbo"
    tier       = "paid"
  }
}
```

#### Argument Reference

- `model_name` (Required, String): The name of the model to manage in LiteLLM.
- `litellm_params` (Required, Map of Strings): Parameters for the model as per LiteLLM API. This should include the underlying model details and any necessary credentials.
- `model_info` (Optional, Map of Strings): Additional model information, such as `id`, `base_model`, and `tier`.

#### Attributes Reference

- `id` (Computed): The ID of the model resource in Terraform. This is set to the value of `model_name`.

### Importing Models

If you have existing models in LiteLLM, you can import them into Terraform:

```bash
terraform import litellm_model.example example-model
```

Replace `example-model` with the `model_name` of your existing model.

## Building the Provider

If you want to contribute or modify the provider, follow these steps to build it from source:

1. **Clone the repository:**

   ```bash
   git clone https://github.com/gzamboni/terraform-provider-litellm.git
   cd terraform-provider-litellm
   ```

2. **Build the provider:**

   ```bash
   go build -o terraform-provider-litellm
   ```

3. **Install the provider:**

   Follow the installation instructions above to place the binary in the correct directory.

4. **Initialize Terraform in your working directory:**

   ```bash
   terraform init
   ```

## Contributing

Contributions are welcome! To collaborate on this project:

1. **Fork the repository** on GitHub.
2. **Create a new branch** for your feature or bug fix:

   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Commit your changes** with clear messages:

   ```bash
   git commit -am "Add new feature"
   ```

4. **Push to your fork:**

   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a pull request** on the main repository.

### Running Tests

To run unit tests:

```bash
go test -v ./...
```

To generate a coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Reporting Issues

If you encounter any issues or have suggestions, please open an issue on GitHub with detailed information.

### Coding Standards

- Follow Go's coding conventions.
- Ensure the code compiles without errors.
- Write clear and concise commit messages.
- Add documentation for any new features.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

**Disclaimer:** This provider is not officially supported by LiteLLM. Use it at your own risk.

---

Feel free to reach out if you have any questions or need assistance!
