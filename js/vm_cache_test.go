package js

import (
	"testing"

	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
	"github.com/robertkrimen/otto"
)

var crockTime = crock.NewTime(time.Now())

func init() {
	time.TimeProxy = crockTime
}

func TestVMCacheGetMiss(t *testing.T) {
	cache := newVMCache()
	defer cache.Close()
	val := cache.Acquire("missing")
	if val != nil {
		t.Fatal("non-nil value")
	}
}

func TestVMCacheGetHit(t *testing.T) {
	cache := newVMCache()
	defer cache.Close()
	vm := otto.New()
	cache.Init("foo", vm)
	cache.reap()
	val := cache.Acquire("foo")
	if val == nil {
		t.Fatal("cache miss when should be hit")
	}
}

func TestVMCacheExpire(t *testing.T) {
	cache := newVMCache()
	defer cache.Close()
	vm := otto.New()
	cache.Init("foo", vm)
	crockTime.Set(crockTime.Now().Add(time.Hour * 2))
	cache.reap()
	val := cache.Acquire("foo")
	if val != nil {
		t.Fatal("non-nil value")
	}
}
