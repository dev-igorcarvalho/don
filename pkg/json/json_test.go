package json_test

import (
	"bytes"
	"testing"

	"github.com/dev-igorcarvalho/don/pkg/json"
	"github.com/stretchr/testify/assert"
)

func TestSerializer_Marshal(t *testing.T) {
	s := json.Serializer{}
	data := map[string]string{"foo": "bar"}

	bytes, err := s.Marshal(data)

	assert.NoError(t, err)
	assert.JSONEq(t, `{"foo":"bar"}`, string(bytes))
}

func TestSerializer_Unmarshal(t *testing.T) {
	s := json.Serializer{}
	input := []byte(`{"foo":"bar"}`)
	var output map[string]string

	err := s.Unmarshal(input, &output)

	assert.NoError(t, err)
	assert.Equal(t, "bar", output["foo"])
}

func TestSerializer_Encode(t *testing.T) {
	s := json.Serializer{}
	data := map[string]string{"foo": "bar"}
	var buf bytes.Buffer

	err := s.Encode(&buf, data)

	assert.NoError(t, err)

	assert.JSONEq(t, `{"foo":"bar"}`, buf.String())
}

func TestSerializer_Decode(t *testing.T) {
	s := json.Serializer{}
	input := []byte(`{"foo":"bar"}`)
	var output map[string]string
	r := bytes.NewReader(input)

	err := s.Decode(r, &output)

	assert.NoError(t, err)
	assert.Equal(t, "bar", output["foo"])
}
