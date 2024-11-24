package provider

import (
	"io"
	"net/http"
)

type LitellmClient struct {
	ApiToken   string `tfsdk:"api_token"`
	ApiBaseURL string `tfsdk:"api_base_url"`
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
