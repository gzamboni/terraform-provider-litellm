package litellm

type ModelInfo struct {
	Id        string `json:"id,omitempty"`
	DbModel   bool   `json:"db_model"`
	Tier      string `json:"tier,omitempty"`
	BaseModel string `json:"base_model,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	CreatedBy string `json:"create_by,omitempty"`
}

type LitellmParams struct {
	Model                            string   `json:"model"`
	CustomLlmProvider                string   `json:"custom_llm_provider,omitempty"`
	Tpm                              int      `json:"tpm,omitempty"`
	Rpm                              int      `json:"rpm,omitempty"`
	ApiKey                           string   `json:"api_key,omitempty"`
	ApiBase                          string   `json:"api_base,omitempty"`
	ApiVersion                       string   `json:"api_version,omitempty"`
	Timeout                          float64  `json:"timeout,omitempty"`
	StreamTimeout                    float64  `json:"stream_timeout,omitempty"`
	MaxRetries                       int      `json:"max_retries,omitempty"`
	Organization                     string   `json:"organization,omitempty"`
	ConfigurableClientsideAuthParams []string `json:"configurable_clientside_auth_params,omitempty"`
	RegionName                       string   `json:"region_name,omitempty"`
	VertexProject                    string   `json:"vertex_project,omitempty"`
	VertexLocation                   string   `json:"vertex_location,omitempty"`
	VertexCredentials                string   `json:"vertex_credentials,omitempty"`
	AwsAccessKeyId                   string   `json:"aws_access_key_id,omitempty"`
	AwsSecretKeyId                   string   `json:"aws_secret_key_id,omitempty"`
	AwsRegionName                    string   `json:"aws_region_name,omitempty"`
	WatsonxRegionName                string   `json:"watsonx_region_name,omitempty"`
	InputCostPerToken                float64  `json:"input_cost_per_token,omitempty"`
	OutputCostPerToken               float64  `json:"output_cost_per_token,omitempty"`
	InputCostPerSecond               float64  `json:"input_cost_per_second,omitempty"`
	OutputCostPerSecond              float64  `json:"output_cost_per_second,omitempty"`
	MaxFileSizeMB                    float64  `json:"max_file_size_mb,omitempty"`
}

type Model struct {
	ModelName     string        `json:"model_name"`
	LitellmParams LitellmParams `json:"litellm_params"`
	ModelInfo     ModelInfo     `json:"model_info"`
}

type ModelInfoWoModelId struct {
	Data []Model `json:"data"`
}

type ModelInfoWModelId struct {
	Data Model `json:"data"`
}
