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
	MapToIR(*slog.Logger) ([]Request, error)
}

type NcloudCommonRequestType struct {
	DetailedRequestType
	Method string `json:"method,omitempty"`
	Path   string `json:"path,omitempty"`
}

type RequestOperations struct {
	Create NcloudCommonRequestType    `json:"create,omitempty"`
	Read   NcloudCommonRequestType    `json:"read,omitempty"`
	Update []*NcloudCommonRequestType `json:"update,omitempty"`
	Delete NcloudCommonRequestType    `json:"delete,omitempty"`
}

type Request struct {
	RequestOperations
	Name string `json:"name,omitempty"`
}

type requestMapper struct {
	resources   map[string]explorer.Resource
	dataSources map[string]explorer.DataSource
	cfg         config.Config
}

type NcloudRequestBody struct {
	spec.RequestBody
	Required []*RequestParameterAttributes `json:"required,omitempty"`
	Optional []*RequestParameterAttributes `json:"optional,omitempty"`
}

type RequestParameters struct {
	Required []*RequestParameterAttributes `json:"required,omitempty"`
	Optional []*RequestParameterAttributes `json:"optional,omitempty"`
}

type RequestParameterAttributes struct {
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
}

type DetailedRequestType struct {
	spec.RequestType
	Parameters  *RequestParameters `json:"parameters,omitempty"`
	RequestBody *NcloudRequestBody `json:"request_body,omitempty"`
}

func NewRequestMapper(resources map[string]explorer.Resource, dataSources map[string]explorer.DataSource, cfg config.Config) RequestMapper {
	return requestMapper{
		resources:   resources,
		dataSources: dataSources,
		cfg:         cfg,
	}
}

func (m requestMapper) MapToIR(logger *slog.Logger) ([]Request, error) {
	requestSchemas := []Request{}

	resourceNames := util.SortedKeys(m.resources)
	dataSourceNames := util.SortedKeys(m.dataSources)

	for _, name := range resourceNames {
		explorerResource := m.resources[name]
		rLogger := logger.With("request", name)

		requestType, err := generateRequestType(rLogger, explorerResource, name, m.cfg)
		if err != nil {
			log.WarnLogOnError(rLogger, err, "skipping resource request type mapping")
			continue
		}

		requestSchemas = append(requestSchemas, requestType)
	}

	for _, name := range dataSourceNames {
		explorerDataSource := m.dataSources[name]
		dsLogger := logger.With("request", name)

		requestType, err := generateRequestDataSourceType(dsLogger, explorerDataSource, name, m.cfg)
		if err != nil {
			log.WarnLogOnError(dsLogger, err, "skipping data source request type mapping")
			continue
		}

		requestSchemas = append(requestSchemas, requestType)
	}

	return requestSchemas, nil
}

func generateRequestType(logger *slog.Logger, explorerResource explorer.Resource, name string, config config.Config) (Request, error) {
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
	createRequest := NcloudCommonRequestType{
		DetailedRequestType: DetailedRequestType{
			RequestType: spec.RequestType{
				Response: response,
			},
			Parameters:  extractParametersInfo(explorerResource.CreateOp),
			RequestBody: requestBody,
		},
		Method: config.Resources[name].Create.Method,
		Path:   config.Resources[name].Create.Path,
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
	readRequest := NcloudCommonRequestType{
		DetailedRequestType: DetailedRequestType{
			RequestType: spec.RequestType{
				Response: response,
			},
			Parameters:  extractParametersInfo(explorerResource.ReadOp),
			RequestBody: requestBody,
		},
		Method: config.Resources[name].Read.Method,
		Path:   config.Resources[name].Read.Path,
	}

	logger.Debug("searching for update operation parameters and request body")
	var updateRequest []*NcloudCommonRequestType
	for _, updateOp := range explorerResource.UpdateOps {
		requestBody, err = extractRequestBody(updateOp, schemaOpts)
		if err != nil {
			log.WarnLogOnError(logger, err, "skipping mapping of update operation rquest body")
		}
		response, err = extractResponse(updateOp, schemaOpts)
		if err != nil {
			log.WarnLogOnError(logger, err, "skipping mappin gof update operation response")
		}
		updateRequest = append(updateRequest, &NcloudCommonRequestType{
			DetailedRequestType: DetailedRequestType{
				RequestType: spec.RequestType{
					Response: response,
				},
				Parameters:  extractParametersInfo(updateOp),
				RequestBody: requestBody,
			},
			Method: config.Resources[name].Update[0].Method,
			Path:   config.Resources[name].Update[0].Path,
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
	deleteRequest := NcloudCommonRequestType{
		DetailedRequestType: DetailedRequestType{
			RequestType: spec.RequestType{
				Response: response,
			},
			Parameters:  extractParametersInfo(explorerResource.DeleteOp),
			RequestBody: requestBody,
		},
		Method: config.Resources[name].Delete.Method,
		Path:   config.Resources[name].Delete.Path,
	}

	return Request{
		Name: name,
		RequestOperations: RequestOperations{
			Create: createRequest,
			Read:   readRequest,
			Update: updateRequest,
			Delete: deleteRequest,
		},
	}, nil
}

func generateRequestDataSourceType(logger *slog.Logger, explorerDataSource explorer.DataSource, name string, config config.Config) (Request, error) {
	schemaOpts := oas.SchemaOpts{
		Ignores: explorerDataSource.SchemaOptions.Ignores,
	}

	logger.Debug("searching for read operation parameters and request body")
	requestBody, err := extractRequestBody(explorerDataSource.ReadOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mapping of read operation request body")
	}
	response, err := extractResponse(explorerDataSource.ReadOp, schemaOpts)
	if err != nil {
		log.WarnLogOnError(logger, err, "skipping mappin gof read operation response")
	}
	readRequest := NcloudCommonRequestType{
		DetailedRequestType: DetailedRequestType{
			RequestType: spec.RequestType{
				Response: response,
			},
			Parameters:  extractParametersInfo(explorerDataSource.ReadOp),
			RequestBody: requestBody,
		},
		Method: config.DataSources[name].Read.Method,
		Path:   config.DataSources[name].Read.Path,
	}

	return Request{
		Name: name,
		RequestOperations: RequestOperations{
			Read: readRequest,
		},
	}, nil
}

func extractParametersInfo(op *high.Operation) *RequestParameters {
	if op == nil || op.Parameters == nil {
		return nil
	}

	var requiredParams []*RequestParameterAttributes
	var optionalParams []*RequestParameterAttributes
	for _, param := range op.Parameters {
		p := &RequestParameterAttributes{
			Name:   param.Name,
			Type:   param.Schema.Schema().Type[0],
			Format: param.Schema.Schema().Format,
		}
		if param.Required != nil && *param.Required {
			requiredParams = append(requiredParams, p)
		} else {
			optionalParams = append(optionalParams, p)
		}
	}
	return &RequestParameters{
		Required: requiredParams,
		Optional: optionalParams,
	}
}

func extractRequestBody(op *high.Operation, schemaOpts oas.SchemaOpts) (*NcloudRequestBody, error) {
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

	var optionalRequestBody []*RequestParameterAttributes
	var requiredRequestBody []*RequestParameterAttributes

	// Get all property keys
	if requestSchema.Schema.Properties != nil {
		for pair := range orderedmap.Iterate(context.TODO(), requestSchema.Schema.Properties) {
			propKey := pair.Key()
			p := &RequestParameterAttributes{
				Name:   propKey,
				Type:   pair.Value().Schema().Type[0],
				Format: pair.Value().Schema().Format,
			}

			// If the property is not in Required slice, it's optional
			if !contains(requestSchema.Schema.Required, propKey) {
				optionalRequestBody = append(optionalRequestBody, p)
			} else {
				requiredRequestBody = append(requiredRequestBody, p)
			}
		}
	}

	return &NcloudRequestBody{
		RequestBody: spec.RequestBody{
			Name: name,
		},
		Required: requiredRequestBody,
		Optional: optionalRequestBody,
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

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}
