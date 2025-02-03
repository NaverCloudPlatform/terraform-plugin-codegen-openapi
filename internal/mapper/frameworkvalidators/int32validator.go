// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package frameworkvalidators

import (
	"strconv"
	"strings"

	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/code"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/schema"
)

const (
	// Int32ValidatorPackage is the name of the int32 validation package in
	// the framework validators module.
	Int32ValidatorPackage = "int32validator"
)

var (
	// Int32ValidatorCodeImport is a single allocation of the framework
	// validators module int32validator package import.
	Int32ValidatorCodeImport code.Import = CodeImport(Int32ValidatorPackage)
)

// Int32ValidatorAtLeast returns a custom validator mapped to the
// int32validator package AtLeast function.
func Int32ValidatorAtLeast(minimum int32) *schema.CustomValidator {
	var schemaDefinition strings.Builder

	schemaDefinition.WriteString(Int32ValidatorPackage)
	schemaDefinition.WriteString(".AtLeast(")
	schemaDefinition.WriteString(strconv.Itoa(int(minimum)))
	schemaDefinition.WriteString(")")

	return &schema.CustomValidator{
		Imports: []code.Import{
			Int32ValidatorCodeImport,
		},
		SchemaDefinition: schemaDefinition.String(),
	}
}

// Int32ValidatorAtMost returns a custom validator mapped to the
// int32validator package AtMost function.
func Int32ValidatorAtMost(maximum int32) *schema.CustomValidator {
	var schemaDefinition strings.Builder

	schemaDefinition.WriteString(Int32ValidatorPackage)
	schemaDefinition.WriteString(".AtMost(")
	schemaDefinition.WriteString(strconv.Itoa(int(maximum)))
	schemaDefinition.WriteString(")")

	return &schema.CustomValidator{
		Imports: []code.Import{
			Int32ValidatorCodeImport,
		},
		SchemaDefinition: schemaDefinition.String(),
	}
}

// Int32ValidatorBetween returns a custom validator mapped to the
// int32validator package Between function.
func Int32ValidatorBetween(minimum, maximum int32) *schema.CustomValidator {
	var schemaDefinition strings.Builder

	schemaDefinition.WriteString(Int32ValidatorPackage)
	schemaDefinition.WriteString(".Between(")
	schemaDefinition.WriteString(strconv.Itoa(int(minimum)))
	schemaDefinition.WriteString(", ")
	schemaDefinition.WriteString(strconv.Itoa(int(maximum)))
	schemaDefinition.WriteString(")")

	return &schema.CustomValidator{
		Imports: []code.Import{
			Int32ValidatorCodeImport,
		},
		SchemaDefinition: schemaDefinition.String(),
	}
}

// Int32ValidatorOneOf returns a custom validator mapped to the int32validator
// package OneOf function. If the values are nil or empty, nil is returned.
func Int32ValidatorOneOf(values []int32) *schema.CustomValidator {
	if len(values) == 0 {
		return nil
	}

	var schemaDefinition strings.Builder

	schemaDefinition.WriteString(Int32ValidatorPackage)
	schemaDefinition.WriteString(".OneOf(\n")

	for _, value := range values {
		schemaDefinition.WriteString(strconv.Itoa(int(value)) + ",\n")
	}

	schemaDefinition.WriteString(")")

	return &schema.CustomValidator{
		Imports: []code.Import{
			Int32ValidatorCodeImport,
		},
		SchemaDefinition: schemaDefinition.String(),
	}
}
