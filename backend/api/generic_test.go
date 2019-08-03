package api

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

func TestGenericClient(t *testing.T) {
	tests := []struct {
		Name      string
		Client    *genericClient
		ExpError  bool
		CreateVal corev2.Resource
		UpdateVal corev2.Resource
		GetVal    corev2.Resource
		ListVal   []corev2.Resource
		ListPred  *store.SelectionPredicate
		GetName   string
		DelName   string
		Ctx       context.Context
	}{}

	for _, test := range tests {
		t.Run(test.Name+"_create", func(t *testing.T) {
			err := test.Client.Create(test.Ctx, test.CreateVal)
			if err != nil && !test.ExpError {
				t.Fatal(err)
			}
			if err == nil && test.ExpError {
				t.Fatal("expected non-nil error")
			}
		})
		t.Run(test.Name+"_update", func(t *testing.T) {
			err := test.Client.Update(test.Ctx, test.UpdateVal)
			if err != nil && !test.ExpError {
				t.Fatal(err)
			}
			if err == nil && test.ExpError {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_get", func(t *testing.T) {
			err := test.Client.Get(test.Ctx, test.GetName, test.GetVal)
			if err != nil && !test.ExpError {
				t.Fatal(err)
			}
			if err == nil && test.ExpError {
				t.Fatal(err)
			}
			if err == nil && test.GetVal.Validate() != nil {
				t.Fatal(test.GetVal.Validate())
			}
		})
		t.Run(test.Name+"_del", func(t *testing.T) {
			err := test.Client.Delete(test.Ctx, test.DelName)
			if err != nil && !test.ExpError {
				t.Fatal(err)
			}
			if err == nil && test.ExpError {
				t.Fatal(err)
			}
		})
		t.Run(test.Name+"_list", func(t *testing.T) {
			err := test.Client.List(test.Ctx, &test.ListVal, test.ListPred)
			if err != nil && !test.ExpError {
				t.Fatal(err)
			}
			if err == nil && test.ExpError {
				t.Fatal(err)
			}
			for _, val := range test.ListVal {
				if err == nil && val.Validate() != nil {
					t.Fatal(val.Validate())
				}
			}
		})
	}
}
