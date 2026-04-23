package json

import (
	"encoding/json"
	"io"
)

// Serializer is a wrapper around the standard library's encoding/json.
// It can be used by consumers to define their own interfaces for serialization.
type Serializer struct{}

// Marshal wraps json.Marshal.
func (s Serializer) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal wraps json.Unmarshal.
func (s Serializer) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// Encode wraps json.NewEncoder(w).Encode(v).
func (s Serializer) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

// Decode wraps json.NewDecoder(r).Decode(v).
func (s Serializer) Decode(r io.Reader, v any) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
