package v2_test

import (
	"errors"
	"testing"

	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/stretchr/testify/mock"
)

func TestProxyCreateOrUpdate(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	got := proxy.CreateOrUpdate(storev2.ResourceRequest{}, nil)
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyUpdateIfExists(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("UpdateIfExists", mock.Anything, mock.Anything).Return(want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	got := proxy.UpdateIfExists(storev2.ResourceRequest{}, nil)
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyCreateIfNotExists(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	got := proxy.CreateIfNotExists(storev2.ResourceRequest{}, nil)
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyGet(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("Get", mock.Anything, mock.Anything).Return((storev2.Wrapper)(nil), want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	_, got := proxy.Get(storev2.ResourceRequest{})
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyDelete(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("Delete", mock.Anything).Return(want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	got := proxy.Delete(storev2.ResourceRequest{})
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyList(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("List", mock.Anything, mock.Anything).Return(make(wrap.List, 0), want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	_, got := proxy.List(storev2.ResourceRequest{}, nil)
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyExists(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("Exists", mock.Anything, mock.Anything).Return(false, want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	_, got := proxy.Exists(storev2.ResourceRequest{})
	if got != want {
		t.Fatal(got)
	}
}

func TestProxyPatch(t *testing.T) {
	s := new(storetest.Store)
	want := errors.New("expected")
	s.On("Patch", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(want)
	var proxy storev2.Proxy
	proxy.UpdateStore(s)
	got := proxy.Patch(storev2.ResourceRequest{}, nil, nil, nil)
	if got != want {
		t.Fatal(got)
	}
}
