package resource

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
	"sigs.k8s.io/yaml"
)

// Parse is a rather heroic function that will parse any number of valid
// JSON or YAML resources. Since it attempts to be intelligent, it likely
// contains bugs.
//
// The general approach is:
// 1. detect if the stream is JSON by sniffing the first non-whitespace byte.
// 2. If the stream is JSON, goto 4.
// 3. If the stream is YAML, split it on '---' to support multiple yaml documents.
// 3. Convert the YAML to JSON document-by-document.
// 4. Unmarshal the JSON one resource at a time.
func Parse(in io.Reader) ([]*types.Wrapper, error) {
	var resources []*types.Wrapper
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("error parsing resources: %s", err)
	}
	// Support concatenated yaml documents separated by '---'
	array := bytes.Split(b, []byte("\n---\n"))
	count := 0
	for _, b := range array {
		var jsonBytes []byte
		if jsonRe.Match(b) {
			// We are dealing with JSON data
			jsonBytes = b
		} else {
			// We are dealing with YAML data
			var err error
			jsonBytes, err = yaml.YAMLToJSON(b)
			if err != nil {
				return nil, fmt.Errorf("error parsing resources: %s", err)
			}
		}
		dec := json.NewDecoder(bytes.NewReader(jsonBytes))
		dec.DisallowUnknownFields()
		errCount := 0
		for dec.More() {
			var w types.Wrapper
			if rerr := dec.Decode(&w); rerr != nil {
				// Write out as many errors as possible before bailing,
				// but cap it at 10.
				err = errors.New("some resources couldn't be parsed")
				if errCount > 10 {
					err = errors.New("too many errors")
					break
				}
				describeError(count, rerr)
				errCount++
				continue
			}

			// Mark the resource as managed by sensuctl in the outer labels
			if len(w.ObjectMeta.Labels) == 0 {
				w.ObjectMeta.Labels = map[string]string{}
			}
			w.ObjectMeta.Labels[corev2.ManagedByLabel] = "sensuctl"

			// Mark the resource as managed by sensuctl in the inner labels
			innerMeta := w.Value.GetObjectMeta()
			if len(innerMeta.Labels) == 0 {
				innerMeta.Labels = map[string]string{}
			}
			innerMeta.Labels[corev2.ManagedByLabel] = "sensuctl"
			w.Value.SetObjectMeta(innerMeta)

			resources = append(resources, &w)
			count++
		}
	}

	// TODO(echlebek): remove this
	filterCheckSubdue(resources)

	return resources, err
}

// filterCheckSubdue nils out any check subdue fields that are supplied.
// TODO(echlebek): this is temporary; remove it after fixing check subdue.
func filterCheckSubdue(resources []*types.Wrapper) {
	for i := range resources {
		switch val := resources[i].Value.(type) {
		case *corev2.CheckConfig:
			val.Subdue = nil
		case *corev2.Check:
			val.Subdue = nil
		case *corev2.EventFilter:
			val.When = nil
		}
	}
}

// Validate loops through a list of resources, appends a namespace
// if one is not already declared, and validates the resource.
func Validate(resources []*types.Wrapper, namespace string) error {
	errCount := 0
	for i, r := range resources {
		resource := r.Value
		if resource == nil {
			errCount++
			fmt.Fprintf(
				os.Stderr,
				"error validating resource #%d with name %q and namespace %q: resource is nil\n",
				i, r.ObjectMeta.Name, r.ObjectMeta.Namespace,
			)
			continue
		}
		if resource.GetObjectMeta().Namespace == "" {
			resource.SetNamespace(namespace)
			// We just set the namespace within the underlying wrapped value. We also
			// need to set it to the outer ObjectMeta for consistency, but only if the
			// resource has a namespace; some resources are cluster-wide and should
			// not be namespaced
			if ns := resource.GetObjectMeta().Namespace; ns != "" {
				r.ObjectMeta.Namespace = ns
			}
		}
	}

	return nil
}

var jsonRe = regexp.MustCompile(`^(\s)*[\{\[]`)

func describeError(index int, err error) {
	jsonErr, ok := err.(*json.UnmarshalTypeError)
	if !ok {
		fmt.Fprintf(os.Stderr, "resource %d: %s\n", index, err)
		return
	}
	fmt.Fprintf(os.Stderr, "resource %d: (offset %d): %s\n", index, jsonErr.Offset, err)
}
