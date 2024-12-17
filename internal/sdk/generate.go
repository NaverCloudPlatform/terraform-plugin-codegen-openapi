package sdk

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pb33f/libopenapi"

	v3high "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const (
	VERSION = "v0.0.1-beta"
)

func Generate(v3Doc *libopenapi.DocumentModel[v3high.Document]) {

	// Generate directory
	err := os.MkdirAll(filepath.Join(MustAbs("./"), "ncloudsdk"), os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create dir: %v", err)
	}

	// Write down version information
	err = os.MkdirAll(filepath.Join(MustAbs("./ncloudsdk"), ".swagger-codegen"), os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create dir: %v", err)
	}

	v, err := os.Create(filepath.Join(MustAbs("./ncloudsdk"), ".swagger-codegen", "VERSION"))
	if err != nil {
		log.Fatalf("failed to create test file: %v", err)
	}

	v.WriteString(VERSION)

	c, err := os.Create(filepath.Join(MustAbs("./ncloudsdk"), "client.go"))
	if err != nil {
		log.Fatalf("failed to create sclient file: %v", err)
	}

	c.Write(WriteClient())

	// Get a specific operation to test
	pathItems := v3Doc.Model.Paths.PathItems.FromNewest()

	for key, item := range pathItems {

		if err := GenerateFile(item.Get, "GET", key); err != nil {
			log.Fatalf("Error with generating get method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Post, "Post", key); err != nil {
			log.Fatalf("Error with generating post method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Put, "PUT", key); err != nil {
			log.Fatalf("Error with generating put method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Delete, "DELETE", key); err != nil {
			log.Fatalf("Error with generating delete method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Patch, "PATCH", key); err != nil {
			log.Fatalf("Error with generating patch method with key %s: %v", key, err)
		}
	}
}

func GenerateFile(op *v3high.Operation, method, key string) error {
	if op == nil {
		return nil
	}

	f, err := os.Create(filepath.Join(MustAbs("./"), "ncloudsdk", fmt.Sprintf("%s.go", PathToFilename(key))))
	if err != nil {
		return err
	}

	template := New(op, method, key)

	_, err = f.Write(template.WriteTemplate())
	if err != nil {
		return err
	}

	return nil
}
