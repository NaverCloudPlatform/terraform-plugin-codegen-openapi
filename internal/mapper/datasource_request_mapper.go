package mapper

import (
	"log/slog"

	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/config"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/explorer"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/log"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-openapi/internal/mapper/oas"
	"github.com/NaverCloudPlatform/terraform-plugin-codegen-spec/spec"
)

var _ RequestMapper = datasourceRequestMapper{}

type datasourceRequestMapper struct {
	datasource explorer.DataSource
	name       string
	cfg        config.Config
}

func NewDataSourceRequestMapper(datasource explorer.DataSource, name string, cfg config.Config) datasourceRequestMapper {
	return datasourceRequestMapper{
		datasource: datasource,
		name:       name,
		cfg:        cfg,
	}
}

func (m datasourceRequestMapper) MapToIR(logger *slog.Logger) (CRUDParameters, error) {
	rLogger := logger.With("request", m.name)

	requestType, err := generateRequestDataSourceType(rLogger, m.datasource, m.name, m.cfg)
	if err != nil {
		log.WarnLogOnError(rLogger, err, "skipping resource request type mapping")
	}

	return requestType, nil
}

func generateRequestDataSourceType(logger *slog.Logger, explorerDataSource explorer.DataSource, name string, config config.Config) (CRUDParameters, error) {
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

	return CRUDParameters{
		Read: readRequest,
	}, nil
}
