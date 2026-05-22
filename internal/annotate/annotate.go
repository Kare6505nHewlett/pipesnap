// Package annotate provides chunk annotation by injecting key-value metadata
// fields into JSON-encoded snapshot chunks before they are written downstream.
package annotate

import (
	"encoding/json"
	"fmt"
)

// Annotation is a key-value pair to inject into a JSON chunk.
type Annotation struct {
	Key   string
	Value any
}

// Annotator holds a set of annotations to apply to each chunk.
type Annotator struct {
	annotations []Annotation
}

// New creates an Annotator with the provided annotations.
// Returns an error if any key is empty.
func New(annotations []Annotation) (*Annotator, error) {
	if len(annotations) == 0 {
		return nil, fmt.Errorf("annotate: at least one annotation is required")
	}
	for _, a := range annotations {
		if a.Key == "" {
			return nil, fmt.Errorf("annotate: annotation key must not be empty")
		}
	}
	return &Annotator{annotations: annotations}, nil
}

// Apply injects the annotator's key-value pairs into the JSON object
// represented by p. Returns the modified JSON or an error if p is not a
// valid JSON object.
func (a *Annotator) Apply(p []byte) ([]byte, error) {
	var obj map[string]any
	if err := json.Unmarshal(p, &obj); err != nil {
		return nil, fmt.Errorf("annotate: chunk is not a JSON object: %w", err)
	}
	for _, ann := range a.annotations {
		obj[ann.Key] = ann.Value
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("annotate: failed to marshal annotated chunk: %w", err)
	}
	return out, nil
}

// Len returns the number of annotations held by the Annotator.
func (a *Annotator) Len() int { return len(a.annotations) }
