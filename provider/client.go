package provider

type LitellmClient struct {
	ApiToken   string `tfsdk:"api_token"`
	ApiBaseURL string `tfsdk:"api_base_url"`
}
