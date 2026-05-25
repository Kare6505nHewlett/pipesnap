// Package pivot transforms JSON chunks by hoisting a chosen field's value to
// the top-level key of the output object.
//
// Given a chunk like:
//
//	{"kind":"metric","name":"cpu","value":0.87}
//
// pivoting on "kind" produces:
//
//	{"metric":{"kind":"metric","name":"cpu","value":0.87}}
//
// This is useful when fan-out routing or downstream consumers need to dispatch
// on a single top-level key without inspecting the full payload.
//
// Chunks that are not valid JSON objects, or that do not contain the pivot
// field as a string value, are silently dropped and counted via Dropped().
package pivot
