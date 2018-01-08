package dynamic

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(data, &m)
	}
}

func BenchmarkMarshal(b *testing.B) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	if err := json.Unmarshal(data, &m); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(&m)
	}
}

func TestMarshal(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	extendedBytes := []byte(`{"a":1,"b":2.0,"c":true,"d":"false","e":[1,2,3],"f":{"foo":"bar"}}`)
	expBytes := []byte(`{"bar":null,"foo":"hello world!","a":1,"b":2.0,"c":true,"d":"false","e":[1,2,3],"f":{"foo":"bar"}}`)

	m := &MyType{
		Foo:   "hello world!",
		Bar:   nil,
		Attrs: extendedBytes,
	}

	b, err := Marshal(m)
	require.NoError(err)
	assert.Equal(expBytes, b)
}

func TestMarshalEmptyAttrs(t *testing.T) {
	var m MyType
	b, err := Marshal(&m)
	require.NoError(t, err)
	assert.Equal(t, `{"bar":null,"foo":""}`, string(b))
}

func TestUnmarshalEmptyAttrs(t *testing.T) {
	var m MyType
	err := Unmarshal([]byte("{}"), &m)
	require.NoError(t, err)
}

func TestMarshalUnmarshal(t *testing.T) {
	data := []byte(`{"bar":null,"foo":"hello","a":10,"b":"c"}`)
	var m MyType
	err := json.Unmarshal(data, &m)
	require.NoError(t, err)
	assert.Equal(t, MyType{Foo: "hello", Attrs: []byte(`{"a":10,"b":"c"}`)}, m)
	b, err := json.Marshal(&m)
	require.NoError(t, err)
	assert.Equal(t, data, b)
}

type Problematic struct {
	Foo   string
	Bar   *string
	Attrs []byte `json:"-"`
}

func (p *Problematic) GetExtendedAttributes() []byte {
	return p.Attrs
}

func (p *Problematic) SetExtendedAttributes(b []byte) {
	p.Attrs = b
}

func TestMarshalNilField(t *testing.T) {
	p := &Problematic{
		Foo: "yes",
		Bar: nil,
	}
	b, err := Marshal(p)
	require.NoError(t, err)
	require.Equal(t, []byte(`{"Bar":null,"Foo":"yes"}`), b)
}

func TestUnmarshalNilField(t *testing.T) {
	p := &Problematic{}
	err := Unmarshal([]byte(`{"Bar":null,"Foo":"yes"}`), p)
	require.NoError(t, err)
}

func TestMarshalNilProblematic(t *testing.T) {
	var p *Problematic
	b, err := Marshal(p)
	require.NoError(t, err)
	require.Equal(t, []byte("null"), b)
}

func TestUnmarshalNilProblematic(t *testing.T) {
	var p *Problematic
	err := Unmarshal([]byte("{}"), p)
	require.Error(t, err)
}
