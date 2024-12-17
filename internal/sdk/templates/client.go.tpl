{{ define "client" }}

package sdk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type NClient struct {
	BaseURL    string
	HTTPClient *http.Client
	ACCESS_KEY string
	SECRET_KEY string
}

func NewClient(baseURL, accessKey, secretKey string) *NClient {
	return &NClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
		ACCESS_KEY: accessKey,
		SECRET_KEY: secretKey,
	}
}

{{ end }}