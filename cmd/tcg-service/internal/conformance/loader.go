// Package conformance implements conformance-only validation of inbound FOP
// payloads against the committed OAS 3.0 spec.
//
// Source: BRD FR-019 ("validate inbound payloads against committed FOP OAS
// 3.0 spec... Check JSON Schema, required fields, data types, value
// constraints... Quarantine malformed payloads without data loss"); BR-003
// ("Conformance validation only; no ETL on ingress... never reshapes,
// filters, or enriches payload").
package conformance

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// LoadSpec loads the committed FOP OAS document. It deliberately does not
// run kin-openapi's strict doc.Validate() self-check: the real, TCG-committed
// v0.4.0 spec has at least one schema (Promotion) using an "examples" sibling
// field, which is valid OAS 3.1 / JSON Schema 2020-12 but not strict OAS
// 3.0.x — a characteristic of the upstream contract, not something this
// service invents or silently rewrites. LoadFromFile still fully parses the
// document and resolves $refs, which is all ValidateComponentSchema and
// ValidateOperationResponse need.
func LoadSpec(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("conformance: load spec %s: %w", path, err)
	}
	return doc, nil
}
