package sdk

import (
	_ "embed"
)

//go:embed templates/method.go.tpl
var MethodTemplate string

//go:embed templates/client.go.tpl
var ClientTemplate string

//go:embed templates/refresh.go.tpl
var RefreshTemplate string
