// Package schema validates that incoming chunks conform to a declared
// structure (JSON field presence / type checks) before they are written
// to a snapshot.
package schema

import (
	"encoding/json"
	"fmt"
)

// FieldType represents the expected JSON type for a field.
type FieldType string

const (
	TypeString  FieldType = "string"
	TypeNumber  FieldType = "number"
	TypeBoolean FieldType = "boolean"
	TypeObject  FieldType = "object"
	TypeArray   FieldType = "array"
)

// Field describes a single expected field in a JSON chunk.
type Field struct {
	Name     string
	Type     FieldType
	Required bool
}

// Schema holds a set of field rules used to validate JSON chunks.
type Schema struct {
	fields []Field
}

// New creates a Schema from the provided field definitions.
func New(fields []Field) (*Schema, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("schema: at least one field is required")
	}
	return &Schema{fields: fields}, nil
}

// Validate checks whether p is valid JSON that satisfies the schema rules.
// It returns nil on success or a descriptive error on the first violation.
func (s *Schema) Validate(p []byte) error {
	var obj map[string]any
	if err := json.Unmarshal(p, &obj); err != nil {
		return fmt.Errorf("schema: invalid JSON: %w", err)
	}
	for _, f := range s.fields {
		val, ok := obj[f.Name]
		if !ok {
			if f.Required {
				return fmt.Errorf("schema: required field %q missing", f.Name)
			}
			continue
		}
		if err := checkType(f.Name, f.Type, val); err != nil {
			return err
		}
	}
	return nil
}

func checkType(name string, want FieldType, val any) error {
	switch want {
	case TypeString:
		if _, ok := val.(string); !ok {
			return fmt.Errorf("schema: field %q: expected string, got %T", name, val)
		}
	case TypeNumber:
		if _, ok := val.(float64); !ok {
			return fmt.Errorf("schema: field %q: expected number, got %T", name, val)
		}
	case TypeBoolean:
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("schema: field %q: expected boolean, got %T", name, val)
		}
	case TypeObject:
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("schema: field %q: expected object, got %T", name, val)
		}
	case TypeArray:
		if _, ok := val.([]any); !ok {
			return fmt.Errorf("schema: field %q: expected array, got %T", name, val)
		}
	}
	return nil
}
