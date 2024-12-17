package sdk

import (
	_ "embed"
)

//go:embed templates/method.go.tpl
var MethodTemplate string

//go:embed templates/client.go.tpl
var ClientTemplate string
