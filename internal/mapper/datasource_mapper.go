// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapper

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/config"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/log"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/oas"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/util"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/datasource"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/schema"
)

var _ DataSourceMapper = dataSourceMapper{}

type DataSourceMapper interface {
	MapToIR(*slog.Logger) ([]DetailDataSourceInfo, error)
}

type DetailDataSourceInfo struct {
	datasource.DataSource
	CRUDParameters      CRUDParameters `json:"crud_parameters"`
	RefreshObjectName   string         `json:"refresh_object_name"`
	ImportStateOverride string         `json:"import_state_override"`
	Id                  string         `json:"id"`
}

type dataSourceMapper struct {
	dataSources map[string]explorer.DataSource
	//nolint:unused // Might be useful later!
	cfg config.Config
}

func NewDataSourceMapper(dataSources map[string]explorer.DataSource, cfg config.Config) DataSourceMapper {
	return dataSourceMapper{
		dataSources: dataSources,
		cfg:         cfg,
	}
}

func (m dataSourceMapper) MapToIR(logger *slog.Logger) ([]DetailDataSourceInfo, error) {
	dataSourceSchemas := []DetailDataSourceInfo{}

	// Guarantee the order of processing
	dataSourceNames := util.SortedKeys(m.dataSources)
	for _, name := range dataSourceNames {
		dataSource := m.dataSources[name]
		dLogger := logger.With("data_source", name)
		id := m.cfg.DataSources[name].Id
		importStateOverride := m.cfg.DataSources[name].ImportStateOverride

		requestMapper := NewDataSourceRequestMapper(dataSource, name, m.cfg)
		datasourceRequestIR, err := requestMapper.MapToIR(logger)
		if err != nil {
			log.WarnLogOnError(dLogger, err, "skipping resource request mapping")
			continue
		}

		var refreshObjectName string
		t := m.cfg.DataSources[name].RefreshObjectName

		if t != "" {
			refreshObjectName = t
		} else {
			g, pre := m.dataSources[name].ReadOp.Responses.Codes.Get("200")
			if !pre {
				log.WarnLogOnError(dLogger, errors.New("error in parsing openapi: "), "couldn't find GET response with status code 200")
			}

			c, pre := g.Content.OrderedMap.Get("application/json;charset=UTF-8")
			if !pre {
				log.WarnLogOnError(dLogger, errors.New("error in parsing openapi: "), "couldn't find valid content with application/json;charset=UTF-8 header")
			}

			s := strings.Split(c.Schema.GetReference(), "/")
			refreshObjectName = s[len(s)-1]
		}

		schema, err := generateDataSourceSchema(dLogger, name, dataSource)
		if err != nil {
			log.WarnLogOnError(dLogger, err, "skipping data source schema mapping")
			continue
		}

		dataSourceSchemas = append(dataSourceSchemas, DetailDataSourceInfo{
			DataSource: datasource.DataSource{
				Name:   name,
				Schema: schema,
			},
			CRUDParameters:      datasourceRequestIR,
			RefreshObjectName:   refreshObjectName,
			ImportStateOverride: importStateOverride,
			Id:                  id,
		})
	}

	return dataSourceSchemas, nil
}

func generateDataSourceSchema(logger *slog.Logger, name string, dataSource explorer.DataSource) (*datasource.Schema, error) {
	dataSourceSchema := &datasource.Schema{
		Attributes: []datasource.Attribute{},
	}

	// ********************
	// READ Response Body (required)
	// ********************
	logger.Debug("searching for read operation response body")

	schemaOpts := oas.SchemaOpts{
		Ignores: dataSource.SchemaOptions.Ignores,
	}
	globalSchemaOpts := oas.GlobalSchemaOpts{
		OverrideComputability: schema.Computed,
	}
	readResponseSchema, err := oas.BuildSchemaFromResponse(dataSource.ReadOp, schemaOpts, globalSchemaOpts)
	if err != nil {
		return nil, err
	}

	readResponseAttributes := attrmapper.DataSourceAttributes{}
	if readResponseSchema.Type == util.OAS_type_array {
		logger.Debug(fmt.Sprintf("response body is an array, building '%s' set attribute", name))

		// API's generally don't guarantee ordering of results for collection/query responses, default mapping to set
		readResponseSchema.Format = util.TF_format_set

		collectionAttribute, schemaErr := readResponseSchema.BuildDataSourceAttribute(name, schema.Computed)
		if schemaErr != nil {
			return nil, schemaErr
		}

		readResponseAttributes = append(readResponseAttributes, collectionAttribute)
	} else {
		attributes, schemaErr := readResponseSchema.BuildDataSourceAttributes()
		if schemaErr != nil {
			return nil, schemaErr
		}

		readResponseAttributes = attributes
	}

	// ****************
	// READ Parameters (optional)
	// ****************
	readParameterAttributes := attrmapper.DataSourceAttributes{}
	for _, param := range dataSource.ReadOpParameters() {
		if param.In != util.OAS_param_path && param.In != util.OAS_param_query {
			continue
		}

		pLogger := logger.With("param", param.Name)
		schemaOpts := oas.SchemaOpts{
			Ignores:             dataSource.SchemaOptions.Ignores,
			OverrideDescription: param.Description,
		}

		s, schemaErr := oas.BuildSchema(param.Schema, schemaOpts, oas.GlobalSchemaOpts{})
		if schemaErr != nil {
			log.WarnLogOnError(pLogger, schemaErr, "skipping mapping of read operation parameter")
			continue
		}

		computability := schema.ComputedOptional
		if param.Required != nil && *param.Required {
			computability = schema.Required
		}

		// Check for any aliases and replace the paramater name if found
		paramName := param.Name
		if aliasedName, ok := dataSource.SchemaOptions.AttributeOptions.Aliases[param.Name]; ok {
			pLogger = pLogger.With("param_alias", aliasedName)
			paramName = aliasedName
		}

		if s.IsPropertyIgnored(paramName) {
			continue
		}

		parameterAttribute, schemaErr := s.BuildDataSourceAttribute(paramName, computability)
		if schemaErr != nil {
			log.WarnLogOnError(pLogger, schemaErr, "skipping mapping of read operation parameter")
			continue
		}

		readParameterAttributes = append(readParameterAttributes, parameterAttribute)
	}

	// TODO: currently, no errors can be returned from merging, but in the future we should consider raising errors/warnings for unexpected scenarios, like type mismatches between attribute schemas
	dataSourceAttributes, _ := readParameterAttributes.Merge(readResponseAttributes)

	// TODO: handle error for overrides
	dataSourceAttributes, _ = dataSourceAttributes.ApplyOverrides(dataSource.SchemaOptions.AttributeOptions.Overrides)

	dataSourceSchema.Attributes = dataSourceAttributes.ToSpec()
	return dataSourceSchema, nil
}
