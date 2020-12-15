package resource

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/compat"
)

// Processor is an interface that processes resources through the API.
type Processor interface {
	Process(client client.GenericClient, resources []*types.Wrapper) error
}

type httpDirectory struct {
	XMLName xml.Name `xml:"pre"`
	Files   []string `xml:"a"`
}

// Process processes the input.
func Process(cli *cli.SensuCli, client *http.Client, inputs []string, recurse bool, processor Processor) error {
	var resources []*types.Wrapper
	for _, input := range inputs {
		res, err := process(client, input, recurse)
		if err != nil {
			return err
		}
		resources = append(resources, res...)
	}
	if err := Validate(resources, cli.Config.Namespace()); err != nil {
		return err
	}
	return processor.Process(cli.Client, resources)
}

func process(client *http.Client, input string, recurse bool) ([]*types.Wrapper, error) {
	var resources []*types.Wrapper
	urly, err := url.Parse(input)
	if err != nil {
		return resources, err
	}
	var res []*types.Wrapper
	if urly.Scheme == "" || len(urly.Scheme) == 1 {
		// We are dealing with a file path
		res, err = ProcessFile(input, recurse)
		if err != nil {
			return resources, err
		}
	} else {
		res, err = ProcessURL(client, urly, input, recurse)
		if err != nil {
			return resources, err
		}
	}
	resources = append(resources, res...)
	return resources, nil
}

// ProcessFile processes a file.
func ProcessFile(input string, recurse bool) ([]*types.Wrapper, error) {
	var resources []*types.Wrapper
	var tld = true
	err := filepath.Walk(input, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Resolve symbolic link
		if info.Mode()&os.ModeSymlink != 0 {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			info, err = os.Lstat(path)
			if err != nil {
				return err
			}
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
		res, err := Parse(f)
		if err != nil {
			return fmt.Errorf("in %s: %s", input, err)
		}
		resources = append(resources, res...)
		return nil
	})
	return resources, err
}

// ProcessURL processes a url.
func ProcessURL(client *http.Client, urly *url.URL, input string, recurse bool) ([]*types.Wrapper, error) {
	var resources []*types.Wrapper
	req, err := http.NewRequest("GET", urly.String(), nil)
	if err != nil {
		return resources, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return resources, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf := new(bytes.Buffer)
		_, _ = io.Copy(buf, resp.Body)
		return resources, errors.New(buf.String())
	}

	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/html") {
		// The server returned us a directory listing
		if !recurse {
			return resources, errors.New("use -r to enable directory recursion")
		}
		dec := xml.NewDecoder(resp.Body)
		var dir httpDirectory
		if err := dec.Decode(&dir); err != nil {
			return resources, err
		}
		for _, file := range dir.Files {
			res, err := process(client, filepath.Join(input, file), recurse)
			if err != nil {
				return resources, err
			}
			resources = append(resources, res...)
		}
	} else {
		resources, err = Parse(resp.Body)
		if err != nil {
			return resources, fmt.Errorf("in %s: %s", input, err)
		}
	}
	return resources, nil
}

// ProcessStdin processes standard in.
func ProcessStdin(cli *cli.SensuCli, client *http.Client, processor Processor) error {
	resources, err := Parse(os.Stdin)
	if err != nil {
		return fmt.Errorf("in stdin: %s", err)
	}
	if err := Validate(resources, cli.Config.Namespace()); err != nil {
		return err
	}
	return processor.Process(cli.Client, resources)
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
				i, resource.ObjectMeta.Name, resource.ObjectMeta.Namespace, compat.URIPath(resource.Value), err,
			)
		}
	}
	return nil
}

// ManagedByLabelPutter is a Processor that applies a corev2.ManagedByLabel
// label with the chosen value to resources before passing them to a Putter.
type ManagedByLabelPutter struct {
	putter *Putter
	Label  string
}

func NewManagedByLabelPutter(label string) *ManagedByLabelPutter {
	return &ManagedByLabelPutter{
		Label:  label,
		putter: NewPutter(),
	}
}

func (p *ManagedByLabelPutter) Process(client client.GenericClient, resources []*types.Wrapper) error {
	for _, resource := range resources {
		p.label(resource)
	}
	return p.putter.Process(client, resources)
}

func (p *ManagedByLabelPutter) label(resource *types.Wrapper) {
	innerMeta := compat.GetObjectMeta(resource.Value)

	if resource.ObjectMeta.Labels == nil {
		resource.ObjectMeta.Labels = map[string]string{}
	}
	outerLabels := resource.ObjectMeta.Labels

	if innerMeta.Labels == nil {
		innerMeta.Labels = map[string]string{}
	}
	innerLabels := innerMeta.Labels

	// By default the resource should be managed by sensuctl
	managedBy := p.Label

	// Mark the resource as managed by `label` in the outer labels if none is
	// already set
	if outerLabels[corev2.ManagedByLabel] != "sensu-agent" {
		outerLabels[corev2.ManagedByLabel] = managedBy
	} else {
		managedBy = outerLabels[corev2.ManagedByLabel]
	}

	// Mark the resource as managed by `label` in the inner labels
	if innerLabels[corev2.ManagedByLabel] != "sensu-agent" || innerLabels[corev2.ManagedByLabel] != outerLabels[corev2.ManagedByLabel] {
		innerLabels[corev2.ManagedByLabel] = managedBy
	}

	compat.SetObjectMeta(resource.Value, innerMeta)
}
