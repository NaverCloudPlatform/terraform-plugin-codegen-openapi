package sdk

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pb33f/libopenapi"

	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const (
	VERSION = "EXPERIMENTAL"
)

type ResponseDetails struct {
	RefreshLogic                       string
	Model                              string
	ConvertValueWithNull               string
	PossibleTypes                      string
	ConvertValueWithNullInEmptyArrCase string
}

func Generate(v3Doc *libopenapi.DocumentModel[v3high.Document]) error {

	// Generate directory
	// TODO - This Value should be gain from cmd flag. Need to be refactored when SDK Layer's architecture is fixed.
	err := os.MkdirAll(filepath.Join(MustAbs("./"), "ncloudsdk"), os.ModePerm)
	if err != nil {
		return err
	}

	// Write down version information
	err = os.MkdirAll(filepath.Join(MustAbs("./ncloudsdk"), ".swagger-codegen"), os.ModePerm)
	if err != nil {
		return err
	}

	v, err := os.Create(filepath.Join(MustAbs("./ncloudsdk"), ".swagger-codegen", "VERSION"))
	if err != nil {
		return err
	}

	v.WriteString(VERSION)

	c, err := os.Create(filepath.Join(MustAbs("./ncloudsdk"), "client.go"))
	if err != nil {
		return err
	}

	c.Write(WriteClient())

	// Get a specific operation to test
	pathItems := v3Doc.Model.Paths.PathItems.FromNewest()

	for key, item := range pathItems {

		if err := GenerateFile(item.Get, http.MethodGet, key); err != nil {
			return fmt.Errorf("error generating GET in key %s: %w", key, err)
		}

		if err := GenerateFile(item.Post, http.MethodPost, key); err != nil {
			return fmt.Errorf("error generating POST in key %s: %w", key, err)
		}

		if err := GenerateFile(item.Put, http.MethodPut, key); err != nil {
			return fmt.Errorf("error generating PUT in key %s: %w", key, err)
		}

		if err := GenerateFile(item.Delete, http.MethodDelete, key); err != nil {
			return fmt.Errorf("error generating DELETE in key %s: %w", key, err)
		}

		if err := GenerateFile(item.Patch, http.MethodPatch, key); err != nil {
			return fmt.Errorf("error generating PATCH in key %s: %w", key, err)
		}
	}

	return nil
}

func GenerateFile(op *v3high.Operation, method, key string) error {
	if op == nil {
		return nil
	}

	f, err := os.Create(filepath.Join(MustAbs("./"), "ncloudsdk", fmt.Sprintf("%s.go", method+"_"+PathToFilename(key))))
	if err != nil {
		return err
	}

	refreshDetails, err := GenerateStructs(op.Responses, method+getMethodName(key))
	if err != nil {
		return err
	}

	template := New(op, method, key, refreshDetails)

	_, err = f.Write(template.WriteTemplate())
	if err != nil {
		return err
	}

	_, err = f.Write(template.WriteRefresh())
	if err != nil {
		return err
	}

	return nil
}

// Generate terraform-spec type based struct with *v3high.Responses input
func GenerateStructs(responses *v3high.Responses, responseName string) (*ResponseDetails, error) {

	codes := []string{
		"200",
		"201",
	}

	for _, code := range codes {
		g, pre := responses.Codes.Get(code)
		if !pre {
			// Skip when expected status code does not exists.
			continue
		}

		c, pre := g.Content.OrderedMap.Get("application/json;charset=UTF-8")
		if !pre {
			// Skip when expected status code does not exists.
			continue
		}

		refreshLogic, model, convertValueWithNull, possibleTypes, convertValueWithNullInEmptyArrCase := Gen_ConvertOAStoTFTypes(c.Schema.Schema(), c.Schema.Schema().Type[0], c.Schema.Schema().Format, responseName)

		return &ResponseDetails{
			RefreshLogic:                       refreshLogic,
			Model:                              model,
			ConvertValueWithNull:               convertValueWithNull,
			PossibleTypes:                      possibleTypes,
			ConvertValueWithNullInEmptyArrCase: convertValueWithNullInEmptyArrCase,
		}, nil
	}

	return nil, fmt.Errorf("no suitable responses found")
}
