package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

// tokenSubstitution evaluates the input template, that possibly contains
// tokens, with the provided data object and returns a slice of bytes
// representing the result along with any error encountered
func tokenSubstitution(data, input interface{}) ([]byte, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the provided template: %s", err.Error())
	}

	// buf2 := new(bytes.Buffer)
	// encoder := json.NewEncoder(buf2)
	// encoder.SetEscapeHTML(false)
	// if err := encoder.Encode(input); err != nil {
	// 	return nil, err
	// }
	// fmt.Println("buf2.String: " + buf2.String())

	// fmt.Printf("%s\n", string(inputBytes))
	// s, err := strconv.Unquote(string(inputBytes))
	// fmt.Println(err)
	// fmt.Println(s)
	// fmt.Println("rawMessage: " + string(json.RawMessage(string(inputBytes))))
	// check := &types.CheckConfig{Command: `{{ defaultValue "foo" }}`}
	// fmt.Println(*check)
	// fmt.Println("inputByes: " + string(inputBytes))

	inputString := strings.Replace(string(inputBytes), `\"`, `"`, -1)

	tmpl := template.New("")
	tmpl.Funcs(funcMap())

	tmpl, err = tmpl.Parse(inputString)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template: %s", err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("could not execute the template: %s", err.Error())
	}
	fmt.Println(buf.String())
	return buf.Bytes(), nil
}

// funcMap ...
func funcMap() template.FuncMap {
	return template.FuncMap{
		"default": defaultFunc,
	}
}

func defaultFunc(v ...interface{}) interface{} {
	if len(v) == 1 {
		return v[0]
	} else if len(v) == 2 {
		return v[1]
	}
	return nil
	// fmt.Println(arg)
	// fmt.Println(value)
	// fmt.Printf("%+v\n", v)

	// s := reflect.ValueOf(v)
	// kv := reflect.ValueOf(arg)
	// // name := s.FieldByName(arg)
	// // fmt.Println(name)
	// fmt.Println(s.MapIndex(kv))

	// return "foo"
}
