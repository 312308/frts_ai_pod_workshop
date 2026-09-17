package conformance

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

// Result is the outcome of validating one payload against one OAS schema.
type Result struct {
	Valid  bool
	Errors []string
}

// Validator validates entity payloads against schemas embedded in the
// operation responses of a loaded FOP OAS document (AC-006a–d).
type Validator struct {
	doc *openapi3.T
}

// NewValidator loads specPath and returns a Validator over it.
func NewValidator(specPath string) (*Validator, error) {
	doc, err := LoadSpec(specPath)
	if err != nil {
		return nil, err
	}
	return &Validator{doc: doc}, nil
}

// ValidateOperationResponse validates payload against the response schema
// declared for method+path+statusCode in the loaded spec (e.g. GET
// /api/v1/wholesale-products, 200). This is how master-data list payloads
// (which the spec inlines per-operation rather than under
// components.schemas) get validated.
func (v *Validator) ValidateOperationResponse(path, method, statusCode string, payload []byte) (*Result, error) {
	item := v.doc.Paths.Find(path)
	if item == nil {
		return nil, fmt.Errorf("conformance: no path %q in spec", path)
	}
	op := item.GetOperation(method)
	if op == nil {
		return nil, fmt.Errorf("conformance: no operation %s %q in spec", method, path)
	}
	respRef := op.Responses.Value(statusCode)
	if respRef == nil || respRef.Value == nil {
		return nil, fmt.Errorf("conformance: no %q response for %s %q in spec", statusCode, method, path)
	}
	mediaType := respRef.Value.Content.Get("application/json")
	if mediaType == nil || mediaType.Schema == nil || mediaType.Schema.Value == nil {
		return nil, fmt.Errorf("conformance: no application/json schema for %s %q %s response", method, path, statusCode)
	}
	return validateAgainstSchema(mediaType.Schema.Value, payload)
}

// ValidateComponentSchema validates payload against a named
// components.schemas entry (e.g. "SalesOrderRequest").
func (v *Validator) ValidateComponentSchema(name string, payload []byte) (*Result, error) {
	schemaRef, ok := v.doc.Components.Schemas[name]
	if !ok || schemaRef.Value == nil {
		return nil, fmt.Errorf("conformance: no component schema %q in spec", name)
	}
	return validateAgainstSchema(schemaRef.Value, payload)
}

func validateAgainstSchema(schema *openapi3.Schema, payload []byte) (*Result, error) {
	var data interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		// Not valid JSON at all — this is itself a conformance failure
		// (AC-006c: data types checked), not a caller error.
		return &Result{Valid: false, Errors: []string{fmt.Sprintf("payload is not valid JSON: %v", err)}}, nil
	}

	// MultiErrors: report every violated constraint, not just the first —
	// AC-006b–d ("required fields... data types... value constraints
	// checked") implies the caller should see the full picture of why a
	// payload was quarantined, not have to fix-and-resubmit one field at a
	// time.
	if err := schema.VisitJSON(data, openapi3.MultiErrors()); err != nil {
		return &Result{Valid: false, Errors: flattenSchemaError(err)}, nil
	}
	return &Result{Valid: true}, nil
}

// flattenSchemaError turns kin-openapi's (possibly nested) schema error into
// a flat, human-readable list — one entry per violated constraint, so a
// single quarantined record can show every reason it failed (AC-006b–d),
// not just the first.
func flattenSchemaError(err error) []string {
	if me, ok := err.(openapi3.MultiError); ok {
		var out []string
		for _, e := range me {
			out = append(out, flattenSchemaError(e)...)
		}
		return out
	}
	if se, ok := err.(*openapi3.SchemaError); ok {
		return []string{se.Error()}
	}
	return []string{err.Error()}
}
