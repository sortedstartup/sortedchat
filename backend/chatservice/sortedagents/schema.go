package sortedagents

import (
	"encoding/json"
	"reflect"
	"strings"
)

// JSONSchema represents a subset of JSON Schema draft 7 for OpenAI functions
type JSONSchema struct {
	Type                 string                `json:"type"`
	Description          string                `json:"description,omitempty"`
	Properties           map[string]JSONSchema `json:"properties,omitempty"`
	Required             []string              `json:"required,omitempty"`
	Items                *JSONSchema           `json:"items,omitempty"`
	Enum                 []any                 `json:"enum,omitempty"`
	AdditionalProperties *bool                 `json:"additionalProperties,omitempty"`
}

// MarshalJSON ensures that an empty properties map is rendered as {} for objects,
// which is required by OpenAI even for functions with no parameters.
func (s JSONSchema) MarshalJSON() ([]byte, error) {
	type Alias JSONSchema
	aux := struct {
		Alias
		// Use pointer to map to strictly control omission vs empty object
		Properties *map[string]JSONSchema `json:"properties,omitempty"`
	}{
		Alias: Alias(s),
	}

	if s.Type == "object" {
		if s.Properties == nil {
			// Point to an empty map to force "{}"
			empty := make(map[string]JSONSchema)
			aux.Properties = &empty
		} else {
			// Point to the existing map
			aux.Properties = &s.Properties
		}
	} else {
		// For leaf nodes, nil pointer triggers omitempty
		aux.Properties = nil
	}

	return json.Marshal(aux)
}

// GenerateSchema creates a JSON schema parameters map from a struct type
// Returns the schema and a boolean indicating if the schema is compatible with Strict Mode
func GenerateSchema[T any]() (*JSONSchema, bool) {
	var zero T
	return GenerateSchemaReflect(reflect.TypeOf(zero))
}

// GenerateSchemaReflect generates a JSON schema from a reflect.Type
func GenerateSchemaReflect(t reflect.Type) (*JSONSchema, bool) {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// If it's not a struct, return a basic object (non-strict)
	if t.Kind() != reflect.Struct {
		return &JSONSchema{
			Type:       "object",
			Properties: make(map[string]JSONSchema),
		}, false
	}

	properties := make(map[string]JSONSchema)
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get JSON name from tag
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue
		}

		name := jsonTag
		if commaIdx := strings.Index(name, ","); commaIdx != -1 {
			name = name[:commaIdx]
		}
		if name == "" {
			name = field.Name
		}

		// Get description and enum from tags
		description := field.Tag.Get("description")
		enumTag := field.Tag.Get("enum")

		// Map Go types to JSON types
		schema := JSONSchema{
			Description: description,
		}

		// Handle Enum
		if enumTag != "" {
			enums := strings.Split(enumTag, ",")
			schema.Enum = make([]any, len(enums))
			for j, e := range enums {
				schema.Enum[j] = strings.TrimSpace(e)
			}
		}

		switch field.Type.Kind() {
		case reflect.String:
			schema.Type = "string"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			schema.Type = "number"
		case reflect.Bool:
			schema.Type = "boolean"
		case reflect.Slice, reflect.Array:
			schema.Type = "array"
			schema.Items = &JSONSchema{Type: "string"}
		case reflect.Map, reflect.Struct:
			schema.Type = "object"
		default:
			schema.Type = "string"
		}

		properties[name] = schema

		// In Strict Mode, all properties MUST be required.
		required = append(required, name)
	}

	schema := &JSONSchema{
		Type:       "object",
		Properties: properties,
	}

	// OpenAI Strict Mode limitation: Objects must have at least one property.
	isStrictCompatible := len(properties) > 0

	if isStrictCompatible {
		f := false
		schema.AdditionalProperties = &f
	}

	if len(required) > 0 {
		schema.Required = required
	}

	return schema, isStrictCompatible
}
