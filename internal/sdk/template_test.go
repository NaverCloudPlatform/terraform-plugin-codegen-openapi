package sdk

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestTemplateGen_basic(t *testing.T) {

	path := MustAbs("./apigw_v1.json")

	// Generate directory
	err := os.MkdirAll(filepath.Join(MustAbs("./"), "ncloudsdk"), os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	// Write down version information
	err = os.MkdirAll(filepath.Join(MustAbs("./ncloudsdk"), ".swagger-codegen"), os.ModePerm)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	v, err := os.Create(filepath.Join(MustAbs("./ncloudsdk"), ".swagger-codegen", "VERSION"))
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	v.WriteString(VERSION)

	c, err := os.Create(filepath.Join(MustAbs("./ncloudsdk"), "client.go"))
	if err != nil {
		t.Fatalf("failed to create sclient file: %v", err)
	}

	c.Write(WriteClient())

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	// Parse OpenAPI document
	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		t.Fatalf("failed to parse OpenAPI doc: %v", err)
	}

	// Convert to v3 model
	v3Doc, errors := doc.BuildV3Model()
	if len(errors) > 0 {
		t.Fatalf("errors building V3 model: %v", errors)
	}

	// Get a specific operation to test
	pathItems := v3Doc.Model.Paths.PathItems.FromNewest()

	for key, item := range pathItems {

		if err := GenerateFile(item.Get, "GET", key); err != nil {
			t.Fatalf("Error with generating get method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Post, "Post", key); err != nil {
			t.Fatalf("Error with generating post method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Put, "PUT", key); err != nil {
			t.Fatalf("Error with generating put method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Delete, "DELETE", key); err != nil {
			t.Fatalf("Error with generating delete method with key %s: %v", key, err)
		}

		if err := GenerateFile(item.Patch, "PATCH", key); err != nil {
			t.Fatalf("Error with generating patch method with key %s: %v", key, err)
		}
	}
}
