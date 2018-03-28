package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
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

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var in io.Reader
		if len(args) > 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		fp, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		if fp == "" {
			if err := detectEmptyStdin(os.Stdin); err != nil {
				_ = cmd.Help()
				return err
			}
			in = os.Stdin
		} else {
			f, err := os.Open(fp)
			if err != nil {
				return err
			}
			stat, err := f.Stat()
			if err != nil {
				return err
			}
			if stat.IsDir() {
				return errors.New("directories not supported yet")
			}
			in = f
		}
		resources, err := parseResources(in)
		if err != nil {
			return err
		}
		if err := validateResources(resources); err != nil {
			return err
		}
		return putResources(cli.Client, resources)
	}
}

func detectEmptyStdin(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		if fi.Mode()&os.ModeNamedPipe == 0 {
			return errors.New("empty stdin")
		}
	}
	return nil
}

func parseResources(in io.Reader) ([]types.Resource, error) {
	var resources []types.Resource
	dec := json.NewDecoder(in)
	dec.DisallowUnknownFields()
	var err error
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
	return resources, err
}

func validateResources(resources []types.Resource) error {
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

func putResources(client client.GenericClient, resources []types.Resource) error {
	for _, resource := range resources {
		if err := client.PutResource(resource); err != nil {
			return err
		}
	}
	return nil
}
