package sdk

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/pb33f/libopenapi"
)

func TestTemplateGen_basic(t *testing.T) {

	path := MustAbs("./apigw_v1.json")

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
	targetPath := "/api-keys/{api-key-id}/unsubscribe"
	pathItems := v3Doc.Model.Paths.PathItems.FromNewest()

	for key, item := range pathItems {

		fmt.Println(key)
		if key != targetPath {
			continue
		}

		fmt.Println(item.Post)

		t := New(item.Post, "POST", key)

		b := t.WriteTemplate()

		fmt.Println(string(b))
	}
}

func MustAbs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Error getting absolute path for %s: %v", path, err)
	}
	return absPath
}
