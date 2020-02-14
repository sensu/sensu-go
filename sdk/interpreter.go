package sdk

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication"

	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/repl"
)

const helpText = `
The Sensu Interactive Console
=============================

This interactive console is a Javascript REPL. You can use this tool to write
Sensu filters, inspect events, entities or other stored resources, or craft
scripts that use parts of the Go standard library.


Store access
------------

Access the Sensu object store with the 'sensu' object. Example::

  ∫∫∫ var event = sensu.GetEventByEntityCheck("router", "ping");
  ∫∫∫ event.check.status
  0


Test Fixtures
-------------

Load test fixtures from the corev2 package. Example::

  ∫∫∫ var event = corev2.FixtureEvent("default", "ping");
  ∫∫∫ event.entity.name
  default
`

func newCoreAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	// TODO(eric): add the rest of the corev2 things with codegen
	return map[string]interface{}{
		"Asset":              call(ctx, vm, with(new(corev2.Asset))),
		"Entity":             call(ctx, vm, with(new(corev2.Entity))),
		"Check":              call(ctx, vm, with(new(corev2.Check))),
		"CheckConfig":        call(ctx, vm, with(new(corev2.CheckConfig))),
		"Event":              call(ctx, vm, with(new(corev2.Event))),
		"EventFilter":        call(ctx, vm, with(new(corev2.EventFilter))),
		"Handler":            call(ctx, vm, with(new(corev2.Handler))),
		"Hook":               call(ctx, vm, with(new(corev2.Hook))),
		"Mutator":            call(ctx, vm, with(new(corev2.Mutator))),
		"Silenced":           call(ctx, vm, with(new(corev2.Silenced))),
		"FixtureAsset":       call(ctx, vm, corev2.FixtureAsset),
		"FixtureEntity":      call(ctx, vm, corev2.FixtureEntity),
		"FixtureCheck":       call(ctx, vm, corev2.FixtureCheck),
		"FixtureCheckConfig": call(ctx, vm, corev2.FixtureCheckConfig),
		"FixtureEvent":       call(ctx, vm, corev2.FixtureEvent),
		"FixtureEventFilter": call(ctx, vm, corev2.FixtureEventFilter),
		"FixtureHandler":     call(ctx, vm, corev2.FixtureHandler),
		"FixtureHook":        call(ctx, vm, corev2.FixtureHook),
		"FixtureMutator":     call(ctx, vm, corev2.FixtureMutator),
		"FixtureSilenced":    call(ctx, vm, corev2.FixtureSilenced),
	}
}

func newIOAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"Copy":             call(ctx, vm, io.Copy),
		"CopyBuffer":       call(ctx, vm, io.CopyBuffer),
		"CopyN":            call(ctx, vm, io.CopyN),
		"Pipe":             call(ctx, vm, io.Pipe),
		"ReadAtLeast":      call(ctx, vm, io.ReadAtLeast),
		"ReadFull":         call(ctx, vm, io.ReadFull),
		"WriteString":      call(ctx, vm, io.WriteString),
		"LimitReader":      call(ctx, vm, io.LimitReader),
		"MultiReader":      call(ctx, vm, io.MultiReader),
		"TeeReader":        call(ctx, vm, io.TeeReader),
		"NewSectionReader": call(ctx, vm, io.NewSectionReader),
		"MultiWriter":      call(ctx, vm, io.MultiWriter),
	}
}

func newBufIOAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"ScanBytes":     call(ctx, vm, bufio.ScanBytes),
		"ScanLines":     call(ctx, vm, bufio.ScanLines),
		"ScanRunes":     call(ctx, vm, bufio.ScanRunes),
		"ScanWords":     call(ctx, vm, bufio.ScanWords),
		"NewReadWriter": call(ctx, vm, bufio.NewReadWriter),
		"NewReader":     call(ctx, vm, bufio.NewReader),
		"NewReaderSize": call(ctx, vm, bufio.NewReaderSize),
		"NewScanner":    call(ctx, vm, bufio.NewScanner),
		"NewWriter":     call(ctx, vm, bufio.NewWriter),
		"NewWriterSize": call(ctx, vm, bufio.NewWriterSize),
	}
}

// n.b., this package is just provided to make opening files possible
func newPartialOSAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"Open":   call(ctx, vm, os.Open),
		"Create": call(ctx, vm, os.Create),
	}
}

func newNetAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"JoinHnettPort":    call(ctx, vm, net.JoinHostPort),
		"LookupAddr":       call(ctx, vm, net.LookupAddr),
		"LookupCNAME":      call(ctx, vm, net.LookupCNAME),
		"LookupHnett":      call(ctx, vm, net.LookupHost),
		"LookupPort":       call(ctx, vm, net.LookupPort),
		"LookupTXT":        call(ctx, vm, net.LookupTXT),
		"ParseCIDR":        call(ctx, vm, net.ParseCIDR),
		"Pipe":             call(ctx, vm, net.Pipe),
		"SplitHnettPort":   call(ctx, vm, net.SplitHostPort),
		"InterfaceAddrs":   call(ctx, vm, net.InterfaceAddrs),
		"Dial":             call(ctx, vm, net.Dial),
		"ParseMAC":         call(ctx, vm, net.ParseMAC),
		"IPv4":             call(ctx, vm, net.IPv4),
		"LookupIP":         call(ctx, vm, net.LookupIP),
		"ParseIP":          call(ctx, vm, net.ParseIP),
		"ResolveIPAddr":    call(ctx, vm, net.ResolveIPAddr),
		"DialIP":           call(ctx, vm, net.DialIP),
		"ListenIP":         call(ctx, vm, net.ListenIP),
		"CIDRMask":         call(ctx, vm, net.CIDRMask),
		"IPv4Mask":         call(ctx, vm, net.IPv4Mask),
		"InterfaceByIndex": call(ctx, vm, net.InterfaceByIndex),
		"InterfaceByName":  call(ctx, vm, net.InterfaceByName),
		"Interfaces":       call(ctx, vm, net.Interfaces),
		// Don't include Listen - the SDK is not appropriate for servers
		// Also don't include things like Listen, like ListenTCP
		//"Listen": call(ctx, vm, net.Listen),
		"LookupMX":        call(ctx, vm, net.LookupMX),
		"LookupNS":        call(ctx, vm, net.LookupNS),
		"LookupSRV":       call(ctx, vm, net.LookupSRV),
		"ResolveTCPAddr":  call(ctx, vm, net.ResolveTCPAddr),
		"DialTCP":         call(ctx, vm, net.DialTCP),
		"ResolveUnixAddr": call(ctx, vm, net.ResolveUnixAddr),
		"DialUnix":        call(ctx, vm, net.DialUnix),
	}
}

func newIOUtilAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"NopCloser": call(ctx, vm, ioutil.NopCloser),
		"ReadAll":   call(ctx, vm, ioutil.ReadAll),
		"ReadDir":   call(ctx, vm, ioutil.ReadDir),
		"ReadFile":  call(ctx, vm, ioutil.ReadFile),
		"TempDir":   call(ctx, vm, ioutil.TempDir),
		"TempFile":  call(ctx, vm, ioutil.TempFile),
		"WriteFile": call(ctx, vm, ioutil.WriteFile),
	}
}

// the HTTP API includes basic client functionality
func newPartialHTTPAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"Get":          call(ctx, vm, http.Get),
		"Head":         call(ctx, vm, http.Head),
		"Post":         call(ctx, vm, http.Post),
		"PostForm":     call(ctx, vm, http.PostForm),
		"ReadResponse": call(ctx, vm, http.ReadResponse),
		"NewRequest":   call(ctx, vm, http.NewRequest),
		"ReadRequest":  call(ctx, vm, http.ReadRequest),
		"Client":       call(ctx, vm, with(new(http.Client))),
		"Header":       call(ctx, vm, with(make(http.Header))),
	}
}

func newFmtAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"Sprintf":  call(ctx, vm, fmt.Sprintf),
		"Sprint":   call(ctx, vm, fmt.Sprint),
		"Sprintln": call(ctx, vm, fmt.Sprintln),
		"Fprintf":  call(ctx, vm, fmt.Fprintf),
		"Fprint":   call(ctx, vm, fmt.Fprint),
		"Fprintln": call(ctx, vm, fmt.Fprintln),
		"Printf":   call(ctx, vm, fmt.Sprintf),  // this is not a mistake
		"Println":  call(ctx, vm, fmt.Sprintln), // this is not a mistake
		"Print":    call(ctx, vm, fmt.Sprint),
		"Errorf":   call(ctx, vm, fmt.Errorf),
		"Sscan":    call(ctx, vm, fmt.Sscan),
		"Sscanf":   call(ctx, vm, fmt.Sscanf),
		"Sscanln":  call(ctx, vm, fmt.Sscanln),
		"Fscan":    call(ctx, vm, fmt.Fscan),
		"Fscanf":   call(ctx, vm, fmt.Fscanf),
		"Fscanln":  call(ctx, vm, fmt.Fscanln),
	}
}

func newJSONAPI(ctx context.Context, vm *otto.Otto) map[string]interface{} {
	return map[string]interface{}{
		"Marshal":    call(ctx, vm, json.Marshal),
		"Unmarshal":  call(ctx, vm, json.Unmarshal),
		"NewDecoder": call(ctx, vm, json.NewDecoder),
		"NewEncoder": call(ctx, vm, json.NewEncoder),
	}
}

func with(value interface{}) interface{} {
	return func() interface{} {
		return value
	}
}

func toInterface(values []reflect.Value) interface{} {
	switch len(values) {
	case 0:
		return otto.UndefinedValue()
	case 1:
		return values[0].Interface()
	default:
		result := make([]interface{}, len(values))
		for i := range result {
			result[i] = values[i].Interface()
		}
		return result
	}

}

func call(ctx context.Context, vm *otto.Otto, fn interface{}) func(...interface{}) interface{} {
	value := reflect.ValueOf(fn)
	typ := reflect.TypeOf(fn)
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	ctxValue := reflect.ValueOf(ctx)

	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("call with non-func type %s", typ.Kind()))
	}
	return func(args ...interface{}) (result interface{}) {
		defer func() {
			if e := recover(); e != nil {
				s := fmt.Sprintf("%s", e)
				result = otto.UndefinedValue()
				if strings.HasPrefix(s, "reflect: ") {
					s = strings.TrimPrefix(s, "reflect: ")
					panic(vm.MakeTypeError(s))
				} else {
					panic(vm.MakeCustomError("Error", s))
				}
			}
		}()
		numArgs := typ.NumIn()
		argValues := make([]reflect.Value, numArgs)
		argsIdx := 0
		for i := 0; i < numArgs; i++ {
			if i == 0 && typ.In(0).Implements(ctxType) {
				argValues[0] = ctxValue
				continue
			}
			if argsIdx >= len(args) || args[argsIdx] == nil {
				argValues[i] = reflect.New(typ.In(i)).Elem()
			} else {
				argValues[i] = reflect.ValueOf(args[argsIdx])
			}
			argsIdx++
		}
		callResults := value.Call(argValues)
		if len(callResults) == 0 {
			return otto.UndefinedValue()
		}
		if !typ.Out(typ.NumOut() - 1).Implements(errorType) {
			return toInterface(callResults)
		}
		errVal := callResults[len(callResults)-1].Interface()
		if errVal != nil {
			err := errVal.(error)
			panic(vm.MakeCustomError("sensu", err.Error()))
		}
		return toInterface(callResults[:len(callResults)-1])
	}
}

type Interpreter struct {
	apis map[string]interface{}
	auth *authentication.Authenticator
	vm   *otto.Otto
}

func NewInterpreter(auth *authentication.Authenticator, apis map[string]interface{}) *Interpreter {
	vm := otto.New()
	vm.Set("help", helpText)
	vm.Set("exit", "type Ctrl+D to exit")
	intr := &Interpreter{apis: apis, vm: vm}
	return intr
}

func (p *Interpreter) makeAPIs(ctx context.Context) {
	for k, v := range p.apis {
		value := reflect.ValueOf(v)
		typ := reflect.TypeOf(v)
		numMethod := typ.NumMethod()
		if numMethod == 0 {
			continue
		}
		methodMap := make(map[string]interface{}, numMethod)
		for i := 0; i < numMethod; i++ {
			methName := typ.Method(i).Name
			meth := value.Method(i).Interface()
			methodMap[methName] = call(ctx, p.vm, meth)
		}
		p.vm.Set(k, methodMap)
	}
}

func toBytes(v interface{}) []byte {
	switch v := v.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	default:
		panic(fmt.Sprintf("can't convert %T to bytea", v))
	}
}

func (p *Interpreter) Run(ctx context.Context) error {
	ctx = context.WithValue(ctx, corev2.NamespaceKey, "default")

	p.vm.Set("corev2", newCoreAPI(ctx, p.vm))
	p.vm.Set("io", newIOAPI(ctx, p.vm))
	p.vm.Set("ioutil", newIOUtilAPI(ctx, p.vm))
	p.vm.Set("bufio", newBufIOAPI(ctx, p.vm))
	p.vm.Set("os", newPartialOSAPI(ctx, p.vm))
	p.vm.Set("net", newNetAPI(ctx, p.vm))
	p.vm.Set("http", newPartialHTTPAPI(ctx, p.vm))
	p.vm.Set("fmt", newFmtAPI(ctx, p.vm))
	p.vm.Set("json", newJSONAPI(ctx, p.vm))

	p.makeAPIs(ctx)

	options := repl.Options{
		Prompt:       "∫∫∫ ",
		Prelude:      "Welcome to the Sensu Software Development Kit! Type 'help' for help.",
		Autocomplete: true,
	}
	return repl.RunWithOptions(p.vm, options)
}
