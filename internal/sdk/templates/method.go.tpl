{{ define "Method" }}
/* =================================================================================
 * NCLOUD SDK LAYER FOR TERRAFORM CODEGEN - DO NOT EDIT
 * =================================================================================
 * Refresh Template
 * Required data are as follows
 *
 *		MethodName             string
 *		RequestQueryParameters string
 *		RequestBodyParameters  string
 *		FunctionName           string
 *		Query                  string
 *		Body                   string
 *		Path                   string
 *		Method                 string
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

{{.RequestQueryParameters}}

{{.RequestBodyParameters}}

{{.FunctionName}}
	query := map[string]string{}

 	{{.Query}}

    {{.Body}}

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

{{ end }}
