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

	refreshLogic, model, newConvertValueWithNull, possibleTypes := GenerateStructs(op.Responses, method+getMethodName(key))
	if err != nil {
		return err
	}

	template := New(op, method, key, refreshLogic, model, newConvertValueWithNull, possibleTypes)

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
