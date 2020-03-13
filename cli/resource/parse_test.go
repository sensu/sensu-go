package resource

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name          string
		resource      *types.Wrapper
		namespace     string
		wantNamespace string
	}{
		{
			name: "a namespaced resource with a configured namespace should not be modified",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
				},
			},
			namespace:     "dev",
			wantNamespace: "default",
		},
		{
			name: "a namespaced resource without a configured namespace should use the provided namespace",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("check-cpu", ""),
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.NewObjectMeta("check-cpu", ""),
				},
			},
			namespace:     "dev",
			wantNamespace: "dev",
		},
		{
			name: "a global resource should not have a namespace configured",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("admin-role", ""),
				Value: &corev2.ClusterRole{
					ObjectMeta: corev2.NewObjectMeta("admin-role", ""),
				},
			},
			namespace:     "dev",
			wantNamespace: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := []*types.Wrapper{tt.resource}
			_ = Validate(resources, tt.namespace)

			if tt.resource.ObjectMeta.Namespace != tt.wantNamespace {
				t.Errorf("Validate() wrapper namespace = %q, want namespace %q", tt.resource.ObjectMeta.Namespace, tt.wantNamespace)
			}
			if tt.resource.Value != nil && tt.resource.Value.GetObjectMeta().Namespace != tt.wantNamespace {
				t.Errorf("Validate() wrapper's resource namespace = %q, want namespace %q", tt.resource.Value.GetObjectMeta().Namespace, tt.wantNamespace)
			}
		})
	}
}

func TestValidateStderr(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ch := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		ch <- buf.String()
	}()

	resources := []*types.Wrapper{&types.Wrapper{
		ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
	}}
	_ = Validate(resources, "default")

	// Reset stderr
	w.Close()
	os.Stderr = oldStderr

	errMsg := <-ch
	errMsg = strings.TrimSpace(errMsg)
	wantErr := `error validating resource #0 with name "check-cpu" and namespace "default": resource is nil`
	if errMsg != wantErr {
		t.Errorf("Validate() err = %s, want %s", errMsg, wantErr)
	}
}

func TestCreatedByLabel(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Wrapper
		want     map[string]string
	}{
		{
			name: "the label is set on both the inner and outer labels",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Name: "foo",
				},
			}),
			want: map[string]string{
				corev2.ManagedByLabel: "sensuctl",
			},
		},
		{
			name: "the label is appended to an existing list of labels",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Labels: map[string]string{
						"region": "us-west-2",
					},
				},
			}),
			want: map[string]string{
				"region":              "us-west-2",
				corev2.ManagedByLabel: "sensuctl",
			},
		},
		{
			name: "the label overwrites any existing value",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Labels: map[string]string{
						corev2.ManagedByLabel: "web-ui",
					},
				},
			}),
			want: map[string]string{
				corev2.ManagedByLabel: "sensuctl",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdin bytes.Buffer
			bytes, _ := json.Marshal(&tt.resource)
			stdin.Write(bytes)

			got, err := Parse(&stdin)
			if err != nil {
				t.Errorf("Parse() error = %v", err)
				return
			}
			if !reflect.DeepEqual(got[0].ObjectMeta.Labels, tt.want) {
				t.Errorf("Parse() inner labels = %v, want %v", got[0].ObjectMeta.Labels, tt.want)
			}
			if !reflect.DeepEqual(got[0].Value.GetObjectMeta().Labels, tt.want) {
				t.Errorf("Parse() outer labels = %v, want %v", got[0].Value.GetObjectMeta().Labels, tt.want)
			}
		})
	}
}
