package primitives

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestParseDefaultResponse(t *testing.T) {
	t.Run("nil target returns error", func(t *testing.T) {
		err := parseDefaultResponse([]byte("hello"), nil)
		if err == nil {
			t.Fatal("expected error for nil target, got nil")
		}
	})

	t.Run("string target receives raw output verbatim", func(t *testing.T) {
		var s string
		out := []byte(`{"not":"json parsed"}`)
		if err := parseDefaultResponse(out, &s); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != string(out) {
			t.Fatalf("expected %q, got %q", string(out), s)
		}
	})

	t.Run("any target receives raw output verbatim as string", func(t *testing.T) {
		var a any
		out := []byte("some raw output")
		if err := parseDefaultResponse(out, &a); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		s, ok := a.(string)
		if !ok {
			t.Fatalf("expected any to hold a string, got %T", a)
		}
		if s != string(out) {
			t.Fatalf("expected %q, got %q", string(out), s)
		}
	})

	t.Run("struct pointer target is unmarshaled as JSON", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
		}
		var p payload
		out := []byte(`{"name":"caporegime"}`)
		if err := parseDefaultResponse(out, &p); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "caporegime" {
			t.Fatalf("expected Name %q, got %q", "caporegime", p.Name)
		}
	})

	t.Run("struct pointer target with malformed JSON returns error", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
		}
		var p payload
		out := []byte(`{"name":`)
		err := parseDefaultResponse(out, &p)
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})

	t.Run("non-pointer target falls through to JSON path and errors", func(t *testing.T) {
		err := parseDefaultResponse([]byte(`{}`), 42)
		if err == nil {
			t.Fatal("expected error for non-pointer target, got nil")
		}
	})
}

func TestParseAsRawString(t *testing.T) {
	t.Run("nil pointer returns error", func(t *testing.T) {
		err := parseAsRawString(nil, []byte("anything"))
		if err == nil {
			t.Fatal("expected error for nil *string, got nil")
		}
	})

	t.Run("stores output verbatim", func(t *testing.T) {
		var s string
		out := []byte("raw content, not necessarily JSON")
		if err := parseAsRawString(&s, out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != string(out) {
			t.Fatalf("expected %q, got %q", string(out), s)
		}
	})

	t.Run("empty output produces empty string", func(t *testing.T) {
		var s string
		if err := parseAsRawString(&s, []byte{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "" {
			t.Fatalf("expected empty string, got %q", s)
		}
	})

	t.Run("nil output produces empty string", func(t *testing.T) {
		var s string
		if err := parseAsRawString(&s, nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s != "" {
			t.Fatalf("expected empty string, got %q", s)
		}
	})
}

func TestParseAsRawAny(t *testing.T) {
	t.Run("nil pointer returns error", func(t *testing.T) {
		err := parseAsRawAny(nil, []byte("anything"))
		if err == nil {
			t.Fatal("expected error for nil *any, got nil")
		}
	})

	t.Run("stores output verbatim as string", func(t *testing.T) {
		var a any
		out := []byte("raw content")
		if err := parseAsRawAny(&a, out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a != string(out) {
			t.Fatalf("expected %q, got %v", string(out), a)
		}
	})

	t.Run("overwrites any pre-existing value", func(t *testing.T) {
		var a any = 12345
		out := []byte("replacement value")
		if err := parseAsRawAny(&a, out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a != string(out) {
			t.Fatalf("expected %q, got %v", string(out), a)
		}
	})
}

func TestParseAsJSON(t *testing.T) {
	t.Run("non-pointer target returns error", func(t *testing.T) {
		err := parseAsJSON(42, []byte(`{}`))
		if err == nil {
			t.Fatal("expected error for non-pointer target, got nil")
		}
		if !strings.Contains(err.Error(), "unknown target type") {
			t.Fatalf("expected 'unknown target type' error, got: %v", err)
		}
	})

	t.Run("nil pointer target returns error", func(t *testing.T) {
		var p *struct{ Name string }
		err := parseAsJSON(p, []byte(`{}`))
		if err == nil {
			t.Fatal("expected error for nil pointer target, got nil")
		}
		if !strings.Contains(err.Error(), "target pointer is nil") {
			t.Fatalf("expected 'target pointer is nil' error, got: %v", err)
		}
	})

	t.Run("valid JSON unmarshals into struct", func(t *testing.T) {
		type payload struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}
		var p payload
		out := []byte(`{"name":"agent","count":3}`)
		if err := parseAsJSON(&p, out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "agent" || p.Count != 3 {
			t.Fatalf("unexpected payload: %+v", p)
		}
	})

	t.Run("malformed JSON returns wrapped error", func(t *testing.T) {
		type payload struct {
			Name string `json:"name"`
		}
		var p payload
		err := parseAsJSON(&p, []byte(`not json`))
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
		if !strings.Contains(err.Error(), "failed unmarshal json into") {
			t.Fatalf("expected wrapped unmarshal error, got: %v", err)
		}
		var syntaxErr *json.SyntaxError
		if !errors.As(err, &syntaxErr) {
			t.Fatalf("expected wrapped error to unwrap to *json.SyntaxError, got: %v", err)
		}
	})

	t.Run("map pointer target unmarshals successfully", func(t *testing.T) {
		var m map[string]any
		out := []byte(`{"key":"value"}`)
		if err := parseAsJSON(&m, out); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["key"] != "value" {
			t.Fatalf("unexpected map contents: %+v", m)
		}
	})
}
