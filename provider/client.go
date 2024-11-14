package provider

type LitellmClient struct {
	ApiToken          string            `tfsdk:"api_token"`
	ApiBaseURL        string            `tfsdk:"api_base_url"`
	JwtTokenEndpoint  string            `tfsdk:"jwt_token_endpoint"`
	JwtRequestHeader  map[string]string `tfsdk:"jwt_request_header"`
	JwtRequestPayload map[string]string `tfsdk:"jwt_request_payload"`
	JwtTokenAttribute string            `tfsdk:"jwt_token_attribute"`
}
