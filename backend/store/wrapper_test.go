package store

import (
	"encoding/json"
	fmt "fmt"
	"testing"

	proto "github.com/golang/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

func TestUnmarshalWrapper(t *testing.T) {
	v := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name: "foo",
		},
	}
	a, err := proto.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	w := &Wrapper{
		Metadata: &corev2.TypeMeta{
			APIVersion: "core/v3",
			Type:       "entityconfig",
		},
		Value: a,
	}
	b, err := json.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s\n", string(b))
}
