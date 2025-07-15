package jwtauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
)

const MockedTokenUrl = "http://localhost:4000/token"
const MockedTokenAttribute = "access_token"
const ExpectedClientId = "mock_client_id"
const ExpectedClientSecret = "mock_client_secret"
const ExpectedClientScope = "mock_scope"

func TestGetJwtAuthUrlFormEncoded(t *testing.T) {
	expectedFormContentType := "application/x-www-form-urlencoded"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", MockedTokenUrl,
		func(req *http.Request) (*http.Response, error) {
			contentType := string(req.Header.Get("Content-Type"))

			if contentType != expectedFormContentType {
				t.Errorf("Expected header Content-Type to be " + expectedFormContentType + " received : " + contentType)
			}

			err := req.ParseForm()
			if err != nil {
				t.Errorf("Can't parse form")
			}

			client_id := req.FormValue("client_id")
			client_secret := req.FormValue("client_secret")
			scope := req.FormValue("scope")

			if client_id != ExpectedClientId {
				t.Errorf("Received client id is " + client_id + " expected : " + ExpectedClientId)
			}
			if client_secret != ExpectedClientSecret {
				t.Errorf("Received client id is " + client_secret + " expected : " + ExpectedClientSecret)
			}
			if scope != ExpectedClientScope {
				t.Errorf("Received client id is " + scope + " expected : " + ExpectedClientScope)
			}

			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				MockedTokenAttribute: "mock_access_token",
			})
			return resp, err
		},
	)

	jwtAuthForm := &JwtAuth{
		TokenEndpoint:  MockedTokenUrl,
		TokenAttribute: MockedTokenAttribute,
		RequestHeader: map[string]string{
			"Content-Type": expectedFormContentType,
		},
		RequestPayload: map[string]string{
			"client_id":     ExpectedClientId,
			"client_secret": ExpectedClientSecret,
			"scope":         ExpectedClientScope,
			"grant_type":    "client_credentials",
		},
	}
	access_token, err := GetApiTokenFromJwt(jwtAuthForm)
	fmt.Println(access_token)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(access_token) <= 0 {
		t.Errorf(err.Error())
	}
}

func TestGetJwtAuthJson(t *testing.T) {
	expectedJsonContentType := "application/json"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", MockedTokenUrl,
		func(req *http.Request) (*http.Response, error) {
			contentType := string(req.Header.Get("Content-Type"))

			if contentType != expectedJsonContentType {
				t.Errorf("Expected header Content-Type to be " + expectedJsonContentType + " received : " + contentType)
			}

			var payload map[string]interface{}
			err := json.NewDecoder(req.Body).Decode(&payload)

			client_id := fmt.Sprintf("%v", payload["client_id"])
			client_secret := fmt.Sprintf("%v", payload["client_secret"])
			scope := fmt.Sprintf("%v", payload["scope"])

			if client_id != ExpectedClientId {
				t.Errorf("Received client id is " + client_id + " expected : " + ExpectedClientId)
			}
			if client_secret != ExpectedClientSecret {
				t.Errorf("Received client id is " + client_secret + " expected : " + ExpectedClientSecret)
			}
			if scope != ExpectedClientScope {
				t.Errorf("Received client id is " + scope + " expected : " + ExpectedClientScope)
			}

			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				MockedTokenAttribute: "mock_access_token",
			})
			return resp, err
		},
	)

	jwtAuthJson := &JwtAuth{
		TokenEndpoint:  MockedTokenUrl,
		TokenAttribute: MockedTokenAttribute,
		RequestHeader: map[string]string{
			"Content-Type": expectedJsonContentType,
		},
		RequestPayload: map[string]string{
			"client_id":     ExpectedClientId,
			"client_secret": ExpectedClientSecret,
			"scope":         ExpectedClientScope,
			"grant_type":    "client_credentials",
		},
	}
	access_token, err := GetApiTokenFromJwt(jwtAuthJson)
	fmt.Println(access_token)
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(access_token) <= 0 {
		t.Errorf(err.Error())
	}
}

func TestGetJwtAuthBadTokenAttribute(t *testing.T) {
	expectedFormContentType := "application/x-www-form-urlencoded"

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", MockedTokenUrl,
		func(req *http.Request) (*http.Response, error) {
			contentType := string(req.Header.Get("Content-Type"))

			if contentType != expectedFormContentType {
				t.Errorf("Expected header Content-Type to be " + expectedFormContentType + " received : " + contentType)
			}

			err := req.ParseForm()
			if err != nil {
				t.Errorf("Can't parse form")
			}

			client_id := req.FormValue("client_id")
			client_secret := req.FormValue("client_secret")
			scope := req.FormValue("scope")

			if client_id != ExpectedClientId {
				t.Errorf("Received client id is " + client_id + " expected : " + ExpectedClientId)
			}
			if client_secret != ExpectedClientSecret {
				t.Errorf("Received client id is " + client_secret + " expected : " + ExpectedClientSecret)
			}
			if scope != ExpectedClientScope {
				t.Errorf("Received client id is " + scope + " expected : " + ExpectedClientScope)
			}

			resp, err := httpmock.NewJsonResponse(200, map[string]interface{}{
				"bad_attribute_token": "mock_access_token",
			})
			return resp, err
		},
	)

	jwtAuthForm := &JwtAuth{
		TokenEndpoint:  MockedTokenUrl,
		TokenAttribute: MockedTokenAttribute,
		RequestHeader: map[string]string{
			"Content-Type": expectedFormContentType,
		},
		RequestPayload: map[string]string{
			"client_id":     ExpectedClientId,
			"client_secret": ExpectedClientSecret,
			"scope":         ExpectedClientScope,
			"grant_type":    "client_credentials",
		},
	}
	access_token, err := GetApiTokenFromJwt(jwtAuthForm)
	fmt.Println(access_token)
	if err == nil {
		t.Errorf("Bad Token attribute test should have returned an error")
	}
	if len(access_token) > 0 {
		t.Errorf("Bad token attribute test should have returned a 0 length access token")
	}
}
