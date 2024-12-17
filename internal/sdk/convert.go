package sdk

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// generateStructs 함수는 OpenAPI 모델을 기반으로 Go 구조체를 생성합니다.
func GenerateStructs(responses *v3high.Responses, responseName string) ([]byte, string, string, error) {
	var s bytes.Buffer
	var ns string
	var nm string
	codes := []string{
		"200",
		"201",
	}

	for _, code := range codes {
		g, pre := responses.Codes.Get(code)
		if !pre {
			continue
			// return nil, errors.New("error in parsing openapi: couldn't find GET response with status code 200")
		}

		c, pre := g.Content.OrderedMap.Get("application/json;charset=UTF-8")
		if !pre {
			return nil, "", "", errors.New("error in parsing openapi: couldn't find valid content with application/json;charset=UTF-8 header")
		}

		s.Write(buildStructFromSchema(c.Schema.Schema().Properties, responseName+"Response"))
		newS, newM := Gen_ConvertOAStoTFTypes(c.Schema.Schema(), c.Schema.Schema().Type[0], c.Schema.Schema().Format, responseName)

		ns = newS
		nm = newM
	}

	return s.Bytes(), ns, nm, nil
}

// buildStructFromSchema 함수는 주어진 스키마를 기반으로 Go 구조체 코드를 생성합니다.
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

// mapOpenAPITypeToGoType 함수는 OpenAPI 타입을 Go 타입으로 매핑합니다.
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
		}`, innerArray.String()) // 배열 타입은 추가 처리가 필요함
	case "object":
		var innerArray strings.Builder
		for propName, propSchema := range propreties.Properties.FromNewest() {
			goType := mapOpenAPITypeToGoType(propSchema.Schema(), propSchema.Schema().Type[0], propSchema.Schema().Format)
			innerArray.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n",
				toCamelCase(propName), goType, propName))
		}

		return fmt.Sprintf(` struct{
			%s
		}`, innerArray.String()) // 배열 타입은 추가 처리가 필요함
	default:
		return "interface{}"
	}
}

// toCamelCase 함수는 문자열을 CamelCase로 변환합니다.
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
