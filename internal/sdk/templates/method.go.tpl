{{ define "Method" }}
/* =================================================================================
 * NCLOUD SDK LAYER FOR TERRAFORM CODEGEN - DO NOT EDIT
 * =================================================================================
 * Refresh Template
 * Required data are as follows
 *
 *		MethodName         string
 *		RequestParameters  string
 *		Query              string
 *		Body               string
 *		Path               string
 *		Method             string
 * ================================================================================= */

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
    {{.RequestParameters}}
}

func (n *NClient) {{.MethodName}}(ctx context.Context, r *{{.MethodName}}Request) (map[string]interface{}, error) {
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

	response, err := n.MakeRequestWithContext(ctx, "{{.Method}}", url, body, query)
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("output is nil")
	}

	snake_case_response := convertKeys(response).(map[string]interface{})

	return snake_case_response, nil
}

func (n *NClient) {{.MethodName}}_TF(ctx context.Context, r *{{.MethodName}}Request) (*{{.MethodName}}Response, error) {
	t, err := n.{{.MethodName}}(ctx, r)
	if err != nil {
		return nil, err
	}

	res, err := ConvertToFrameworkTypes_{{.MethodName}}(context.TODO(), t)
	if err != nil {
		return nil, err
	}

	return res, nil
}

{{ end }}
