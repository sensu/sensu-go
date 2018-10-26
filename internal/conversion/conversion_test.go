package conversion

import (
	"errors"
	"testing"
)

func mockConversionFunc(dst, src interface{}) error {
	return nil
}

func mockConversionFuncErr(dst, src interface{}) error {
	return errors.New("an error")
}

func mockConversionFuncPanic(dst, src interface{}) error {
	panic("omg")
}

type mockType struct {
	APIVersion string
	Kind       string
}

func (m mockType) GetAPIVersion() string {
	return m.APIVersion
}

func (m mockType) GetKind() string {
	return m.Kind
}

func TestConvertTypes(t *testing.T) {
	t1 := mockType{APIVersion: "core/v1", Kind: "check"}
	t2 := mockType{APIVersion: "core", Kind: "check"}
	registry[key{
		SourceAPIVersion: "core",
		DestAPIVersion:   "core/v1",
		Kind:             "check",
	}] = mockConversionFunc
	if err := ConvertTypes(&t1, &t2); err != nil {
		t.Fatal(err)
	}
}

func TestConvertTypesErr(t *testing.T) {
	t1 := mockType{APIVersion: "core/v1", Kind: "entity"}
	t2 := mockType{APIVersion: "core", Kind: "entity"}
	registry[key{
		SourceAPIVersion: "core",
		DestAPIVersion:   "core/v1",
		Kind:             "entity",
	}] = mockConversionFuncErr
	if err := ConvertTypes(&t1, &t2); err == nil {
		t.Fatal("expected error")
	} else if err == ErrImpossibleConversion {
		t.Fatal("wrong error")
	}
}

func TestConvertTypesPanic(t *testing.T) {
	t1 := mockType{APIVersion: "core/v1", Kind: "event"}
	t2 := mockType{APIVersion: "core", Kind: "event"}
	registry[key{
		SourceAPIVersion: "core",
		DestAPIVersion:   "core/v1",
		Kind:             "event",
	}] = mockConversionFuncPanic
	if err := ConvertTypes(&t1, &t2); err == nil {
		t.Fatal("expected error")
	} else if err == ErrImpossibleConversion {
		t.Fatalf("wrong error: %q", err)
	}
}

func TestConvertTypesMissing(t *testing.T) {
	t1 := mockType{APIVersion: "core/v1", Kind: "mutator"}
	t2 := mockType{APIVersion: "core", Kind: "mutator"}
	if err := ConvertTypes(&t1, &t2); err != ErrImpossibleConversion {
		t.Fatalf("wrong error: %q", err)
	}
}
