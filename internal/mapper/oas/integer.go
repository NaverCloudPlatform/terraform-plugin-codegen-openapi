// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oas

import (
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/frameworkvalidators"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/datasource"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/provider"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/resource"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/schema"
)

func (s *OASSchema) BuildInt32Resource(name string, computability schema.ComputedOptionalRequired) (attrmapper.ResourceAttribute, *SchemaError) {
	result := &attrmapper.ResourceInt32Attribute{
		Name: name,
		Int32Attribute: resource.Int32Attribute{
			ComputedOptionalRequired: computability,
			DeprecationMessage:       s.GetDeprecationMessage(),
			Description:              s.GetDescription(),
		},
	}

	if s.Schema.Default != nil {
		var staticDefault int32
		if err := s.Schema.Default.Decode(&staticDefault); err == nil {
			if computability == schema.Required {
				result.ComputedOptionalRequired = schema.ComputedOptional
			}

			result.Default = &schema.Int32Default{
				Static: &staticDefault,
			}
		}
	}

	if computability != schema.Computed {
		result.Validators = s.GetInt32Validators()
	}

	return result, nil
}

func (s *OASSchema) BuildInt64Resource(name string, computability schema.ComputedOptionalRequired) (attrmapper.ResourceAttribute, *SchemaError) {
	result := &attrmapper.ResourceInt64Attribute{
		Name: name,
		Int64Attribute: resource.Int64Attribute{
			ComputedOptionalRequired: computability,
			DeprecationMessage:       s.GetDeprecationMessage(),
			Description:              s.GetDescription(),
		},
	}

	if s.Schema.Default != nil {
		var staticDefault int64
		if err := s.Schema.Default.Decode(&staticDefault); err == nil {
			if computability == schema.Required {
				result.ComputedOptionalRequired = schema.ComputedOptional
			}

			result.Default = &schema.Int64Default{
				Static: &staticDefault,
			}
		}
	}

	if computability != schema.Computed {
		result.Validators = s.GetInt64Validators()
	}

	return result, nil
}

func (s *OASSchema) BuildInt32DataSource(name string, computability schema.ComputedOptionalRequired) (attrmapper.DataSourceAttribute, *SchemaError) {
	result := &attrmapper.DataSourceInt32Attribute{
		Name: name,
		Int32Attribute: datasource.Int32Attribute{
			ComputedOptionalRequired: computability,
			DeprecationMessage:       s.GetDeprecationMessage(),
			Description:              s.GetDescription(),
		},
	}

	if computability != schema.Computed {
		result.Validators = s.GetInt32Validators()
	}

	return result, nil
}

func (s *OASSchema) BuildInt64DataSource(name string, computability schema.ComputedOptionalRequired) (attrmapper.DataSourceAttribute, *SchemaError) {
	result := &attrmapper.DataSourceInt64Attribute{
		Name: name,
		Int64Attribute: datasource.Int64Attribute{
			ComputedOptionalRequired: computability,
			DeprecationMessage:       s.GetDeprecationMessage(),
			Description:              s.GetDescription(),
		},
	}

	if computability != schema.Computed {
		result.Validators = s.GetInt64Validators()
	}

	return result, nil
}

func (s *OASSchema) BuildInt32Provider(name string, optionalOrRequired schema.OptionalRequired) (attrmapper.ProviderAttribute, *SchemaError) {
	result := &attrmapper.ProviderInt32Attribute{
		Name: name,
		Int32Attribute: provider.Int32Attribute{
			OptionalRequired:   optionalOrRequired,
			DeprecationMessage: s.GetDeprecationMessage(),
			Description:        s.GetDescription(),
			Validators:         s.GetInt32Validators(),
		},
	}

	return result, nil
}

func (s *OASSchema) BuildInt64Provider(name string, optionalOrRequired schema.OptionalRequired) (attrmapper.ProviderAttribute, *SchemaError) {
	result := &attrmapper.ProviderInt64Attribute{
		Name: name,
		Int64Attribute: provider.Int64Attribute{
			OptionalRequired:   optionalOrRequired,
			DeprecationMessage: s.GetDeprecationMessage(),
			Description:        s.GetDescription(),
			Validators:         s.GetInt64Validators(),
		},
	}

	return result, nil
}

func (s *OASSchema) BuildInt32ElementType() (schema.ElementType, *SchemaError) {
	return schema.ElementType{
		Int32: &schema.Int32Type{},
	}, nil
}

func (s *OASSchema) BuildInt64ElementType() (schema.ElementType, *SchemaError) {
	return schema.ElementType{
		Int64: &schema.Int64Type{},
	}, nil
}

func (s *OASSchema) GetInt32Validators() []schema.Int32Validator {
	var result []schema.Int32Validator

	if len(s.Schema.Enum) > 0 {
		var enum []int32

		for _, valueNode := range s.Schema.Enum {
			var value int32
			if err := valueNode.Decode(&value); err != nil {
				// could consider error/panic here to notify developers
				continue
			}

			enum = append(enum, value)
		}

		customValidator := frameworkvalidators.Int32ValidatorOneOf(enum)

		if customValidator != nil {
			result = append(result, schema.Int32Validator{
				Custom: customValidator,
			})
		}
	}

	minimum := s.Schema.Minimum
	maximum := s.Schema.Maximum

	if minimum != nil && maximum != nil {
		result = append(result, schema.Int32Validator{
			Custom: frameworkvalidators.Int32ValidatorBetween(int32(*minimum), int32(*maximum)),
		})
	} else if minimum != nil {
		result = append(result, schema.Int32Validator{
			Custom: frameworkvalidators.Int32ValidatorAtLeast(int32(*minimum)),
		})
	} else if maximum != nil {
		result = append(result, schema.Int32Validator{
			Custom: frameworkvalidators.Int32ValidatorAtMost(int32(*maximum)),
		})
	}

	return result
}

func (s *OASSchema) GetInt64Validators() []schema.Int64Validator {
	var result []schema.Int64Validator

	if len(s.Schema.Enum) > 0 {
		var enum []int64

		for _, valueNode := range s.Schema.Enum {
			var value int64
			if err := valueNode.Decode(&value); err != nil {
				// could consider error/panic here to notify developers
				continue
			}

			enum = append(enum, value)
		}

		customValidator := frameworkvalidators.Int64ValidatorOneOf(enum)

		if customValidator != nil {
			result = append(result, schema.Int64Validator{
				Custom: customValidator,
			})
		}
	}

	minimum := s.Schema.Minimum
	maximum := s.Schema.Maximum

	if minimum != nil && maximum != nil {
		result = append(result, schema.Int64Validator{
			Custom: frameworkvalidators.Int64ValidatorBetween(int64(*minimum), int64(*maximum)),
		})
	} else if minimum != nil {
		result = append(result, schema.Int64Validator{
			Custom: frameworkvalidators.Int64ValidatorAtLeast(int64(*minimum)),
		})
	} else if maximum != nil {
		result = append(result, schema.Int64Validator{
			Custom: frameworkvalidators.Int64ValidatorAtMost(int64(*maximum)),
		})
	}

	return result
}
