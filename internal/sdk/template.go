package sdk

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type Template struct {
	OAS        *v3high.Operation
	funcMap    template.FuncMap
	methodName string
	method     string
	path       string
	request    string
	query      string
	body       string
}

func New(oas *v3high.Operation, method, path string) *Template {

	t := &Template{
		OAS:    oas,
		method: method,
	}

	funcMap := CreateFuncMap()

	r, q, b := getAll(oas.Parameters, oas.RequestBody)

	t.methodName = getMethodName(path)
	t.path = getPath(path)
	t.request = r
	t.query = q
	t.body = b
	t.funcMap = funcMap

	return t
}

func (t *Template) WriteTemplate() []byte {
	var b bytes.Buffer

	methodTemplate, err := template.New("").Funcs(t.funcMap).Parse(MethodTemplate)
	if err != nil {
		log.Fatalf("error occurred with baseTemplate at rendering create: %v", err)
	}

	data := struct {
		MethodName string
		Request    string
		Query      string
		Body       string
		Path       string
		Method     string
	}{
		MethodName: t.methodName,
		Method:     t.method,
		Request:    t.request,
		Query:      t.query,
		Body:       t.body,
		Path:       t.path,
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
			s = s + fmt.Sprintf(`r.%s`, PathToPascal(val))
		}
	}

	return s
}

func getAll(params []*v3high.Parameter, body *v3high.RequestBody) (string, string, string) {
	var r strings.Builder
	var q strings.Builder

	for _, params := range params {

		key := params.Name

		// In Default, all parameters needs to be in requeest struct
		r.WriteString(fmt.Sprintf("%[1]s string `json:\"%[2]s\"`", PathToPascal(key), key) + "\n")

		// In case  of query parameters
		if params.In == "query" {
			q.WriteString(fmt.Sprintf(`"%[1]s": r.%[2]s`, key, FirstAlphabetToUpperCase(key)) + "\n")
		}
	}

	b, bodyForRequest := getBody(body)

	r.WriteString(bodyForRequest)

	return r.String(), q.String(), b
}

func getBody(body *v3high.RequestBody) (string, string) {
	var b strings.Builder
	var r strings.Builder

	content, ok := body.Content.OrderedMap.Get("application/json;charset=UTF-8")
	if !ok {
		return b.String(), r.String()
	}

	schema := content.Schema.Schema()
	keys := schema.Properties.KeysFromNewest()

	for key := range keys {
		b.WriteString(fmt.Sprintf(`"%[1]s": r.%[2]s`, key, FirstAlphabetToUpperCase(key)) + "\n")
		r.WriteString(fmt.Sprintf("%[1]s string `json:\"%[2]s\"`", FirstAlphabetToUpperCase(key), key) + "\n")
	}

	return b.String(), r.String()
}
