package sdk

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// generate converter that convert openapi.json schema to terraform type
func Gen_ConvertOAStoTFTypes(propreties *base.Schema, openapiType, format, resourceName string) (s string, m string) {

	for name, propSchema := range propreties.Properties.FromNewest() {
		switch propSchema.Schema().Type[0] {
		case "string":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.StringValue(data["%[2]s"].(string))
			}`, ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.String `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "integer":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.Int64Value(data["%[2]s"].(string))
			}`, ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Bool `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "boolean":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.BoolValue(data["%[2]s"].(string))
			}`, ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Int64 `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "array":

			// Case for List Nested
			if propSchema.Schema().Items.A.Schema().Type[0] == "object" {
				s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				temp%[1]s := data["%[2]s"].([]interface{})
				dto.%[1]s = diagOff(types.ListValueFrom, context.TODO(), types.ListType{ElemType:
					%[3]s
				}}.ElementType(), temp%[1]s)
			}`, CamelToPascalCase(name), PascalToSnakeCase(name), GenArray(propSchema.Schema().Items.A.Schema(), name)+"\n")
			}

			if propSchema.Schema().Items.A.Schema().Type[0] == "string" {
				s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.StringType},`, name) + "\n"
			} else if propSchema.Schema().Items.A.Schema().Type[0] == "boolean" {
				s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.BoolType},`, name) + "\n"
			}

			m = m + fmt.Sprintf("%[1]s         types.List `tfsdk:\"%[2]s\"`", CamelToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "object":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				temp%[1]s := data["%[2]s"].(map[string]interface{})
				convertedTemp%[1]s, err := convertMapToObject(context.TODO(), temp%[1]s)
				if err != nil {
					fmt.Println("ConvertMapToObject Error")
				}

				dto.%[1]s = diagOff(types.ObjectValueFrom, context.TODO(), types.ObjectType{AttrTypes: map[string]attr.Type{
					%[3]s
				}}.AttributeTypes(), convertedTemp%[1]s)
			}`, CamelToPascalCase(name), PascalToSnakeCase(name), GenObject(propSchema.Schema(), name)) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Object `tfsdk:\"%[2]s\"`", CamelToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		}
	}

	return s, m
}

func PascalToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

func CamelToPascalCase(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func GenArray(d *base.Schema, pName string) string {
	var r string
	var s string
	var t string

	for n, schema := range d.Properties.FromNewest() {

		switch schema.Schema().Type[0] {
		case "string":
			t = t + fmt.Sprintf(`"%[1]s": types.StringType,`, n) + "\n"
		case "boolean":
			t = t + fmt.Sprintf(`"%[1]s": types.BoolType,`, n) + "\n"
		case "integer":
			t = t + fmt.Sprintf(`"%[1]s": types.Int64Type,`, n) + "\n"
		case "object":
			s = s + fmt.Sprintf(`
			"%[1]s": types.ObjectType{AttrTypes: map[string]attr.Type{
				%[2]s
			}},`, n, GenObject(schema.Schema().Properties.Newest().Value.Schema(), n)) + "\n"
		}
	}

	r = r + fmt.Sprintf(`
	types.ObjectType{AttrTypes: map[string]attr.Type{
		%[1]s
		%[2]s
	},`, s, t)

	return r
}

func GenObject(d *base.Schema, pName string) string {
	var s string

	fmt.Println(pName)

	for n, schema := range d.Properties.FromNewest() {
		fmt.Println(n, schema.Schema().Type[0])

		switch schema.Schema().Type[0] {
		case "string":
			s = s + fmt.Sprintf(`"%[1]s": types.StringType,`, n) + "\n"
		case "boolean":
			s = s + fmt.Sprintf(`"%[1]s": types.BoolType,`, n) + "\n"
		case "integer":
			s = s + fmt.Sprintf(`"%[1]s": types.Int64Type,`, n) + "\n"
		case "object":
			// In case of `properties: { }`
			if schema.Schema().Properties == nil {
				s = s + fmt.Sprintf(`
				"%[1]s": types.ObjectType{AttrTypes: map[string]attr.Type{
				}},`, n) + "\n"
			} else {
			s = s + fmt.Sprintf(`
			"%[1]s": types.ObjectType{AttrTypes: map[string]attr.Type{
				%[2]s
			}},`, n, GenObject(schema.Schema().Properties.Newest().Value.Schema(), n)) + "\n"
			}
		case "array":
			if schema.Schema().Items.A.Schema().Type[0] == "object" {
				s = s + fmt.Sprintf(`
			"%[1]s": types.ListType{ElemType:
				%[2]s
			}},`, n, GenArray(schema.Schema().Items.A.Schema(), n)) + "\n"
			} else {
				if schema.Schema().Items.A.Schema().Type[0] == "string" {
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.StringType},`, n) + "\n"
				} else if schema.Schema().Items.A.Schema().Type[0] == "boolean" {
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.BoolType},`, n) + "\n"
				}
			}
		}
	}
	return s
}
