package sdk

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// generate converter that convert openapi.json schema to terraform type
func Gen_ConvertOAStoTFTypes(propreties *base.Schema, openapiType, format, resourceName string) (s, m, convertValueWithNull, possibleTypes, convertValueWithNullInEmptyArrCase string) {

	for name, propSchema := range propreties.Properties.FromNewest() {
		switch propSchema.Schema().Type[0] {
		case "string":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.StringValue(data["%[2]s"].(string))
			}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.String `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"

		case "integer":
			switch propSchema.Schema().Format {
			case "int64":
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					dto.%[1]s = types.Int64Value(data["%[2]s"].(int64))
				}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
				m = m + fmt.Sprintf("%[1]s         types.Int64`tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"

			case "int32":
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					dto.%[1]s = types.Int32Value(data["%[2]s"].(int32))
				}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
				m = m + fmt.Sprintf("%[1]s         types.Int32`tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"
			}

		case "number":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.Float64Value(data["%[2]s"].(float64))
			}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Float64 `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"

		case "boolean":
			s = s + fmt.Sprintf(`
			if data["%[2]s"] != nil {
				dto.%[1]s = types.BoolValue(data["%[2]s"].(bool))
			}`, ToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
			m = m + fmt.Sprintf("%[1]s         types.Bool `tfsdk:\"%[2]s\"`", ToPascalCase(name), PascalToSnakeCase(name)) + "\n"

		case "array":

			switch propSchema.Schema().Items.A.Schema().Type[0] {
			case "object":
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					temp%[1]s := data["%[2]s"].([]interface{})
					dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType:
						%[3]s
					}}.ElementType(), temp%[1]s)
				}`, CamelToPascalCase(name), PascalToSnakeCase(CamelToPascalCase(name)), GenArray(propSchema.Schema().Items.A.Schema(), name)+"\n")

			case "string":
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					temp%[1]s := data["%[2]s"].([]interface{})
					dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.StringType}.ElementType(), temp%[1]s)
				}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"

			case "boolean":
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					temp%[1]s := data["%[2]s"].([]interface{})
					dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.BoolType}.ElementType(), temp%[1]s)
				}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"

			case "integer":
				switch propSchema.Schema().Items.A.Schema().Format {
				case "int64":
					s = s + fmt.Sprintf(`
					if data["%[2]s"] != nil {
						temp%[1]s := data["%[2]s"].([]interface{})
						dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.Int64Type}.ElementType(), temp%[1]s)
					}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"

				case "int32":
					s = s + fmt.Sprintf(`
					if data["%[2]s"] != nil {
						temp%[1]s := data["%[2]s"].([]interface{})
						dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.Int32Type}.ElementType(), temp%[1]s)
					}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
				}

			case "number":
				s = s + fmt.Sprintf(`
				if data["%[2]s"] != nil {
					temp%[1]s := data["%[2]s"].([]interface{})
					dto.%[1]s = diagOff(types.ListValueFrom, ctx, types.ListType{ElemType: types.Float64Type}.ElementType(), temp%[1]s)
				}`, ToPascalCase(PascalToSnakeCase(name)), PascalToSnakeCase(CamelToPascalCase(name))) + "\n"
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
			convertValueWithNull, convertValueWithNullInEmptyArrCase = GenConvertValueWithNull(propSchema.Schema(), name)
		}
	}

	return s, m, convertValueWithNull, possibleTypes, convertValueWithNullInEmptyArrCase
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

			switch schema.Schema().Format {
			case "int64":
				t = t + fmt.Sprintf(`"%[1]s": types.Int64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

			case "int32":
				t = t + fmt.Sprintf(`"%[1]s": types.Int32Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

			}

		case "number":
			t = t + fmt.Sprintf(`"%[1]s": types.Float64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

		case "array":

			switch schema.Schema().Items.A.Schema().Type[0] {
			case "object":
				t = t + fmt.Sprintf(`
				"%[1]s": types.ListType{ElemType:
					types.ObjectType{AttrTypes: map[string]attr.Type{
						%[2]s
					},
				}},`, PascalToSnakeCase(CamelToPascalCase(n)), GenObject(schema.Schema().Items.A.Schema(), n)) + "\n"

			case "string":
				t = t + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.StringType},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

			case "boolean":
				t = t + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.BoolType},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

			case "integer":
				switch schema.Schema().Format {
				case "int64":
					t = t + fmt.Sprintf(`"%[1]s": types.Int64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

				case "int32":
					t = t + fmt.Sprintf(`"%[1]s": types.Int32Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
				}

			case "number":
				t = t + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.Float64Type},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
			}

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
			switch schema.Schema().Format {
			case "int64":
				s = s + fmt.Sprintf(`"%[1]s": types.Int64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

			case "int32":
				s = s + fmt.Sprintf(`"%[1]s": types.Int32Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
			}

		case "number":
			s = s + fmt.Sprintf(`"%[1]s": types.Float64Type,`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

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
				switch schema.Schema().Items.A.Schema().Type[0] {
				case "string":
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.StringType},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

				case "boolean":
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.BoolType},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

				case "integer":
					switch schema.Schema().Items.A.Schema().Format {
					case "int64":
						s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.Int64Type},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"

					case "int32":
						s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.Int32Type},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
					}

				case "number":
					s = s + fmt.Sprintf(`"%[1]s": types.ListType{ElemType: types.Float64Type},`, PascalToSnakeCase(CamelToPascalCase(n))) + "\n"
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

func GenConvertValueWithNull(d *base.Schema, pName string) (s string, v string) {
	for n, schema := range d.Properties.FromNewest() {
		switch schema.Schema().Type[0] {
		case "array":
			// in case of empty array, logic assumes it non-null
			// so explicitly check it
			v = v + fmt.Sprintf(`
			if field == "%[1]s" && len(value.([]interface{})) == 0 {
				listV := types.ListNull(types.ObjectNull(map[string]attr.Type{
					%[2]s
				}).Type(ctx))
				attrValues[field] = listV
				continue
			}`, PascalToSnakeCase(n), GenObject(schema.Schema().Items.A.Schema(), n)) + "\n"

			switch schema.Schema().Items.A.Schema().Type[0] {
			case "object":
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.ObjectNull(map[string]attr.Type{
						%[2]s
					}).Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n), GenObject(schema.Schema().Items.A.Schema(), n)) + "\n"

			case "string":
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.StringNull().Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n)) + "\n"

			case "boolean":
				s = s + fmt.Sprintf(`
					if field == "%[1]s" {
						listV := types.ListNull(types.BoolNull().Type(ctx))
						attrValues[field] = listV
						continue
					}`, PascalToSnakeCase(n)) + "\n"

			case "integer":
				switch schema.Schema().Items.A.Schema().Format {
				case "int64":
					s = s + fmt.Sprintf(`
					if field == "%[1]s" {
						listV := types.ListNull(types.Int64Null().Type(ctx))
						attrValues[field] = listV
						continue
					}`, PascalToSnakeCase(n)) + "\n"

				case "int32":
					s = s + fmt.Sprintf(`
					if field == "%[1]s" {
						listV := types.ListNull(types.Int32Null().Type(ctx))
						attrValues[field] = listV
						continue
					}`, PascalToSnakeCase(n)) + "\n"
				}

			case "number":
				s = s + fmt.Sprintf(`
				if field == "%[1]s" {
					listV := types.ListNull(types.Float64Null().Type(ctx))
					attrValues[field] = listV
					continue
				}`, PascalToSnakeCase(n)) + "\n"
			}

		case "object":
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

	return s, v
}
