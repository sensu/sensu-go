package bonsai_test

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/bonsai"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetOutput(ioutil.Discard)
}

func TestFetchAsset(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasSuffix(req.URL.String(), "/default/asset") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		asset := bonsai.Asset{
			Name:        "asset",
			Description: "some asset",
			URL:         "http://example.com/default/asset",
		}
		_ = json.NewEncoder(w).Encode(asset)
	}))
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	tlsConfig := &tls.Config{
		RootCAs: pool,
	}
	client := bonsai.New(bonsai.Config{EndpointURL: server.URL, TLSConfig: tlsConfig})
	asset, err := client.FetchAsset("default", "asset")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := asset.Name, "asset"; got != want {
		t.Fatalf("bad asset name: got %q, want %q", got, want)
	}
	if _, err := client.FetchAsset("default", "notexists"); err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestFetchAssetVersion(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if !strings.HasSuffix(req.URL.String(), "/default/asset/v0.1.0/release_asset_builds") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = fmt.Fprint(w, "asset definition")
	}))
	pool := x509.NewCertPool()
	pool.AddCert(server.Certificate())
	tlsConfig := &tls.Config{
		RootCAs: pool,
	}
	client := bonsai.New(bonsai.Config{EndpointURL: server.URL, TLSConfig: tlsConfig})
	defn, err := client.FetchAssetVersion("default", "asset", "v0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := defn, "asset definition"; got != want {
		t.Fatalf("bad asset defn: got %q, want %q", got, want)
	}
	if _, err := client.FetchAssetVersion("default", "asset", "v0.2.0"); err == nil {
		t.Fatal("expected non-nil error")
	}
}
