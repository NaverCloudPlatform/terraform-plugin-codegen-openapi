package sdk

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Generate terraform-spec type based struct with *v3high.Responses input
func GenerateStructs(responses *v3high.Responses, responseName string) (refreshLogic string, model string, convertValueWithNull string, possibleTypes string) {

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

		newRefreshLogic, newModel, newConvertValueWithNull, newPossibleTypes := Gen_ConvertOAStoTFTypes(c.Schema.Schema(), c.Schema.Schema().Type[0], c.Schema.Schema().Format, responseName)

		refreshLogic = newRefreshLogic
		model = newModel
		convertValueWithNull = newConvertValueWithNull
		possibleTypes = newPossibleTypes
	}

	return refreshLogic, model, convertValueWithNull, possibleTypes
}

// Generate struct based on given schema
func buildStructFromSchema(propreties *orderedmap.Map[string, *base.SchemaProxy], structName string) []byte {
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("type %s struct {\n", structName))

	for propName, propSchema := range propreties.FromNewest() {
		goType := mapOpenAPITypeToGoType(propSchema.Schema(), propSchema.Schema().Type[0], propSchema.Schema().Format)
		b.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n",
			toCamelCase(propName), goType, propName))
	}

	b.WriteString("}\n")
	return b.Bytes()
}

// mapOpenAPITypeToGoType maps OpenAPI types to Go types
func mapOpenAPITypeToGoType(propreties *base.Schema, openapiType, format string) string {

	switch openapiType {
	case "string":
		if format == "date-time" {
			return "time.Time"
		}
		return "string"
	case "integer":
		if format == "int64" {
			return "int64"
		}
		return "int"
	case "boolean":
		return "bool"
	case "array":
		var innerArray strings.Builder
		for propName, propSchema := range propreties.Items.A.Schema().Properties.FromNewest() {
			goType := mapOpenAPITypeToGoType(propSchema.Schema(), propSchema.Schema().Type[0], propSchema.Schema().Format)
			innerArray.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n",
				toCamelCase(propName), goType, propName))
		}

		return fmt.Sprintf(` []struct{
			%s
		}`, innerArray.String())
	case "object":
		var innerArray strings.Builder
		for propName, propSchema := range propreties.Properties.FromNewest() {
			goType := mapOpenAPITypeToGoType(propSchema.Schema(), propSchema.Schema().Type[0], propSchema.Schema().Format)
			innerArray.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n",
				toCamelCase(propName), goType, propName))
		}

		return fmt.Sprintf(` struct{
			%s
		}`, innerArray.String())
	default:
		return "interface{}"
	}
}

func toCamelCase(input string) string {
	var result string
	capitalizeNext := true

	for _, char := range input {
		if char == '_' || char == '-' || char == ' ' {
			capitalizeNext = true
			continue
		}

		if capitalizeNext {
			result += strings.ToUpper(string(char))
			capitalizeNext = false
		} else {
			result += strings.ToLower(string(char))
		}
	}

	return result
}

type DashboardsApikeysApikeyidProductidsResponse struct {
	Products []struct {
		ActionName string `json:"actionName"`
		Disabled   bool   `json:"disabled"`
		IsDeleted  bool   `json:"isDeleted"`
	} `json:"products"`
}

type AuthorizersResponse struct {
	Tenantid              string    `json:"tenantId"`
	Modifier              string    `json:"modifier"`
	Modtime               time.Time `json:"modTime"`
	Domaincode            string    `json:"domainCode"`
	Cachettlsec           int       `json:"cacheTtlSec"`
	Authorizertype        string    `json:"authorizerType"`
	Authorizername        string    `json:"authorizerName"`
	Authorizerid          string    `json:"authorizerId"`
	Authorizerdescription string    `json:"authorizerDescription"`
	Authorizerconfig      struct {
		Payload []struct {
			Name string `json:"name"`
			In   string `json:"in"`
		} `json:"payload"`
		Functionid string `json:"functionId"`
		Region     string `json:"region"`
	} `json:"authorizerConfig"`
}
