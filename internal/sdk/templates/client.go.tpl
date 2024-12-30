{{ define "Client" }}
package ncloudsdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
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

// MakeRequest() - Streamlined core logic of abstracted api call
//
// Manufacture main request call
func (n *NClient) MakeRequest(method, endpoint, reqBody string, query map[string]string) (map[string]interface{}, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	// Set default request with query parsing
	req, err := n.SetRequest(url, query, reqBody, strings.ToUpper(method))
	if err != nil {
		return nil, err
	}

	// Make signature & set headers
	n.SetHeader(req, url, strings.ToUpper(method))

	// Execute api call
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check if delete succeeded
	if method == "DELETE" && resp.StatusCode == 204 {
		return map[string]interface{}{}, nil
	}

	// Parse response into map[string]interface{}
	var respBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, err
	}

	return respBody, nil
}

func (n *NClient) SetRequest(url *url.URL, queryParams map[string]string, reqBody, method string) (*http.Request, error) {
	q := url.Query()
	for key, value := range queryParams {
		q.Add(key, value)
	}

	url.RawQuery = q.Encode()

	b := bytes.NewBuffer([]byte(reqBody))

	req, err := http.NewRequest(method, url.String(), b)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (n *NClient) SetHeader(req *http.Request, url *url.URL, method string) {

	// Check if query string exists.
	// If then, do not even add "?".
	queryString := ""
	if len(url.RawQuery) > 0 {
		queryString = "?" + url.RawQuery
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	signature := makeSignature(method, url.Path+queryString, timestamp, n.ACCESS_KEY, n.SECRET_KEY)

	headers := map[string]string{
		"Content-Type":             "application/json",
		"x-ncp-apigw-timestamp":    timestamp,
		"x-ncp-iam-access-key":     n.ACCESS_KEY,
		"x-ncp-apigw-signature-v2": signature,
		"cache-control":            "no-cache",
		"pragma":                   "no-cache",
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}
}

// For curl request
func makeSignature(method, url, timestamp, accessKey, secretKey string) string {
	message := fmt.Sprintf("%s %s\n%s\n%s",
		method,
		url,
		timestamp,
		accessKey,
	)

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(message))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}


func diagOff[V, T interface{}](input func(ctx context.Context, elementType T, elements any) (V, diag.Diagnostics), ctx context.Context, elementType T, elements any) V {
	var emptyReturn V

	v, diags := input(ctx, elementType, elements)

	if diags.HasError() {
		diags.AddError("REFRESHING ERROR", "invalid diagOff operation")
		return emptyReturn
	}

	return v
}

// convertKeys recursively converts all keys in a map from camelCase to snake_case
func convertKeys(input interface{}) interface{} {
	switch v := input.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			// Convert the key to snake_case
			newKey := camelToSnake(key)
			// Recursively convert nested values
			newMap[newKey] = convertKeys(value)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, value := range v {
			newSlice[i] = convertKeys(value)
		}
		return newSlice
	default:
		return v
	}
}

func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func clearDoubleQuote(s string) string {
	return strings.Replace(strings.Replace(strings.Replace(s, "\\", "", -1), "\"", "", -1), `"`, "", -1)
}
{{ end }}