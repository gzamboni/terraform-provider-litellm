// provider.go
package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/gzamboni/terraform-provider-litellm/provider/jwtauth"
)

type AuthType string

const (
	JWT_AUTH AuthType = "jwt"
	API_AUTH AuthType = "api_token"
)

func getJwtAuth(d *schema.ResourceData) *jwtauth.JwtAuth {
	jwtTokenEndpoint := d.Get("jwt_token_endpoint").(string)
	jwtRequestPayload := d.Get("jwt_request_payload").(map[string]interface{})
	jwtRequestHeader := d.Get("jwt_request_header").(map[string]interface{})
	jwtTokenAttribute := d.Get("jwt_token_attribute").(string)

	//Convert map[string]interface{} to map[string]string
	headerConversion := make(map[string]string)
	payloadConversion := make(map[string]string)
	for k, v := range jwtRequestHeader {
		headerConversion[k] = fmt.Sprintf("%v", v)
	}
	for k, v := range jwtRequestPayload {
		payloadConversion[k] = fmt.Sprintf("%v", v)
	}

	jwtInfo := &jwtauth.JwtAuth{
		TokenEndpoint:  jwtTokenEndpoint,
		RequestHeader:  headerConversion,
		RequestPayload: payloadConversion,
		TokenAttribute: jwtTokenAttribute,
	}
	return jwtInfo
}

func hasSetCredentials(d *schema.ResourceData) (bool, AuthType) {
	apiToken := d.Get("api_token").(string)
	jwtInfo := getJwtAuth(d)

	apiTokenIsSet := jwtauth.IsApiTokenSet(apiToken)
	jwtIsSet := jwtauth.IsJwtSet(jwtInfo)

	if !apiTokenIsSet && !jwtIsSet {
		return false, "api_token or jwt attributes should be set"
	}
	if apiTokenIsSet && jwtIsSet {
		return false, "either api_token or jwt attributes should be set, but not both"
	}

	var authenticationMethod AuthType
	if apiTokenIsSet == true {
		authenticationMethod = API_AUTH
	}
	if jwtIsSet == true {
		authenticationMethod = JWT_AUTH
	}

	return true, authenticationMethod
}

func NewProvider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				Computed:    true,
				DefaultFunc: schema.EnvDefaultFunc("LITELLM_API_TOKEN", nil),
				Description: "The API token (bearer token) for accessing the LiteLLM API.",
			},
			"api_base_url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("LITELLM_API_BASE_URL", nil),
				Description: "The base URL for the LiteLLM API (e.g., 'https://your-litellm-instance.com').",
			},
			"jwt_token_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				DefaultFunc: schema.EnvDefaultFunc("LITELLM_JWT_TOKEN_ENDPOINT", nil),
				Description: "IdP Token endpoint. Use this parameter if authenticating with a jwt auth token. You should also set `jwt_request_header` and `jwt_request_payload`.",
			},
			"jwt_request_header": {
				Type:        schema.TypeMap,
				Optional:    true,
				Required:    false,
				Description: "IdP Headers to put in the request to get the token. You should also set `jwt_token_endpoint` and `jwt_request_payload`.",
			},
			"jwt_request_payload": {
				Type:        schema.TypeMap,
				Optional:    true,
				Required:    false,
				Description: "IdP Headers to put in the request to get the token. You should also set `jwt_token_endpoint` and `jwt_request_header`.",
			},
			"jwt_token_attribute": {
				Type:        schema.TypeString,
				Optional:    true,
				Required:    false,
				Description: "Describe in which attribute is the token in the HTTP Response from the IdP to get the token.",
				Default:     "access_token",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"litellm_model": resourceModel(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	isCredentialSet, authenticationMethod := hasSetCredentials(d)
	if isCredentialSet == false {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Could define authentication method",
			Detail:   fmt.Sprintf("%v", authenticationMethod),
		})
		return nil, diags
	}

	apiBaseURL := d.Get("api_base_url").(string)
	if apiBaseURL == "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "API base URL is missing",
			Detail:   "An API base URL is required to access the LiteLLM API.",
		})
		return nil, diags
	}

	client := &LitellmClient{}
	client.ApiBaseURL = apiBaseURL

	switch authenticationMethod {
	case API_AUTH:
		client.ApiToken = d.Get("api_token").(string)
	case JWT_AUTH:
		jwtInfo := getJwtAuth(d)
		token, err := jwtauth.GetApiTokenFromJwt(jwtInfo)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Could not get token from jwt token endpoint : %v", err.Error()),
				Detail:   err.Error(),
			})
			return nil, diags
		}
		client.ApiToken = token
		client.JwtTokenEndpoint = jwtInfo.TokenEndpoint
		client.JwtRequestHeader = jwtInfo.RequestHeader
		client.JwtRequestPayload = jwtInfo.RequestPayload
		client.JwtTokenAttribute = jwtInfo.TokenAttribute
	default:
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unknown authentication method",
			Detail:   fmt.Sprintf("%s is an unknown authentication method", authenticationMethod),
		})
		return nil, diags
	}

	return client, diags
}
