package sdk

import (
	_ "embed"
)

//go:embed templates/method.go.tpl
var MethodTemplate string
