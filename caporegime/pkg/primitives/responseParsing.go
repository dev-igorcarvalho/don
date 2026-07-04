package primitives

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// parseDefaultResponse deserializes raw provider output into target. *string and *any
// targets receive the raw output verbatim; any other pointer type is unmarshaled as JSON,
// so new FoundationModelResponse types work without adding a case here.
func parseDefaultResponse(out []byte, target any) error {
	if target == nil {
		return errors.New("parse target is nil")
	}
	switch t := target.(type) {
	case *string:
		return parseAsRawString(t, out)
	case *any:
		return parseAsRawAny(t, out)
	default:
		return parseAsJSON(target, out)
	}
}

// parseAsRawString stores out verbatim in t, without attempting any deserialization.
func parseAsRawString(t *string, out []byte) error {
	if t == nil {
		return errors.New("target string pointer is nil")
	}
	*t = string(out)
	return nil
}

// parseAsRawAny stores out verbatim in t, without attempting any deserialization.
func parseAsRawAny(t *any, out []byte) error {
	if t == nil {
		return errors.New("target any pointer is nil")
	}
	*t = string(out)
	return nil
}

// parseAsJSON unmarshals out as JSON into target, which must be a non-nil pointer.
func parseAsJSON(target any, out []byte) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("unknown target type: %T", target)
	}
	if rv.IsNil() {
		return errors.New("target pointer is nil")
	}
	if err := json.Unmarshal(out, target); err != nil {
		return fmt.Errorf("failed unmarshal json into %T: %w", target, err)
	}
	return nil
}
