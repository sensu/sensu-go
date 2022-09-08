package vet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"

	"cuelang.org/go/cue/cuecontext"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/compat"
	"github.com/spf13/cobra"
)

var ChunkSize = 100
var description = `sensuctl vet

Vet resources Example:
$ sensuctl vet checks --spec "ttl: > 0"

The tool also supports naming types by their fully-qualified names:
$ sensuctl vet core/v2.CheckConfig --spec "ttl: > 0"

The tool uses cue (cuelang.org) to constrain types
`

func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vet [RESOURCE TYPE] [-spec SPEC]",
		Short: "Vet resources",
		Long:  description,
		RunE:  execute(cli),
	}

	_ = cmd.Flags().StringP("spec", "s", "{}", "spec for resources to follow")
	helpers.AddAllNamespace(cmd.Flags())
	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		resourceSpec, err := cmd.Flags().GetString("spec")
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		// parse the comma separated resource types and match against the defined actions
		requests, err := resource.GetResourceRequests(args[0], resource.All)
		if err != nil {
			return err
		}

		var vetErrors []vetErr

		for _, resourceReq := range requests {
			ok, err := cmd.Flags().GetBool(flags.AllNamespaces)
			if err != nil {
				return err
			}
			if ok {
				resourceReq.SetNamespace(corev2.NamespaceTypeAll)
			} else {
				resourceReq.SetNamespace(cli.Config.Namespace())
			}
			var resourceSlice reflect.Value
			if proxy, ok := resourceReq.(*corev3.V2ResourceProxy); ok {
				resourceSlice = reflect.New(reflect.SliceOf(reflect.TypeOf(proxy.Resource)))
			} else {
				resourceSlice = reflect.New(reflect.SliceOf(reflect.TypeOf(resourceReq)))
			}
			err = cli.Client.List(
				fmt.Sprintf("%s?types=%s", resourceReq.URIPath(), url.QueryEscape(types.WrapResource(resourceReq).Type)),
				resourceSlice.Interface(), &client.ListOptions{
					ChunkSize: ChunkSize,
				}, nil)
			if err != nil {
				// We want to ignore non-nil errors that are a result of
				// resources not existing, or features being licensed.
				return fmt.Errorf("API error: %s", err)
			}

			resourceSlice = reflect.Indirect(resourceSlice)
			if resourceSlice.Len() == 0 {
				continue
			}
			ctx := cuecontext.New()
			constraint := ctx.CompileString(resourceSpec)
			for i := 0; i < resourceSlice.Len(); i++ {
				resource := compat.V2Resource(resourceSlice.Index(i).Interface())
				resourceJson := bytes.Buffer{}
				helpers.PrintJSON(resource, &resourceJson)
				if cerr := constraint.Err(); cerr != nil {
					return cerr
				}
				cueval := ctx.CompileString(resourceJson.String())
				if cErr := constraint.Subsume(cueval); cErr != nil {
					wrapper := types.WrapResource(resource)
					vetErrors = append(vetErrors, vetErr{Error: cErr.Error(), Resource: corev2.ResourceReference{
						Name:       resource.GetObjectMeta().Name,
						Type:       wrapper.Type,
						APIVersion: wrapper.APIVersion,
					}})
				}
			}

		}
		if len(vetErrors) > 0 {
			encoder := json.NewEncoder(os.Stderr)
			if err := encoder.Encode(vetErrors); err != nil {
				panic(err)
			}
			os.Exit(2)
		}
		return nil

	}
}

type vetErr struct {
	Error    string
	Resource corev2.ResourceReference
}
