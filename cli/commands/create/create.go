package create

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
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// CreateCommand creates generic Sensu resources.
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [-r] [[-f URL] ... ]",
		Short: "create or replace resources from file or URL (path, file://, http[s]://), or STDIN otherwise.",
		RunE:  execute(cli),
	}

	_ = cmd.Flags().StringSliceP("file", "f", nil, "Files, directories, or URLs to create resources from")
	_ = cmd.Flags().BoolP("recursive", "r", false, "Follow subdirectories")

	return cmd
}

type httpDirectory struct {
	XMLName xml.Name `xml:"pre"`
	Files   []string `xml:"a"`
}

func processFile(cli *cli.SensuCli, input string, recurse bool) error {
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
		return PutResources(cli.Client, resources)
	})
}

func process(cli *cli.SensuCli, client *http.Client, input string, recurse bool) error {
	urly, err := url.Parse(input)
	if err != nil {
		return err
	}
	if urly.Scheme == "" || len(urly.Scheme) == 1 {
		// We are dealing with a file path
		return processFile(cli, input, recurse)
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
			if err := process(cli, client, filepath.Join(input, file), recurse); err != nil {
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
	return PutResources(cli.Client, resources)
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		t := &http.Transport{}
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		client := &http.Client{Transport: t}
		inputs, err := cmd.Flags().GetStringSlice("file")
		if err != nil {
			return err
		}
		if len(inputs) == 0 {
			return processStdin(cli, client)
		}
		recurse, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}
		for _, input := range inputs {
			if err := process(cli, client, input, recurse); err != nil {
				return err
			}
		}
		return nil
	}
}

func processStdin(cli *cli.SensuCli, client *http.Client) error {
	resources, err := ParseResources(os.Stdin)
	if err != nil {
		return fmt.Errorf("in stdin: %s", err)
	}
	if err := ValidateResources(resources, cli.Config.Namespace()); err != nil {
		return err
	}
	return PutResources(cli.Client, resources)
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
func ParseResources(in io.Reader) ([]types.Wrapper, error) {
	var resources []types.Wrapper
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
			resources = append(resources, w)
			count++
		}
	}

	// TODO(echlebek): remove this
	filterCheckSubdue(resources)

	return resources, err
}

// filterCheckSubdue nils out any check subdue fields that are supplied.
// TODO(echlebek): this is temporary; remove it after fixing check subdue.
func filterCheckSubdue(resources []types.Wrapper) {
	for i := range resources {
		switch val := resources[i].Value.(type) {
		case *types.CheckConfig:
			val.Subdue = nil
		case *types.Check:
			val.Subdue = nil
		case *types.EventFilter:
			val.When = nil
		}
	}
}

// ValidateResources loops through a list of resources, appends a namespace
// if one is not already declared, and validates the resource.
func ValidateResources(resources []types.Wrapper, namespace string) error {
	var err error
	errCount := 0
	for i, r := range resources {
		resource := r.Value
		if resource == nil {
			errCount++
			fmt.Fprintf(os.Stderr, "error validating resource %d: resource is nil\n", i)
			continue
		}
		if resource.GetObjectMeta().Namespace == "" {
			resource.SetNamespace(namespace)
		}
	}
	return err
}

func describeError(index int, err error) {
	jsonErr, ok := err.(*json.UnmarshalTypeError)
	if !ok {
		fmt.Fprintf(os.Stderr, "resource %d: %s\n", index, err)
		return
	}
	fmt.Fprintf(os.Stderr, "resource %d: (offset %d): %s\n", index, jsonErr.Offset, err)
}

// PutResources uses the GenericClient to PUT a resource at the inferred URI path.
func PutResources(client client.GenericClient, resources []types.Wrapper) error {
	for i, resource := range resources {
		if err := client.PutResource(resource); err != nil {
			return fmt.Errorf("error putting resource %d (%s): %s", i, resource.Value.URIPath(), err)
		}
	}
	return nil
}
