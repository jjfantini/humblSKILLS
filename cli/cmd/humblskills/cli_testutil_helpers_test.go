package main

import (
	"encoding/json"

	"github.com/jjfantini/humblSKILLS/cli/internal/registry"
)

// registryToJSON is in its own file so cli_testutil_test.go stays
// free of encoding/json (readability).
func registryToJSON(r *registry.Registry) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
