package create

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"

	"github.com/ghodss/yaml"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [-f FILE]",
		Short: "create new resources from file or STDIN",
		RunE:  execute(cli),
	}

	_ = cmd.Flags().StringP("file", "f", "", "File or directory to create resources from")

	return cmd
}

// returns true if --namespace is specified to be anything other than "default"
func namespaceFlagsSet(cmd *cobra.Command) bool {
	namespace, err := cmd.Flags().GetString("namespace")
	if err == nil && namespace != config.DefaultNamespace {
		return true
	}
	return false
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var in io.Reader
		if len(args) > 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		if namespaceFlagsSet(cmd) {
			cli.Logger.Warn("namespace flags have no effect for create command")
		}
		fp, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		in, err = helpers.InputData(fp)
		if err != nil {
			return err
		}

		resources, err := ParseResources(in)
		if err != nil {
			return err
		}
		if err := ValidateResources(resources); err != nil {
			return err
		}
		return PutResources(cli.Client, resources)
	}
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
func ParseResources(in io.Reader) ([]types.Resource, error) {
	var resources []types.Resource
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("error parsing resources: %s", err)
	}
	// Support concatenated yaml documents separated by '---'
	array := bytes.Split(b, []byte("\n---\n"))
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
				describeError(rerr)
				errCount++
			}
			resources = append(resources, w.Value)
		}
	}
	return resources, err
}

func ValidateResources(resources []types.Resource) error {
	var err error
	errCount := 0
	for i, resource := range resources {
		if verr := resource.Validate(); verr != nil {
			errCount++
			fmt.Fprintf(os.Stderr, "error validating resource %d (%s): %s\n", i, resource.URIPath(), verr)
			if errCount >= 10 {
				err = errors.New("too many errors")
				break
			}
			err = errors.New("resource validation failed")
		}
	}
	return err
}

func describeError(err error) {
	jsonErr, ok := err.(*json.UnmarshalTypeError)
	if !ok {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintf(os.Stderr, "error parsing resource (offset %d): %s\n", jsonErr.Offset, err)
}

func PutResources(client client.GenericClient, resources []types.Resource) error {
	for _, resource := range resources {
		if err := client.PutResource(resource); err != nil {
			return err
		}
	}
	return nil
}
