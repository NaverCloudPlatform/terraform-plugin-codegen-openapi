// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attrmapper

import (
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/util"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/datasource"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/provider"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/resource"
)

type ResourceInt32Attribute struct {
	resource.Int32Attribute

	Name string
}

func (a *ResourceInt32Attribute) GetName() string {
	return a.Name
}

func (a *ResourceInt32Attribute) Merge(mergeAttribute ResourceAttribute) (ResourceAttribute, error) {
	int32Attribute, ok := mergeAttribute.(*ResourceInt32Attribute)
	// TODO: return error if types don't match?
	if ok && (a.Description == nil || *a.Description == "") {
		a.Description = int32Attribute.Description
	}

	return a, nil
}

func (a *ResourceInt32Attribute) ApplyOverride(override explorer.Override) (ResourceAttribute, error) {
	a.Description = &override.Description

	cor, err := ApplyComputedOptionalRequiredOverride(override.ComputedOptionalRequired)
	if err != nil {
		return nil, err
	}
	a.ComputedOptionalRequired = cor

	return a, nil
}

func (a *ResourceInt32Attribute) ToSpec() resource.Attribute {
	return resource.Attribute{
		Name:  util.TerraformIdentifier(a.Name),
		Int32: &a.Int32Attribute,
	}
}

type DataSourceInt32Attribute struct {
	datasource.Int32Attribute

	Name string
}

func (a *DataSourceInt32Attribute) GetName() string {
	return a.Name
}

func (a *DataSourceInt32Attribute) Merge(mergeAttribute DataSourceAttribute) (DataSourceAttribute, error) {
	int32Attribute, ok := mergeAttribute.(*DataSourceInt32Attribute)
	// TODO: return error if types don't match?
	if ok && (a.Description == nil || *a.Description == "") {
		a.Description = int32Attribute.Description
	}

	return a, nil
}

func (a *DataSourceInt32Attribute) ApplyOverride(override explorer.Override) (DataSourceAttribute, error) {
	a.Description = &override.Description

	return a, nil
}

func (a *DataSourceInt32Attribute) ToSpec() datasource.Attribute {
	return datasource.Attribute{
		Name:  util.TerraformIdentifier(a.Name),
		Int32: &a.Int32Attribute,
	}
}

type ProviderInt32Attribute struct {
	provider.Int32Attribute

	Name string
}

func (a *ProviderInt32Attribute) ToSpec() provider.Attribute {
	return provider.Attribute{
		Name:  util.TerraformIdentifier(a.Name),
		Int32: &a.Int32Attribute,
	}
}
