package sdk

import (
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type Template struct {
	OAS                                *v3high.Operation
	funcMap                            template.FuncMap
	methodName                         string
	method                             string
	model                              string
	path                               string
	requestQueryParameters             string
	requestBodyParameters              string
	functionName                       string
	refreshLogic                       string
	possibleTypes                      string
	conditionalObjectFieldsWithNull    string
	convertValueWithNullInEmptyArrCase string
	query                              string
	body                               string
}

func New(oas *v3high.Operation, method, path string, refreshDetails *ResponseDetails) *Template {

	t := &Template{
		OAS:    oas,
		method: method,
	}

	funcMap := CreateFuncMap()

	t.methodName = t.method + getMethodName(path)
	t.model = refreshDetails.Model
	t.refreshLogic = refreshDetails.RefreshLogic
	t.path = getPath(path)

	requestQueryParameters, initQuery := getQueryParameters(oas.Parameters, t.methodName)
	requestBodyParameters, initBody := getBodyParameters(oas.RequestBody, t.methodName)
	t.requestQueryParameters = requestQueryParameters
	t.requestBodyParameters = requestBodyParameters
	t.query = initQuery
	t.body = initBody

	t.functionName = getFunctionName(t.methodName, requestQueryParameters, requestBodyParameters)

	t.funcMap = funcMap
	t.possibleTypes = refreshDetails.PossibleTypes
	t.conditionalObjectFieldsWithNull = refreshDetails.ConvertValueWithNull
	t.convertValueWithNullInEmptyArrCase = refreshDetails.ConvertValueWithNullInEmptyArrCase

	return t
}

func WriteClient() []byte {
	var b bytes.Buffer

	clientTemplate, err := template.New("").Parse(ClientTemplate)
	if err != nil {
		log.Fatalf("error occurred with baseTemplate at rendering create: %v", err)
	}

	err = clientTemplate.ExecuteTemplate(&b, "Client", nil)
	if err != nil {
		log.Fatalf("error occurred with Generating Method: %v", err)
	}

	return b.Bytes()
}

func (t *Template) WriteRefresh() []byte {
	var b bytes.Buffer

	refreshTemplate, err := template.New("").Funcs(t.funcMap).Parse(RefreshTemplate)
	if err != nil {
		log.Fatalf("error occurred with baseTemplate at rendering create: %v", err)
	}

	data := struct {
		MethodName                         string
		Model                              string
		RefreshLogic                       string
		PossibleTypes                      string
		ConditionalObjectFieldsWithNull    string
		ConvertValueWithNullInEmptyArrCase string
	}{
		MethodName:                         t.methodName,
		Model:                              t.model,
		RefreshLogic:                       t.refreshLogic,
		PossibleTypes:                      t.possibleTypes,
		ConditionalObjectFieldsWithNull:    t.conditionalObjectFieldsWithNull,
		ConvertValueWithNullInEmptyArrCase: t.convertValueWithNullInEmptyArrCase,
	}

	err = refreshTemplate.ExecuteTemplate(&b, "Refresh", data)
	if err != nil {
		log.Fatalf("error occurred with Generating Refresh: %v", err)
	}

	return b.Bytes()
}

func (t *Template) WriteTemplate() []byte {
	var b bytes.Buffer

	methodTemplate, err := template.New("").Funcs(t.funcMap).Parse(MethodTemplate)
	if err != nil {
		log.Fatalf("error occurred with baseTemplate at rendering create: %v", err)
	}

	data := struct {
		MethodName             string
		RequestQueryParameters string
		RequestBodyParameters  string
		FunctionName           string
		Query                  string
		Body                   string
		Path                   string
		Method                 string
	}{
		MethodName:             t.methodName,
		Method:                 t.method,
		RequestQueryParameters: t.requestQueryParameters,
		RequestBodyParameters:  t.requestBodyParameters,
		FunctionName:           t.functionName,
		Query:                  t.query,
		Body:                   t.body,
		Path:                   t.path,
	}

	err = methodTemplate.ExecuteTemplate(&b, "Method", data)
	if err != nil {
		log.Fatalf("error occurred with Generating Method: %v", err)
	}

	return b.Bytes()
}

func getMethodName(s string) string {
	parts := strings.Split(s, "/")
	var result []string

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Remove curly braces if present
		part = strings.TrimPrefix(part, "{")
		part = strings.TrimSuffix(part, "}")

		// Remove hyphens and convert to uppercase
		part = strings.ReplaceAll(part, "-", "")
		part = FirstAlphabetToUpperCase(part)

		result = append(result, part)
	}

	return strings.Join(result, "")
}

func getPath(path string) string {
	parts := strings.Split(path, "/")
	s := ``

	for _, val := range parts {

		if len(val) < 1 {
			continue
		}

		s = s + `+"/"+`

		start := strings.Index(val, "{")

		// if val doesn't wrapped with curly brace
		if start == -1 {
			s = s + fmt.Sprintf(`"%s"`, val)
		} else {
			s = s + fmt.Sprintf(`ClearDoubleQuote(*q.%s)`, PathToPascal(val))
		}
	}

	return s
}

func getQueryParameters(params []*v3high.Parameter, methodName string) (string, string) {
	var requestParameters strings.Builder
	var initQuery strings.Builder

	if params == nil {
		return "", ""
	}

	requestParameters.WriteString(fmt.Sprintf("type %sRequestQuery struct {", methodName) + "\n")

	for _, params := range params {
		key := params.Name

		// In Default, all parameters needs to be in request struct
		switch params.Schema.Schema().Type[0] {
		case "string":
			requestParameters.WriteString(fmt.Sprintf("%[1]s *string `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")

		case "boolean":
			requestParameters.WriteString(fmt.Sprintf("%[1]s *bool `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")

		case "integer":
			if params.Schema.Schema().Format == "int64" {
				requestParameters.WriteString(fmt.Sprintf("%[1]s *int64 `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")
			} else if params.Schema.Schema().Format == "int32" {
				requestParameters.WriteString(fmt.Sprintf("%[1]s *int32 `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")
			}

		case "number":
			requestParameters.WriteString(fmt.Sprintf("%[1]s *float64 `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")

		case "array":
			requestParameters.WriteString(fmt.Sprintf("%[1]s []*string `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")

		case "object":
			requestParameters.WriteString(fmt.Sprintf("%[1]s []*string `json:\"%[2]s,omitempty\"`", PathToPascal(key), key) + "\n")
		}

		// In case of query parameters
		if params.In == "query" {
			if params.Required == nil {
				// optional query parameters
				initQuery.WriteString(fmt.Sprintf(`
				if q.%[1]s!= nil {
					query["%[2]s"] = *q.%[1]s
				}`, FirstAlphabetToUpperCase(key), key) + "\n")
			} else {
				// required query parameters
				initQuery.WriteString(fmt.Sprintf(`
				query["%[1]s"] = *q.%[2]s`, key, FirstAlphabetToUpperCase(key)) + "\n")
			}
		}
	}

	requestParameters.WriteString(fmt.Sprintf("}") + "\n")

	return requestParameters.String(), initQuery.String()
}

func getBodyParameters(body *v3high.RequestBody, methodName string) (string, string) {
	var requestParameters strings.Builder
	var initBody strings.Builder

	// return if requestBody does not needed.
	if body == nil {
		return "", ""
	}

	content, ok := body.Content.OrderedMap.Get("application/json;charset=UTF-8")
	if !ok {
		return "", ""
	}

	schema := content.Schema.Schema()
	keys := schema.Properties.KeysFromNewest()

	requestParameters.WriteString(fmt.Sprintf("type %sRequestBody struct {", methodName) + "\n")

	for key := range keys {
		if slices.Contains(schema.Required, key) {
			initBody.WriteString(fmt.Sprintf(`initBody["%[1]s"] = *b.%[2]s`, key, FirstAlphabetToUpperCase(key)) + "\n")
		} else {
			initBody.WriteString(fmt.Sprintf(`
			if b.%[1]s != nil {
				initBody["%[2]s"] = *b.%[1]s
			}`, FirstAlphabetToUpperCase(key), key) + "\n")
		}

		schemaValue, ok := schema.Properties.Get(key)
		if !ok {
			return requestParameters.String(), initBody.String()
		}

		switch schemaValue.Schema().Type[0] {
		case "string":
			requestParameters.WriteString(fmt.Sprintf("%[1]s *string `json:\"%[2]s,,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "boolean":
			requestParameters.WriteString(fmt.Sprintf("%[1]s *bool `json:\"%[2]s,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "integer":
			if schemaValue.Schema().Format == "int64" {
				requestParameters.WriteString(fmt.Sprintf("%[1]s *int64 `json:\"%[2]s,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
			} else if schemaValue.Schema().Format == "int32" {
				requestParameters.WriteString(fmt.Sprintf("%[1]s *int32 `json:\"%[2]s,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
			}
		case "number":
			requestParameters.WriteString(fmt.Sprintf("%[1]s *float64 `json:\"%[2]s,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "array":
			requestParameters.WriteString(fmt.Sprintf("%[1]s []*string `json:\"%[2]s,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "object":
			requestParameters.WriteString(fmt.Sprintf("%[1]s []*string `json:\"%[2]s,omitempty\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		}
	}
	requestParameters.WriteString(fmt.Sprintf("}") + "\n")

	return requestParameters.String(), initBody.String()
}

func getFunctionName(methodName string, queryParameters string, bodyParameters string) string {
	var functionName string

	if len(queryParameters) <= 0 {
		functionName = fmt.Sprintf("func (n *NClient) %[1]s(ctx context.Context, b *%[1]sRequestBody) (map[string]interface{}, error) {\n", methodName)
	} else if len(bodyParameters) <= 0 {
		functionName = fmt.Sprintf("func (n *NClient) %[1]s(ctx context.Context, q *%[1]sRequestQuery) (map[string]interface{}, error) {\n", methodName)
	} else {
		functionName = fmt.Sprintf("func (n *NClient) %[1]s(ctx context.Context, q *%[1]sRequestQuery, b *%[1]sRequestBody) (map[string]interface{}, error) {\n", methodName)
	}

	return functionName
}

func MustAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Error getting absolute path for %s: %v", path, err)
	}
	return absPath
}
