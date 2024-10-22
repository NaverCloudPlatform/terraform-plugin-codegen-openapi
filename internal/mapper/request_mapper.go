package mapper

import (
	"context"
	"log/slog"
	"strings"

	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/config"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/log"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/oas"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/util"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/spec"
	high "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

var _ RequestMapper = requestMapper{}

type RequestMapper interface {
	MapToIR(*slog.Logger) ([]spec.Request, error)
}

type requestMapper struct {
	resources map[string]explorer.Resource
	cfg       config.Config
}

func NewRequestMapper(resources map[string]explorer.Resource, cfg config.Config) RequestMapper {
	return requestMapper{
		resources: resources,
		cfg:       cfg,
	}
}

func (m requestMapper) MapToIR(logger *slog.Logger) ([]spec.Request, error) {
	requestSchemas := []spec.Request{}

	resourceNames := util.SortedKeys(m.resources)
	for _, name := range resourceNames {
		explorerResource := m.resources[name]
		rLogger := logger.With("request", name)

		requestType, err := generateRequestType(rLogger, explorerResource)
		if err != nil {
			log.WarnLogOnError(rLogger, err, "skipping resource request type mapping")
			continue
		}

		requestSchemas = append(requestSchemas, requestType)
	}

	return requestSchemas, nil
}

func generateRequestType(logger *slog.Logger, explorerResource explorer.Resource) (spec.Request, error) {
	schemaOpts := oas.SchemaOpts{
		Ignores: explorerResource.SchemaOptions.Ignores,
	}

	logger.Debug("searching for create operation parameters and request body")
	requestBody, err := extractRequestBody(explorerResource.CreateOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mapping of create operation rquest body")
	}
	response, err := extractResponse(explorerResource.CreateOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mappin gof create operation response")
	}
	createRequest := spec.RequestType{
		Parameters:  extractParameterNames(explorerResource.CreateOp),
		RequestBody: requestBody,
		Response:    response,
	}

	logger.Debug("searching for read operation parameters and request body")
	requestBody, err = extractRequestBody(explorerResource.ReadOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mapping of read operation request body")
	}
	response, err = extractResponse(explorerResource.ReadOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mappin gof read operation response")
	}
	readRequest := spec.RequestType{
		Parameters:  extractParameterNames(explorerResource.ReadOp),
		RequestBody: requestBody,
		Response:    response,
	}

	logger.Debug("searching for update operation parameters and request body")
	var updateRequest []*spec.RequestType
	for _, updateOp := range explorerResource.UpdateOps {
		requestBody, err = extractRequestBody(updateOp, schemaOpts)
		if err != nil {
			log.WarnLogOnError(logger, err, "skipping mapping of update operation rquest body")
		}
		response, err = extractResponse(updateOp, schemaOpts)
		if err != nil {
			log.WarnLogOnError(logger, err, "skipping mappin gof update operation response")
		}
		updateRequest = append(updateRequest, &spec.RequestType{
			Parameters:  extractParameterNames(updateOp),
			RequestBody: requestBody,
			Response:    response,
		})
	}

	logger.Debug("searching for delete operation parameters and request body")
	requestBody, err = extractRequestBody(explorerResource.DeleteOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mapping of delete operation rquest body")
	}
	response, err = extractResponse(explorerResource.DeleteOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mappin gof delete operation response")
	}
	deleteRequest := spec.RequestType{
		Parameters:  extractParameterNames(explorerResource.DeleteOp),
		RequestBody: requestBody,
		Response:    response,
	}

	return spec.Request{
		Create: createRequest,
		Read:   readRequest,
		Update: updateRequest,
		Delete: deleteRequest,
	}, nil
}

func extractParameterNames(op *high.Operation) []string {
	if op == nil || op.Parameters == nil {
		return nil
	}

	var paramNames []string
	for _, param := range op.Parameters {
		paramNames = append(paramNames, param.Name)
	}
	return paramNames
}

func extractRequestBody(op *high.Operation, schemaOpts oas.SchemaOpts) (*spec.RequestBody, error) {
	requestSchema, err := oas.BuildSchemaFromRequest(op, schemaOpts, oas.GlobalSchemaOpts{})
	if err != nil {
		if err == oas.ErrSchemaNotFound {
			return nil, nil
		}
		return nil, err
	}

	name := ""

	jsonMediaType, ok := op.RequestBody.Content.Get(util.OAS_mediatype_json)
	if ok && jsonMediaType.Schema != nil {
		if jsonMediaType.Schema.IsReference() {
			parts := strings.Split(jsonMediaType.Schema.GetReference(), "/")
			if len(parts) > 0 {
				name = parts[len(parts)-1]
			}
		}
	}

	return &spec.RequestBody{
		Name:     name,
		Required: requestSchema.Schema.Required,
	}, nil
}

func extractResponse(op *high.Operation, schemaOpts oas.SchemaOpts) (string, error) {
	_, err := oas.BuildSchemaFromResponse(op, schemaOpts, oas.GlobalSchemaOpts{})
	if err != nil {
		if err == oas.ErrSchemaNotFound {
			return "", nil
		}
		return "", err
	}

	sortedCodes := orderedmap.SortAlpha(op.Responses.Codes)
	for pair := range orderedmap.Iterate(context.TODO(), sortedCodes) {
		responseCode := pair.Value()
		content := responseCode.Content

		if jsonMediaType, ok := content.Get(util.OAS_mediatype_json); ok {
			if jsonMediaType.Schema != nil && jsonMediaType.Schema.IsReference() {
				parts := strings.Split(jsonMediaType.Schema.GetReference(), "/")
				if len(parts) > 0 {
					return parts[len(parts)-1], nil
				}
			}
		}
	}

	return "", nil
}
