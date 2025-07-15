package jwtauth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type JwtAuth struct {
	TokenEndpoint  string
	RequestHeader  map[string]string
	RequestPayload map[string]string
	TokenAttribute string
}

func IsApiTokenSet(apiToken string) bool {
	return len(apiToken) > 0
}

func IsJwtSet(jwtAuth *JwtAuth) bool {
	return len(jwtAuth.TokenEndpoint) > 0
}

func BasicAuth(username string, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func GetApiTokenFromJwt(jwtAuth *JwtAuth) (string, error) {
	// jsonBody := map[string]string{"grant_type": "client_credentials", "scope": jwtCredentials.JwtScope, "client_id": jwtCredentials.JwtClientId, "client_secret": jwtCredentials.JwtClientSecret}
	contentType := jwtAuth.RequestHeader["Content-Type"]
	if contentType == "" {
		contentType = "application/x-www-form-urlencoded"
	}

	var requestPayload io.Reader
	if contentType == "application/json" {
		stringPayload, err := json.Marshal(jwtAuth.RequestPayload)
		if err != nil {
			return "", err
		}
		requestPayload = strings.NewReader(string(stringPayload))
	} else {
		data := url.Values{}
		for k, v := range jwtAuth.RequestPayload {
			data.Set(k, v)
		}
		requestPayload = strings.NewReader(data.Encode())
	}
	r, err := http.NewRequest(http.MethodPost, jwtAuth.TokenEndpoint, requestPayload)
	if err != nil {
		return "", err
	}

	r.Header.Add("Content-Type", contentType)

	response, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", errors.New("Jwt auth get access token status code is not 200, returned value " + strconv.Itoa(response.StatusCode))
	}
	rawBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	jsonResponsePayload := make(map[string]interface{})
	if err := json.Unmarshal(rawBody, &jsonResponsePayload); err != nil {
		return "", err
	}
	token := fmt.Sprintf("%v", jsonResponsePayload[jwtAuth.TokenAttribute])
	if token == "<nil>" {
		return "", errors.New("Token attribute in jwt auth get access token is null")
	}
	return token, nil
}
