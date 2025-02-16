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
	primitiveRequest                   string
	stringifiedRequest                 string
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

	primitiveRequest, stringifiedRequest, q, b := getAll(oas.Parameters, oas.RequestBody)

	t.methodName = t.method + getMethodName(path)
	t.model = refreshDetails.Model
	t.refreshLogic = refreshDetails.RefreshLogic
	t.path = getPath(path)
	t.primitiveRequest = primitiveRequest
	t.stringifiedRequest = stringifiedRequest
	t.query = q
	t.body = b
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
		MethodName         string
		PrimitiveRequest   string
		StringifiedRequest string
		Query              string
		Body               string
		Path               string
		Method             string
	}{
		MethodName:         t.methodName,
		Method:             t.method,
		PrimitiveRequest:   t.primitiveRequest,
		StringifiedRequest: t.stringifiedRequest,
		Query:              t.query,
		Body:               t.body,
		Path:               t.path,
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
			s = s + fmt.Sprintf(`ClearDoubleQuote(r.%s)`, PathToPascal(val))
		}
	}

	return s
}

func getAll(params []*v3high.Parameter, body *v3high.RequestBody) (string, string, string, string) {
	var primitiveRequest strings.Builder
	var stringifiedRequest strings.Builder
	var q strings.Builder

	for _, params := range params {

		key := params.Name

		stringifiedRequest.WriteString(fmt.Sprintf("%[1]s string `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")
		// In Default, all parameters needs to be in request struct
		switch params.Schema.Schema().Type[0] {
		case "string":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s string `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")

		case "boolean":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s bool `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")

		case "integer":
			if params.Schema.Schema().Format == "int64" {
				primitiveRequest.WriteString(fmt.Sprintf("%[1]s int64 `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")
			} else if params.Schema.Schema().Format == "int32" {
				primitiveRequest.WriteString(fmt.Sprintf("%[1]s int32 `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")
			}

		case "number":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s float64 `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")

		case "array":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s types.List `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")

		case "object":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s types.Object `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")
		}

		// In case of query parameters
		if params.In == "query" {
			if params.Required == nil {
				// optional query parameters
				q.WriteString(fmt.Sprintf(`
				if r.%[1]s!= "" {
					query["%[2]s"] = r.%[1]s
				}`, FirstAlphabetToUpperCase(key), key) + "\n")
			} else {
				// required query parameters
				q.WriteString(fmt.Sprintf(`
				query["%[1]s"] = r.%[2]s`, key, FirstAlphabetToUpperCase(key)) + "\n")
			}

		}
	}

	b, bodyForStringifiedRequest, bodyForPrimitiveRequest := getBody(body)

	primitiveRequest.WriteString(bodyForPrimitiveRequest)
	stringifiedRequest.WriteString(bodyForStringifiedRequest)

	return primitiveRequest.String(), stringifiedRequest.String(), q.String(), b
}

func getBody(body *v3high.RequestBody) (string, string, string) {
	var b strings.Builder
	var primitiveRequest strings.Builder
	var stringifiedRequest strings.Builder

	// return if requestBody does not needed.
	if body == nil {
		return "", "", ""
	}

	content, ok := body.Content.OrderedMap.Get("application/json;charset=UTF-8")
	if !ok {
		return b.String(), stringifiedRequest.String(), primitiveRequest.String()
	}

	schema := content.Schema.Schema()
	keys := schema.Properties.KeysFromNewest()

	for key := range keys {
		if slices.Contains(schema.Required, key) {
			b.WriteString(fmt.Sprintf(`initBody["%[1]s"] = r.%[2]s`, key, FirstAlphabetToUpperCase(key)) + "\n")
		} else {
			b.WriteString(fmt.Sprintf(`
			if r.%[1]s != "" {
				initBody["%[2]s"] = r.%[1]s
			}`, FirstAlphabetToUpperCase(key), key) + "\n")
		}

		schemaValue, ok := schema.Properties.Get(key)
		if !ok {
			return b.String(), stringifiedRequest.String(), primitiveRequest.String()
		}

		switch schemaValue.Schema().Type[0] {
		case "string":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s string `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "boolean":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s bool `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "integer":
			if schemaValue.Schema().Format == "int64" {
				primitiveRequest.WriteString(fmt.Sprintf("%[1]s int64 `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
			} else if schemaValue.Schema().Format == "int32" {
				primitiveRequest.WriteString(fmt.Sprintf("%[1]s int32 `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
			}
		case "number":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s float64 `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "array":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s types.List `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		case "object":
			primitiveRequest.WriteString(fmt.Sprintf("%[1]s types.Object `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
		}

		stringifiedRequest.WriteString(fmt.Sprintf("%[1]s string `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
	}

	return b.String(), stringifiedRequest.String(), primitiveRequest.String()
}

func MustAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Error getting absolute path for %s: %v", path, err)
	}
	return absPath
}
