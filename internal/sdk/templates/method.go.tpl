{{ define "Method" }}
package ncloudsdk

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type {{.MethodName}}Request struct {
    {{.Request}}
}

func (n *NClient) {{.MethodName}}(r *{{.MethodName}}Request) (map[string]interface{}, error) {
	query := map[string]string{}
	initBody := map[string]string{}

 	{{.Query}}

	{{.Body}}

	rawBody, err := json.Marshal(initBody)
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

	snake_case_response := convertKeys(response).(map[string]interface{})

	return snake_case_response, nil
}

func (n *NClient) {{.MethodName}}_TF(r *{{.MethodName}}Request) (*{{.MethodName}}Response, error) {
	t, err := n.{{.MethodName}}(r)
	if err != nil {
		return nil, err
	}

	res, err := ConvertToFrameworkTypes_{{.MethodName}}(t)
	if err != nil {
		return nil, err
	}

	return res, nil
}

{{ end }}