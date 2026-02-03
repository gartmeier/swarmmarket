package task

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// JSONSchemaValidator validates JSON data against JSON Schema.
type JSONSchemaValidator struct{}

// NewJSONSchemaValidator creates a new JSON Schema validator.
func NewJSONSchemaValidator() *JSONSchemaValidator {
	return &JSONSchemaValidator{}
}

// Validate validates data against a JSON Schema.
func (v *JSONSchemaValidator) Validate(schema, data json.RawMessage) error {
	if len(schema) == 0 {
		// No schema defined, allow any data
		return nil
	}

	// Create a new compiler for each validation to avoid caching issues
	compiler := jsonschema.NewCompiler()
	compiler.Draft = jsonschema.Draft2020

	// Add schema resource
	schemaURL := "schema://input"
	if err := compiler.AddResource(schemaURL, bytes.NewReader(schema)); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Compile schema
	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		return fmt.Errorf("failed to compile schema: %w", err)
	}

	// Parse data
	var dataValue any
	if err := json.Unmarshal(data, &dataValue); err != nil {
		return fmt.Errorf("invalid JSON data: %w", err)
	}

	// Validate
	if err := compiled.Validate(dataValue); err != nil {
		// Format validation errors nicely
		if validationErr, ok := err.(*jsonschema.ValidationError); ok {
			return fmt.Errorf("validation failed: %s", formatValidationError(validationErr))
		}
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}

// formatValidationError extracts a human-readable message from validation errors.
func formatValidationError(err *jsonschema.ValidationError) string {
	if len(err.Causes) > 0 {
		// Get the first cause for a more specific error
		return formatValidationError(err.Causes[0])
	}

	msg := err.Message
	if err.InstanceLocation != "" {
		msg = fmt.Sprintf("%s: %s", err.InstanceLocation, msg)
	}
	return msg
}
