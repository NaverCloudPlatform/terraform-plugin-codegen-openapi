{{ define "Refresh" }}
// Template for generating Terraform provider Refresh operation code
// Needed data is as follows.
// Model string
// MethodName string
// RefreshLogic string

type {{.MethodName}}Response struct {
    {{.Model}}
}

func ConvertToFrameworkTypes_{{.MethodName}}(data map[string]interface{}) (*{{.MethodName}}Response, error) {
	var dto {{.MethodName}}Response

    {{.RefreshLogic}}

	return &dto, nil
}

{{ end }}