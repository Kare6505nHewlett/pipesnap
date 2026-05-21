// Package schema provides lightweight JSON-schema validation for snapshot
// chunks in the pipesnap pipeline.
//
// # Overview
//
// A Schema is built from a slice of Field descriptors. Each Field names a
// JSON key, declares its expected type (string, number, boolean, object, or
// array), and marks it as required or optional.
//
// # Usage
//
//	s, err := schema.New([]schema.Field{
//		{Name: "level",   Type: schema.TypeString,  Required: true},
//		{Name: "ts",      Type: schema.TypeNumber,  Required: true},
//		{Name: "payload", Type: schema.TypeObject,  Required: false},
//	})
//
//	// Validate a raw chunk:
//	if err := s.Validate(chunk); err != nil {
//		log.Println("invalid chunk:", err)
//	}
//
//	// Or wrap any io.Writer to drop non-conforming chunks automatically:
//	w := schema.NewWriter(dest, s)
//
// Chunks that fail validation are silently dropped; the Writer still returns
// the original byte length so it remains compatible with io.Writer callers.
package schema
