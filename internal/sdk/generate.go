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
	basePath := MustAbs("./")

	// Generate directories
	err := createDirectories(basePath)
	if err != nil {
		return err
	}

	// Write down version information
	err = writeVersionInfo(basePath)
	if err != nil {
		return err
	}

	// Create client file
	err = createClientFile(basePath)
	if err != nil {
		return err
	}

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

	// To figure out intended response code
	codes := []string{
		"200",
		"201",
		"204",
	}

	for _, code := range codes {
		g, pre := responses.Codes.Get(code)
		if !pre {
			// Skip when expected status code does not exists.
			continue
		}

		// Case of http.StatusNoContent
		if code == "204" {
			return &ResponseDetails{
				RefreshLogic:                       "",
				Model:                              "",
				ConvertValueWithNull:               "",
				PossibleTypes:                      "",
				ConvertValueWithNullInEmptyArrCase: "",
			}, nil
		}

		c, pre := g.Content.OrderedMap.Get("application/json;charset=UTF-8")
		if !pre {
			// Skip when expected status code does not exists.
			// Case of empty content with 200
			return &ResponseDetails{
				RefreshLogic:                       "",
				Model:                              "",
				ConvertValueWithNull:               "",
				PossibleTypes:                      "",
				ConvertValueWithNullInEmptyArrCase: "",
			}, nil
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

// Helper function to create directories
func createDirectories(basePath string) error {
	dirs := []string{
		filepath.Join(basePath, "ncloudsdk"),
		filepath.Join(basePath, "ncloudsdk", ".codegen"),
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// Helper function to write version information
func writeVersionInfo(basePath string) error {
	v, err := os.Create(filepath.Join(basePath, "ncloudsdk", ".codegen", "VERSION"))
	if err != nil {
		return err
	}
	defer v.Close()

	_, err = v.WriteString(VERSION)
	return err
}

// Helper function to create client file
func createClientFile(basePath string) error {
	c, err := os.Create(filepath.Join(basePath, "ncloudsdk", "client.go"))
	if err != nil {
		return err
	}
	defer c.Close()

	_, err = c.Write(WriteClient())
	return err
}
