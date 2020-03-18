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
	return ProcessURL(cli, client, urly, input, recurse, processor)
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
		resources, err := Parse(f)
		if err != nil {
			return fmt.Errorf("in %s: %s", input, err)
		}
		if err := Validate(resources, cli.Config.Namespace()); err != nil {
			return err
		}
		return processor.Process(cli.Client, resources)
	})
}

// ProcessURL processes a url.
func ProcessURL(cli *cli.SensuCli, client *http.Client, urly *url.URL, input string, recurse bool, processor Processor) error {
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

	resources, err := Parse(resp.Body)
	if err != nil {
		return fmt.Errorf("in %s: %s", input, err)
	}
	if err := Validate(resources, cli.Config.Namespace()); err != nil {
		return err
	}
	return processor.Process(cli.Client, resources)
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
				i, resource.ObjectMeta.Name, resource.ObjectMeta.Namespace, resource.Value.URIPath(), err,
			)
		}
	}
	return nil
}
