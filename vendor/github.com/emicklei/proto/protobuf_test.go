package proto

import (
	"net/http"
	"os"
	"testing"
)

// PB=y go test -v -run ^TestParseTheTest$
func TestParseTheTest(t *testing.T) {
	if len(os.Getenv("PB")) == 0 {
		t.Skip("PB test not run")
	}
	fetchAndParse(t, "https://raw.githubusercontent.com/gogo/protobuf/master/test/thetest.proto")
}

func fetchAndParse(t *testing.T, url string) {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	parser := NewParser(resp.Body)
	def, err := parser.Parse()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("elements:", len(def.Elements))
}

// PB=y go test -v -run ^TestParseTheProto3$
func TestParseTheProto3(t *testing.T) {
	if len(os.Getenv("PB")) == 0 {
		t.Skip("PB test not run")
	}
	fetchAndParse(t, "https://raw.githubusercontent.com/gogo/protobuf/master/test/theproto3/theproto3.proto")
}
