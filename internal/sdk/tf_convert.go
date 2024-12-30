package sdk

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// generate converter that convert openapi.json schema to terraform type
func Gen_ConvertOAStoTFTypes(propreties *base.Schema, openapiType, format, resourceName string) (s, m, convertValueWithNull, possibleTypes string) {

	for name, propSchema := range propreties.Properties.FromNewest() {
		switch propSchema.Schema().Type[0] {
		case "string":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.StringValue(data["%[2]s"].(string))
			}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.String `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "integer":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.Int64Value(data["%[2]s"].(int64))
			}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Int64`tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "boolean":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.BoolValue(data["%[2]s"].(bool))
			}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Bool `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "array":

			// Case for List Nested
			if propSchema.Schema().Items.A.Schema().Type[0] == "object" {
				s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				temp%[1]s := data["%[2]s"].([]interface{})
				dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType:
					%[3]s
				}}.ElementType(), temp%[1]s)
			}`, CamelToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name)), GenArray(propSchema.Schema().Items.A.Schema(), name)+"\n")
			}

			if propSchema.Schema().Items.A.Schema().Type[0] == "string" {
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					temp%[1]s := data["%[2]s"].([]interface{})
					dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.StringType}.ElementType(), temp%[1]s)
				}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
				// s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.StringType},`, name) + "\n"
			} else if propSchema.Schema().Items.A.Schema().Type[0] == "boolean" {
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					temp%[1]s := data["%[2]s"].([]interface{})
					dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.BoolType}.ElementType(), temp%[1]s)
				}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
				// s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.BoolType},`, name) + "\n"
			}

			m = m + fmt.Sprintf("%[1]s         types.List `tfsdk:\"%[2]s\"`", CamelToPascalCase(name), PascalToSnakeCase(name)) + "\n"
		case "object":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				temp%[1]s := data["%[2]s"].(map[string]interface{})

				allFields := []string{
					%[5]s
				}

				convertedMap := make(map[string]interface{})
				for _, field := range allFields {
					if val, ok := temp%[1]s[field]; ok {
						convertedMap[field] = val
					}
				}

				convertedTemp%[1]s, err := convertToObject_%[4]s(ctx, convertedMap)
				if err != nil {
					return nil, err
				}

				dto.%[1]s = diagOff(types.ObjectValueFrom, ctx, types.ObjectType{AttrTypes: map[string]attr.Type{
					%[3]s
				}}.AttributeTypes(), convertedTemp%[1]s)
			}`, CamelToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name)), GenObject(propSchema.Schema(), name), resourceName, GenAllFields(propSchema.Schema())) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Object `tfsdk:\"%[2]s\"`", CamelToPascalCase(name), PascalToSnakeCase(name)) + "\n"
			possibleTypes = possibleTypes + GenObject(propSchema.Schema(), name) + "\n"
			convertValueWithNull = GenConvertValueWithNull(propSchema.Schema(), name)
		}
	}

	return s, m, convertValueWithNull, possibleTypes
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
			t = t + fmt.Sprintf(`"%[1]s": types.StringType,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
		case "boolean":
			t = t + fmt.Sprintf(`"%[1]s": types.BoolType,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
		case "integer":
			t = t + fmt.Sprintf(`"%[1]s": types.Int64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
		case "object":
			s = s + fmt.Sprintf(`
			"%[1]s": types.ObjectType{AttrTypes: map[string]attr.Type{
				%[2]s
			}},`, PascalToSnakeCase(CamelToPascalCase(n)), GenObject(schema.Schema(), n)) + "\n"
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

	for n, schema := range d.Properties.FromNewest() {
		switch schema.Schema().Type[0] {
		case "string":
			s = s + fmt.Sprintf(`"%[1]s": types.StringType,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
		case "boolean":
			s = s + fmt.Sprintf(`"%[1]s": types.BoolType,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
		case "integer":
			s = s + fmt.Sprintf(`"%[1]s": types.Int64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
		case "object":
			// In case of `properties: { }`
			if schema.Schema().Properties == nil {
				s = s + fmt.Sprintf(`
				"%[1]s": types.ObjectType{AttrTypes: map[string]attr.Type{
				}},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
			} else {
				s = s + fmt.Sprintf(`
				"%[1]s": types.ObjectType{AttrTypes: map[string]attr.Type{
					%[2]s
				}},`, PascalToSnakeCase(CamelToPascalCase(n)), GenObject(schema.Schema(), n)) + "\n"
			}
		case "array":
			if schema.Schema().Items.A.Schema().Type[0] == "object" {
				s = s + fmt.Sprintf(`
			"%[1]s": types.ListType{ElemType:
				%[2]s
			}},`, PascalToSnakeCase(CamelToPascalCase(n)), GenArray(schema.Schema().Items.A.Schema(), n)) + "\n"
			} else {
				if schema.Schema().Items.A.Schema().Type[0] == "string" {
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.StringType},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
				} else if schema.Schema().Items.A.Schema().Type[0] == "boolean" {
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.BoolType},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
				}
			}
		}
	}
	return s
}

func GenAllFields(d *base.Schema) string {
	var s string

	for n := range d.Properties.FromNewest() {
		s = s + fmt.Sprintf(`"%[1]s",`, PascalToSnakeCase(n)) + "\n"
	}
	return s
}

func GenConvertValueWithNull(d *base.Schema, pName string) string {
	var s string

	for n, schema := range d.Properties.FromNewest() {
		if schema.Schema().Type[0] == "array" {
			if schema.Schema().Items.A.Schema().Type[0] == "object" {
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.ObjectNull(map[string]attr.Type{
						%[2]s
					}).Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n), GenObject(schema.Schema().Items.A.Schema(), n)) + "\n"
			} else if schema.Schema().Items.A.Schema().Type[0] == "string" {
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.StringNull().Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n)) + "\n"
			} else if schema.Schema().Items.A.Schema().Type[0] == "boolean" {
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.BoolNull().Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n)) + "\n"
			} else if schema.Schema().Items.A.Schema().Type[0] == "integer" {
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.Int64Null().Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n)) + "\n"
			}
		} else if schema.Schema().Type[0] == "object" {
			// In case of `properties: { }`
			if schema.Schema().Properties == nil {
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ObjectNull(map[string]attr.Type{})
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n)) + "\n"
			} else {
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ObjectNull(map[string]attr.Type{
						%[2]s
					})
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n), GenObject(schema.Schema(), n)) + "\n"
			}
		}
	}

	return s
}
