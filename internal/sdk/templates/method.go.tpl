{{ define "Method" }}
package ncloudsdk

import (
	"encoding/json"
	"fmt"
	"strings"
)

type {{.MethodName}}Request struct {
    {{.Request}}
}

func (n *NClient) {{.MethodName}}(r *{{.MethodName}}Request) (map[string]interface{}, error) {
	query := map[string]string{
        {{.Query}}
	}

	rawBody, err := json.Marshal(map[string]string{
        {{.Body}}
    })
	if err != nil {
		return nil, err
	}

	body := strings.Replace(string(rawBody), `\"`, "", -1)

	url := n.BaseURL {{.Path}}

	response, err := n.MakeRequest("{{.Method}}", url, body, query)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("output is nil")
	}

	return response, nil
}

{{ end }}