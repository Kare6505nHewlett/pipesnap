// Package routing provides a content-based chunk router that dispatches
// incoming chunks to one of several named destinations based on filter matches.
package routing

import (
	"errors"
	"fmt"
	"io"
)

// Rule pairs a filter function with a destination writer.
type Rule struct {
	// Name is a human-readable label used in error messages.
	Name string
	// Match returns true when the chunk should be sent to Dst.
	Match func([]byte) bool
	// Dst receives chunks that satisfy Match.
	Dst io.Writer
}

// Router dispatches each chunk to the first matching Rule.
// If no rule matches and Fallback is non-nil the chunk is written there;
// otherwise the chunk is silently dropped.
type Router struct {
	rules    []Rule
	fallback io.Writer
}

// New creates a Router from the supplied rules.
// At least one rule must be provided.
func New(fallback io.Writer, rules ...Rule) (*Router, error) {
	if len(rules) == 0 {
		return nil, errors.New("routing: at least one rule is required")
	}
	for i, r := range rules {
		if r.Match == nil {
			return nil, fmt.Errorf("routing: rule[%d] %q has nil Match function", i, r.Name)
		}
		if r.Dst == nil {
			return nil, fmt.Errorf("routing: rule[%d] %q has nil Dst writer", i, r.Name)
		}
	}
	return &Router{rules: rules, fallback: fallback}, nil
}

// Route evaluates each rule in order and writes p to the first matching
// destination. Returns (len(p), nil) on success so it satisfies io.Writer.
func (r *Router) Route(p []byte) (int, error) {
	for _, rule := range r.rules {
		if rule.Match(p) {
			if _, err := rule.Dst.Write(p); err != nil {
				return 0, fmt.Errorf("routing: rule %q write error: %w", rule.Name, err)
			}
			return len(p), nil
		}
	}
	if r.fallback != nil {
		if _, err := r.fallback.Write(p); err != nil {
			return 0, fmt.Errorf("routing: fallback write error: %w", err)
		}
	}
	return len(p), nil
}

// Write implements io.Writer so Router can be used as a drop-in writer.
func (r *Router) Write(p []byte) (int, error) {
	return r.Route(p)
}
