{{ define "Refresh" }}
// Template for generating Terraform provider Refresh operation code
// Needed data is as follows.
// Model string
// MethodName string
// RefreshLogic string
// PossibleTypes string
// ConditionalObjectFieldsWithNull string

type {{.MethodName}}Response struct {
    {{.Model}}
}

func ConvertToFrameworkTypes_{{.MethodName}}(data map[string]interface{}) (*{{.MethodName}}Response, error) {
	var dto {{.MethodName}}Response

    ctx := context.TODO()

    {{.RefreshLogic}}

	return &dto, nil
}

func convertToObject_{{.MethodName}}(data map[string]interface{}) (types.Object, error) {
	attrTypes := make(map[string]attr.Type)
	attrValues := make(map[string]attr.Value)

    possibleTypes := map[string]attr.Type{
        {{.PossibleTypes}}
	}

	for field, fieldType := range possibleTypes {
		attrTypes[field] = fieldType

		if value, exists := data[field]; exists {

			attrValue, err := convertValueToAttr_{{.MethodName}}(value)
			if err != nil {
				return types.Object{}, fmt.Errorf("error converting field %s: %v", field, err)
			}
			attrValues[field] = attrValue
		} else {
            {{.ConditionalObjectFieldsWithNull}}

			switch fieldType {
			case types.StringType:
				attrValues[field] = types.StringNull()
			case types.Int64Type:
				attrValues[field] = types.Int64Null()
			case types.BoolType:
				attrValues[field] = types.BoolNull()
			}
		}
	}
}

func convertValueToAttr_{{.MethodName}}(value interface{}) (attr.Value, error) {
	switch v := value.(type) {
	case string:
		return types.StringValue(v), nil
	case float64:
		return types.Int64Value(int64(v)), nil
	case bool:
		return types.BoolValue(v), nil
	case nil:
		return types.StringNull(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}
}

{{ end }}