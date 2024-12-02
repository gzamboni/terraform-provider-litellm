package provider

import (
	"io"
	"net/http"
)

type LitellmClient struct {
	ApiToken          string            `tfsdk:"api_token"`
	ApiBaseURL        string            `tfsdk:"api_base_url"`
	JwtTokenEndpoint  string            `tfsdk:"jwt_token_endpoint"`
	JwtRequestHeader  map[string]string `tfsdk:"jwt_request_header"`
	JwtRequestPayload map[string]string `tfsdk:"jwt_request_payload"`
	JwtTokenAttribute string            `tfsdk:"jwt_token_attribute"`
}

func (c *LitellmClient) NewRequest(method string, url string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", "Bearer "+c.ApiToken)
	return request, nil
}
