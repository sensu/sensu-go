package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTransformBasicMessage(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		message Simple {
			double field_double = 1;
			float field_float = 2;
			int32 field_int32 = 3;
			int64 field_int64 = 4;
			uint64 field_uint64 = 5;
			sint32 field_sint32 = 6;
			sint64 field_sint64 = 7;
			fixed32 field_fixed32 = 8;
			fixed64 field_fixed64 = 9;
			sfixed32 field_sfixed32 = 10;
			sfixed64 field_sfixed64 = 11;
			bool field_bool = 12;
			string field_string = 13;
			bytes field_bytes = 14;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestSimple {
    field_double: Float
    field_float: Float
    field_int32: Int
    field_int64: Int
    field_uint64: Int
    field_sint32: Int
    field_sint64: Int
    field_fixed32: Int
    field_fixed64: Int
    field_sfixed32: Int
    field_sfixed64: Int
    field_bool: Boolean
    field_string: String
    field_bytes: [String]
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformNestedMessages(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package api.myapp;

		message Outer {                  // Level 0
			int32 field_int32 = 1;

	        message MiddleAA {  // Level 1
				string field_string = 1;

	            message Inner {   // Level 2
	                int64 ival = 1;
				  	bool booly = 2;
				}
			}

			message MiddleBB {  // Level 1
				int64 field_int64 = 1;

				message Inner {   // Level 2
					int32 ival = 1;
					bool booly = 2;
				}
			}
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type ApiMyappOuter {
    field_int32: Int
}

type ApiMyappOuterMiddleAA {
    field_string: String
}

type ApiMyappOuterMiddleAAInner {
    ival: Int
    booly: Boolean
}

type ApiMyappOuterMiddleBB {
    field_int64: Int
}

type ApiMyappOuterMiddleBBInner {
    ival: Int
    booly: Boolean
}
	`

	expected = expected
	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformUseOfNestedTypes(t *testing.T) {
	schema := []byte(`
syntax = "proto3";
package api.myapp;

message SearchResponse {
	message Result {
    	string url = 1;
    	string title = 2;
    	repeated string snippets = 3;
  	}

	repeated Result results = 1;
}

message SomeOtherMessage {
    SearchResponse.Result result = 1;
}

	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type ApiMyappSearchResponse {
    results: [ApiMyappSearchResponseResult]
}

type ApiMyappSearchResponseResult {
    url: String
    title: String
    snippets: [String]
}

type ApiMyappSomeOtherMessage {
    result: ApiMyappSearchResponseResult
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformNestedMessagesHoisting(t *testing.T) {
	schema := []byte(`
syntax = "proto3";
package api.myapp;

message SearchResponse {
	repeated Result results = 1;

	message Result {
    	string url = 1;
    	string title = 2;
    	repeated string snippets = 3;
  	}
}

	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type ApiMyappSearchResponse {
    results: [ApiMyappSearchResponseResult]
}

type ApiMyappSearchResponseResult {
    url: String
    title: String
    snippets: [String]
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformNestedTypes(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		message A {
			int64 field_int64 = 1;
		}

		message B {
			A field_a = 1;
		}

		message C {
			repeated B field_b = 1;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestA {
    field_int64: Int
}

type TestB {
    field_a: TestA
}

type TestC {
    field_b: [TestB]
}
`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformEnums(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		enum Corpus {
			UNIVERSAL = 0;
			WEB = 1;
			IMAGES = 2;
			LOCAL = 3;
			NEWS = 4;
			PRODUCTS = 5;
			VIDEO = 6;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
enum TestCorpus {
    UNIVERSAL
    WEB
    IMAGES
    LOCAL
    NEWS
    PRODUCTS
    VIDEO
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformNestedEnums(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		message Person {
			string firstName = 1;
			string lastName = 2;

			enum Gender {
				UNKNOWN = 0;
				MALE = 1;
				FEMALE = 2;
			}
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestPerson {
    firstName: String
    lastName: String
}

enum TestPersonGender {
    UNKNOWN
    MALE
    FEMALE
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformImportedTypes(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		import "google/protobuf/any.proto";

		message ErrorStatus {
		  string message = 1;
		  repeated google.protobuf.Any details = 2;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestErrorStatus {
    message: String
    details: [GoogleProtobufAny]
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformImportedNestedTypes(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		import "google/protobuf/any.proto";

		message OtherMessage {
		  google.protobuf.Any.Type details = 1;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestOtherMessage {
    details: GoogleProtobufAnyType
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformImportedTypesFromSamePackage(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		import "any.proto";

		message ErrorStatus {
		  string message = 1;
		  repeated Any details = 2;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestErrorStatus {
    message: String
    details: [TestAny]
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformResolveImportedTypes(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		import "google/protobuf/timestamp.proto";

		message ErrorStatus {
		  google.protobuf.Timestamp time = 1;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)

	transformer.Import("google/protobuf/timestamp.proto", "https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/timestamp.proto")

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type TestErrorStatus {
    time: GoogleProtobufTimestamp
}

type GoogleProtobufTimestamp {
    seconds: Int
    nanos: Int
}
	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformWithNoTypePrefix(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		import "google/protobuf/timestamp.proto";

		message ErrorStatus {
		  google.protobuf.Timestamp time = 1;
		}

		message SearchResponse {
			message Result {
    			string url = 1;
    			string title = 2;
    			repeated string snippets = 3;
  			}

			repeated Result results = 1;
		}

		message SomeOtherMessage {
    		SearchResponse.Result result = 1;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)
	transformer.DisablePrefix(true)
	transformer.Import("google/protobuf/timestamp.proto", "https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/timestamp.proto")

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type ErrorStatus {
    time: Timestamp
}

type SearchResponse {
    results: [SearchResponseResult]
}

type SearchResponseResult {
    url: String
    title: String
    snippets: [String]
}

type SomeOtherMessage {
    result: SearchResponseResult
}

type Timestamp {
    seconds: Int
    nanos: Int
}

	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformWithPackageAliases(t *testing.T) {
	schema := []byte(`
		syntax = "proto3";
		package test;

		import "google/protobuf/timestamp.proto";

		message ErrorStatus {
		  google.protobuf.Timestamp time = 1;
		}

		message SearchResponse {
			message Result {
    			string url = 1;
    			string title = 2;
    			repeated string snippets = 3;
  			}

			repeated Result results = 1;
		}

		message SomeOtherMessage {
    		SearchResponse.Result result = 1;
		}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)
	transformer.SetPackageAlias("test", "Dashboard")
	transformer.SetPackageAlias("google.protobuf", "")
	transformer.Import("google/protobuf/timestamp.proto", "https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/timestamp.proto")

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
type DashboardErrorStatus {
    time: Timestamp
}

type DashboardSearchResponse {
    results: [DashboardSearchResponseResult]
}

type DashboardSearchResponseResult {
    url: String
    title: String
    snippets: [String]
}

type DashboardSomeOtherMessage {
    result: DashboardSearchResponseResult
}

type Timestamp {
    seconds: Int
    nanos: Int
}

	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}

func TestTransformWithPackageAliases2(t *testing.T) {
	schema := []byte(`
syntax = "proto3";
package com.users.api;

enum ContactType {
    PHONE = 0;
    EMAIL = 1;
    WEBSITE = 2;
    FACEBOOK = 3;
    TWITTER = 4;
    INSTAGRAM = 5;
    YOUTUBE = 6;
    FLICKR = 7;
    MEDIUM = 8;
}

message Contact {
    uint64 id = 1;
    ContactType type = 2;
    string value = 3;
}
	`)

	input := new(bytes.Buffer)
	input.Write(schema)

	output := new(bytes.Buffer)
	transformer := NewTransformer(output)
	transformer.SetPackageAlias("com.users.api", "User")

	if err := transformer.Transform(input); err != nil {
		t.Fatal(err)
	}

	expected := `
enum UserContactType {
    PHONE
    EMAIL
    WEBSITE
    FACEBOOK
    TWITTER
    INSTAGRAM
    YOUTUBE
    FLICKR
    MEDIUM
}

type UserContact {
    id: Int
    type: UserContactType
    value: String
}

	`

	expected = strings.TrimSpace(expected)
	actual := strings.TrimSpace(output.String())

	if expected != actual {
		t.Fatalf("Expected %s to equal to %s", expected, actual)
	}
}