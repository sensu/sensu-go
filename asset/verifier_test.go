package asset_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/sensu/sensu-go/asset"
)

func TestSuccessfulVerify(t *testing.T) {
	t.Parallel()

	assetPath := getFixturePath("rubby-on-rails.tar")
	assetSHA, err := ioutil.ReadFile(fmt.Sprintf("%s.sha512", assetPath))
	if err != nil {
		t.Fatalf("could not read asset sha, error: %v", err)
	}

	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("could not open asset, error: %v", err)
	}
	defer f.Close()

	verifier := &asset.SHA512Verifier{}
	if err := verifier.Verify(f, string(assetSHA)); err != nil {
		t.Logf("expected no error, got %v", err)
		t.Fail()
	}
}

func TestFailedVerify(t *testing.T) {
	t.Parallel()

	assetPath := getFixturePath("rubby-on-rails.tar")

	f, err := os.Open(assetPath)
	if err != nil {
		t.Fatalf("could not open asset, error: %v", err)
	}
	defer f.Close()

	verifier := &asset.SHA512Verifier{}
	if err := verifier.Verify(f, "obviously not a SHA512"); err == nil {
		t.Log("expected error, got nil")
		t.Fail()
	}
}
