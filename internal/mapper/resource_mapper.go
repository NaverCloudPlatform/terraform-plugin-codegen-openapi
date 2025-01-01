// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mapper

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/config"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/log"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/attrmapper"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/oas"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/util"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/resource"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/schema"
)

var _ ResourceMapper = resourceMapper{}

type ResourceMapper interface {
	MapToIR(*slog.Logger) ([]ResourceWithRefreshObjectName, error)
}

type ResourceWithRefreshObjectName struct {
	resource.Resource
	RefreshObjectName   string `json:"refresh_object_name"`
	ImportStateOverride string `json:"import_state_override"`
	Id                  string `json:"id"`
}

type resourceMapper struct {
	resources map[string]explorer.Resource
	//nolint:unused // Might be useful later!
	cfg config.Config
}

func NewResourceMapper(resources map[string]explorer.Resource, cfg config.Config) ResourceMapper {
	return resourceMapper{
		resources: resources,
		cfg:       cfg,
	}
}

func (m resourceMapper) MapToIR(logger *slog.Logger) ([]ResourceWithRefreshObjectName, error) {
	resourceSchemas := []ResourceWithRefreshObjectName{}

	// Guarantee the order of processing
	resourceNames := util.SortedKeys(m.resources)
	for _, name := range resourceNames {
		explorerResource := m.resources[name]
		rLogger := logger.With("resource", name)
		id := m.cfg.Resources[name].Id
		importStateOverride := m.cfg.Resources[name].ImportStateOverride

		var refreshObjectName string
		t := m.cfg.Resources[name].RefreshObjectName

		if t != "" {
			refreshObjectName = t
		} else {
			g, pre := m.resources[name].ReadOp.Responses.Codes.Get("200")
			if !pre {
				log.WarnLogOnError(rLogger, errors.New("error in parsing openapi: "), "couldn't find GET response with status code 200")
			}

			c, pre := g.Content.OrderedMap.Get("application/json;charset=UTF-8")
			if !pre {
				log.WarnLogOnError(rLogger, errors.New("error in parsing openapi: "), "couldn't find valid content with application/json;charset=UTF-8 header")
			}

			s := strings.Split(c.Schema.GetReference(), "/")
			refreshObjectName = s[len(s)-1]
		}

		schema, err := generateResourceSchema(rLogger, explorerResource)
		if err != nil {
			log.WarnLogOnError(rLogger, err, "skipping resource schema mapping")
			continue
		}

		resourceSchemas = append(resourceSchemas, ResourceWithRefreshObjectName{
			Resource: resource.Resource{
				Name:   name,
				Schema: schema,
			},
			RefreshObjectName:   refreshObjectName,
			ImportStateOverride: importStateOverride,
			Id:                  id,
		})
	}

	return resourceSchemas, nil
}

func generateResourceSchema(logger *slog.Logger, explorerResource explorer.Resource) (*resource.Schema, error) {
	resourceSchema := &resource.Schema{
		Attributes: []resource.Attribute{},
	}

	// ********************
	// Create Request Body (required)
	// ********************
	logger.Debug("searching for create operation request body")

	schemaOpts := oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	createRequestSchema, err := oas.BuildSchemaFromRequest(explorerResource.CreateOp, schemaOpts, oas.GlobalSchemaOpts{})
	if err != nil {
		return nil, err
	}
	createRequestAttributes, schemaErr := createRequestSchema.BuildResourceAttributes()
	if schemaErr != nil {
		return nil, schemaErr
	}

	// *********************
	// Create Response Body (optional)
	// *********************
	logger.Debug("searching for create operation response body")

	createResponseAttributes := attrmapper.ResourceAttributes{}
	schemaOpts = oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	globalSchemaOpts := oas.GlobalSchemaOpts{
		OverrideComputability: schema.Computed,
	}
	createResponseSchema, err := oas.BuildSchemaFromResponse(explorerResource.CreateOp, schemaOpts, globalSchemaOpts)
	if err != nil {
		if errors.Is(err, oas.ErrSchemaNotFound) {
			// Demote log to INFO if there was no schema found
			logger.Info("skipping mapping of create operation response body", "err", err)
		} else {
			logger.Warn("skipping mapping of create operation response body", "err", err)
		}
	} else {
		createResponseAttributes, schemaErr = createResponseSchema.BuildResourceAttributes()
		if schemaErr != nil {
			log.WarnLogOnError(logger, schemaErr, "skipping mapping of create operation response body")
		}
	}

	// *******************
	// READ Response Body (optional)
	// *******************
	logger.Debug("searching for read operation response body")

	readResponseAttributes := attrmapper.ResourceAttributes{}

	schemaOpts = oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	globalSchemaOpts = oas.GlobalSchemaOpts{
		OverrideComputability: schema.Computed,
	}
	readResponseSchema, err := oas.BuildSchemaFromResponse(explorerResource.ReadOp, schemaOpts, globalSchemaOpts)
	if err != nil {
		if errors.Is(err, oas.ErrSchemaNotFound) {
			// Demote log to INFO if there was no schema found
			logger.Info("skipping mapping of read operation response body", "err", err)
		} else {
			logger.Warn("skipping mapping of read operation response body", "err", err)
		}
	} else {
		readResponseAttributes, schemaErr = readResponseSchema.BuildResourceAttributes()
		if schemaErr != nil {
			log.WarnLogOnError(logger, schemaErr, "skipping mapping of read operation response body")
		}
	}

	// ****************
	// READ Parameters (optional)
	// ****************
	readParameterAttributes := attrmapper.ResourceAttributes{}
	for _, param := range explorerResource.ReadOpParameters() {
		if param.In != util.OAS_param_path && param.In != util.OAS_param_query {
			continue
		}

		pLogger := logger.With("param", param.Name)
		schemaOpts := oas.SchemaOpts{
			Ignores:             explorerResource.SchemaOptions.Ignores,
			OverrideDescription: param.Description,
		}
		globalSchemaOpts := oas.GlobalSchemaOpts{OverrideComputability: schema.ComputedOptional}

		s, schemaErr := oas.BuildSchema(param.Schema, schemaOpts, globalSchemaOpts)
		if schemaErr != nil {
			log.WarnLogOnError(pLogger, schemaErr, "skipping mapping of read operation parameter")
			continue
		}

		// Check for any aliases and replace the paramater name if found
		paramName := param.Name
		if aliasedName, ok := explorerResource.SchemaOptions.AttributeOptions.Aliases[param.Name]; ok {
			pLogger = pLogger.With("param_alias", aliasedName)
			paramName = aliasedName
		}

		if s.IsPropertyIgnored(paramName) {
			continue
		}

		parameterAttribute, schemaErr := s.BuildResourceAttribute(paramName, schema.ComputedOptional)
		if schemaErr != nil {
			log.WarnLogOnError(pLogger, schemaErr, "skipping mapping of read operation parameter")
			continue
		}

		readParameterAttributes = append(readParameterAttributes, parameterAttribute)
	}

	// ********************
	// Update Request Body (optional)
	// ********************
	logger.Debug("searching for update operation request body")

	schemaOpts = oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}
	updateRequestSchema, err := oas.BuildSchemaFromRequest(explorerResource.UpdateOps[0], schemaOpts, oas.GlobalSchemaOpts{})
	if err != nil {
		return nil, err
	}
	updateRequestAttributes, schemaErr := updateRequestSchema.BuildResourceAttributes()
	if schemaErr != nil {
		return nil, schemaErr
	}

	// TODO: currently, no errors can be returned from merging, but in the future we should consider raising errors/warnings for unexpected scenarios, like type mismatches between attribute schemas
	resourceAttributes, _ := createRequestAttributes.Merge(createResponseAttributes, readResponseAttributes, readParameterAttributes, updateRequestAttributes)

	// TODO: handle error for overrides
	resourceAttributes, _ = resourceAttributes.ApplyOverrides(explorerResource.SchemaOptions.AttributeOptions.Overrides)

	resourceSchema.Attributes = resourceAttributes.ToSpec()
	return resourceSchema, nil
}
