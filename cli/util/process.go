package util

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ghodss/yaml"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
)

// Processor is an interface that processes resources through the API.
type Processor interface {
	Process(client client.GenericClient, resources []*types.Wrapper) error
}

type httpDirectory struct {
	XMLName xml.Name `xml:"pre"`
	Files   []string `xml:"a"`
}

// ProcessFile processes a file.
func ProcessFile(cli *cli.SensuCli, input string, recurse bool, processor Processor) error {
	var tld = true
	return filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !recurse && info.IsDir() && !tld {
			return filepath.SkipDir
		} else if info.IsDir() && tld || info.IsDir() && recurse {
			tld = false
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		resources, err := ParseResources(f)
		if err != nil {
			return fmt.Errorf("in %s: %s", input, err)
		}
		if err := ValidateResources(resources, cli.Config.Namespace()); err != nil {
			return err
		}
		return processor.Process(cli.Client, resources)
	})
}

// Process processes the input.
func Process(cli *cli.SensuCli, client *http.Client, input string, recurse bool, processor Processor) error {
	urly, err := url.Parse(input)
	if err != nil {
		return err
	}
	if urly.Scheme == "" || len(urly.Scheme) == 1 {
		// We are dealing with a file path
		return ProcessFile(cli, input, recurse, processor)
	}
	req, err := http.NewRequest("GET", urly.String(), nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		_, _ = io.Copy(buf, resp.Body)
		return errors.New(buf.String())
	}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		// The server returned us a directory listing
		if !recurse {
			return errors.New("use -r to enable directory recursion")
		}
		dec := xml.NewDecoder(resp.Body)
		var dir httpDirectory
		if err := dec.Decode(&dir); err != nil {
			return err
		}
		for _, file := range dir.Files {
			if err := Process(cli, client, filepath.Join(input, file), recurse, processor); err != nil {
				return err
			}
		}
	}

	resources, err := ParseResources(resp.Body)
	if err != nil {
		return fmt.Errorf("in %s: %s", input, err)
	}
	if err := ValidateResources(resources, cli.Config.Namespace()); err != nil {
		return err
	}
	return processor.Process(cli.Client, resources)
}

// ProcessStdin processes standard in.
func ProcessStdin(cli *cli.SensuCli, client *http.Client, processor Processor) error {
	resources, err := ParseResources(os.Stdin)
	if err != nil {
		return fmt.Errorf("in stdin: %s", err)
	}
	if err := ValidateResources(resources, cli.Config.Namespace()); err != nil {
		return err
	}
	return processor.Process(cli.Client, resources)
}

var jsonRe = regexp.MustCompile(`^(\s)*[\{\[]`)

// ParseResources is a rather heroic function that will parse any number of valid
// JSON or YAML resources. Since it attempts to be intelligent, it likely
// contains bugs.
//
// The general approach is:
// 1. detect if the stream is JSON by sniffing the first non-whitespace byte.
// 2. If the stream is JSON, goto 4.
// 3. If the stream is YAML, split it on '---' to support multiple yaml documents.
// 3. Convert the YAML to JSON document-by-document.
// 4. Unmarshal the JSON one resource at a time.
func ParseResources(in io.Reader) ([]*types.Wrapper, error) {
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

// ValidateResources loops through a list of resources, appends a namespace
// if one is not already declared, and validates the resource.
func ValidateResources(resources []*types.Wrapper, namespace string) error {
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

func describeError(index int, err error) {
	jsonErr, ok := err.(*json.UnmarshalTypeError)
	if !ok {
		fmt.Fprintf(os.Stderr, "resource %d: %s\n", index, err)
		return
	}
	fmt.Fprintf(os.Stderr, "resource %d: (offset %d): %s\n", index, jsonErr.Offset, err)
}

// Putter is a Processor that puts resources in the API.
type Putter struct{}

// NewPutter instantiates a new Putter Processor.
func NewPutter() *Putter {
	return &Putter{}
}

// Process puts resources in the API.
func (p *Putter) Process(client client.GenericClient, resources []*types.Wrapper) error {
	for i, resource := range resources {
		if err := client.PutResource(*resource); err != nil {
			return fmt.Errorf(
				"error putting resource #%d with name %q and namespace %q (%s): %s",
				i, resource.ObjectMeta.Name, resource.ObjectMeta.Namespace, resource.Value.URIPath(), err,
			)
		}
	}
	return nil
}
