package sdk

import (
	"fmt"

	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Generate terraform-spec type based struct with *v3high.Responses input
func GenerateStructs(responses *v3high.Responses, responseName string) (string, string, string, string, string, error) {

	codes := []string{
		"200",
		"201",
	}

	for _, code := range codes {
		g, pre := responses.Codes.Get(code)
		if !pre {
			// Skip when expected status code does not exists.
			continue
		}

		c, pre := g.Content.OrderedMap.Get("application/json;charset=UTF-8")
		if !pre {
			// Skip when expected status code does not exists.
			continue
		}

		refreshLogic, model, convertValueWithNull, possibleTypes, convertValueWithNullInEmptyArrCase := Gen_ConvertOAStoTFTypes(c.Schema.Schema(), c.Schema.Schema().Type[0], c.Schema.Schema().Format, responseName)

		return refreshLogic, model, convertValueWithNull, possibleTypes, convertValueWithNullInEmptyArrCase, nil
	}

	return "", "", "", "", "", fmt.Errorf("no suitable responses found")
}
